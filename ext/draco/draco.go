package draco

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math"

	"github.com/flywave/gltf"
	"github.com/flywave/go-draco"
	"github.com/flywave/go3d/vec2"
	"github.com/flywave/go3d/vec3"
	"github.com/flywave/go3d/vec4"
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

	// 清理标记为nil的缓冲区视图
	cleanUpUnusedResources(doc) // 改为调用统一的清理函数

	doc.RemoveExtension(ExtensionName)
	// 确保从ExtensionsUsed中移除
	for i, ext := range doc.ExtensionsUsed {
		if ext == ExtensionName {
			doc.ExtensionsUsed = append(doc.ExtensionsUsed[:i], doc.ExtensionsUsed[i+1:]...)
			break
		}
	}
	return nil
}

// Decode 对文档中的所有网格应用Draco解压缩
func Decode(doc *gltf.Document) error {
	return DecodeAll(doc)
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
	defer mesh.Free()

	err := decoder.DecodeMesh(mesh, compressedData)
	if err != nil {
		return fmt.Errorf("draco解码失败: %v", err)
	}

	// 处理索引数据
	var indices []uint32
	if mesh.NumFaces() > 0 {
		faceCount := mesh.NumFaces()
		indices = make([]uint32, faceCount*3)
		mesh.Faces(indices)
	}

	// 更新索引访问器
	if primitive.Indices != nil && len(indices) > 0 {
		origAccessor := doc.Accessors[*primitive.Indices]
		if origAccessor == nil {
			return fmt.Errorf("索引访问器不存在")
		}

		if err := updateIndexAccessor(doc, origAccessor, indices); err != nil {
			return fmt.Errorf("更新索引数据失败: %w", err)
		}
	}

	// 获取顶点数量
	vertCount := int(mesh.NumPoints())
	if vertCount <= 0 {
		return fmt.Errorf("无效的顶点数量: %d", vertCount)
	}

	// 更新顶点属性
	for name, id := range ext.Attributes {
		attr := mesh.Attr(int32(id))
		if attr == nil {
			return fmt.Errorf("找不到属性: %s", name)
		}

		// 获取原始访问器ID
		origAccessorID, exists := primitive.Attributes[name]
		if !exists {
			return fmt.Errorf("找不到原始访问器: %s", name)
		}

		origAccessor := doc.Accessors[origAccessorID]
		if origAccessor == nil {
			return fmt.Errorf("无效的访问器索引: %d", origAccessorID)
		}

		// 获取属性类型和组件数
		components := componentsPerType(origAccessor.Type)
		if components == 0 {
			return fmt.Errorf("无效的属性类型: %s", origAccessor.Type)
		}

		// 获取属性数据
		outVert := make([]float32, vertCount*int(components))
		if _, ok := mesh.AttrData(attr, outVert); !ok {
			return fmt.Errorf("获取属性数据失败")
		}

		// 更新位置属性的min/max
		if name == "POSITION" || name == "NORMAL" {
			min, max := calculateMinMax(outVert, int(components))
			origAccessor.Min = min
			origAccessor.Max = max
		}

		// 更新访问器计数
		origAccessor.Count = uint32(vertCount)

		// 转换数据为字节并更新缓冲区
		byteData := float32ToBytes(outVert)
		// 重置原始访问器的BufferView，确保创建新的缓冲区存储解压后数据
		origAccessor.BufferView = nil
		if err := updateAccessorDataWithBytes(doc, origAccessor, byteData); err != nil {
			return fmt.Errorf("更新属性 %s 失败: %w", name, err)
		}
	}

	// 移除Draco扩展
	delete(primitive.Extensions, ExtensionName)

	// 标记压缩数据BufferView为待删除
	removeCompressedData(doc, ext.BufferView)

	return nil
}

