package draco

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math"

	"github.com/flywave/gltf"
	"github.com/flywave/go-draco"
)

const ExtensionName = "KHR_draco_mesh_compression"

func init() {
	gltf.RegisterExtension(ExtensionName, Unmarshal)
}

type DracoExtension struct {
	BufferView uint32            `json:"bufferView"`
	Attributes map[string]uint32 `json:"attributes"`
}

func Unmarshal(data []byte) (interface{}, error) {
	ext := &DracoExtension{}
	if err := json.Unmarshal(data, ext); err != nil {
		return nil, fmt.Errorf("KHR_draco解析失败: %w", err)
	}
	return ext, nil
}

func DecodeAll(doc *gltf.Document) error {
	for _, mesh := range doc.Meshes {
		for i := range mesh.Primitives {
			if err := decodePrimitive(doc, mesh.Primitives[i]); err != nil {
				return fmt.Errorf("图元解码失败: %w", err)
			}
		}
	}
	doc.RemoveExtension(ExtensionName)
	return nil
}

func decodePrimitive(doc *gltf.Document, primitive *gltf.Primitive) error {
	extData, exists := primitive.Extensions[ExtensionName]
	if !exists {
		return nil
	}

	ext, ok := extData.(*DracoExtension)
	if !ok {
		return fmt.Errorf("无效的Draco扩展格式")
	}

	// 获取压缩数据
	bufferView := doc.BufferViews[ext.BufferView]
	if bufferView.Buffer >= uint32(len(doc.Buffers)) {
		return fmt.Errorf("缓冲区索引越界")
	}
	bufferData := doc.Buffers[bufferView.Buffer].Data
	start := bufferView.ByteOffset
	end := start + bufferView.ByteLength
	if end > uint32(len(bufferData)) {
		return fmt.Errorf("缓冲区视图超出范围")
	}
	compressedData := bufferData[start:end]

	// Draco解码
	decoder := draco.NewDecoder()

	mesh := draco.NewMesh()

	err := decoder.DecodeMesh(mesh, compressedData)
	if err != nil {
		return fmt.Errorf("draco解码失败: %o", err)
	}
	defer mesh.Free()

	// 使用共享缓冲区优化
	sharedBuffer := &bytes.Buffer{}
	var bufferIndex, bufferViewIndex uint32

	// 构建属性映射
	attrs := make(map[string]uint32)
	for name, id := range ext.Attributes {
		attr := mesh.Attr(int32(id))
		if attr == nil {
			return fmt.Errorf("找不到属性: %s", name)
		}

		// 获取顶点数量
		vertCount := mesh.NumPoints()
		if vertCount <= 0 {
			return fmt.Errorf("无效的顶点数量: %d", vertCount)
		}

		// 获取属性数据
		outVert := make([]float32, vertCount*3)
		if _, ok := mesh.AttrData(attr, outVert); !ok {
			return fmt.Errorf("获取属性数据失败: %w", err)
		}

		// 创建访问器
		accessorIdx, err := createAccessor(
			doc,
			name,
			outVert,
			sharedBuffer,
			&bufferIndex,
			&bufferViewIndex,
		)
		if err != nil {
			return err
		}
		attrs[name] = accessorIdx
	}

	// 处理索引数据
	if mesh.NumFaces() > 0 {
		faceCount := mesh.NumFaces()
		indices := make([]uint32, faceCount*3)
		if err := mesh.Faces(indices); err != nil {
			return fmt.Errorf("获取索引数据失败: %o", err)
		}

		idxAccessor, err := createIndexAccessor(doc, indices)
		if err != nil {
			return fmt.Errorf("创建索引访问器失败: %w", err)
		}
		primitive.Indices = gltf.Index(idxAccessor)
	}

	// 如果共享缓冲区有数据，添加到文档
	if sharedBuffer.Len() > 0 {
		doc.Buffers = append(doc.Buffers, &gltf.Buffer{
			Data: sharedBuffer.Bytes(),
		})
	}

	// 更新图元
	primitive.Attributes = attrs
	delete(primitive.Extensions, ExtensionName)
	return nil
}

