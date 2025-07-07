package splatting

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"

	"github.com/flywave/gltf"
	"github.com/flywave/gltf/ext/meshopt"
)

const (
	ExtensionName = "KHR_gaussian_splatting"
)

func init() {
	gltf.RegisterExtension(ExtensionName, UnmarshalGaussianSplatting)
}

type GaussianSplatting struct {
	Attributes         map[string]uint32   `json:"attributes"`
	SphericalHarmonics *SphericalHarmonics `json:"sphericalHarmonics,omitempty"`
	BufferView         int                 `json:"bufferView,omitempty"`
}

type SphericalHarmonics struct {
	Coefficients []float32 `json:"coefficients"`
}

func UnmarshalGaussianSplatting(data []byte) (interface{}, error) {
	gs := &GaussianSplatting{}
	if err := json.Unmarshal(data, gs); err != nil {
		return nil, fmt.Errorf("KHR_gaussian_splatting解析失败: %w", err)
	}

	// 验证必要属性
	required := map[string]struct{}{
		"POSITION":  {},
		"COLOR_0":   {},
		"_SCALE":    {},
		"_ROTATION": {},
	}

	for attr := range required {
		if _, exists := gs.Attributes[attr]; !exists {
			return nil, fmt.Errorf("缺少必要属性: %s", attr)
		}
	}

	return gs, nil
}

func (g *GaussianSplatting) MarshalJSON() ([]byte, error) {
	type Alias GaussianSplatting
	return json.Marshal(&struct {
		*Alias
	}{
		Alias: (*Alias)(g),
	})
}

// 压缩图元数据到bufferView
func (g *GaussianSplatting) Compress(doc *gltf.Document, data interface{}) error {
	buf := new(bytes.Buffer)

	// 支持多种数据类型压缩
	switch v := data.(type) {
	case []float32:
		if err := binary.Write(buf, binary.LittleEndian, v); err != nil {
			return fmt.Errorf("浮点数据写入失败: %w", err)
		}
	case []uint8:
		if err := binary.Write(buf, binary.LittleEndian, v); err != nil {
			return fmt.Errorf("字节数据写入失败: %w", err)
		}
	case []uint16:
		if err := binary.Write(buf, binary.LittleEndian, v); err != nil {
			return fmt.Errorf("短整型数据写入失败: %w", err)
		}
	default:
		return fmt.Errorf("不支持的数据类型: %T", data)
	}

	// 添加压缩标记并创建bufferView
	viewIndex, err := addBufferView(doc, buf.Bytes(), true)
	if err != nil {
		return fmt.Errorf("创建压缩bufferView失败: %w", err)
	}

	// 添加必要的扩展声明
	addExtensionUsed(doc, "KHR_mesh_quantization")
	addExtensionUsed(doc, "EXT_meshopt_compression")

	g.BufferView = viewIndex
	return nil
}

func addExtensionUsed(doc *gltf.Document, ext string) {
	for _, existing := range doc.ExtensionsUsed {
		if existing == ext {
			return
		}
	}
	doc.ExtensionsUsed = append(doc.ExtensionsUsed, ext)
}

// 添加缓冲区视图辅助函数
func addBufferView(doc *gltf.Document, data []byte, compress bool) (int, error) {
	if compress {
		compressed, err := compressWithMeshopt(data)
		if err == nil {
			data = compressed
			addExtensionUsed(doc, "EXT_meshopt_compression")
		}
	}
	if len(doc.Buffers) == 0 {
		doc.Buffers = append(doc.Buffers, &gltf.Buffer{})
	}

	buffer := doc.Buffers[0]
	buffer.ByteLength += uint32(len(data))

	view := &gltf.BufferView{
		Buffer:     0,
		ByteOffset: buffer.ByteLength - uint32(len(data)),
		ByteLength: uint32(len(data)),
	}

	doc.BufferViews = append(doc.BufferViews, view)
	return len(doc.BufferViews) - 1, nil
}

