package quantization

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math"
	"strings"

	"github.com/flywave/gltf"
)

const ExtensionName = "KHR_mesh_quantization"

type QuantizationExtension struct {
	PositionBits uint8 `json:"POSITION,omitempty"`
	NormalBits   uint8 `json:"NORMAL,omitempty"`
	TangentBits  uint8 `json:"TANGENT,omitempty"`
	TexCoordBits uint8 `json:"TEXCOORD,omitempty"`
	ColorBits    uint8 `json:"COLOR,omitempty"`
	GenericBits  uint8 `json:"GENERIC,omitempty"`
	JointBits    uint8 `json:"JOINTS,omitempty"`
	WeightBits   uint8 `json:"WEIGHTS,omitempty"`
}

func init() {
	gltf.RegisterExtension(ExtensionName, Unmarshal)
}

func Unmarshal(data []byte) (interface{}, error) {
	ext := &QuantizationExtension{}
	if err := json.Unmarshal(data, ext); err != nil {
		return nil, fmt.Errorf("KHR_mesh_quantization unmarshal failed: %w", err)
	}
	return ext, nil
}

type Dequantizer struct {
	doc *gltf.Document
}

func NewDequantizer(doc *gltf.Document) *Dequantizer {
	return &Dequantizer{doc: doc}
}

func (d *Dequantizer) Process() error {
	processed := false

	for _, mesh := range d.doc.Meshes {
		for _, primitive := range mesh.Primitives {
			extValue, exists := primitive.Extensions[ExtensionName]
			if !exists {
				continue
			}

			ext, ok := extValue.(*QuantizationExtension)
			if !ok {
				return fmt.Errorf("invalid quantization extension format")
			}

			processed = true

			for attr, accessorIdx := range primitive.Attributes {
				if accessorIdx >= uint32(len(d.doc.Accessors)) {
					continue
				}
				accessor := d.doc.Accessors[accessorIdx]

				// 跳过浮点类型和JOINTS属性
				if accessor.ComponentType == gltf.ComponentFloat ||
					strings.HasPrefix(attr, "JOINTS_") {
					continue
				}

				bits := d.getQuantizationBits(attr, ext)
				if bits == 0 {
					continue
				}

				newIdx, err := d.dequantizeAccessor(accessor, bits)
				if err != nil {
					return fmt.Errorf("dequantize attribute %s failed: %w", attr, err)
				}

				primitive.Attributes[attr] = newIdx
			}

			delete(primitive.Extensions, ExtensionName)
		}
	}

	if processed {
		d.removeTopLevelExtension()
	}

	return nil
}

func (d *Dequantizer) getQuantizationBits(attributeName string, ext *QuantizationExtension) uint8 {
	switch {
	case strings.HasPrefix(attributeName, "POSITION"):
		return firstNonZero(ext.PositionBits, 12)
	case strings.HasPrefix(attributeName, "NORMAL"):
		return firstNonZero(ext.NormalBits, 10)
	case strings.HasPrefix(attributeName, "TANGENT"):
		return firstNonZero(ext.TangentBits, 10)
	case strings.HasPrefix(attributeName, "TEXCOORD"):
		return firstNonZero(ext.TexCoordBits, 12)
	case strings.HasPrefix(attributeName, "COLOR"):
		return firstNonZero(ext.ColorBits, 8)
	case strings.HasPrefix(attributeName, "WEIGHTS"):
		return firstNonZero(ext.WeightBits, 8)
	default:
		return firstNonZero(ext.GenericBits, 8)
	}
}

