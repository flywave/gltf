package instance

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"

	"github.com/flywave/gltf"
)

const ExtensionName = "EXT_mesh_gpu_instancing"

func init() {
	gltf.RegisterExtension(ExtensionName, Unmarshal)
}

// 扩展数据结构
type InstanceAttributes struct {
	Attributes map[string]uint32 `json:"attributes"`
}

// 用于解析的包装结构
type instanceEnvelope struct {
	Attributes *InstanceAttributes `json:"attributes"`
}

// Unmarshal 将 JSON 数据解析为扩展数据结构
func Unmarshal(data []byte) (interface{}, error) {
	// 尝试解析为完整扩展对象
	var fullExt instanceEnvelope
	if err := json.Unmarshal(data, &fullExt); err == nil && fullExt.Attributes != nil {
		return fullExt.Attributes, nil
	}

	// 尝试解析为直接的 attributes 对象
	var directAttrs InstanceAttributes
	if err := json.Unmarshal(data, &directAttrs); err == nil && directAttrs.Attributes != nil {
		return &directAttrs, nil
	}

	// 两种格式都失败，返回错误
	return nil, fmt.Errorf("无法解析 EXT_mesh_gpu_instancing 数据")
}

type InstanceConfig struct {
	TranslationType gltf.ComponentType
	RotationType    gltf.ComponentType
	ScaleType       gltf.ComponentType
	Normalized      bool
}

func DefaultConfig() *InstanceConfig {
	return &InstanceConfig{
		TranslationType: gltf.ComponentFloat,
		RotationType:    gltf.ComponentFloat,
		ScaleType:       gltf.ComponentFloat,
		Normalized:      false,
	}
}

type InstanceData struct {
	Translations [][3]float32
	Rotations    [][4]float32
	Scales       [][3]float32
}

func WriteInstancing(doc *gltf.Document, data *InstanceData, config *InstanceConfig) error {
	attrs := make(map[string]uint32)

	properties := []struct {
		name       string
		data       interface{}
		compType   gltf.ComponentType
		accType    gltf.AccessorType
		normalized bool
	}{
		{"TRANSLATION", data.Translations, config.TranslationType, gltf.AccessorVec3, config.Normalized},
		{"ROTATION", data.Rotations, config.RotationType, gltf.AccessorVec4,
			config.RotationType != gltf.ComponentFloat}, // 整数类型需要归一化
		{"SCALE", data.Scales, config.ScaleType, gltf.AccessorVec3, config.Normalized},
	}

	for _, prop := range properties {
		idx, err := createVectorAccessor(doc, prop.data, prop.compType, prop.accType, prop.normalized)
		if err != nil {
			return fmt.Errorf("%s属性创建失败: %w", prop.name, err)
		}
		attrs[prop.name] = idx
	}

	if doc.Extensions == nil {
		doc.Extensions = make(gltf.Extensions)
	}
	doc.Extensions[ExtensionName] = map[string]interface{}{"attributes": attrs}
	doc.AddExtensionUsed(ExtensionName)

	return nil
}

