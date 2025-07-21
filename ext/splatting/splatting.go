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
	SH0           = 0.28209479177387814
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

func addExtensionUsed(doc *gltf.Document, ext string) {
	for _, existing := range doc.ExtensionsUsed {
		if existing == ext {
			return
		}
	}
	doc.ExtensionsUsed = append(doc.ExtensionsUsed, ext)
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

// PrepareSplatData 根据规范转换原始数据
func PrepareSplatData(data *VertexData) {
	// 1. 颜色处理 (RGB 分量乘以 SH0 常数)
	for i := 0; i < len(data.Colors); i += 4 {
		// RGB 分量乘以 SH0
		data.Colors[i+0] *= SH0
		data.Colors[i+1] *= SH0
		data.Colors[i+2] *= SH0

		// 2. 不透明度处理 (sigmoid 激活)
		data.Colors[i+3] = 1 / (1 + float32(math.Exp(-float64(data.Colors[i+3]))))
	}

	// 3. 缩放处理 (指数激活)
	for i := 0; i < len(data.Scales); i++ {
		data.Scales[i] = float32(math.Exp(float64(data.Scales[i])))
	}

	// 4. 旋转归一化
	for i := 0; i < len(data.Rotations); i += 4 {
		q := data.Rotations[i : i+4]
		lenSq := q[0]*q[0] + q[1]*q[1] + q[2]*q[2] + q[3]*q[3]

		if lenSq > 1e-12 { // 避免除以零
			lenInv := 1 / float32(math.Sqrt(float64(lenSq)))
			q[0] *= lenInv
			q[1] *= lenInv
			q[2] *= lenInv
			q[3] *= lenInv
		}
	}
}

// ValidateRotation 验证旋转数据是否为单位四元数
func ValidateRotation(rotations []float32) error {
	for i := 0; i < len(rotations); i += 4 {
		q := rotations[i : i+4]
		lenSq := q[0]*q[0] + q[1]*q[1] + q[2]*q[2] + q[3]*q[3]

		// 允许 1e-4 的误差范围
		if math.Abs(float64(lenSq-1.0)) > 1e-4 {
			return fmt.Errorf("非单位四元数在索引 %d: 长度平方 = %f", i, lenSq)
		}
	}
	return nil
}

func WireGaussianSplatting(
	doc *gltf.Document,
	vertexData *VertexData,
	shCoefficients []float32,
	config *QuantizationConfig,
	compress bool,
) (*GaussianSplatting, error) {
	// 验证旋转数据
	if err := ValidateRotation(vertexData.Rotations); err != nil {
		// 尝试修复非单位四元数
		fmt.Printf("警告: %v, 尝试自动修复\n", err)
		for i := 0; i < len(vertexData.Rotations); i += 4 {
			q := vertexData.Rotations[i : i+4]
			lenSq := q[0]*q[0] + q[1]*q[1] + q[2]*q[2] + q[3]*q[3]

			if lenSq > 1e-12 { // 避免除以零
				lenInv := 1 / float32(math.Sqrt(float64(lenSq)))
				q[0] *= lenInv
				q[1] *= lenInv
				q[2] *= lenInv
				q[3] *= lenInv
			}
		}

		// 再次验证
		if err := ValidateRotation(vertexData.Rotations); err != nil {
			return nil, fmt.Errorf("旋转数据验证失败: %w", err)
		}
	}

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

	// 定义属性布局
	attributes := []struct {
		name       string
		data       []float32
		compType   gltf.ComponentType
		dataType   gltf.AccessorType
		normalized bool
	}{
		{"POSITION", vertexData.Positions, config.PositionType, gltf.AccessorVec3, false},
		{"COLOR_0", vertexData.Colors, config.ColorType, gltf.AccessorVec4, config.Normalized},
		{"_SCALE", vertexData.Scales, config.ScaleType, gltf.AccessorVec3, config.Normalized},
		{"_ROTATION", vertexData.Rotations, config.RotationType, gltf.AccessorVec4, config.Normalized},
	}

	// 计算每个属性的大小和步长
	attrSizes := make(map[string]int)
	totalStride := 0
	for _, attr := range attributes {
		compSize := componentSize(attr.compType)
		comps := int(attr.dataType.Components())
		size := compSize * comps
		attrSizes[attr.name] = size
		totalStride += size
	}

	// 创建交错缓冲区
	buf := bytes.NewBuffer(make([]byte, 0, vertexCount*totalStride))

	// 初始化min/max数组
	mins := make(map[string][]float32)
	maxs := make(map[string][]float32)
	for _, attr := range attributes {
		comps := attr.dataType.Components()
		mins[attr.name] = make([]float32, comps)
		maxs[attr.name] = make([]float32, comps)
		for i := range mins[attr.name] {
			mins[attr.name][i] = math.MaxFloat32
			maxs[attr.name][i] = -math.MaxFloat32
			if attr.name == "_ROTATION" {
				mins[attr.name][i] = -1
				maxs[attr.name][i] = 1
			}
		}
	}

	// 第一次遍历：计算实际范围（旋转数据除外）
	for i := 0; i < vertexCount; i++ {
		for _, attr := range attributes {
			if attr.name == "_ROTATION" {
				continue
			}

			comps := int(attr.dataType.Components())
			idx := i * comps

			for j := 0; j < comps; j++ {
				val := attr.data[idx+j]
				if val < mins[attr.name][j] {
					mins[attr.name][j] = val
				}
				if val > maxs[attr.name][j] {
					maxs[attr.name][j] = val
				}
			}
		}
	}

	// 第二次遍历：填充缓冲区
	for i := 0; i < vertexCount; i++ {
		for _, attr := range attributes {
			comps := int(attr.dataType.Components())
			idx := i * comps

			for j := 0; j < comps; j++ {
				val := attr.data[idx+j]

				// 处理归一化
				if attr.normalized && attr.name != "_ROTATION" {
					rng := maxs[attr.name][j] - mins[attr.name][j]
					if rng > 0 {
						val = (val - mins[attr.name][j]) / rng
					} else {
						val = 0
					}
				}

				// 特殊处理旋转数据
				if attr.name == "_ROTATION" {
					val = clamp(val, -1, 1)
				}

				// 写入缓冲区
				switch attr.compType {
				case gltf.ComponentUbyte:
					buf.WriteByte(uint8(val * 255))
				case gltf.ComponentUshort:
					binary.Write(buf, binary.LittleEndian, uint16(val*65535))
				case gltf.ComponentShort:
					binary.Write(buf, binary.LittleEndian, int16(val*32767))
				default:
					binary.Write(buf, binary.LittleEndian, val)
				}
			}
		}
	}

	// 添加缓冲视图
	dataBytes := buf.Bytes()
	bvIndex, err := addBufferView(doc, dataBytes, compress, totalStride, vertexCount)
	if err != nil {
		return nil, err
	}

	// 创建访问器并设置属性映射
	attrs := make(map[string]uint32)
	offset := 0
	for _, attr := range attributes {
		accessor := &gltf.Accessor{
			BufferView:    gltf.Index(uint32(bvIndex)),
			ByteOffset:    uint32(offset),
			ComponentType: attr.compType,
			Count:         uint32(vertexCount),
			Type:          attr.dataType,
			Normalized:    attr.normalized,
			Min:           mins[attr.name],
			Max:           maxs[attr.name],
		}

		doc.Accessors = append(doc.Accessors, accessor)
		attrs[attr.name] = uint32(len(doc.Accessors) - 1)
		offset += attrSizes[attr.name]
	}

	// 处理球谐系数
	var shAccessor *uint32
	if len(shCoefficients) > 0 {
		// 球谐系数不需要交错存储，单独处理
		buf := new(bytes.Buffer)
		for _, v := range shCoefficients {
			binary.Write(buf, binary.LittleEndian, v)
		}

		// 添加缓冲视图
		data := buf.Bytes()
		bvIndex, err := addBufferView(doc, data, false, 0, 0)
		if err != nil {
			return nil, fmt.Errorf("球谐系数缓冲视图创建失败: %w", err)
		}

		// 创建访问器
		count := len(shCoefficients)
		accessor := &gltf.Accessor{
			BufferView:    gltf.Index(uint32(bvIndex)),
			ComponentType: gltf.ComponentFloat,
			Count:         uint32(count),
			Type:          gltf.AccessorScalar,
		}

		doc.Accessors = append(doc.Accessors, accessor)
		idx := uint32(len(doc.Accessors) - 1)
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

func addBufferView(doc *gltf.Document, data []byte, compress bool, stride, vertexCount int) (int, error) {
	originalData := data
	originalSize := len(originalData)

	// 添加步长对齐填充（4字节边界）
	if stride > 0 {
		paddedStride := (stride + 3) &^ 3
		if paddedStride != stride {
			// 创建带填充字节的对齐缓冲区
			alignedData := make([]byte, vertexCount*paddedStride)
			for i := 0; i < vertexCount; i++ {
				src := originalData[i*stride : (i+1)*stride]
				dst := alignedData[i*paddedStride : (i+1)*paddedStride]
				copy(dst, src)
			}
			data = alignedData
			stride = paddedStride
			originalSize = len(alignedData) // 更新原始大小为对齐后的大小
		}
	}

	if compress {
		if stride == 0 {
			return 0, fmt.Errorf("压缩需要非零步长")
		}

		compressed, ext, err := meshopt.MeshoptEncode(
			data,
			uint32(vertexCount),
			uint32(stride),
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
		ext.ByteLength = uint32(originalSize) // 使用更新后的原始大小
		ext.ByteStride = uint32(stride)
		ext.Count = uint32(vertexCount)
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

	if stride > 0 {
		view.ByteStride = uint32(stride)
	}

	// 添加压缩扩展
	if compress {
		view.Extensions = make(gltf.Extensions)
		view.Extensions["EXT_meshopt_compression"] = &meshopt.CompressionExtension{
			Buffer:     0,
			ByteOffset: byteOffset,
			ByteLength: uint32(originalSize), // 使用更新后的原始大小
			ByteStride: uint32(stride),
			Count:      uint32(vertexCount),
			Mode:       meshopt.ModeAttributes,
			Filter:     meshopt.FilterNone,
		}
	}

	doc.BufferViews = append(doc.BufferViews, view)
	return len(doc.BufferViews) - 1, nil
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