func (d *Dequantizer) dequantizeAccessor(accessor *gltf.Accessor, bits uint8) (uint32, error) {
	// 验证访问器数据
	if accessor.Min == nil || accessor.Max == nil {
		return 0, fmt.Errorf("accessor missing min/max values")
	}

	minValues := accessor.Min
	maxValues := accessor.Max

	// 获取分量数量
	componentCount, ok := map[gltf.AccessorType]int{
		gltf.AccessorScalar: 1,
		gltf.AccessorVec2:   2,
		gltf.AccessorVec3:   3,
		gltf.AccessorVec4:   4,
	}[accessor.Type]

	if !ok || componentCount < 1 {
		return 0, fmt.Errorf("unsupported accessor type: %s", accessor.Type)
	}

	// 验证min/max长度
	if len(minValues) < componentCount || len(maxValues) < componentCount {
		return 0, fmt.Errorf("min/max length mismatch")
	}

	// 获取缓冲视图和缓冲区
	if accessor.BufferView == nil {
		return 0, fmt.Errorf("accessor missing buffer view")
	}
	bvIndex := *accessor.BufferView
	if bvIndex >= uint32(len(d.doc.BufferViews)) {
		return 0, fmt.Errorf("buffer view index out of range")
	}
	bv := d.doc.BufferViews[bvIndex]

	if bv.Buffer >= uint32(len(d.doc.Buffers)) {
		return 0, fmt.Errorf("buffer index out of range")
	}
	buffer := d.doc.Buffers[bv.Buffer]

	// 计算数据范围和偏移
	start := bv.ByteOffset + accessor.ByteOffset
	stride := bv.ByteStride
	if stride == 0 {
		stride = uint32(componentCount * gltf.SizeOfComponent(accessor.ComponentType))
	}

	// 验证数据范围
	count := accessor.Count
	elementSize := uint32(gltf.SizeOfComponent(accessor.ComponentType)) * uint32(componentCount)
	end := start + (count-1)*stride + elementSize
	if end > uint32(len(buffer.Data)) {
		return 0, fmt.Errorf("accessor data exceeds buffer range")
	}

	// 准备浮点数据存储
	floatData := make([]float32, count*uint32(componentCount))

	// 解量化参数
	maxInteger := float32(math.Pow(2, float64(bits)) - 1)
	ranges := make([]float32, componentCount)
	for i := 0; i < componentCount; i++ {
		ranges[i] = maxValues[i] - minValues[i]
		if ranges[i] == 0 {
			ranges[i] = 1e-6 // 避免除以零
		}
	}

	// 解量化数据
	componentSize := gltf.SizeOfComponent(accessor.ComponentType)
	for i := uint32(0); i < count; i++ {
		offset := start + i*stride

		for c := 0; c < componentCount; c++ {
			// 读取量化值
			rawValue, err := readComponent(
				buffer.Data[offset:],
				accessor.ComponentType,
				accessor.Normalized,
			)
			if err != nil {
				return 0, err
			}

			// 应用解量化公式: value = min + (raw / maxInteger) * range
			normalized := rawValue / maxInteger
			floatValue := minValues[c] + normalized*ranges[c]

			// 存储解量化后的值
			idx := i*uint32(componentCount) + uint32(c)
			floatData[idx] = floatValue

			offset += uint32(componentSize)
		}
	}

	// 创建新缓冲区
	byteData := float32ToBytes(floatData)
	newBuffer := gltf.Buffer{
		ByteLength: uint32(len(byteData)),
		Data:       byteData,
	}
	newBufferIndex := uint32(len(d.doc.Buffers))
	d.doc.Buffers = append(d.doc.Buffers, &newBuffer)

	// 创建新缓冲视图
	newBufferView := gltf.BufferView{
		Buffer:     newBufferIndex,
		ByteOffset: 0,
		ByteLength: newBuffer.ByteLength,
		ByteStride: uint32(componentCount * 4), // 每个浮点数4字节
		Target:     bv.Target,
	}
	newBufferViewIndex := uint32(len(d.doc.BufferViews))
	d.doc.BufferViews = append(d.doc.BufferViews, &newBufferView)

	// 创建新访问器
	newAccessor := &gltf.Accessor{
		BufferView:    &newBufferViewIndex,
		ByteOffset:    0,
		ComponentType: gltf.ComponentFloat,
		Count:         accessor.Count,
		Type:          accessor.Type,
		Min:           minValues,
		Max:           maxValues,
		Normalized:    false,
	}

	// 添加新访问器并返回索引
	d.doc.Accessors = append(d.doc.Accessors, newAccessor)
	return uint32(len(d.doc.Accessors) - 1), nil
}