// 创建高斯泼溅图元扩展
func CreateGaussianPrimitive(doc *gltf.Document, attributes map[string]uint32, coefficients []float32) (*GaussianSplatting, error) {
	gs := &GaussianSplatting{
		Attributes: attributes,
	}

	if len(coefficients) > 0 {
		gs.SphericalHarmonics = &SphericalHarmonics{
			Coefficients: coefficients,
		}
	}

	// 自动添加扩展声明
	if !hasExtensionUsed(doc, ExtensionName) {
		doc.ExtensionsUsed = append(doc.ExtensionsUsed, ExtensionName)
	}

	return gs, nil
}

func hasExtensionUsed(doc *gltf.Document, ext string) bool {
	for _, e := range doc.ExtensionsUsed {
		if e == ext {
			return true
		}
	}
	return false
}

// QuantizationConfig 定义量化配置
type QuantizationConfig struct {
	PositionType gltf.ComponentType // 默认FLOAT
	ColorType    gltf.ComponentType // 推荐UNSIGNED_BYTE
	RotationType gltf.ComponentType // 推荐SHORT或FLOAT
	ScaleType    gltf.ComponentType // 推荐UNSIGNED_SHORT
	Normalized   bool               // 是否启用归一化
}

// DefaultQuantization 返回推荐的量化配置
func DefaultQuantization() *QuantizationConfig {
	return &QuantizationConfig{
		PositionType: gltf.ComponentFloat,
		ColorType:    gltf.ComponentUbyte,
		RotationType: gltf.ComponentShort,
		ScaleType:    gltf.ComponentUshort,
		Normalized:   true,
	}
}

// VertexData 定义高斯泼溅顶点数据结构
type VertexData struct {
	Positions []float32 `json:"-"` // 位置数据 (x, y, z)
	Colors    []float32 `json:"-"` // 颜色数据 (r, g, b, a)
	Scales    []float32 `json:"-"` // 缩放数据 (sx, sy, sz)
	Rotations []float32 `json:"-"` // 旋转数据 (rx, ry, rz, rw)
}

// WireGaussianSplatting 创建并关联高斯泼溅图元扩展
func WireGaussianSplatting(
	doc *gltf.Document,
	attributes map[string]uint32,
	coefficients []float32,
	vertexData *VertexData, // 顶点数据数组
	config *QuantizationConfig,
	compress bool,
) (*GaussianSplatting, error) {
	// 创建属性访问器
	attrs := make(map[string]uint32)

	// 创建位置属性访问器
	posIdx, err := createAccessor(doc, vertexData.Positions,
		config.PositionType, gltf.AccessorVec3, false, compress)
	if err != nil {
		return nil, fmt.Errorf("位置属性创建失败: %w", err)
	}
	attrs["POSITION"] = posIdx

	// 创建颜色属性访问器
	colorIdx, err := createAccessor(doc, vertexData.Colors,
		config.ColorType, gltf.AccessorVec4, config.Normalized, compress)
	if err != nil {
		return nil, fmt.Errorf("颜色属性创建失败: %w", err)
	}
	attrs["COLOR_0"] = colorIdx

	// 创建缩放属性访问器
	scaleIdx, err := createAccessor(doc, vertexData.Scales,
		config.ScaleType, gltf.AccessorVec3, config.Normalized, compress)
	if err != nil {
		return nil, fmt.Errorf("缩放属性创建失败: %w", err)
	}
	attrs["_SCALE"] = scaleIdx

	// 创建旋转属性访问器
	rotIdx, err := createAccessor(doc, vertexData.Rotations,
		config.RotationType, gltf.AccessorVec4, config.Normalized, compress)

	if err != nil {
		return nil, fmt.Errorf("旋转属性创建失败: %w", err)
	}
	attrs["_ROTATION"] = rotIdx

	// 合并顶点数据并压缩
	mergedData := mergeVertexData(vertexData)
	gs, err := CreateGaussianPrimitive(doc, attrs, coefficients)
	if err != nil {
		return nil, err
	}

	if err := gs.Compress(doc, mergedData); err != nil {
		return nil, err
	}
	// 自动创建图元并关联扩展
	primitive := &gltf.Primitive{
		Attributes: attributes,
		Extensions: make(gltf.Extensions),
	}
	primitive.Extensions[ExtensionName] = gs

	// 确保mesh存在
	if len(doc.Meshes) == 0 {
		doc.Meshes = append(doc.Meshes, &gltf.Mesh{})
	}
	doc.Meshes[0].Primitives = append(doc.Meshes[0].Primitives, primitive)

	return gs, nil
}