// 为访问器创建新的缓冲区和视图
func createNewBufferForAccessor(doc *gltf.Document, accessor *gltf.Accessor, data []byte) error {
	// 创建新的缓冲区
	buffer := &gltf.Buffer{
		Data:       data,
		ByteLength: uint32(len(data)),
	}
	doc.Buffers = append(doc.Buffers, buffer)
	bufferIndex := uint32(len(doc.Buffers) - 1)

	// 创建新的BufferView
	view := &gltf.BufferView{
		Buffer:     bufferIndex,
		ByteOffset: 0,
		ByteLength: uint32(len(data)),
	}

	// 根据访问器类型设置目标
	switch accessor.Type {
	case gltf.AccessorScalar:
		view.Target = gltf.TargetElementArrayBuffer
	default:
		view.Target = gltf.TargetArrayBuffer
	}

	doc.BufferViews = append(doc.BufferViews, view)
	accessor.BufferView = gltf.Index(uint32(len(doc.BufferViews) - 1))
	accessor.ByteOffset = 0
	return nil
}

// 更新缓冲区视图数据
func updateBufferViewData(doc *gltf.Document, view *gltf.BufferView, byteOffset uint32, data []byte) error {
	if view.Buffer >= uint32(len(doc.Buffers)) {
		return fmt.Errorf("无效的缓冲区索引: %d", view.Buffer)
	}

	buffer := doc.Buffers[view.Buffer]
	start := view.ByteOffset + byteOffset
	end := start + uint32(len(data))

	// 检查是否需要扩展缓冲区
	if end > uint32(len(buffer.Data)) {
		// 扩展缓冲区
		newSize := end
		if newSize < uint32(len(buffer.Data))*2 {
			newSize = uint32(len(buffer.Data)) * 2
		}
		newData := make([]byte, newSize)
		copy(newData, buffer.Data)
		buffer.Data = newData
		buffer.ByteLength = uint32(len(newData))
	}

	// 更新数据
	copy(buffer.Data[start:end], data)
	return nil
}

// 更新索引访问器 - 处理没有bufferView的情况
func updateIndexAccessor(doc *gltf.Document, accessor *gltf.Accessor, indices []uint32) error {
	// 确定组件类型
	maxIndex := uint32(0)
	for _, idx := range indices {
		if idx > maxIndex {
			maxIndex = idx
		}
	}

	var componentType gltf.ComponentType
	switch {
	case maxIndex <= math.MaxUint8:
		componentType = gltf.ComponentUbyte
	case maxIndex <= math.MaxUint16:
		componentType = gltf.ComponentUshort
	default:
		componentType = gltf.ComponentUint
	}

	// 更新访问器元数据
	accessor.ComponentType = componentType
	accessor.Count = uint32(len(indices))

	// 转换索引数据为字节
	byteData := indicesToBytes(indices, componentType)

	// 更新缓冲区数据
	return updateAccessorDataWithBytes(doc, accessor, byteData)
}

// 使用字节数据更新访问器
func updateAccessorDataWithBytes(doc *gltf.Document, accessor *gltf.Accessor, data []byte) error {
	// 如果访问器没有bufferView，创建新的
	if accessor.BufferView == nil {
		return createNewBufferForAccessor(doc, accessor, data)
	}

	// 获取关联的BufferView
	viewIdx := uint32(*accessor.BufferView)
	if viewIdx >= uint32(len(doc.BufferViews)) {
		return fmt.Errorf("无效的BufferView索引: %d", viewIdx)
	}
	view := doc.BufferViews[viewIdx]

	// 更新缓冲区数据
	return updateBufferViewData(doc, view, accessor.ByteOffset, data)
}

