package splatting

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math"

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
	Attributes         map[string]uint32 `json:"attributes"`
	SphericalHarmonics *uint32           `json:"sphericalHarmonics,omitempty"`
}

func UnmarshalGaussianSplatting(data []byte) (interface{}, error) {
	gs := &GaussianSplatting{}
	if err := json.Unmarshal(data, gs); err != nil {
		return nil, fmt.Errorf("KHR_gaussian_splatting解析失败: %w", err)
	}

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

func CreateGaussianPrimitive(doc *gltf.Document, attributes map[string]uint32, shAccessor *uint32) *GaussianSplatting {
	gs := &GaussianSplatting{
		Attributes: attributes,
	}

	if shAccessor != nil {
		gs.SphericalHarmonics = shAccessor
	}

	if !hasExtensionUsed(doc, ExtensionName) {
		doc.ExtensionsUsed = append(doc.ExtensionsUsed, ExtensionName)
	}

	return gs
}

func hasExtensionUsed(doc *gltf.Document, ext string) bool {
	for _, e := range doc.ExtensionsUsed {
		if e == ext {
			return true
		}
	}
	return false
}

type QuantizationConfig struct {
	PositionType gltf.ComponentType
	ColorType    gltf.ComponentType
	RotationType gltf.ComponentType
	ScaleType    gltf.ComponentType
	Normalized   bool
}

func DefaultQuantization() *QuantizationConfig {
	return &QuantizationConfig{
		PositionType: gltf.ComponentFloat,
		ColorType:    gltf.ComponentUbyte,
		RotationType: gltf.ComponentShort,
		ScaleType:    gltf.ComponentUshort,
		Normalized:   true,
	}
}

type VertexData struct {
	Positions []float32
	Colors    []float32
	Scales    []float32
	Rotations []float32
}

func WireGaussianSplatting(
	doc *gltf.Document,
	vertexData *VertexData,
	shCoefficients []float32,
	config *QuantizationConfig,
	compress bool,
) (*GaussianSplatting, error) {
	attrs := make(map[string]uint32)

	vertexCount := len(vertexData.Positions) / 3
	if len(vertexData.Colors)/4 != vertexCount ||
		len(vertexData.Scales)/3 != vertexCount ||
		len(vertexData.Rotations)/4 != vertexCount {
		return nil, fmt.Errorf("顶点属性长度不一致")
	}

	// 添加压缩扩展声明
	if compress {
		addExtensionUsed(doc, "EXT_meshopt_compression")
	}

	// 创建访问器
	posIdx, err := createAccessor(doc, vertexData.Positions,
		config.PositionType, gltf.AccessorVec3, false, compress, config, "POSITION")
	if err != nil {
		return nil, fmt.Errorf("位置属性创建失败: %w", err)
	}
	attrs["POSITION"] = posIdx

	colorIdx, err := createAccessor(doc, vertexData.Colors,
		config.ColorType, gltf.AccessorVec4, config.Normalized, compress, config, "COLOR_0")
	if err != nil {
		return nil, fmt.Errorf("颜色属性创建失败: %w", err)
	}
	attrs["COLOR_0"] = colorIdx

	scaleIdx, err := createAccessor(doc, vertexData.Scales,
		config.ScaleType, gltf.AccessorVec3, config.Normalized, compress, config, "_SCALE")
	if err != nil {
		return nil, fmt.Errorf("缩放属性创建失败: %w", err)
	}
	attrs["_SCALE"] = scaleIdx

	rotIdx, err := createAccessor(doc, vertexData.Rotations,
		config.RotationType, gltf.AccessorVec4, config.Normalized, compress, config, "_ROTATION")
	if err != nil {
		return nil, fmt.Errorf("旋转属性创建失败: %w", err)
	}
	attrs["_ROTATION"] = rotIdx

	var shAccessor *uint32
	if len(shCoefficients) > 0 {
		idx, err := createAccessor(doc, shCoefficients,
			gltf.ComponentFloat, gltf.AccessorScalar, false, compress, config, "SH")
		if err != nil {
			return nil, fmt.Errorf("球谐系数创建失败: %w", err)
		}
		shAccessor = &idx
	}

	gs := CreateGaussianPrimitive(doc, attrs, shAccessor)

	primitive := &gltf.Primitive{
		Attributes: attrs,
		Extensions: make(gltf.Extensions),
	}
	primitive.Extensions[ExtensionName] = gs

	if len(doc.Meshes) == 0 {
		doc.Meshes = append(doc.Meshes, &gltf.Mesh{})
	}
	doc.Meshes[0].Primitives = append(doc.Meshes[0].Primitives, primitive)

	// 添加量化扩展声明
	if config.PositionType != gltf.ComponentFloat ||
		config.ColorType != gltf.ComponentFloat ||
		config.ScaleType != gltf.ComponentFloat ||
		config.RotationType != gltf.ComponentFloat {
		addExtensionUsed(doc, "KHR_mesh_quantization")
	}

	return gs, nil
}

func createAccessor(
	doc *gltf.Document,
	data []float32,
	compType gltf.ComponentType,
	dataType gltf.AccessorType,
	normalized bool,
	compress bool,
	config *QuantizationConfig,
	attrName string,
) (uint32, error) {
	components := int(dataType.Components())
	count := len(data) / components
	if len(data)%components != 0 {
		return 0, fmt.Errorf("数据长度不匹配访问器类型")
	}

	buf := new(bytes.Buffer)
	min := make([]float32, components)
	max := make([]float32, components)

	// 初始化min/max
	for i := range min {
		min[i] = math.MaxFloat32
		max[i] = -math.MaxFloat32
	}

	// 特殊处理旋转数据
	if attrName == "_ROTATION" {
		for i := range min {
			min[i] = -1
			max[i] = 1
		}
	}

	// 第一次遍历计算实际范围（旋转数据除外）
	if attrName != "_ROTATION" {
		for i, v := range data {
			idx := i % components
			if v < min[idx] {
				min[idx] = v
			}
			if v > max[idx] {
				max[idx] = v
			}
		}
	}

	// 第二次遍历处理量化
	for i, v := range data {
		idx := i % components
		var quantizedValue float32

		switch attrName {
		case "_ROTATION":
			// 旋转数据特殊处理，保持单位四元数
			quantizedValue = clamp(v, -1, 1)
		default:
			if normalized {
				// 正确归一化公式
				rangeVal := max[idx] - min[idx]
				if rangeVal > 0 {
					quantizedValue = (v - min[idx]) / rangeVal
				} else {
					quantizedValue = 0
				}
			} else {
				quantizedValue = v
			}
		}

		// 根据类型处理量化
		switch compType {
		case gltf.ComponentUbyte:
			val := uint8(quantizedValue * 255)
			buf.WriteByte(val)
		case gltf.ComponentUshort:
			val := uint16(quantizedValue * 65535)
			binary.Write(buf, binary.LittleEndian, val)
		case gltf.ComponentShort:
			val := int16(quantizedValue * 32767)
			binary.Write(buf, binary.LittleEndian, val)
		default:
			binary.Write(buf, binary.LittleEndian, v)
		}
	}

	// 添加缓冲区视图
	bufViewIdx, err := addBufferView(doc, buf.Bytes(), compress, config)
	if err != nil {
		return 0, err
	}

	// 创建访问器
	accessor := &gltf.Accessor{
		BufferView:    gltf.Index(uint32(bufViewIdx)),
		ComponentType: compType,
		Count:         uint32(count),
		Type:          dataType,
		Normalized:    normalized,
		Min:           min,
		Max:           max,
	}

	doc.Accessors = append(doc.Accessors, accessor)
	return uint32(len(doc.Accessors) - 1), nil
}

func calculateVertexLayout(config *QuantizationConfig) int {
	size := 0
	size += 3 * componentSize(config.PositionType)
	size += 4 * componentSize(config.ColorType)
	size += 3 * componentSize(config.ScaleType)
	size += 4 * componentSize(config.RotationType)
	return size
}

func componentSize(ct gltf.ComponentType) int {
	switch ct {
	case gltf.ComponentUbyte, gltf.ComponentByte:
		return 1
	case gltf.ComponentUshort, gltf.ComponentShort:
		return 2
	default: // float
		return 4
	}
}

func addBufferView(doc *gltf.Document, data []byte, compress bool, config *QuantizationConfig) (int, error) {
	originalData := data
	if compress {
		vertexSize := calculateVertexLayout(config)
		if len(data)%vertexSize != 0 {
			return 0, fmt.Errorf("数据大小不符合顶点布局")
		}
		count := uint32(len(data) / vertexSize)

		compressed, ext, err := meshopt.MeshoptEncode(
			data,
			count,
			uint32(vertexSize),
			meshopt.ModeAttributes,
			meshopt.FilterNone,
		)

		if err != nil {
			return 0, fmt.Errorf("meshopt压缩失败: %w", err)
		}

		// 使用压缩后的数据
		data = compressed

		// 添加压缩扩展
		ext.Buffer = 0
		ext.ByteOffset = 0
		ext.ByteLength = uint32(len(data))
		ext.ByteStride = uint32(vertexSize)
		ext.Count = count
	}

	if len(doc.Buffers) == 0 {
		doc.Buffers = append(doc.Buffers, &gltf.Buffer{})
	}

	buffer := doc.Buffers[0]
	byteOffset := buffer.ByteLength
	buffer.ByteLength += uint32(len(data))

	view := &gltf.BufferView{
		Buffer:     0,
		ByteOffset: byteOffset,
		ByteLength: uint32(len(data)),
	}

	// 添加压缩扩展
	if compress {
		view.Extensions = make(gltf.Extensions)
		view.Extensions["EXT_meshopt_compression"] = &meshopt.CompressionExtension{
			Buffer:     0,
			ByteOffset: byteOffset,
			ByteLength: uint32(len(originalData)),
			ByteStride: uint32(calculateVertexLayout(config)),
			Count:      uint32(len(originalData)) / uint32(calculateVertexLayout(config)),
			Mode:       meshopt.ModeAttributes,
			Filter:     meshopt.FilterNone,
		}
	}

	doc.BufferViews = append(doc.BufferViews, view)
	return len(doc.BufferViews) - 1, nil
}

func addExtensionUsed(doc *gltf.Document, ext string) {
	for _, existing := range doc.ExtensionsUsed {
		if existing == ext {
			return
		}
	}
	doc.ExtensionsUsed = append(doc.ExtensionsUsed, ext)
}

func clamp(value, min, max float32) float32 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}
