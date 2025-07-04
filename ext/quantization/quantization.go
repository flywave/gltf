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
	PositionBits  uint8  `json:"POSITION,omitempty"`
	NormalBits    uint8  `json:"NORMAL,omitempty"`
	TangentBits   uint8  `json:"TANGENT,omitempty"`
	TexCoordBits  uint8  `json:"TEXCOORD,omitempty"`
	ColorBits     uint8  `json:"COLOR,omitempty"`
	GenericBits   uint8  `json:"GENERIC,omitempty"`
	JointBits     uint8  `json:"JOINTS,omitempty"`
	WeightBits    uint8  `json:"WEIGHTS,omitempty"`
	ComponentType string `json:"componentType,omitempty"`
}

func init() {
	gltf.RegisterExtension(ExtensionName, Unmarshal)
}

func Unmarshal(data []byte) (interface{}, error) {
	ext := &QuantizationExtension{}
	if err := json.Unmarshal(data, ext); err != nil {
		return nil, fmt.Errorf("KHR_mesh_quantization解析失败: %w", err)
	}
	return ext, nil
}

// 量化编码函数
func QuantizeFloat32(data []float32, bits uint8, normalized bool) ([]byte, error) {
	if bits == 0 || bits > 16 {
		return nil, fmt.Errorf("无效的量化位数: %d", bits)
	}

	maxValue := float32(math.Pow(2, float64(bits)) - 1)
	quantized := make([]byte, len(data)*2) // 假设使用uint16存储

	for i, v := range data {
		if normalized {
			v = (v + 1) * 0.5 * maxValue // [-1,1] -> [0,maxValue]
		}
		val := uint16(math.Round(float64(v)))
		quantized[i*2] = byte(val & 0xFF)
		quantized[i*2+1] = byte(val >> 8)
	}
	return quantized, nil
}

// 量化解码函数
func DequantizeUint16(data []byte, bits uint8, normalized bool) []float32 {
	maxValue := float32(math.Pow(2, float64(bits)) - 1)
	result := make([]float32, len(data)/2)

	for i := 0; i < len(data); i += 2 {
		val := uint16(data[i]) | uint16(data[i+1])<<8
		f := float32(val)
		if normalized {
			f = (f/maxValue)*2 - 1 // [0,maxValue] -> [-1,1]
		}
		result[i/2] = f
	}
	return result
}

// 解量化网格数据的主函数
func DequantizeMeshData(doc *gltf.Document) error {
	for m := range doc.Meshes {
		mesh := doc.Meshes[m]
		for p := range mesh.Primitives {
			primitive := mesh.Primitives[p]
			for attr, accessorIndex := range primitive.Attributes {
				// 只处理特定属性
				if attr == "POSITION" || attr == "NORMAL" ||
					attr == "TANGENT" || strings.HasPrefix(attr, "TEXCOORD_") {

					if accessorIndex >= uint32(len(doc.Accessors)) {
						continue
					}
					accessor := doc.Accessors[accessorIndex]
					if accessor.ComponentType == gltf.ComponentFloat {
						continue
					}
					if err := dequantizeAccessor(doc, accessor); err != nil {
						return fmt.Errorf("解量化访问器失败: %w", err)
					}
				}
			}
		}
	}
	return nil
}

// 解量化单个访问器
func dequantizeAccessor(doc *gltf.Document, accessor *gltf.Accessor) error {
	// 获取分量数量
	componentSize, ok := map[gltf.AccessorType]int{
		gltf.AccessorScalar: 1,
		gltf.AccessorVec2:   2,
		gltf.AccessorVec3:   3,
		gltf.AccessorVec4:   4,
	}[accessor.Type]

	if !ok || componentSize < 2 {
		return nil // 不处理其他类型或无效分量
	}

	// 获取缓冲视图和缓冲区
	if accessor.BufferView == nil {
		return fmt.Errorf("访问器缺少缓冲视图")
	}
	bvIndex := *accessor.BufferView
	if bvIndex >= uint32(len(doc.BufferViews)) {
		return fmt.Errorf("缓冲视图索引越界")
	}
	bv := doc.BufferViews[bvIndex]

	if bv.Buffer >= uint32(len(doc.Buffers)) {
		return fmt.Errorf("缓冲区索引越界")
	}
	buffer := doc.Buffers[bv.Buffer]
	data := buffer.Data

	// 计算原始数据位置和步长
	start := bv.ByteOffset + accessor.ByteOffset
	stride := bv.ByteStride
	if stride == 0 {
		stride = uint32(componentSize * gltf.SizeOfComponent(accessor.ComponentType))
	}

	// 验证数据范围
	count := accessor.Count
	end := start + uint32(count-1)*stride + uint32(componentSize*gltf.SizeOfComponent(accessor.ComponentType))
	if end > uint32(len(data)) {
		return fmt.Errorf("访问器数据超出缓冲区范围")
	}

	// 准备浮点数据存储
	floatData := make([]float32, count*uint32(componentSize))

	// 解量化数据
	if err := convertToFloat(
		floatData,
		data[start:],
		count,
		stride,
		accessor.ComponentType,
		accessor.Normalized,
		componentSize,
	); err != nil {
		return err
	}

	// 更新min/max值
	updateMinMax(accessor)

	// 创建新缓冲区
	byteData := float32ToBytes(floatData)
	newBuffer := gltf.Buffer{
		ByteLength: uint32(len(byteData)),
		Data:       byteData,
	}
	newBufferIndex := uint32(len(doc.Buffers))
	doc.Buffers = append(doc.Buffers, &newBuffer)

	// 创建新缓冲视图
	newBufferView := gltf.BufferView{
		Buffer:     newBufferIndex,
		ByteOffset: 0,
		ByteLength: newBuffer.ByteLength,
		ByteStride: uint32(componentSize * 4), // 每个浮点数4字节
		Target:     bv.Target,
	}
	newBufferViewIndex := uint32(len(doc.BufferViews))
	doc.BufferViews = append(doc.BufferViews, &newBufferView)

	// 更新访问器
	accessor.BufferView = &newBufferViewIndex
	accessor.ByteOffset = 0
	accessor.ComponentType = gltf.ComponentFloat
	accessor.Normalized = false

	return nil
}