func ReadInstancing(doc *gltf.Document, nodeIndex uint32) (*InstanceData, error) {
	if int(nodeIndex) >= len(doc.Nodes) {
		return nil, fmt.Errorf("节点索引超出范围")
	}

	node := doc.Nodes[nodeIndex]
	if node.Extensions == nil {
		return nil, fmt.Errorf("节点未使用 %s 扩展", ExtensionName)
	}

	ext, ok := node.Extensions[ExtensionName]
	if !ok {
		return nil, fmt.Errorf("节点缺少 %s 扩展", ExtensionName)
	}

	attrsData, err := parseAttributes(ext)
	if err != nil {
		return nil, fmt.Errorf("解析实例属性失败: %w", err)
	}

	data := &InstanceData{}
	var instanceCount uint32
	var countMismatch bool

	// 处理TRANSLATION属性
	if idx, exists := attrsData["TRANSLATION"]; exists {
		data.Translations, instanceCount, err = readVec3Data(doc, idx)
		if err != nil {
			return nil, fmt.Errorf("读取TRANSLATION失败: %w", err)
		}
	}

	// 处理ROTATION属性
	if idx, exists := attrsData["ROTATION"]; exists {
		accessor := doc.Accessors[idx]
		// 根据组件类型使用不同的读取方法
		switch accessor.ComponentType {
		case gltf.ComponentFloat:
			data.Rotations, _, err = readVec4Data(doc, idx)
		case gltf.ComponentByte:
			data.Rotations, _, err = readNormalizedVec4Byte(doc, idx)
		case gltf.ComponentShort:
			data.Rotations, _, err = readNormalizedVec4Short(doc, idx)
		default:
			err = fmt.Errorf("不支持的ROTATION组件类型: %v", accessor.ComponentType)
		}

		if err != nil {
			return nil, fmt.Errorf("读取ROTATION失败: %w", err)
		}

		if instanceCount > 0 && uint32(len(data.Rotations)) != instanceCount {
			countMismatch = true
		}
	}

	// 处理SCALE属性
	if idx, exists := attrsData["SCALE"]; exists {
		data.Scales, _, err = readVec3Data(doc, idx)
		if err != nil {
			return nil, fmt.Errorf("读取SCALE失败: %w", err)
		}
		if instanceCount > 0 && uint32(len(data.Scales)) != instanceCount {
			countMismatch = true
		}
	}

	if countMismatch {
		return nil, fmt.Errorf("实例属性数量不一致")
	}

	if len(data.Translations)+len(data.Rotations)+len(data.Scales) == 0 {
		return nil, fmt.Errorf("未找到有效的实例属性")
	}

	return data, nil
}

// 解析属性映射
func parseAttributes(ext interface{}) (map[string]uint32, error) {
	switch v := ext.(type) {
	case *InstanceAttributes:
		return v.Attributes, nil
	case InstanceAttributes:
		return v.Attributes, nil
	case map[string]interface{}:
		attrs := make(map[string]uint32)
		for key, val := range v {
			if idx, ok := val.(float64); ok {
				attrs[key] = uint32(idx)
			} else if idx, ok := val.(uint32); ok {
				attrs[key] = idx
			} else {
				return nil, fmt.Errorf("无效的属性索引类型: %T", val)
			}
		}
		return attrs, nil
	default:
		return nil, fmt.Errorf("未知的扩展类型: %T", ext)
	}
}

// 读取三维向量数据 (FLOAT)
func readVec3Data(doc *gltf.Document, accessorIdx uint32) ([][3]float32, uint32, error) {
	accessor := doc.Accessors[accessorIdx]
	if accessor.Type != gltf.AccessorVec3 {
		return nil, 0, fmt.Errorf("访问器类型应为VEC3，实际为%s", accessor.Type)
	}

	if accessor.ComponentType != gltf.ComponentFloat {
		return nil, 0, fmt.Errorf("TRANSLATION仅支持FLOAT组件类型")
	}

	data, err := readBufferData(doc, accessor)
	if err != nil {
		return nil, 0, err
	}

	count := accessor.Count
	result := make([][3]float32, count)
	for i := uint32(0); i < count; i++ {
		result[i] = [3]float32{
			data[3*i],
			data[3*i+1],
			data[3*i+2],
		}
	}

	return result, count, nil
}

// 读取四维向量数据 (FLOAT)
func readVec4Data(doc *gltf.Document, accessorIdx uint32) ([][4]float32, uint32, error) {
	accessor := doc.Accessors[accessorIdx]
	if accessor.Type != gltf.AccessorVec4 {
		return nil, 0, fmt.Errorf("访问器类型应为VEC4，实际为%s", accessor.Type)
	}

	if accessor.ComponentType != gltf.ComponentFloat {
		return nil, 0, fmt.Errorf("ROTATION FLOAT类型要求FLOAT组件类型")
	}

	data, err := readBufferData(doc, accessor)
	if err != nil {
		return nil, 0, err
	}

	count := accessor.Count
	result := make([][4]float32, count)
	for i := uint32(0); i < count; i++ {
		result[i] = [4]float32{
			data[4*i],
			data[4*i+1],
			data[4*i+2],
			data[4*i+3],
		}
	}

	return result, count, nil
}