func (d *Dequantizer) removeTopLevelExtension() {
	// 从扩展中移除
	delete(d.doc.Extensions, ExtensionName)

	// 从已使用扩展中移除
	for i, ext := range d.doc.ExtensionsUsed {
		if ext == ExtensionName {
			d.doc.ExtensionsUsed = append(d.doc.ExtensionsUsed[:i], d.doc.ExtensionsUsed[i+1:]...)
			break
		}
	}

	// 从必需扩展中移除
	for i, ext := range d.doc.ExtensionsRequired {
		if ext == ExtensionName {
			d.doc.ExtensionsRequired = append(d.doc.ExtensionsRequired[:i], d.doc.ExtensionsRequired[i+1:]...)
			break
		}
	}
}

// ====================== 量化处理器 ======================
type Quantizer struct {
	doc    *gltf.Document
	config *QuantizationExtension
}

func NewQuantizer(doc *gltf.Document, config *QuantizationExtension) *Quantizer {
	if config == nil {
		config = &QuantizationExtension{
			PositionBits: 12,
			NormalBits:   10,
			TangentBits:  10,
			TexCoordBits: 12,
			ColorBits:    8,
			WeightBits:   8,
			GenericBits:  8,
		}
	}
	return &Quantizer{doc: doc, config: config}
}

func (q *Quantizer) Process() error {
	q.ensureExtensionDeclared()

	for _, mesh := range q.doc.Meshes {
		for _, primitive := range mesh.Primitives {
			primitiveExt := *q.config // 克隆配置

			for attr, accessorIdx := range primitive.Attributes {
				if accessorIdx >= uint32(len(q.doc.Accessors)) {
					continue
				}
				accessor := q.doc.Accessors[accessorIdx]

				// 跳过非浮点类型和JOINTS属性
				if accessor.ComponentType != gltf.ComponentFloat ||
					strings.HasPrefix(attr, "JOINTS_") {
					continue
				}

				bits := q.getQuantizationBits(attr, &primitiveExt)
				if bits == 0 {
					continue
				}

				// 确定组件类型
				componentType := gltf.ComponentUbyte
				if bits > 8 {
					componentType = gltf.ComponentUshort
				}

				newIdx, err := q.quantizeAccessor(accessor, bits, componentType)
				if err != nil {
					return fmt.Errorf("quantize attribute %s failed: %w", attr, err)
				}

				primitive.Attributes[attr] = newIdx
			}

			// 添加扩展到primitive
			if primitive.Extensions == nil {
				primitive.Extensions = make(map[string]interface{})
			}
			primitive.Extensions[ExtensionName] = &primitiveExt
		}
	}

	return nil
}

func (q *Quantizer) ensureExtensionDeclared() {
	// 添加扩展声明
	if !contains(q.doc.ExtensionsUsed, ExtensionName) {
		q.doc.ExtensionsUsed = append(q.doc.ExtensionsUsed, ExtensionName)
	}

	if !contains(q.doc.ExtensionsRequired, ExtensionName) {
		q.doc.ExtensionsRequired = append(q.doc.ExtensionsRequired, ExtensionName)
	}
}

func (q *Quantizer) getQuantizationBits(attributeName string, ext *QuantizationExtension) uint8 {
	switch {
	case strings.HasPrefix(attributeName, "POSITION"):
		return firstNonZero(ext.PositionBits, 12)
	case strings.HasPrefix(attributeName, "NORMAL"):
		return firstNonZero(ext.NormalBits, 10)
	case strings.HasPrefix(attributeName, "TANGENT"):
		return firstNonZero(ext.TangentBits, 10)
	case strings.HasPrefix(attributeName, "TEXCOORD"):
		return firstNonZero(ext.TexCoordBits, 12)
	case strings.HasPrefix(attributeName, "COLOR"):
		return firstNonZero(ext.ColorBits, 8)
	case strings.HasPrefix(attributeName, "WEIGHTS"):
		return firstNonZero(ext.WeightBits, 8)
	default:
		return firstNonZero(ext.GenericBits, 8)
	}
}