// 移除压缩数据
func removeCompressedData(doc *gltf.Document, bufferViewID uint32) {
	if int(bufferViewID) >= len(doc.BufferViews) {
		return
	}

	// 标记压缩缓冲区视图为已删除
	doc.BufferViews[bufferViewID] = nil
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
	buf.Grow(len(indices) * gltf.SizeOfComponent(componentType))

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

// Encode 对文档中的所有网格应用Draco压缩
func Encode(doc *gltf.Document) error {
	return EncodeAll(doc, nil)
}

// EncodeWithOptions 对文档中的所有网格应用Draco压缩，并允许指定编码选项
func EncodeWithOptions(doc *gltf.Document, options map[string]interface{}) error {
	return EncodeAll(doc, options)
}

// EncodePrimitive 对单个图元应用Draco压缩
func EncodePrimitive(doc *gltf.Document, primitive *gltf.Primitive) error {
	// 注意：根据go-draco库的实现，Encoder可能不需要手动释放
	// 如果需要释放资源，请参考go-draco的文档
	encoder := draco.NewEncoder()
	return encodePrimitive(doc, encoder, primitive, nil)
}

// EncodePrimitiveWithOptions 对单个图元应用Draco压缩，并允许指定编码选项
func EncodePrimitiveWithOptions(doc *gltf.Document, primitive *gltf.Primitive, options map[string]interface{}) error {
	// 注意：根据go-draco库的实现，Encoder可能不需要手动释放
	// 如果需要释放资源，请参考go-draco的文档
	encoder := draco.NewEncoder()
	return encodePrimitive(doc, encoder, primitive, options)
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

	// 新增：全局清理未使用的缓冲区资源
	cleanUpUnusedResources(doc)

	doc.AddExtensionUsed(ExtensionName)
	return nil
}

func encodePrimitive(doc *gltf.Document, encoder *draco.Encoder, primitive *gltf.Primitive, options map[string]interface{}) error {
	// 检查必要数据
	if primitive.Indices == nil {
		return fmt.Errorf("图元缺少索引")
	}
	if _, ok := primitive.Attributes["POSITION"]; !ok {
		return fmt.Errorf("缺少位置属性")
	}

	// 获取索引数据
	indexAcc := doc.Accessors[*primitive.Indices]
	indices, err := parseIndexData(doc, indexAcc)
	if err != nil {
		return fmt.Errorf("解析索引数据失败: %w", err)
	}

	// 获取位置数据
	positionAccessor := doc.Accessors[primitive.Attributes["POSITION"]]
	positionData, err := parseAttributeData(doc, positionAccessor)
	if err != nil {
		return fmt.Errorf("解析位置数据失败: %w", err)
	}

	vertexCount := int(positionAccessor.Count)
	if vertexCount <= 0 {
		return fmt.Errorf("无效的顶点数量: %d", vertexCount)
	}

	// 创建Draco网格
	builder := draco.NewMeshBuilder()
	defer builder.Free()
	builder.Start(vertexCount)

	// 添加位置属性
	posData := make([]vec3.T, vertexCount)
	for i := 0; i < vertexCount; i++ {
		posData[i] = vec3.T{
			positionData[i*3],
			positionData[i*3+1],
			positionData[i*3+2],
		}
	}
	posIndex := builder.SetAttribute(vertexCount, posData, draco.GAT_POSITION)

	attrMap := make(map[string]uint32)
	attrMap["POSITION"] = posIndex

	// 添加其他属性
	for name, accessorIdx := range primitive.Attributes {
		if name == "POSITION" {
			continue
		}

		attrType := dracoAttributeType(name)
		if attrType == draco.GAT_INVALID {
			continue
		}

		attrAcc := doc.Accessors[accessorIdx]
		data, cerr := parseAttributeData(doc, attrAcc)
		if cerr != nil {
			return fmt.Errorf("解析属性 %s 失败: %w", name, cerr)
		}

		// 根据属性类型处理数据
		accType, _ := getAccessorType(name)
		components := componentsPerType(accType)

		switch components {
		case 2:
			vec2Data := make([]vec2.T, vertexCount)
			for i := 0; i < vertexCount; i++ {
				vec2Data[i] = vec2.T{data[i*2], data[i*2+1]}
			}
			attrMap[name] = builder.SetAttribute(vertexCount, vec2Data, attrType)
		case 3:
			vec3Data := make([]vec3.T, vertexCount)
			for i := 0; i < vertexCount; i++ {
				vec3Data[i] = vec3.T{data[i*3], data[i*3+1], data[i*3+2]}
			}
			attrMap[name] = builder.SetAttribute(vertexCount, vec3Data, attrType)
		case 4:
			vec4Data := make([]vec4.T, vertexCount)
			for i := 0; i < vertexCount; i++ {
				vec4Data[i] = vec4.T{data[i*4], data[i*4+1], data[i*4+2], data[i*4+3]}
			}
			attrMap[name] = builder.SetAttribute(vertexCount, vec4Data, attrType)
		default:
			return fmt.Errorf("不支持的组件数量: %d", components)
		}
	}

	// 设置面数据
	faceCount := len(indices) / 3
	if faceCount*3 != len(indices) {
		return fmt.Errorf("索引数量必须是3的倍数")
	}

	mesh := builder.GetMesh()
	defer mesh.Free()

	// 配置编码参数
	applyEncoderOptions(encoder, options)

	// 执行编码
	err, encodedData := encoder.EncodeMesh(mesh)
	if err != nil {
		return fmt.Errorf("draco编码失败: %v", err)
	}

	// 创建新的缓冲区存储压缩数据（独立存储）
	buffer := &gltf.Buffer{
		ByteLength: uint32(len(encodedData)),
		Data:       encodedData,
	}
	doc.Buffers = append(doc.Buffers, buffer)
	bufferIndex := uint32(len(doc.Buffers) - 1)

	// 创建缓冲区视图
	view := &gltf.BufferView{
		Buffer:     bufferIndex,
		ByteOffset: 0,
		ByteLength: uint32(len(encodedData)),
	}
	doc.BufferViews = append(doc.BufferViews, view)
	viewIndex := uint32(len(doc.BufferViews) - 1)

	// 添加填充字节以满足4字节对齐
	padding := paddingBytes(len(encodedData))
	if padding > 0 {
		buffer.Data = append(buffer.Data, make([]byte, padding)...)
		buffer.ByteLength += uint32(padding)
	}

	// 构建扩展对象
	ext := &DracoExtension{
		BufferView: viewIndex,
		Attributes: attrMap,
	}

	// 更新图元
	if primitive.Extensions == nil {
		primitive.Extensions = make(map[string]interface{})
	}
	primitive.Extensions[ExtensionName] = ext

	// 11. 将原始访问器的BufferView置空
	// 置空索引访问器
	if primitive.Indices != nil {
		indexAccessor := doc.Accessors[*primitive.Indices]
		indexAccessor.BufferView = nil
		indexAccessor.ByteOffset = 0
	}

	// 置空所有属性访问器
	for _, accessorIdx := range primitive.Attributes {
		attrAccessor := doc.Accessors[accessorIdx]
		attrAccessor.BufferView = nil
		attrAccessor.ByteOffset = 0
	}
	return nil
}

func parseIndexData(doc *gltf.Document, accessor *gltf.Accessor) ([]uint32, error) {
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
	indices := make([]uint32, count)

	stride := int(view.ByteStride)
	if stride == 0 {
		stride = gltf.SizeOfComponent(accessor.ComponentType)
	}

	for i := 0; i < count; i++ {
		segment := data[i*stride:]
		switch accessor.ComponentType {
		case gltf.ComponentUbyte:
			indices[i] = uint32(segment[0])
		case gltf.ComponentUshort:
			indices[i] = uint32(binary.LittleEndian.Uint16(segment))
		case gltf.ComponentUint:
			indices[i] = binary.LittleEndian.Uint32(segment)
		default:
			return nil, fmt.Errorf("不支持的索引组件类型: %s", accessor.ComponentType)
		}
	}
	return indices, nil
}

// 计算需要的填充字节数
func paddingBytes(size int) int {
	remainder := size % 4
	if remainder == 0 {
		return 0
	}
	return 4 - remainder
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
		stride = gltf.SizeOfComponent(accessor.ComponentType) * int(components)
	}

	offset := 0
	for i := 0; i < count; i++ {
		segment := data[i*stride:]
		for j := 0; j < int(components); j++ {
			if len(segment) < gltf.SizeOfComponent(accessor.ComponentType) {
				return nil, fmt.Errorf("数据不足")
			}

			result[offset] = readComponent(segment, accessor.ComponentType)
			segment = segment[gltf.SizeOfComponent(accessor.ComponentType):]
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
	case gltf.ComponentUint:
		return float32(binary.LittleEndian.Uint32(data))
	default:
		return 0
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

func cleanUpUnusedResources(doc *gltf.Document) {
	// 步骤1: 收集所有仍被引用的 bufferView
	usedBufferViews := make(map[uint32]bool)
	for _, accessor := range doc.Accessors {
		if accessor.BufferView != nil {
			usedBufferViews[uint32(*accessor.BufferView)] = true
		}
	}

	// 收集 Draco 扩展引用的 bufferView
	for _, mesh := range doc.Meshes {
		for _, primitive := range mesh.Primitives {
			if extData, exists := primitive.Extensions[ExtensionName]; exists {
				if ext, ok := extData.(*DracoExtension); ok {
					usedBufferViews[ext.BufferView] = true
				}
			}
		}
	}

	// 步骤2: 清理未使用的 bufferView
	var validBufferViews []*gltf.BufferView
	bufferViewRemap := make(map[uint32]uint32) // 旧索引到新索引的映射

	for idx, view := range doc.BufferViews {
		if view != nil && usedBufferViews[uint32(idx)] {
			bufferViewRemap[uint32(idx)] = uint32(len(validBufferViews))
			validBufferViews = append(validBufferViews, view)
		}
	}
	doc.BufferViews = validBufferViews

	// 步骤3: 更新访问器中的 bufferView 引用
	for _, accessor := range doc.Accessors {
		if accessor.BufferView != nil {
			if newIdx, ok := bufferViewRemap[uint32(*accessor.BufferView)]; ok {
				accessor.BufferView = gltf.Index(newIdx)
			} else {
				accessor.BufferView = nil
			}
		}
	}

	for _, mesh := range doc.Meshes {
		for _, primitive := range mesh.Primitives {
			if extData, exists := primitive.Extensions[ExtensionName]; exists {
				if ext, ok := extData.(*DracoExtension); ok {
					if newIdx, ok := bufferViewRemap[ext.BufferView]; ok {
						ext.BufferView = newIdx // 更新为新的BufferView索引
					} else {
						// 如果不在映射表中，说明该BufferView已被移除
						delete(primitive.Extensions, ExtensionName)
					}
				}
			}
		}
	}

	// 步骤4: 收集所有仍被引用的 buffer
	usedBuffers := make(map[uint32]bool)
	for _, view := range doc.BufferViews {
		usedBuffers[view.Buffer] = true
	}

	// 步骤5: 清理未使用的 buffer
	var validBuffers []*gltf.Buffer
	bufferRemap := make(map[uint32]uint32) // 旧索引到新索引的映射

	for idx, buffer := range doc.Buffers {
		if buffer != nil && usedBuffers[uint32(idx)] {
			bufferRemap[uint32(idx)] = uint32(len(validBuffers))
			validBuffers = append(validBuffers, buffer)
		}
	}
	doc.Buffers = validBuffers

	// 步骤6: 更新 bufferView 中的 buffer 引用
	for _, view := range doc.BufferViews {
		if newIdx, ok := bufferRemap[view.Buffer]; ok {
			view.Buffer = newIdx
		} else {
			// 不应该发生，因为我们已经过滤了使用的 buffer
			view.Buffer = 0
		}
	}
}
