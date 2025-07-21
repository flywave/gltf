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
		ScaleType:    gltf.ComponentFloat, // 改为浮点型以保持精度
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
		// 添加安全保护，防止无效数学运算
		if !math.IsNaN(float64(data.Colors[i+3])) && !math.IsInf(float64(data.Colors[i+3]), 0) {
			data.Colors[i+3] = 1 / (1 + float32(math.Exp(-float64(data.Colors[i+3]))))
		} else {
			data.Colors[i+3] = 0.5 // 默认值
		}
	}

	// 3. 缩放处理 (指数激活)
	for i := 0; i < len(data.Scales); i++ {
		// 添加安全保护
		if !math.IsNaN(float64(data.Scales[i])) && !math.IsInf(float64(data.Scales[i]), 0) {
			data.Scales[i] = float32(math.Exp(float64(data.Scales[i])))
		} else {
			data.Scales[i] = 1.0 // 默认值
		}
	}

	// 4. 增强旋转归一化
	for i := 0; i < len(data.Rotations); i += 4 {
		q := data.Rotations[i : i+4]
		lenSq := q[0]*q[0] + q[1]*q[1] + q[2]*q[2] + q[3]*q[3]

		// 添加安全保护
		if lenSq < 1e-12 {
			q[0] = 1.0
			q[1] = 0
			q[2] = 0
			q[3] = 0
			continue
		}

		lenInv := 1 / float32(math.Sqrt(float64(lenSq)))
		// 使用更高精度的归一化计算
		q[0] = float32(math.Round(float64(q[0]*lenInv)*1e6) / 1e6)
		q[1] = float32(math.Round(float64(q[1]*lenInv)*1e6) / 1e6)
		q[2] = float32(math.Round(float64(q[2]*lenInv)*1e6) / 1e6)
		q[3] = float32(math.Round(float64(q[3]*lenInv)*1e6) / 1e6)
	}
}