func createAccessor(
	doc *gltf.Document,
	attrName string,
	data []float32,
	sharedBuffer *bytes.Buffer,
	bufferIndex *uint32,
	bufferViewIndex *uint32,
) (uint32, error) {
	// 确定访问器类型
	accType, componentType := getAccessorType(attrName)

	// 计算元素数量和组件数
	components := componentsPerType(accType)
	count := uint32(len(data)) / components
	if uint32(len(data))%components != 0 {
		return 0, fmt.Errorf("数据长度与属性类型不匹配")
	}

	// 创建访问器
	accessor := &gltf.Accessor{
		ComponentType: componentType,
		Type:          accType,
		Count:         count,
	}

	// 计算最小/最大值 (仅对位置属性需要)
	if attrName == "POSITION" {
		min, max := calculateMinMax(data, int(components))
		accessor.Min = min
		accessor.Max = max
	}

	// 将数据转换为字节
	byteData := float32ToBytes(data)

	// 添加到共享缓冲区
	offset := sharedBuffer.Len()
	if _, err := sharedBuffer.Write(byteData); err != nil {
		return 0, fmt.Errorf("写入缓冲区失败: %w", err)
	}

	// 创建或复用缓冲区视图
	if *bufferIndex == 0 {
		// 首次使用，创建新的缓冲区
		*bufferIndex = uint32(len(doc.Buffers))
		doc.Buffers = append(doc.Buffers, &gltf.Buffer{})
	}

	view := &gltf.BufferView{
		Buffer:     *bufferIndex,
		ByteOffset: uint32(offset),
		ByteLength: uint32(len(byteData)),
		Target:     gltf.TargetArrayBuffer,
	}
	doc.BufferViews = append(doc.BufferViews, view)
	accessor.BufferView = gltf.Index(uint32(len(doc.BufferViews) - 1))
	*bufferViewIndex = uint32(len(doc.BufferViews) - 1)

	// 添加访问器到文档
	doc.Accessors = append(doc.Accessors, accessor)
	return uint32(len(doc.Accessors) - 1), nil
}

func createIndexAccessor(doc *gltf.Document, indices []uint32) (uint32, error) {
	if len(indices) == 0 {
		return 0, fmt.Errorf("索引数据为空")
	}

	// 确定组件类型
	var componentType gltf.ComponentType
	maxIndex := uint32(0)
	for _, idx := range indices {
		if idx > maxIndex {
			maxIndex = idx
		}
	}

	switch {
	case maxIndex <= math.MaxUint8:
		componentType = gltf.ComponentUbyte
	case maxIndex <= math.MaxUint16:
		componentType = gltf.ComponentUshort
	default:
		componentType = gltf.ComponentUint
	}

	// 创建访问器
	accessor := &gltf.Accessor{
		ComponentType: componentType,
		Type:          gltf.AccessorScalar,
		Count:         uint32(len(indices)),
	}

	// 转换索引数据为字节
	byteData := indicesToBytes(indices, componentType)

	// 创建缓冲区
	buffer := &gltf.Buffer{Data: byteData}
	doc.Buffers = append(doc.Buffers, buffer)
	bufferIdx := uint32(len(doc.Buffers) - 1)

	// 创建缓冲区视图
	view := &gltf.BufferView{
		Buffer:     bufferIdx,
		ByteLength: uint32(len(byteData)),
		Target:     gltf.TargetElementArrayBuffer,
	}
	doc.BufferViews = append(doc.BufferViews, view)
	accessor.BufferView = gltf.Index(uint32(len(doc.BufferViews) - 1))

	// 添加访问器
	doc.Accessors = append(doc.Accessors, accessor)
	return uint32(len(doc.Accessors) - 1), nil
}