func (q *Quantizer) quantizeAccessor(accessor *gltf.Accessor, bits uint8, componentType gltf.ComponentType) (uint32, error) {
	// 验证访问器数据
	if accessor.ComponentType != gltf.ComponentFloat {
		return 0, fmt.Errorf("only float accessors can be quantized")
	}

	// 获取分量数量
	componentCount, ok := map[gltf.AccessorType]int{
		gltf.AccessorScalar: 1,
		gltf.AccessorVec2:   2,
		gltf.AccessorVec3:   3,
		gltf.AccessorVec4:   4,
	}[accessor.Type]

	if !ok || componentCount < 1 {
		return 0, fmt.Errorf("unsupported accessor type: %s", accessor.Type)
	}

	// 计算最小值和最大值
	minVals, maxVals := q.calculateMinMax(accessor, componentCount)

	// 获取缓冲视图和缓冲区
	if accessor.BufferView == nil {
		return 0, fmt.Errorf("accessor missing buffer view")
	}
	bvIndex := *accessor.BufferView
	if bvIndex >= uint32(len(q.doc.BufferViews)) {
		return 0, fmt.Errorf("buffer view index out of range")
	}
	bv := q.doc.BufferViews[bvIndex]

	if bv.Buffer >= uint32(len(q.doc.Buffers)) {
		return 0, fmt.Errorf("buffer index out of range")
	}
	buffer := q.doc.Buffers[bv.Buffer]

	// 计算数据范围和偏移
	start := bv.ByteOffset + accessor.ByteOffset
	stride := bv.ByteStride
	if stride == 0 {
		stride = uint32(componentCount * 4) // 浮点数每个分量4字节
	}

	// 验证数据范围
	count := accessor.Count
	end := start + (count-1)*stride + uint32(4*componentCount)
	if end > uint32(len(buffer.Data)) {
		return 0, fmt.Errorf("accessor data exceeds buffer range")
	}

	// 准备量化数据存储
	componentSize := gltf.SizeOfComponent(componentType)
	quantizedData := make([]byte, count*uint32(componentSize)*uint32(componentCount))

	// 量化参数
	maxInteger := float32(math.Pow(2, float64(bits)) - 1)
	ranges := make([]float32, componentCount)
	for i := 0; i < componentCount; i++ {
		ranges[i] = maxVals[i] - minVals[i]
		if ranges[i] == 0 {
			ranges[i] = 1e-6 // 避免除以零
		}
	}

	// 处理量化
	for i := uint32(0); i < count; i++ {
		offset := start + i*stride
		baseIdx := i * uint32(componentSize) * uint32(componentCount)

		for c := 0; c < componentCount; c++ {
			// 读取原始浮点值
			raw := binary.LittleEndian.Uint32(buffer.Data[offset : offset+4])
			value := math.Float32frombits(raw)

			// 应用量化公式: raw = clamp(round((value - min) / range * maxInteger), 0, maxInteger)
			normalized := (value - minVals[c]) / ranges[c]
			quantizedValue := normalized * maxInteger
			quantizedValue = float32(math.Round(float64(quantizedValue)))
			quantizedValue = clampFloat(quantizedValue, 0, maxInteger)

			// 根据组件类型写入数据
			idx := baseIdx + uint32(c*componentSize)
			switch componentType {
			case gltf.ComponentUbyte:
				quantizedData[idx] = byte(quantizedValue)
			case gltf.ComponentUshort:
				binary.LittleEndian.PutUint16(quantizedData[idx:], uint16(quantizedValue))
			default:
				return 0, fmt.Errorf("unsupported quantization component type: %d", componentType)
			}

			offset += 4
		}
	}

	// 创建新缓冲区
	newBuffer := gltf.Buffer{
		ByteLength: uint32(len(quantizedData)),
		Data:       quantizedData,
	}
	newBufferIndex := uint32(len(q.doc.Buffers))
	q.doc.Buffers = append(q.doc.Buffers, &newBuffer)

	// 创建新缓冲视图
	newBufferView := gltf.BufferView{
		Buffer:     newBufferIndex,
		ByteOffset: 0,
		ByteLength: newBuffer.ByteLength,
		ByteStride: uint32(componentCount * componentSize),
		Target:     bv.Target,
	}
	newBufferViewIndex := uint32(len(q.doc.BufferViews))
	q.doc.BufferViews = append(q.doc.BufferViews, &newBufferView)

	// 创建新访问器
	newAccessor := &gltf.Accessor{
		BufferView:    &newBufferViewIndex,
		ByteOffset:    0,
		ComponentType: componentType,
		Count:         accessor.Count,
		Type:          accessor.Type,
		Min:           minVals,
		Max:           maxVals,
		Normalized:    true,
	}

	// 添加新访问器并返回索引
	q.doc.Accessors = append(q.doc.Accessors, newAccessor)
	return uint32(len(q.doc.Accessors) - 1), nil
}