// clamp 确保值在指定范围内
func clamp(v, min, max float32) float32 {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

// ValidateRotation 验证旋转数据是否为单位四元数
func ValidateRotation(rotations []float32) error {
	for i := 0; i < len(rotations); i += 4 {
		q := rotations[i : i+4]
		lenSq := q[0]*q[0] + q[1]*q[1] + q[2]*q[2] + q[3]*q[3]

		// 扩大容差范围到1e-3
		if math.Abs(float64(lenSq-1.0)) > 1e-3 {
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
	PrepareSplatData(vertexData)

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

	// 计算缩放范围（每个分量）
	scaleMins := [3]float32{math.MaxFloat32, math.MaxFloat32, math.MaxFloat32}
	scaleMaxs := [3]float32{-math.MaxFloat32, -math.MaxFloat32, -math.MaxFloat32}
	for i := 0; i < len(vertexData.Scales); i += 3 {
		for j := 0; j < 3; j++ {
			val := vertexData.Scales[i+j]
			if val < scaleMins[j] {
				scaleMins[j] = val
			}
			if val > scaleMaxs[j] {
				scaleMaxs[j] = val
			}
		}
	}

	// 定义属性布局
	attributes := []struct {
		name       string
		data       []float32
		compType   gltf.ComponentType
		dataType   gltf.AccessorType
		normalized bool
		min, max   []float32 // 缩放范围
	}{
		{"POSITION", vertexData.Positions, config.PositionType, gltf.AccessorVec3, config.Normalized, nil, nil},
		{"COLOR_0", vertexData.Colors, config.ColorType, gltf.AccessorVec4, config.Normalized, nil, nil},
		{"_SCALE", vertexData.Scales, config.ScaleType, gltf.AccessorVec3, config.Normalized, scaleMins[:], scaleMaxs[:]},
		{"_ROTATION", vertexData.Rotations, config.RotationType, gltf.AccessorVec4, config.Normalized, nil, nil},
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
		if attr.min != nil {
			// 使用预定义的范围
			mins[attr.name] = attr.min
			maxs[attr.name] = attr.max
		} else {
			// 自动计算范围
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
	}

	// 第二次遍历：填充缓冲区
	for i := 0; i < vertexCount; i++ {
		for _, attr := range attributes {
			comps := int(attr.dataType.Components())
			idx := i * comps

			for j := 0; j < comps; j++ {
				val := attr.data[idx+j]

				// 处理旋转数据（先执行）
				if attr.name == "_ROTATION" {
					// 确保值在[-1, 1]范围内
					val = clamp(val, -1, 1)
				}

				// 处理缩放数据（使用动态范围）
				if attr.name == "_SCALE" && attr.normalized {
					rng := maxs[attr.name][j] - mins[attr.name][j]
					if rng <= 0 {
						val = 0
					} else {
						val = (val - mins[attr.name][j]) / rng
					}
					val = clamp(val, 0, 1)
				}

				// 处理其他属性的归一化
				if attr.normalized && attr.name != "_SCALE" {
					switch attr.compType {
					case gltf.ComponentUbyte, gltf.ComponentUshort:
						// 无符号归一化：clamp到[0, 1]
						val = clamp(val, 0, 1)
					case gltf.ComponentShort:
						// 有符号归一化：clamp到[-1, 1]
						val = clamp(val, -1, 1)
					}
				}

				// 写入缓冲区
				switch attr.compType {
				case gltf.ComponentUbyte:
					if attr.normalized {
						buf.WriteByte(uint8(val * 255)) // [0,1] -> [0,255]
					} else {
						buf.WriteByte(uint8(val)) // 直接截断
					}
				case gltf.ComponentUshort:
					if attr.normalized {
						binary.Write(buf, binary.LittleEndian, uint16(val*65535))
					} else {
						binary.Write(buf, binary.LittleEndian, uint16(val))
					}
				case gltf.ComponentShort:
					if attr.normalized {
						binary.Write(buf, binary.LittleEndian, int16(val*32767))
					} else {
						binary.Write(buf, binary.LittleEndian, int16(val))
					}
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
		}
	}
	padding := (4 - (len(data) % 4)) % 4
	if padding > 0 {
		data = append(data, make([]byte, padding)...)
	}

	view := &gltf.BufferView{
		Buffer: 0,
	}

	var byteOffset uint32
	if !compress {
		// 非压缩模式：将数据追加到主缓冲区
		if len(doc.Buffers) == 0 {
			doc.Buffers = append(doc.Buffers, &gltf.Buffer{})
		}
		buffer := doc.Buffers[0]
		byteOffset = buffer.ByteLength
		buffer.ByteLength += uint32(len(data))
		buffer.Data = append(buffer.Data, data...)
	}

	if compress {
		// 压缩模式：创建独立缓冲区存储压缩数据
		buffer := &gltf.Buffer{ByteLength: uint32(len(data))}
		doc.Buffers = append(doc.Buffers, buffer)
		bufferIndex := len(doc.Buffers) - 1

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
		buffer.Data = append(buffer.Data, compressed...)

		ext.ByteLength = uint32(len(compressed))
		ext.Count = uint32(vertexCount)
		ext.ByteStride = uint32(stride)

		view.Extensions = make(gltf.Extensions)
		view.Extensions["EXT_meshopt_compression"] = &meshopt.CompressionExtension{
			Buffer:     uint32(bufferIndex),
			ByteLength: uint32(len(compressed)),
			// 记录原始数据参数
			Count:      uint32(vertexCount),
			ByteStride: uint32(stride),
		}
	} else {
		// 非压缩模式使用主缓冲区
		view.Buffer = 0
		view.ByteOffset = byteOffset
	}

	if stride > 0 {
		view.ByteStride = uint32(stride)
	}

	doc.BufferViews = append(doc.BufferViews, view)
	return len(doc.BufferViews) - 1, nil
}

// ReadGaussianSplatting 从glTF文档中读取高斯泼溅数据
func ReadGaussianSplatting(doc *gltf.Document, primitive *gltf.Primitive) (*VertexData, []float32, error) {
	// 检查扩展是否存在
	ext, exists := primitive.Extensions[ExtensionName]
	if !exists {
		return nil, nil, fmt.Errorf("primitive does not have KHR_gaussian_splatting extension")
	}

	// 将扩展解析为GaussianSplatting结构
	gs, ok := ext.(*GaussianSplatting)
	if !ok {
		return nil, nil, fmt.Errorf("invalid KHR_gaussian_splatting extension")
	}

	// 读取各个属性
	vertexData := &VertexData{}
	var err error

	// 读取位置
	if posAccIdx, ok := gs.Attributes["POSITION"]; ok {
		vertexData.Positions, err = readAccessorAsFloat32(doc, int(posAccIdx))
		if err != nil {
			return nil, nil, fmt.Errorf("读取位置数据失败: %w", err)
		}
	} else {
		return nil, nil, fmt.Errorf("missing POSITION attribute")
	}

	// 读取颜色
	if colorAccIdx, ok := gs.Attributes["COLOR_0"]; ok {
		vertexData.Colors, err = readAccessorAsFloat32(doc, int(colorAccIdx))
		if err != nil {
			return nil, nil, fmt.Errorf("读取颜色数据失败: %w", err)
		}
	} else {
		return nil, nil, fmt.Errorf("missing COLOR_0 attribute")
	}

	// 读取缩放
	if scaleAccIdx, ok := gs.Attributes["_SCALE"]; ok {
		vertexData.Scales, err = readAccessorAsFloat32(doc, int(scaleAccIdx))
		if err != nil {
			return nil, nil, fmt.Errorf("读取缩放数据失败: %w", err)
		}
	} else {
		return nil, nil, fmt.Errorf("missing _SCALE attribute")
	}

	// 读取旋转
	if rotAccIdx, ok := gs.Attributes["_ROTATION"]; ok {
		vertexData.Rotations, err = readAccessorAsFloat32(doc, int(rotAccIdx))
		if err != nil {
			return nil, nil, fmt.Errorf("读取旋转数据失败: %w", err)
		}
	} else {
		return nil, nil, fmt.Errorf("missing _ROTATION attribute")
	}

	// 验证属性长度一致性
	vertexCount := len(vertexData.Positions) / 3
	if len(vertexData.Colors)/4 != vertexCount ||
		len(vertexData.Scales)/3 != vertexCount ||
		len(vertexData.Rotations)/4 != vertexCount {
		return nil, nil, fmt.Errorf("顶点属性长度不一致")
	}

	// 读取球谐系数（如果有）
	var shCoeffs []float32
	if gs.SphericalHarmonics != nil {
		shCoeffs, err = readAccessorAsFloat32(doc, int(*gs.SphericalHarmonics))
		if err != nil {
			return nil, nil, fmt.Errorf("读取球谐系数失败: %w", err)
		}
	}

	// 对缩放属性进行反归一化（如果需要）
	if scaleAccIdx, ok := gs.Attributes["_SCALE"]; ok {
		scaleAccessor := doc.Accessors[scaleAccIdx]
		if scaleAccessor.Normalized {
			min := scaleAccessor.Min
			max := scaleAccessor.Max
			if min != nil && max != nil && len(min) == 3 && len(max) == 3 {
				for i := 0; i < len(vertexData.Scales); i += 3 {
					for j := 0; j < 3; j++ {
						rng := max[j] - min[j]
						if rng > 0 {
							vertexData.Scales[i+j] = vertexData.Scales[i+j]*rng + min[j]
						}
					}
				}
			}
		}
	}

	// 对数据进行逆处理
	invertSplatData(vertexData)

	return vertexData, shCoeffs, nil
}

// readAccessorAsFloat32 从访问器中读取数据并转换为float32切片
func readAccessorAsFloat32(doc *gltf.Document, accessorIndex int) ([]float32, error) {
	if accessorIndex < 0 || accessorIndex >= len(doc.Accessors) {
		return nil, fmt.Errorf("无效的访问器索引: %d", accessorIndex)
	}
	accessor := doc.Accessors[accessorIndex]

	// 获取缓冲视图和缓冲区数据
	bv, buffer, err := getBufferData(doc, accessor)
	if err != nil {
		return nil, err
	}

	// 计算访问器参数
	componentCount := accessor.Type.Components()
	componentSize := gltf.SizeOfComponent(accessor.ComponentType)
	stride := calculateStride(bv, componentSize, int(componentCount))

	// 预分配结果数组
	data := make([]float32, accessor.Count*uint32(componentCount))

	// 批量读取数据
	if err := batchReadComponents(
		buffer.Data,
		bv.ByteOffset+accessor.ByteOffset,
		stride,
		accessor.ComponentType,
		accessor.Normalized,
		int(componentCount),
		accessor.Count,
		data,
	); err != nil {
		return nil, err
	}

	return data, nil
}

// 辅助函数：获取缓冲视图和缓冲区数据（添加解压支持）
func getBufferData(doc *gltf.Document, accessor *gltf.Accessor) (*gltf.BufferView, *gltf.Buffer, error) {
	if accessor.BufferView == nil {
		return nil, nil, fmt.Errorf("访问器缺少缓冲视图")
	}
	bvIndex := *accessor.BufferView
	if int(bvIndex) >= len(doc.BufferViews) {
		return nil, nil, fmt.Errorf("缓冲视图索引越界")
	}
	bv := doc.BufferViews[bvIndex]

	if int(bv.Buffer) >= len(doc.Buffers) {
		return nil, nil, fmt.Errorf("缓冲区索引越界")
	}
	buffer := doc.Buffers[bv.Buffer]

	// 处理meshopt压缩扩展
	if ext, ok := bv.Extensions["EXT_meshopt_compression"].(*meshopt.CompressionExtension); ok {
		if ext == nil {
			return nil, nil, fmt.Errorf("无效的EXT_meshopt_compression扩展")
		}

		// 获取原始压缩数据
		compressedData := buffer.Data[ext.ByteOffset : ext.ByteOffset+ext.ByteLength]

		// 解码数据
		decoded, err := meshopt.MeshoptDecode(
			uint32(ext.Count),
			uint32(ext.ByteStride),
			compressedData,
			meshopt.ModeAttributes,
			meshopt.FilterNone,
		)
		if err != nil {
			return nil, nil, fmt.Errorf("meshopt解压失败: %w", err)
		}

		// 创建临时缓冲区存放解压后的数据
		bvCopy := *bv
		bufferCopy := *buffer
		bufferCopy.Data = decoded
		bvCopy.Buffer = uint32(len(doc.Buffers))
		doc.Buffers = append(doc.Buffers, &bufferCopy)

		return &bvCopy, &bufferCopy, nil
	}

	return bv, buffer, nil
}

// 辅助函数：计算有效步长
func calculateStride(bv *gltf.BufferView, componentSize, componentCount int) uint32 {
	if bv.ByteStride > 0 {
		return bv.ByteStride
	}
	return uint32(componentSize * componentCount)
}

// 辅助函数：批量读取组件数据
func batchReadComponents(
	buffer []byte,
	startOffset uint32,
	stride uint32,
	compType gltf.ComponentType,
	normalized bool,
	componentCount int,
	count uint32,
	out []float32,
) error {
	totalComponents := int(count) * componentCount

	// 根据组件类型选择处理方式
	switch compType {
	case gltf.ComponentFloat:
		if err := binary.Read(bytes.NewReader(buffer[startOffset:]), binary.LittleEndian, out); err != nil {
			return fmt.Errorf("浮点数据读取失败: %w", err)
		}

	case gltf.ComponentUbyte, gltf.ComponentByte:
		processByteComponents(buffer, startOffset, stride, compType, normalized, componentCount, count, out)

	case gltf.ComponentUshort, gltf.ComponentShort:
		processShortComponents(buffer, startOffset, stride, compType, normalized, componentCount, count, out)

	default:
		return fmt.Errorf("不支持的组件类型: %d", compType)
	}

	// 验证输出长度
	if len(out) != totalComponents {
		return fmt.Errorf("数据长度不匹配，预期 %d 实际 %d", totalComponents, len(out))
	}

	return nil
}

// 处理字节类型组件 (优化内存访问模式)
func processByteComponents(buffer []byte, start uint32, stride uint32, compType gltf.ComponentType, normalized bool, compCount int, count uint32, out []float32) {
	divisor := float32(1.0)
	if normalized {
		divisor = 255.0
		if compType == gltf.ComponentByte {
			divisor = 127.0
		}
	}

	// 修复：正确计算偏移量
	for i := uint32(0); i < count; i++ {
		offset := start + i*stride
		for c := 0; c < compCount; c++ {
			idx := i*uint32(compCount) + uint32(c)
			byteOffset := offset + uint32(c)
			val := buffer[byteOffset]

			if compType == gltf.ComponentByte {
				out[idx] = float32(int8(val)) / divisor
			} else {
				out[idx] = float32(val) / divisor
			}
		}
	}
}

// 处理短整型组件 (使用批量转换)
func processShortComponents(buffer []byte, start uint32, stride uint32, compType gltf.ComponentType, normalized bool, compCount int, count uint32, out []float32) {
	divisor := float32(1.0)
	if normalized {
		divisor = 65535.0
		if compType == gltf.ComponentShort {
			divisor = 32767.0
		}
	}

	// 预计算所有short值
	shorts := make([]int16, count*uint32(compCount))
	// 修复：将 startOffset 改为函数参数中的 start
	if err := binary.Read(bytes.NewReader(buffer[start:]), binary.LittleEndian, shorts); err == nil {
		for i, v := range shorts {
			out[i] = float32(v) / divisor
		}
		return
	}

	// 回退逐元素处理
	for i := uint32(0); i < count; i++ {
		offset := start + i*stride
		for c := 0; c < compCount; c++ {
			idx := i*uint32(compCount) + uint32(c)
			byteOffset := offset + uint32(c*2)
			val := int16(binary.LittleEndian.Uint16(buffer[byteOffset:]))
			out[idx] = float32(val) / divisor
		}
	}
}

// invertSplatData 对读取的高斯泼溅数据进行逆处理
func invertSplatData(data *VertexData) {
	// 1. 颜色逆处理
	for i := 0; i < len(data.Colors); i += 4 {
		// RGB分量除以SH0
		data.Colors[i+0] /= SH0
		data.Colors[i+1] /= SH0
		data.Colors[i+2] /= SH0

		// 不透明度逆处理 (logit函数)
		opacity := data.Colors[i+3]
		// 添加安全保护，防止无效数学运算
		if opacity <= 0 {
			opacity = 1e-6
		} else if opacity >= 1 {
			opacity = 1 - 1e-6
		}
		data.Colors[i+3] = float32(math.Log(float64(opacity / (1 - opacity))))
	}

	// 2. 缩放逆处理 (对数) - 添加安全保护
	for i := 0; i < len(data.Scales); i++ {
		if data.Scales[i] <= 0 {
			data.Scales[i] = 1e-6
		}
		data.Scales[i] = float32(math.Log(float64(data.Scales[i])))
	}

	// 3. 旋转数据已经是单位四元数，不需要逆处理
}
