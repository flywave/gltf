package tile3d

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/flywave/gltf"
	ext_instance "github.com/flywave/gltf/ext/3dtile/instance"
)

// InstanceDataWriter 封装实例化数据写入逻辑
type InstanceDataWriter struct {
	doc          *gltf.Document
	translations [][3]float32
	rotations    [][4]float32
	scales       [][3]float32
	count        int
}

func WriteInstancingData(
	doc *gltf.Document,
	translations [][3]float32,
	rotations [][4]float32,
	scales [][3]float32,
) error {
	writer := &InstanceDataWriter{
		doc:          doc,
		translations: translations,
		rotations:    rotations,
		scales:       scales,
		count:        len(translations),
	}

	if err := writer.validate(); err != nil {
		return err
	}

	if err := writer.prepareBuffer(); err != nil {
		return err
	}

	if err := writer.createAccessors(); err != nil {
		return err
	}

	writer.createExtension()

	return nil
}

// validate 验证输入数据
func (w *InstanceDataWriter) validate() error {
	if len(w.rotations) != w.count || len(w.scales) != w.count {
		return fmt.Errorf("RTS arrays must have the same length")
	}
	return nil
}

// prepareBuffer 准备缓冲区数据
func (w *InstanceDataWriter) prepareBuffer() error {
	// 序列化数据到二进制
	transData, err := w.serializeData(w.translations)
	if err != nil {
		return fmt.Errorf("failed to serialize translations: %v", err)
	}

	scaleData, err := w.serializeData(w.scales)
	if err != nil {
		return fmt.Errorf("failed to serialize scales: %v", err)
	}

	rotData, err := w.serializeData(w.rotations)
	if err != nil {
		return fmt.Errorf("failed to serialize rotations: %v", err)
	}

	// 确保主缓冲区存在
	if len(w.doc.Buffers) == 0 {
		w.doc.Buffers = append(w.doc.Buffers, &gltf.Buffer{})
	}

	// 将数据追加到缓冲区
	buffer := w.doc.Buffers[0]
	byteOffset := buffer.ByteLength

	w.appendBufferData(buffer, transData)
	w.appendBufferData(buffer, scaleData)
	w.appendBufferData(buffer, rotData)

	// 创建缓冲区视图
	bufferView := &gltf.BufferView{
		Buffer:     0,
		ByteOffset: byteOffset,
		ByteLength: uint32(len(transData) + len(scaleData) + len(rotData)),
		Target:     gltf.TargetArrayBuffer,
	}
	w.doc.BufferViews = append(w.doc.BufferViews, bufferView)

	return nil
}

// createAccessors 创建访问器
func (w *InstanceDataWriter) createAccessors() error {
	// 计算各数据段的偏移量
	transSize := w.count * 3 * 4 // 3 floats * 4 bytes each
	scaleSize := w.count * 3 * 4

	// 创建访问器
	baseAccessorIndex := uint32(len(w.doc.Accessors))

	// 平移访问器
	w.createVectorAccessor(baseAccessorIndex, gltf.AccessorVec3, 0)

	// 缩放访问器
	w.createVectorAccessor(baseAccessorIndex+1, gltf.AccessorVec3, uint32(transSize))

	// 旋转访问器
	w.createVectorAccessor(baseAccessorIndex+2, gltf.AccessorVec4, uint32(transSize+scaleSize))

	padBuffer(w.doc)
	return nil
}

// createExtension 创建实例化扩展
func (w *InstanceDataWriter) createExtension() {
	baseAccessorIndex := uint32(len(w.doc.Accessors) - 3)

	instancingConfig := map[string]interface{}{
		"attributes": map[string]uint32{
			ext_instance.Translation: baseAccessorIndex,
			ext_instance.Scale:       baseAccessorIndex + 1,
			ext_instance.Rotation:    baseAccessorIndex + 2,
		},
	}

	w.doc.Extensions[ext_instance.ExtensionName] = instancingConfig
	addExtensionUsed(w.doc, ext_instance.ExtensionName)
}

// serializeData 通用数据序列化函数
func (w *InstanceDataWriter) serializeData(data interface{}) ([]byte, error) {
	buf := new(bytes.Buffer)
	if err := binary.Write(buf, binary.LittleEndian, data); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// appendBufferData 将数据追加到缓冲区
func (w *InstanceDataWriter) appendBufferData(buffer *gltf.Buffer, data []byte) {
	buffer.Data = append(buffer.Data, data...)
	buffer.ByteLength += uint32(len(data))
}

// createVectorAccessor 创建向量访问器
func (w *InstanceDataWriter) createVectorAccessor(index uint32, accType gltf.AccessorType, byteOffset uint32) {
	accessor := &gltf.Accessor{
		ComponentType: gltf.ComponentFloat,
		Type:          accType,
		Count:         uint32(w.count),
		BufferView:    &index,
		ByteOffset:    byteOffset,
	}
	w.doc.Accessors = append(w.doc.Accessors, accessor)
}

// padBuffer 填充缓冲区对齐
