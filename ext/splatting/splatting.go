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
}

func UnmarshalGaussianSplatting(data []byte) (interface{}, error) {
	gs := &GaussianSplatting{}
	if err := json.Unmarshal(data, gs); err != nil {
		return nil, fmt.Errorf("KHR_gaussian_splatting解析失败: %w", err)
	}
	return gs, nil
}

func (g *GaussianSplatting) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
	}{})
}

func CreateGaussianPrimitive(doc *gltf.Document, attributes map[string]uint32) *GaussianSplatting {
	gs := &GaussianSplatting{}

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
		if i+3 >= len(data.Colors) {
			break
		}

		data.Colors[i+0] *= SH0
		data.Colors[i+1] *= SH0
		data.Colors[i+2] *= SH0

		// 应用不透明度 sigmoid
		opacity := data.Colors[i+3]
		data.Colors[i+3] = 1 / (1 + float32(math.Exp(-float64(opacity))))
	}

	// 2. 缩放处理 (带符号的对数变换)
	for i := 0; i < len(data.Scales); i++ {
		// 处理接近零的值
		val := data.Scales[i]
		if math.Abs(float64(val)) < 1e-6 {
			if val < 0 {
				val = -1e-6
			} else {
				val = 1e-6
			}
		}

		// 保留符号的对数变换
		sign := float32(1.0)
		if val < 0 {
			sign = -1.0
		}
		absVal := math.Abs(float64(val))
		data.Scales[i] = sign * float32(math.Log(absVal))
	}

	// 3. 旋转归一化 (处理零旋转)
	for i := 0; i < len(data.Rotations); i += 4 {
		q := data.Rotations[i : i+4]
		lenSq := q[0]*q[0] + q[1]*q[1] + q[2]*q[2] + q[3]*q[3]

		if lenSq > 1e-6 { // 只对非零向量归一化
			lenInv := 1 / float32(math.Sqrt(float64(lenSq)))
			q[0] *= lenInv
			q[1] *= lenInv
			q[2] *= lenInv
			q[3] *= lenInv
		} else {
			// 设置默认单位四元数
			q[0], q[1], q[2], q[3] = 1, 0, 0, 0
		}
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
	compress bool,
) (*GaussianSplatting, error) {
	PrepareSplatData(vertexData)
	vertexCount := len(vertexData.Positions) / 3

	// 添加必要的扩展声明
	if compress {
		addExtensionUsed(doc, "EXT_meshopt_compression")
		addExtensionUsed(doc, "KHR_mesh_quantization")
	}
	var (
		posMin = []float32{math.MaxFloat32, math.MaxFloat32, math.MaxFloat32}
		posMax = []float32{-math.MaxFloat32, -math.MaxFloat32, -math.MaxFloat32}
	)
	for i := 0; i < len(vertexData.Positions); i += 3 {
		for j := 0; j < 3; j++ {
			v := vertexData.Positions[i+j]
			if v < posMin[j] {
				posMin[j] = v
			}
			if v > posMax[j] {
				posMax[j] = v
			}
		}
	}

	// 属性定义
	attributes := []struct {
		name       string
		data       []float32
		compType   gltf.ComponentType
		dataType   gltf.AccessorType
		normalized bool
		min, max   []float32
		filter     meshopt.CompressionFilter
	}{}

	if compress {
		attributes = append(attributes, []struct {
			name       string
			data       []float32
			compType   gltf.ComponentType
			dataType   gltf.AccessorType
			normalized bool
			min, max   []float32
			filter     meshopt.CompressionFilter
		}{
			{
				name:       "POSITION",
				data:       vertexData.Positions,
				compType:   gltf.ComponentUshort,
				dataType:   gltf.AccessorVec3,
				normalized: false,
				min:        posMin,
				max:        posMax,
			},
			{
				name:       "COLOR_0",
				data:       vertexData.Colors,
				compType:   gltf.ComponentUbyte,
				dataType:   gltf.AccessorVec4,
				normalized: true,
			},
			{
				name:       "_ROTATION",
				data:       vertexData.Rotations,
				compType:   gltf.ComponentShort,
				dataType:   gltf.AccessorVec4,
				normalized: false,
				filter:     meshopt.FilterQuaternion,
			},
			{
				name:       "_SCALE",
				data:       vertexData.Scales,
				compType:   gltf.ComponentFloat,
				dataType:   gltf.AccessorVec3,
				normalized: false,
				filter:     meshopt.FilterExponential,
			},
		}...)
	} else {
		attributes = append(attributes, []struct {
			name       string
			data       []float32
			compType   gltf.ComponentType
			dataType   gltf.AccessorType
			normalized bool
			min, max   []float32
			filter     meshopt.CompressionFilter
		}{
			{
				name:       "POSITION",
				data:       vertexData.Positions,
				compType:   gltf.ComponentFloat,
				dataType:   gltf.AccessorVec3,
				normalized: false,
				min:        posMin,
				max:        posMax,
			},
			{
				name:       "COLOR_0",
				data:       vertexData.Colors,
				compType:   gltf.ComponentUbyte,
				dataType:   gltf.AccessorVec4,
				normalized: true,
			},
			{
				name:       "_ROTATION",
				data:       vertexData.Rotations,
				compType:   gltf.ComponentFloat,
				dataType:   gltf.AccessorVec4,
				normalized: false,
				filter:     meshopt.FilterQuaternion,
			},
			{
				name:       "_SCALE",
				data:       vertexData.Scales,
				compType:   gltf.ComponentFloat,
				dataType:   gltf.AccessorVec3,
				normalized: false,
				filter:     meshopt.FilterExponential,
			},
		}...)
	}

	attrs := make(map[string]uint32)

	// 为每个属性单独创建bufferView和accessor
	for _, attr := range attributes {
		// 量化位置到[0,2047]范围
		if attr.name == "POSITION" {
			for i := range attr.data {
				attr.data[i] = clamp(attr.data[i], attr.min[i/3], attr.max[i/3])
			}
		}

		// 计算组件大小和步长
		compSize := componentSize(attr.compType)
		comps := int(attr.dataType.Components())
		stride := compSize * comps

		// 4字节对齐
		if stride%4 != 0 {
			padding := 4 - (stride % 4)
			stride += padding
		}

		// 准备属性数据
		buf := bytes.NewBuffer(nil)
		for i := 0; i < vertexCount; i++ {
			idx := i * comps
			for j := 0; j < comps; j++ {
				val := attr.data[idx+j]

				// 特殊处理
				switch attr.name {
				case "POSITION":
					// 映射到量化范围
					rng := attr.max[j] - attr.min[j]
					if rng == 0 {
						val = 0
					} else {
						val = (val - attr.min[j]) / rng
					}
				}

				// 写入数据
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

			// 添加填充字节
			if padding := stride - (compSize * comps); padding > 0 {
				buf.Write(make([]byte, padding))
			}
		}

		// 创建bufferView
		bvIndex, err := addBufferView(doc, buf.Bytes(), compress, stride, vertexCount, attr.filter)
		if err != nil {
			return nil, err
		}

		// 创建accessor
		accessor := &gltf.Accessor{
			BufferView:    gltf.Index(uint32(bvIndex)),
			ComponentType: attr.compType,
			Count:         uint32(vertexCount),
			Type:          attr.dataType,
			Normalized:    attr.normalized,
		}

		if attr.min != nil {
			accessor.Min = attr.min
			accessor.Max = attr.max
		}

		doc.Accessors = append(doc.Accessors, accessor)
		attrs[attr.name] = uint32(len(doc.Accessors) - 1)
	}

	// 创建高斯泼溅图元
	gs := CreateGaussianPrimitive(doc, attrs)
	primitive := &gltf.Primitive{
		Attributes: attrs,
		Mode:       gltf.PrimitivePoints,
		Extensions: gltf.Extensions{
			ExtensionName: gs,
		},
	}

	// 添加到mesh
	if len(doc.Meshes) == 0 {
		doc.Meshes = append(doc.Meshes, &gltf.Mesh{})
	}
	doc.Meshes[0].Primitives = append(doc.Meshes[0].Primitives, primitive)

	return gs, nil
}

func addBufferView(
	doc *gltf.Document,
	data []byte,
	compress bool,
	stride, vertexCount int,
	filter meshopt.CompressionFilter,
) (int, error) {
	view := &gltf.BufferView{
		ByteStride: uint32(stride),
	}

	// 确保数据长度是4的倍数（GLB要求）
	padding := (4 - (len(data) % 4)) % 4
	if padding > 0 {
		data = append(data, make([]byte, padding)...)
	}

	if compress {
		if filter == "" {
			filter = meshopt.FilterNone
		}

		// 应用meshopt压缩
		compressed, ext, err := meshopt.MeshoptEncode(
			data,
			uint32(vertexCount),
			uint32(stride),
			meshopt.ModeAttributes,
			filter,
		)
		if err != nil {
			return 0, err
		}

		// 创建压缩缓冲区
		compressedBuffer := &gltf.Buffer{
			ByteLength: uint32(len(compressed)),
			Data:       compressed,
		}
		doc.Buffers = append(doc.Buffers, compressedBuffer)
		compressedBufferIndex := len(doc.Buffers) - 1

		doc.Buffers = append(doc.Buffers, &gltf.Buffer{
			ByteLength: uint32(len(data)),
		})

		uncompressedBufferIndex := len(doc.Buffers) - 1

		// 设置扩展信息
		ext.Buffer = uint32(compressedBufferIndex)
		ext.ByteLength = uint32(len(compressed))
		ext.Count = uint32(vertexCount)
		ext.ByteStride = uint32(stride)

		view.Extensions = gltf.Extensions{
			"EXT_meshopt_compression": ext,
		}
		view.Buffer = uint32(uncompressedBufferIndex)
		view.ByteLength = uint32(len(data))
	} else {
		// 非压缩模式
		if len(doc.Buffers) == 0 {
			doc.Buffers = append(doc.Buffers, &gltf.Buffer{})
		}
		buffer := doc.Buffers[0]
		view.Buffer = 0
		view.ByteOffset = buffer.ByteLength
		view.ByteLength = uint32(len(data))
		buffer.ByteLength += view.ByteLength
		buffer.Data = append(buffer.Data, data...)
	}

	doc.BufferViews = append(doc.BufferViews, view)
	return len(doc.BufferViews) - 1, nil
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

// ReadGaussianSplatting 从glTF文档中读取高斯泼溅数据
func ReadGaussianSplatting(doc *gltf.Document, primitive *gltf.Primitive) (*VertexData, error) {
	err := meshopt.DecodeAll(doc)
	if err != nil {
		return nil, fmt.Errorf("meshopt解码失败: %w", err)
	}

	// 检查扩展是否存在（如果存在则解析，但不依赖其Attributes）
	ext, exists := primitive.Extensions[ExtensionName]
	if exists {
		_, ok := ext.(*GaussianSplatting)
		if !ok {
			return nil, fmt.Errorf("invalid KHR_gaussian_splatting extension")
		}
	}

	// 读取各个属性
	vertexData := &VertexData{}

	// 读取位置
	if posAccIdx, ok := primitive.Attributes["POSITION"]; ok {
		posAccessor := doc.Accessors[posAccIdx]
		vertexData.Positions, err = readAccessorAsFloat32(doc, int(posAccIdx))
		if err != nil {
			return nil, fmt.Errorf("读取位置数据失败: %w", err)
		}

		// ✅ 逆量化 POSITION（仅当使用 ComponentUshort 时）
		if posAccessor.ComponentType == gltf.ComponentUshort {
			if posAccessor.Min == nil || posAccessor.Max == nil || len(posAccessor.Min) != 3 || len(posAccessor.Max) != 3 {
				return nil, fmt.Errorf("POSITION accessor 缺少有效的 min/max 用于逆量化")
			}

			min := posAccessor.Min
			max := posAccessor.Max
			for i := 0; i < len(vertexData.Positions); i += 3 {
				for j := 0; j < 3; j++ {
					v := vertexData.Positions[i+j]
					// 反归一化：[0, 65535] → [min, max]
					v = v/65535*(max[j]-min[j]) + min[j]
					vertexData.Positions[i+j] = v / 65535
				}
			}
		}
	} else {
		return nil, fmt.Errorf("missing POSITION attribute")
	}

	// 读取颜色
	if colorAccIdx, ok := primitive.Attributes["COLOR_0"]; ok {
		vertexData.Colors, err = readAccessorAsFloat32(doc, int(colorAccIdx))
		if err != nil {
			return nil, fmt.Errorf("读取颜色数据失败: %w", err)
		}
	} else {
		return nil, fmt.Errorf("missing COLOR_0 attribute")
	}

	var scaleAccIdx uint32
	scaleAttrExists := false

	// 读取缩放
	if accIdx, ok := primitive.Attributes["_SCALE"]; ok {
		scaleAccIdx = accIdx
		scaleAttrExists = true
		vertexData.Scales, err = readAccessorAsFloat32(doc, int(scaleAccIdx))
		if err != nil {
			return nil, fmt.Errorf("读取缩放数据失败: %w", err)
		}
	} else {
		return nil, fmt.Errorf("missing _SCALE attribute")
	}

	// 读取旋转
	if rotAccIdx, ok := primitive.Attributes["_ROTATION"]; ok {
		rotAccessor := doc.Accessors[rotAccIdx]

		vertexData.Rotations, err = readAccessorAsFloat32(doc, int(rotAccIdx))
		if err != nil {
			return nil, fmt.Errorf("读取旋转数据失败: %w", err)
		}
		if rotAccessor.ComponentType == gltf.ComponentShort {
			for i := 0; i < len(vertexData.Rotations); i++ {
				v := vertexData.Rotations[i]
				vertexData.Rotations[i] = v / 32767
			}
		}
	} else {
		return nil, fmt.Errorf("missing _ROTATION attribute")
	}

	// 验证属性长度一致性
	vertexCount := len(vertexData.Positions) / 3
	if len(vertexData.Colors)/4 != vertexCount ||
		len(vertexData.Scales)/3 != vertexCount ||
		len(vertexData.Rotations)/4 != vertexCount {
		return nil, fmt.Errorf("顶点属性长度不一致")
	}

	// 对缩放属性进行反归一化（如果需要）
	if scaleAttrExists {
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
						} else {
							vertexData.Scales[i+j] = min[j]
						}
					}
				}
			}
		}
	}

	// 对数据进行逆处理
	InvertSplatData(vertexData)

	return vertexData, nil
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
func InvertSplatData(data *VertexData) {
	// 1. 颜色逆处理
	for i := 0; i < len(data.Colors); i += 4 {
		data.Colors[i+0] /= SH0
		data.Colors[i+1] /= SH0
		data.Colors[i+2] /= SH0

		opacity := data.Colors[i+3]
		epsilon := float32(1e-6)
		clamped := float64(math.Max(float64(epsilon), math.Min(1.0-float64(epsilon), float64(opacity))))
		data.Colors[i+3] = float32(math.Log(clamped / (1 - clamped)))
	}

	// 2. 缩放逆处理 (指数变换)
	for i := 0; i < len(data.Scales); i++ {
		val := data.Scales[i]
		data.Scales[i] = float32(math.Exp(float64(val)))
	}

	// 3. 旋转数据保持单位四元数，不需要逆处理
}