// 读取归一化的字节四维向量 (ROTATION)
func readNormalizedVec4Byte(doc *gltf.Document, accessorIdx uint32) ([][4]float32, uint32, error) {
	accessor := doc.Accessors[accessorIdx]
	if accessor.Type != gltf.AccessorVec4 {
		return nil, 0, fmt.Errorf("访问器类型应为VEC4，实际为%s", accessor.Type)
	}

	if accessor.ComponentType != gltf.ComponentByte {
		return nil, 0, fmt.Errorf("应为BYTE组件类型")
	}

	if !accessor.Normalized {
		return nil, 0, fmt.Errorf("BYTE ROTATION必须归一化")
	}

	data, count, err := readByteVec4Data(doc, accessor)
	if err != nil {
		return nil, 0, err
	}

	result := make([][4]float32, count)
	for i := uint32(0); i < count; i++ {
		// 将归一化的字节转换为浮点数 [-1, 1]
		result[i] = [4]float32{
			float32(data[4*i]) / 127.0,
			float32(data[4*i+1]) / 127.0,
			float32(data[4*i+2]) / 127.0,
			float32(data[4*i+3]) / 127.0,
		}
	}

	return result, count, nil
}

// 读取归一化的短整型四维向量 (ROTATION)
func readNormalizedVec4Short(doc *gltf.Document, accessorIdx uint32) ([][4]float32, uint32, error) {
	accessor := doc.Accessors[accessorIdx]
	if accessor.Type != gltf.AccessorVec4 {
		return nil, 0, fmt.Errorf("访问器类型应为VEC4，实际为%s", accessor.Type)
	}

	if accessor.ComponentType != gltf.ComponentShort {
		return nil, 0, fmt.Errorf("应为SHORT组件类型")
	}

	if !accessor.Normalized {
		return nil, 0, fmt.Errorf("SHORT ROTATION必须归一化")
	}

	data, count, err := readShortVec4Data(doc, accessor)
	if err != nil {
		return nil, 0, err
	}

	result := make([][4]float32, count)
	for i := uint32(0); i < count; i++ {
		// 将归一化的短整型转换为浮点数 [-1, 1]
		result[i] = [4]float32{
			float32(data[4*i]) / 32767.0,
			float32(data[4*i+1]) / 32767.0,
			float32(data[4*i+2]) / 32767.0,
			float32(data[4*i+3]) / 32767.0,
		}
	}

	return result, count, nil
}

// 读取字节向量数据
func readByteVec4Data(doc *gltf.Document, accessor *gltf.Accessor) ([]int8, uint32, error) {
	bufferView := doc.BufferViews[*accessor.BufferView]
	buffer := doc.Buffers[bufferView.Buffer]

	offset := bufferView.ByteOffset + accessor.ByteOffset
	byteData := buffer.Data[offset : offset+bufferView.ByteLength]

	count := accessor.Count
	result := make([]int8, count*4)
	if err := binary.Read(bytes.NewReader(byteData), binary.LittleEndian, &result); err != nil {
		return nil, 0, fmt.Errorf("二进制读取失败: %w", err)
	}

	return result, count, nil
}

// 读取短整型向量数据
func readShortVec4Data(doc *gltf.Document, accessor *gltf.Accessor) ([]int16, uint32, error) {
	bufferView := doc.BufferViews[*accessor.BufferView]
	buffer := doc.Buffers[bufferView.Buffer]

	offset := bufferView.ByteOffset + accessor.ByteOffset
	byteData := buffer.Data[offset : offset+bufferView.ByteLength]

	count := accessor.Count
	result := make([]int16, count*4)
	if err := binary.Read(bytes.NewReader(byteData), binary.LittleEndian, &result); err != nil {
		return nil, 0, fmt.Errorf("二进制读取失败: %w", err)
	}

	return result, count, nil
}