func (q *Quantizer) calculateMinMax(accessor *gltf.Accessor, componentCount int) ([]float32, []float32) {
	if accessor.Min != nil && accessor.Max != nil &&
		len(accessor.Min) >= componentCount &&
		len(accessor.Max) >= componentCount {
		return accessor.Min, accessor.Max
	}

	minVals := make([]float32, componentCount)
	maxVals := make([]float32, componentCount)
	for i := range minVals {
		minVals[i] = math.MaxFloat32
		maxVals[i] = -math.MaxFloat32
	}

	if accessor.BufferView == nil {
		return minVals, maxVals
	}

	bvIndex := *accessor.BufferView
	if bvIndex >= uint32(len(q.doc.BufferViews)) {
		return minVals, maxVals
	}
	bv := q.doc.BufferViews[bvIndex]

	if bv.Buffer >= uint32(len(q.doc.Buffers)) {
		return minVals, maxVals
	}
	buffer := q.doc.Buffers[bv.Buffer]

	start := bv.ByteOffset + accessor.ByteOffset
	stride := bv.ByteStride
	if stride == 0 {
		stride = uint32(componentCount * 4)
	}

	count := accessor.Count
	for i := uint32(0); i < count; i++ {
		offset := start + i*stride

		for c := 0; c < componentCount; c++ {
			raw := binary.LittleEndian.Uint32(buffer.Data[offset : offset+4])
			value := math.Float32frombits(raw)

			if value < minVals[c] {
				minVals[c] = value
			}
			if value > maxVals[c] {
				maxVals[c] = value
			}
			offset += 4
		}
	}

	return minVals, maxVals
}

func readComponent(data []byte, componentType gltf.ComponentType, normalized bool) (float32, error) {
	switch componentType {
	case gltf.ComponentByte:
		if len(data) < 1 {
			return 0, fmt.Errorf("insufficient data")
		}
		v := int8(data[0])
		if normalized {
			return float32(v) / 127.0, nil
		}
		return float32(v), nil

	case gltf.ComponentUbyte:
		if len(data) < 1 {
			return 0, fmt.Errorf("insufficient data")
		}
		v := data[0]
		if normalized {
			return float32(v) / 255.0, nil
		}
		return float32(v), nil

	case gltf.ComponentShort:
		if len(data) < 2 {
			return 0, fmt.Errorf("insufficient data")
		}
		v := int16(binary.LittleEndian.Uint16(data[0:2]))
		if normalized {
			return float32(v) / 32767.0, nil
		}
		return float32(v), nil

	case gltf.ComponentUshort:
		if len(data) < 2 {
			return 0, fmt.Errorf("insufficient data")
		}
		v := binary.LittleEndian.Uint16(data[0:2])
		if normalized {
			return float32(v) / 65535.0, nil
		}
		return float32(v), nil

	default:
		return 0, fmt.Errorf("unsupported component type: %d", componentType)
	}
}

func float32ToBytes(data []float32) []byte {
	const batchSize = 1024
	bytes := make([]byte, len(data)*4)

	for i := 0; i < len(data); i += batchSize {
		end := i + batchSize
		if end > len(data) {
			end = len(data)
		}

		for j := i; j < end; j++ {
			binary.LittleEndian.PutUint32(bytes[j*4:], math.Float32bits(data[j]))
		}
	}
	return bytes
}

func clampFloat(value, min, max float32) float32 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

func firstNonZero(values ...uint8) uint8 {
	for _, v := range values {
		if v != 0 {
			return v
		}
	}
	return 0
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
