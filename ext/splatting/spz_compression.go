package splatting

import (
	"fmt"

	"github.com/flywave/gltf"
)

const (
	// SpzCompressionExtensionName KHR_gaussian_splatting_compression_spz_2 扩展名
	SpzCompressionExtensionName = "KHR_gaussian_splatting_compression_spz_2"
)

// SpzCompression 表示 KHR_gaussian_splatting_compression_spz_2 扩展
// 将原始 SPZ 压缩数据作为 bufferView 嵌入 GLB
type SpzCompression struct {
	BufferView uint32 `json:"bufferView"`
}

// WireSpzCompression 将SPZ压缩数据嵌入GLB并返回扩展实例
// spzData 为完整 SPZ 文件内容（含SPZ头+压缩高斯数据）
func WireSpzCompression(doc *gltf.Document, spzData []byte) (*GaussianSplatting, error) {
	if len(spzData) == 0 {
		return nil, fmt.Errorf("SPZ数据不能为空")
	}

	if len(doc.Buffers) == 0 {
		doc.Buffers = append(doc.Buffers, &gltf.Buffer{})
	}
	buffer := doc.Buffers[0]

	spzBufferView := &gltf.BufferView{
		Buffer:     0,
		ByteOffset: buffer.ByteLength,
		ByteLength: uint32(len(spzData)),
	}
	doc.BufferViews = append(doc.BufferViews, spzBufferView)
	bvIndex := uint32(len(doc.BufferViews) - 1)

	buffer.ByteLength += spzBufferView.ByteLength
	buffer.Data = append(buffer.Data, spzData...)

	addExtensionUsed(doc, ExtensionName)
	addExtensionUsed(doc, SpzCompressionExtensionName)
	addExtensionRequired(doc, ExtensionName)
	addExtensionRequired(doc, SpzCompressionExtensionName)

	primitive := &gltf.Primitive{
		Mode: gltf.PrimitivePoints,
		Extensions: gltf.Extensions{
			ExtensionName: &GaussianSplatting{
				SpzCompression: &SpzCompression{
					BufferView: bvIndex,
				},
			},
		},
	}

	meshIndex := 0
	if len(doc.Meshes) > 0 {
		for i, m := range doc.Meshes {
			if m.Name == "GaussianSplattingMesh" {
				meshIndex = i
				break
			}
		}
	}
	if meshIndex >= len(doc.Meshes) {
		doc.Meshes = append(doc.Meshes, &gltf.Mesh{Name: "GaussianSplattingMesh"})
		meshIndex = len(doc.Meshes) - 1
	}
	doc.Meshes[meshIndex].Primitives = append(doc.Meshes[meshIndex].Primitives, primitive)

	node := &gltf.Node{
		Name: "GaussianSplattingNode",
		Mesh: gltf.Index(uint32(meshIndex)),
	}
	doc.Nodes = append(doc.Nodes, node)
	nodeIndex := uint32(len(doc.Nodes) - 1)

	if len(doc.Scenes) == 0 {
		doc.Scenes = append(doc.Scenes, &gltf.Scene{Name: "Default Scene"})
	}
	if doc.Scene == nil {
		doc.Scene = gltf.Index(0)
	}
	doc.Scenes[*doc.Scene].Nodes = append(doc.Scenes[*doc.Scene].Nodes, nodeIndex)

	return primitive.Extensions[ExtensionName].(*GaussianSplatting), nil
}

// addExtensionRequired 添加扩展到extensionsRequired
func addExtensionRequired(doc *gltf.Document, ext string) {
	for _, existing := range doc.ExtensionsRequired {
		if existing == ext {
			return
		}
	}
	doc.ExtensionsRequired = append(doc.ExtensionsRequired, ext)
}