func getAccessorType(attrName string) (gltf.AccessorType, gltf.ComponentType) {
	switch attrName {
	case "POSITION", "NORMAL":
		return gltf.AccessorVec3, gltf.ComponentFloat
	case "TEXCOORD_0", "TEXCOORD_1":
		return gltf.AccessorVec2, gltf.ComponentFloat
	case "COLOR_0":
		return gltf.AccessorVec4, gltf.ComponentFloat
	case "TANGENT":
		return gltf.AccessorVec4, gltf.ComponentFloat
	case "JOINTS_0":
		return gltf.AccessorVec4, gltf.ComponentUshort
	case "WEIGHTS_0":
		return gltf.AccessorVec4, gltf.ComponentFloat
	default:
		// 默认为VEC3浮点数
		return gltf.AccessorVec3, gltf.ComponentFloat
	}
}

func componentsPerType(accType gltf.AccessorType) uint32 {
	switch accType {
	case gltf.AccessorScalar:
		return 1
	case gltf.AccessorVec2:
		return 2
	case gltf.AccessorVec3:
		return 3
	case gltf.AccessorVec4:
		return 4
	default:
		return 0
	}
}

func calculateMinMax(data []float32, components int) (min, max []float32) {
	min = make([]float32, components)
	max = make([]float32, components)
	for i := range min {
		min[i] = math.MaxFloat32
		max[i] = -math.MaxFloat32
	}

	for i := 0; i < len(data); i += components {
		for j := 0; j < components; j++ {
			idx := i + j
			if idx >= len(data) {
				break
			}
			val := data[idx]
			if val < min[j] {
				min[j] = val
			}
			if val > max[j] {
				max[j] = val
			}
		}
	}
	return min, max
}

func float32ToBytes(data []float32) []byte {
	buf := new(bytes.Buffer)
	if err := binary.Write(buf, binary.LittleEndian, data); err != nil {
		return nil
	}
	return buf.Bytes()
}

func indicesToBytes(indices []uint32, componentType gltf.ComponentType) []byte {
	buf := new(bytes.Buffer)
	buf.Grow(len(indices) * sizeOfComponent(componentType))

	switch componentType {
	case gltf.ComponentUbyte:
		for _, idx := range indices {
			buf.WriteByte(byte(idx))
		}
	case gltf.ComponentUshort:
		for _, idx := range indices {
			binary.Write(buf, binary.LittleEndian, uint16(idx))
		}
	case gltf.ComponentUint:
		binary.Write(buf, binary.LittleEndian, indices)
	default:
		binary.Write(buf, binary.LittleEndian, indices)
	}
	return buf.Bytes()
}

func EncodeAll(doc *gltf.Document, options map[string]interface{}) error {
	encoder := draco.NewEncoder()

	for _, mesh := range doc.Meshes {
		for i := range mesh.Primitives {
			if err := encodePrimitive(doc, encoder, mesh.Primitives[i], options); err != nil {
				return err
			}
		}
	}

	doc.AddExtensionUsed(ExtensionName)
	return nil
}

