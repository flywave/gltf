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
		return nil, fmt.Errorf("KHR_mesh_quantization解析失败: %w", err)
	}
	return ext, nil
}

// 主解量化函数
func DequantizeMeshData(doc *gltf.Document) error {
	// 记录是否处理了任何扩展
	processedExtensions := false

	for m := range doc.Meshes {
		mesh := doc.Meshes[m]
		for p := range mesh.Primitives {
			primitive := mesh.Primitives[p]

			// 获取扩展参数
			extValue, exists := primitive.Extensions[ExtensionName]
			if !exists {
				continue
			}

			ext, ok := extValue.(*QuantizationExtension)
			if !ok {
				return fmt.Errorf("无效的KHR_mesh_quantization扩展格式")
			}

			processedExtensions = true

			// 处理所有属性
			for attr, accessorIndex := range primitive.Attributes {
				if accessorIndex >= uint32(len(doc.Accessors)) {
					continue
				}
				accessor := doc.Accessors[accessorIndex]

				// 跳过浮点类型和JOINTS属性
				if accessor.ComponentType == gltf.ComponentFloat ||
					strings.HasPrefix(attr, "JOINTS_") {
					continue
				}

				// 获取量化位数
				bits := getQuantizationBits(attr, ext)
				if bits == 0 {
					continue
				}

				newAccessorIndex, err := dequantizeAccessor(doc, accessor, bits)
				if err != nil {
					return fmt.Errorf("解量化属性 %s 失败: %w", attr, err)
				}

				// 更新属性索引
				primitive.Attributes[attr] = newAccessorIndex
			}

			// 移除primitive的扩展
			delete(primitive.Extensions, ExtensionName)
		}
	}

	// 移除顶级扩展声明
	if processedExtensions {
		removeTopLevelExtension(doc)
	}

	return nil
}

// 获取属性的量化位数
func getQuantizationBits(attributeName string, ext *QuantizationExtension) uint8 {
	switch {
	case strings.HasPrefix(attributeName, "POSITION"):
		if ext.PositionBits > 0 {
			return ext.PositionBits
		}
		return 12 // 默认值
	case strings.HasPrefix(attributeName, "NORMAL"):
		if ext.NormalBits > 0 {
			return ext.NormalBits
		}
		return 10 // 默认值
	case strings.HasPrefix(attributeName, "TANGENT"):
		if ext.TangentBits > 0 {
			return ext.TangentBits
		}
		return 10 // 默认值
	case strings.HasPrefix(attributeName, "TEXCOORD"):
		if ext.TexCoordBits > 0 {
			return ext.TexCoordBits
		}
		return 12 // 默认值
	case strings.HasPrefix(attributeName, "COLOR"):
		if ext.ColorBits > 0 {
			return ext.ColorBits
		}
		return 8 // 默认值
	case strings.HasPrefix(attributeName, "WEIGHTS"):
		if ext.WeightBits > 0 {
			return ext.WeightBits
		}
		return 8 // 默认值
	default:
		if ext.GenericBits > 0 {
			return ext.GenericBits
		}
		return 8 // 默认值
	}
}