// 更新createAccessor函数支持量化
func createAccessor(
	doc *gltf.Document,
	data []float32,
	compType gltf.ComponentType,
	dataType gltf.AccessorType,
	normalized bool,
	compress bool,
) (uint32, error) {
	buf := new(bytes.Buffer)

	// 根据组件类型转换数据
	switch compType {
	case gltf.ComponentUbyte:
		quantized := make([]uint8, len(data))
		for i, v := range data {
			quantized[i] = uint8(v * 255) // 归一化到0-255
		}
		binary.Write(buf, binary.LittleEndian, quantized)
	case gltf.ComponentUshort:
		quantized := make([]uint16, len(data))
		for i, v := range data {
			quantized[i] = uint16(v * 65535) // 归一化到0-65535
		}
		binary.Write(buf, binary.LittleEndian, quantized)
	case gltf.ComponentShort:
		quantized := make([]int16, len(data))
		// 假设数据在[-1,1]范围
		for i, v := range data {
			quantized[i] = int16(v * 32767)
		}
		binary.Write(buf, binary.LittleEndian, quantized)
	default:
		if err := binary.Write(buf, binary.LittleEndian, data); err != nil {
			return 0, err
		}
	}

	bufViewIdx, err := addBufferView(doc, buf.Bytes(), compress)
	if err != nil {
		return 0, err
	}

	accessor := &gltf.Accessor{
		BufferView:    gltf.Index(uint32(bufViewIdx)),
		ComponentType: compType,
		Count:         uint32(len(data) / int(dataType.Components())),
		Type:          dataType,
		Normalized:    normalized,
	}

	doc.Accessors = append(doc.Accessors, accessor)
	return uint32(len(doc.Accessors) - 1), nil
}

// 添加压缩函数
func compressWithMeshopt(data []byte) ([]byte, error) {
	// 假设顶点数据格式为每个顶点包含14个float32（位置3 + 颜色4 + 缩放3 + 旋转4）
	const floatSize = 4 // float32占4字节
	const elementsPerVertex = 14
	byteStride := elementsPerVertex * floatSize

	// 计算顶点数量
	if len(data)%byteStride != 0 {
		return nil, fmt.Errorf("invalid data length for vertex compression")
	}
	count := uint32(len(data) / byteStride)

	// 调用meshopt压缩（使用ATTRIBUTES模式）
	compressed, ext, err := meshopt.MeshoptEncode(
		data,
		count,
		uint32(byteStride),
		meshopt.ModeAttributes,
		meshopt.FilterNone,
	)

	if err != nil {
		return nil, fmt.Errorf("meshopt压缩失败: %w", err)
	}

	// 这里可以添加扩展信息处理（如果需要）
	_ = ext // 暂时忽略扩展信息

	return compressed, nil
}

// 修改WireGaussianSplatting函数中的合并逻辑
func mergeVertexData(vd *VertexData) []byte {
	// 创建足够大的缓冲区
	buf := new(bytes.Buffer)

	// 按原始浮点格式写入
	binary.Write(buf, binary.LittleEndian, vd.Positions)
	binary.Write(buf, binary.LittleEndian, vd.Colors)
	binary.Write(buf, binary.LittleEndian, vd.Scales)
	binary.Write(buf, binary.LittleEndian, vd.Rotations)

	return buf.Bytes()
}