// 从缓冲区读取浮点数据
func readBufferData(doc *gltf.Document, accessor *gltf.Accessor) ([]float32, error) {
	bufferView := doc.BufferViews[*accessor.BufferView]
	buffer := doc.Buffers[bufferView.Buffer]

	offset := bufferView.ByteOffset + accessor.ByteOffset
	byteData := buffer.Data[offset : offset+bufferView.ByteLength]

	count := accessor.Count
	var dim uint32
	switch accessor.Type {
	case gltf.AccessorVec3:
		dim = 3
	case gltf.AccessorVec4:
		dim = 4
	default:
		return nil, fmt.Errorf("不支持的访问器类型: %s", accessor.Type)
	}

	result := make([]float32, count*dim)
	if err := binary.Read(bytes.NewReader(byteData), binary.LittleEndian, &result); err != nil {
		return nil, fmt.Errorf("二进制读取失败: %w", err)
	}

	return result, nil
}

// 创建向量访问器
func createVectorAccessor(
	doc *gltf.Document,
	data interface{},
	compType gltf.ComponentType,
	accType gltf.AccessorType,
	normalized bool,
) (uint32, error) {
	var byteData []byte
	var count uint32

	switch v := data.(type) {
	case [][3]float32:
		count = uint32(len(v))
		if count > 0 {
			byteData = flattenFloat32Slice(v, 3)
		}
	case [][4]float32:
		count = uint32(len(v))
		if count > 0 {
			byteData = flattenFloat32Slice(v, 4)
		}
	case [][4]int8: // 支持字节类型的ROTATION
		count = uint32(len(v))
		if count > 0 {
			byteData = flattenInt8Slice(v, 4)
		}
	case [][4]int16: // 支持短整型类型的ROTATION
		count = uint32(len(v))
		if count > 0 {
			byteData = flattenInt16Slice(v, 4)
		}
	default:
		return 0, fmt.Errorf("不支持的向量数据类型: %T", data)
	}

	if count == 0 {
		accessor := &gltf.Accessor{
			ComponentType: compType,
			Type:          accType,
			Count:         0,
			Normalized:    normalized,
		}
		doc.Accessors = append(doc.Accessors, accessor)
		return uint32(len(doc.Accessors) - 1), nil
	}

	viewIdx, err := addBufferView(doc, byteData)
	if err != nil {
		return 0, fmt.Errorf("创建BufferView失败: %w", err)
	}

	accessor := &gltf.Accessor{
		BufferView:    gltf.Index(viewIdx),
		ComponentType: compType,
		Type:          accType,
		Count:         count,
		Normalized:    normalized,
	}

	doc.Accessors = append(doc.Accessors, accessor)
	return uint32(len(doc.Accessors) - 1), nil
}

// 转换float32切片为字节数组
func flattenFloat32Slice(slice interface{}, dim int) []byte {
	buf := bytes.NewBuffer(nil)
	binary.Write(buf, binary.LittleEndian, slice)
	return buf.Bytes()
}

// 转换int8切片为字节数组 (用于ROTATION)
func flattenInt8Slice(slice [][4]int8, dim int) []byte {
	buf := bytes.NewBuffer(nil)
	for _, v := range slice {
		binary.Write(buf, binary.LittleEndian, v[:])
	}
	return buf.Bytes()
}

// 转换int16切片为字节数组 (用于ROTATION)
func flattenInt16Slice(slice [][4]int16, dim int) []byte {
	buf := bytes.NewBuffer(nil)
	for _, v := range slice {
		binary.Write(buf, binary.LittleEndian, v[:])
	}
	return buf.Bytes()
}

// 添加缓冲区视图
func addBufferView(doc *gltf.Document, data []byte) (uint32, error) {
	if len(doc.Buffers) == 0 {
		doc.Buffers = []*gltf.Buffer{{ByteLength: 0}}
	}
	buffer := doc.Buffers[0]

	view := &gltf.BufferView{
		Buffer:     0,
		ByteOffset: buffer.ByteLength,
		ByteLength: uint32(len(data)),
	}

	buffer.Data = append(buffer.Data, data...)
	buffer.ByteLength += uint32(len(data))
	pad := (4 - (buffer.ByteLength % 4)) % 4
	buffer.Data = append(buffer.Data, make([]byte, pad)...)

	doc.BufferViews = append(doc.BufferViews, view)
	return uint32(len(doc.BufferViews) - 1), nil
}