// 解量化单个访问器
func dequantizeAccessor(doc *gltf.Document, accessor *gltf.Accessor, bits uint8) (uint32, error) {
	// 验证访问器数据
	if accessor.Min == nil || accessor.Max == nil {
		return 0, fmt.Errorf("访问器缺少min/max值")
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
		return 0, fmt.Errorf("不支持的访问器类型: %s", accessor.Type)
	}

	// 验证min/max长度
	if len(minValues) < componentCount || len(maxValues) < componentCount {
		return 0, fmt.Errorf("min/max值数量与组件数量不匹配")
	}

	// 获取缓冲视图和缓冲区
	if accessor.BufferView == nil {
		return 0, fmt.Errorf("访问器缺少缓冲视图")
	}
	bvIndex := *accessor.BufferView
	if bvIndex >= uint32(len(doc.BufferViews)) {
		return 0, fmt.Errorf("缓冲视图索引越界")
	}
	bv := doc.BufferViews[bvIndex]

	if bv.Buffer >= uint32(len(doc.Buffers)) {
		return 0, fmt.Errorf("缓冲区索引越界")
	}
	buffer := doc.Buffers[bv.Buffer]

	// 计算数据范围和偏移
	start := bv.ByteOffset + accessor.ByteOffset
	stride := bv.ByteStride
	if stride == 0 {
		stride = uint32(componentCount * gltf.SizeOfComponent(accessor.ComponentType))
	}

	// 验证数据范围
	count := accessor.Count
	end := start + (count-1)*stride + uint32(gltf.SizeOfComponent(accessor.ComponentType))*uint32(componentCount)
	if end > uint32(len(buffer.Data)) {
		return 0, fmt.Errorf("访问器数据超出缓冲区范围")
	}

	// 准备浮点数据存储
	floatData := make([]float32, count*uint32(componentCount))

	// 解量化数据
	maxInteger := float32(math.Pow(2, float64(bits)) - 1)
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

			// 应用解量化公式: value = min + (max - min) * (raw / maxInteger)
			normalized := float32(rawValue) / maxInteger
			floatValue := minValues[c] + (maxValues[c]-minValues[c])*normalized

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
	newBufferIndex := uint32(len(doc.Buffers))
	doc.Buffers = append(doc.Buffers, &newBuffer)

	// 创建新缓冲视图
	newBufferView := gltf.BufferView{
		Buffer:     newBufferIndex,
		ByteOffset: 0,
		ByteLength: newBuffer.ByteLength,
		ByteStride: uint32(componentCount * 4), // 每个浮点数4字节
		Target:     bv.Target,
	}
	newBufferViewIndex := uint32(len(doc.BufferViews))
	doc.BufferViews = append(doc.BufferViews, &newBufferView)

	// 创建新访问器
	newAccessor := &gltf.Accessor{
		BufferView:    &newBufferViewIndex,
		ByteOffset:    0,
		ComponentType: gltf.ComponentFloat,
		Count:         accessor.Count,
		Type:          accessor.Type,
		Min:           minValues, // 保持原始min/max
		Max:           maxValues,
		Normalized:    false,
	}

	// 添加新访问器并返回索引
	doc.Accessors = append(doc.Accessors, newAccessor)
	return uint32(len(doc.Accessors) - 1), nil
}

// 从字节数据读取组件值
func readComponent(data []byte, componentType gltf.ComponentType, normalized bool) (float32, error) {
	switch componentType {
	case gltf.ComponentByte:
		if len(data) < 1 {
			return 0, fmt.Errorf("数据不足")
		}
		v := int8(data[0])
		if normalized {
			return float32(v) / 127.0, nil
		}
		return float32(v), nil

	case gltf.ComponentUbyte:
		if len(data) < 1 {
			return 0, fmt.Errorf("数据不足")
		}
		v := data[0]
		if normalized {
			return float32(v) / 255.0, nil
		}
		return float32(v), nil

	case gltf.ComponentShort:
		if len(data) < 2 {
			return 0, fmt.Errorf("数据不足")
		}
		v := int16(binary.LittleEndian.Uint16(data[0:2]))
		if normalized {
			return float32(v) / 32767.0, nil
		}
		return float32(v), nil

	case gltf.ComponentUshort:
		if len(data) < 2 {
			return 0, fmt.Errorf("数据不足")
		}
		v := binary.LittleEndian.Uint16(data[0:2])
		if normalized {
			return float32(v) / 65535.0, nil
		}
		return float32(v), nil

	default:
		return 0, fmt.Errorf("不支持的组件类型: %d", componentType)
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

// 移除顶级扩展声明
func removeTopLevelExtension(doc *gltf.Document) {
	// 从扩展中移除
	delete(doc.Extensions, ExtensionName)

	// 从已使用扩展中移除
	for i, ext := range doc.ExtensionsUsed {
		if ext == ExtensionName {
			doc.ExtensionsUsed = append(doc.ExtensionsUsed[:i], doc.ExtensionsUsed[i+1:]...)
			break
		}
	}

	// 从必需扩展中移除
	for i, ext := range doc.ExtensionsRequired {
		if ext == ExtensionName {
			doc.ExtensionsRequired = append(doc.ExtensionsRequired[:i], doc.ExtensionsRequired[i+1:]...)
			break
		}
	}
}