// 将原始数据转换为浮点数
func convertToFloat(
	floatData []float32,
	rawData []byte,
	count uint32,
	stride uint32,
	componentType gltf.ComponentType,
	normalized bool,
	componentSize int,
) error {
	index := 0
	for i := uint32(0); i < count; i++ {
		offset := i * stride
		for j := 0; j < componentSize; j++ {
			var value float32
			var err error

			switch componentType {
			case gltf.ComponentByte:
				value, err = convertByte(rawData, offset, normalized)
				offset += 1
			case gltf.ComponentUbyte:
				value, err = convertUbyte(rawData, offset, normalized)
				offset += 1
			case gltf.ComponentShort:
				value, err = convertShort(rawData, offset, normalized)
				offset += 2
			case gltf.ComponentUshort:
				value, err = convertUshort(rawData, offset, normalized)
				offset += 2
			default:
				return fmt.Errorf("不支持的组件类型: %d", componentType)
			}

			if err != nil {
				return err
			}

			floatData[index] = value
			index++
		}
	}
	return nil
}

// 各种数据类型的转换函数
func convertByte(data []byte, offset uint32, normalized bool) (float32, error) {
	if int(offset) >= len(data) {
		return 0, fmt.Errorf("字节偏移超出范围")
	}
	v := int8(data[offset])
	if normalized {
		return float32(math.Max(float64(v)/127.0, -1.0)), nil
	}
	return float32(v), nil
}

func convertUbyte(data []byte, offset uint32, normalized bool) (float32, error) {
	if int(offset) >= len(data) {
		return 0, fmt.Errorf("无符号字节偏移超出范围")
	}
	v := data[offset]
	if normalized {
		return float32(v) / 255.0, nil
	}
	return float32(v), nil
}

func convertShort(data []byte, offset uint32, normalized bool) (float32, error) {
	if int(offset+1) >= len(data) {
		return 0, fmt.Errorf("短整型偏移超出范围")
	}
	v := int16(binary.LittleEndian.Uint16(data[offset : offset+2]))
	if normalized {
		return float32(math.Max(float64(v)/32767.0, -1.0)), nil
	}
	return float32(v), nil
}

func convertUshort(data []byte, offset uint32, normalized bool) (float32, error) {
	if int(offset+1) >= len(data) {
		return 0, fmt.Errorf("无符号短整型偏移超出范围")
	}
	v := binary.LittleEndian.Uint16(data[offset : offset+2])
	if normalized {
		return float32(v) / 65535.0, nil
	}
	return float32(v), nil
}

// 更新访问器的min/max值
func updateMinMax(accessor *gltf.Accessor) {
	convert := func(value float32, compType gltf.ComponentType, normalized bool) float32 {
		if !normalized {
			return float32(value)
		}
		switch compType {
		case gltf.ComponentByte:
			return float32(math.Max(float64(value/127.0), -1.0))
		case gltf.ComponentUbyte:
			return float32(value / 255.0)
		case gltf.ComponentShort:
			return float32(math.Max(float64(value/32767.0), -1.0))
		case gltf.ComponentUshort:
			return float32(value / 65535.0)
		default:
			return float32(value)
		}
	}

	if accessor.Min != nil {
		for i := range accessor.Min {
			accessor.Min[i] = convert(accessor.Min[i], accessor.ComponentType, accessor.Normalized)
		}
	}
	if accessor.Max != nil {
		for i := range accessor.Max {
			accessor.Max[i] = convert(accessor.Max[i], accessor.ComponentType, accessor.Normalized)
		}
	}
}

// 浮点数切片转字节切片
func float32ToBytes(data []float32) []byte {
	bytes := make([]byte, len(data)*4)
	for i, f := range data {
		u := math.Float32bits(f)
		binary.LittleEndian.PutUint32(bytes[i*4:], u)
	}
	return bytes
}
