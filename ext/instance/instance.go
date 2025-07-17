package instance

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/flywave/gltf"
)

const ExtensionName = "EXT_mesh_gpu_instancing"

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

	// 统一处理所有属性
	properties := []struct {
		name     string
		data     interface{}
		compType gltf.ComponentType
		accType  gltf.AccessorType
	}{
		{"TRANSLATION", data.Translations, config.TranslationType, gltf.AccessorVec3},
		{"ROTATION", data.Rotations, config.RotationType, gltf.AccessorVec4},
		{"SCALE", data.Scales, config.ScaleType, gltf.AccessorVec3},
	}

	for _, prop := range properties {
		idx, err := createVectorAccessor(doc, prop.data, prop.compType, prop.accType, config.Normalized)
		if err != nil {
			return fmt.Errorf("%s属性创建失败: %w", prop.name, err)
		}
		attrs[prop.name] = idx
	}

	// 设置扩展配置
	if doc.Extensions == nil {
		doc.Extensions = make(gltf.Extensions)
	}
	doc.Extensions[ExtensionName] = map[string]interface{}{"attributes": attrs}
	doc.AddExtensionUsed(ExtensionName)

	return nil
}

func createVectorAccessor(doc *gltf.Document, data interface{},
	compType gltf.ComponentType, accType gltf.AccessorType, normalized bool) (uint32, error) {

	var byteData []byte
	var count uint32

	// 直接处理底层字节数组，避免反射开销
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
	default:
		return 0, fmt.Errorf("不支持的向量数据类型: %T", data)
	}

	// 处理空数据情况
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

// 直接转换float32切片为字节数组
func flattenFloat32Slice(slice interface{}, dim int) []byte {
	buf := bytes.NewBuffer(nil)
	switch v := slice.(type) {
	case [][3]float32:
		buf.Grow(len(v) * dim * 4)
	case [][4]float32:
		buf.Grow(len(v) * dim * 4)
	}
	binary.Write(buf, binary.LittleEndian, slice)
	return buf.Bytes()
}

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