func encodePrimitive(doc *gltf.Document, encoder *draco.Encoder, primitive *gltf.Primitive, options map[string]interface{}) error {
	// 1. 收集顶点数据
	positionAccessor, ok := primitive.Attributes["POSITION"]
	if !ok {
		return fmt.Errorf("缺少位置属性")
	}

	positionData, err := parseAttributeData(doc, doc.Accessors[positionAccessor])
	if err != nil {
		return fmt.Errorf("解析位置数据失败: %w", err)
	}

	vertexCount := len(positionData) / 3
	if vertexCount <= 0 {
		return fmt.Errorf("无效的顶点数量: %d", vertexCount)
	}

	// 2. 创建Draco网格
	builder := draco.NewMeshBuilder()
	defer builder.Free()
	builder.Start(vertexCount)

	// 2. 添加位置属性
	builder.SetAttribute(vertexCount, positionData, draco.GAT_POSITION)

	// 3. 添加其他属性
	for name, accessorIdx := range primitive.Attributes {
		if name == "POSITION" {
			continue
		}
		attrType := dracoAttributeType(name)
		if attrType == draco.GAT_INVALID {
			continue
		}
		data, err := parseAttributeData(doc, doc.Accessors[accessorIdx])
		if err != nil {
			return fmt.Errorf("解析属性 %s 失败: %w", name, err)
		}

		builder.SetAttribute(vertexCount, data, attrType)
	}

	mesh := builder.GetMesh()

	// 添加面索引
	if primitive.Indices != nil {
		indices := parseIndexData(doc, primitive.Indices) // 新增索引解析
		mesh.Faces(indices)
	}

	// 4. 配置编码参数
	applyEncoderOptions(encoder, options) // 新增选项配置
	enc := draco.NewEncoder()

	// 5. 执行编码
	err, encodedData := enc.EncodeMesh(mesh)
	if err != nil {
		return fmt.Errorf("draco编码失败: %w", err)
	}

	// 5. 创建bufferView存储压缩数据
	buffer := &gltf.Buffer{
		Data: encodedData,
	}
	doc.Buffers = append(doc.Buffers, buffer)
	bufferIndex := uint32(len(doc.Buffers) - 1)

	view := &gltf.BufferView{
		Buffer:     bufferIndex,
		ByteLength: uint32(len(encodedData)),
	}
	doc.BufferViews = append(doc.BufferViews, view)
	viewIndex := uint32(len(doc.BufferViews) - 1)

	// 6. 构建扩展对象
	ext := &DracoExtension{
		BufferView: viewIndex,
		Attributes: make(map[string]uint32),
	}

	// 9. 映射属性ID
	for name := range primitive.Attributes {
		attrType := dracoAttributeType(name)
		if attrType == draco.GAT_INVALID {
			continue
		}

		if attrID := mesh.NamedAttributeID(attrType); attrID != -1 {
			ext.Attributes[name] = uint32(attrID)
		}
	}

	// 10. 更新图元
	primitive.Extensions = make(map[string]interface{})
	primitive.Extensions[ExtensionName] = ext
	primitive.Indices = nil
	primitive.Attributes = nil

	return nil
}

// 辅助函数：将glTF属性名映射为Draco属性类型
func dracoAttributeType(name string) draco.GeometryAttrType {
	switch name {
	case "POSITION":
		return draco.GAT_POSITION
	case "NORMAL":
		return draco.GAT_NORMAL
	case "TEXCOORD_0":
		return draco.GAT_TEX_COORD
	case "COLOR_0":
		return draco.GAT_COLOR
	default:
		return draco.GAT_GENERIC
	}
}

func parseIndexData(doc *gltf.Document, accessorIdx *uint32) []uint32 {
	if accessorIdx == nil {
		return nil
	}

	accessor := doc.Accessors[*accessorIdx]
	if accessor.BufferView == nil {
		return nil
	}

	viewIdx := uint32(*accessor.BufferView)
	if viewIdx >= uint32(len(doc.BufferViews)) {
		return nil
	}

	view := doc.BufferViews[viewIdx]
	if view.Buffer >= uint32(len(doc.Buffers)) {
		return nil
	}

	buffer := doc.Buffers[view.Buffer]
	if buffer.Data == nil {
		return nil
	}

	start := view.ByteOffset + accessor.ByteOffset
	end := start + accessor.Count*uint32(sizeOfComponent(accessor.ComponentType))
	if end > uint32(len(buffer.Data)) {
		return nil
	}

	data := buffer.Data[start:end]
	count := int(accessor.Count)
	indices := make([]uint32, count)

	switch accessor.ComponentType {
	case gltf.ComponentUbyte:
		for i := 0; i < count; i++ {
			indices[i] = uint32(data[i])
		}
	case gltf.ComponentUshort:
		for i := 0; i < count; i++ {
			indices[i] = uint32(binary.LittleEndian.Uint16(data[i*2:]))
		}
	case gltf.ComponentUint:
		for i := 0; i < count; i++ {
			indices[i] = binary.LittleEndian.Uint32(data[i*4:])
		}
	default:
		return nil
	}
	return indices
}

func parseAttributeData(doc *gltf.Document, accessor *gltf.Accessor) ([]float32, error) {
	if accessor.BufferView == nil {
		return nil, fmt.Errorf("访问器缺少BufferView")
	}

	viewIdx := uint32(*accessor.BufferView)
	if viewIdx >= uint32(len(doc.BufferViews)) {
		return nil, fmt.Errorf("无效的BufferView索引: %d", viewIdx)
	}

	view := doc.BufferViews[viewIdx]
	if view.Buffer >= uint32(len(doc.Buffers)) {
		return nil, fmt.Errorf("无效的缓冲区索引: %d", view.Buffer)
	}

	buffer := doc.Buffers[view.Buffer]
	if buffer.Data == nil {
		return nil, fmt.Errorf("缓冲区数据为空")
	}

	start := view.ByteOffset + accessor.ByteOffset
	end := start + view.ByteLength
	if end > uint32(len(buffer.Data)) {
		return nil, fmt.Errorf("缓冲区范围超出: %d-%d (缓冲区大小: %d)",
			start, end, len(buffer.Data))
	}

	data := buffer.Data[start:end]
	count := int(accessor.Count)
	components := componentsPerType(accessor.Type)
	if components == 0 {
		return nil, fmt.Errorf("无效的访问器类型: %s", accessor.Type)
	}

	result := make([]float32, count*int(components))
	stride := int(view.ByteStride)
	if stride == 0 {
		stride = sizeOfComponent(accessor.ComponentType) * int(components)
	}

	offset := 0
	for i := 0; i < count; i++ {
		segment := data[i*stride:]
		for j := 0; j < int(components); j++ {
			if len(segment) < sizeOfComponent(accessor.ComponentType) {
				return nil, fmt.Errorf("数据不足")
			}

			result[offset] = readComponent(segment, accessor.ComponentType)
			segment = segment[sizeOfComponent(accessor.ComponentType):]
			offset++
		}
	}
	return result, nil
}

func readComponent(data []byte, compType gltf.ComponentType) float32 {
	switch compType {
	case gltf.ComponentFloat:
		return math.Float32frombits(binary.LittleEndian.Uint32(data))
	case gltf.ComponentUbyte:
		return float32(data[0]) / 255.0
	case gltf.ComponentByte:
		return float32(int8(data[0])) / 127.0
	case gltf.ComponentUshort:
		return float32(binary.LittleEndian.Uint16(data)) / 65535.0
	case gltf.ComponentShort:
		return float32(int16(binary.LittleEndian.Uint16(data))) / 32767.0
	default:
		return 0
	}
}

// 获取组件大小
func sizeOfComponent(compType gltf.ComponentType) int {
	switch compType {
	case gltf.ComponentByte, gltf.ComponentUbyte:
		return 1
	case gltf.ComponentShort, gltf.ComponentUshort:
		return 2
	case gltf.ComponentFloat, gltf.ComponentUint:
		return 4
	default:
		return 4
	}
}

func applyEncoderOptions(encoder *draco.Encoder, options map[string]interface{}) {
	// 应用量化选项
	if quant, ok := options["quantization"].(map[string]int); ok {
		if bits, ok := quant["position"]; ok {
			encoder.SetAttributeQuantization(draco.GAT_POSITION, int32(bits))
		}
		if bits, ok := quant["normal"]; ok {
			encoder.SetAttributeQuantization(draco.GAT_NORMAL, int32(bits))
		}
		if bits, ok := quant["texcoord"]; ok {
			encoder.SetAttributeQuantization(draco.GAT_TEX_COORD, int32(bits))
		}
		if bits, ok := quant["color"]; ok {
			encoder.SetAttributeQuantization(draco.GAT_COLOR, int32(bits))
		}
		if bits, ok := quant["generic"]; ok {
			encoder.SetAttributeQuantization(draco.GAT_GENERIC, int32(bits))
		}
	}
}
