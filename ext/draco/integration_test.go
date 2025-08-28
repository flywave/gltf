package draco

import (
	"math"
	"testing"

	"github.com/flywave/gltf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDracoIntegration 测试Draco压缩和解压缩的完整集成流程
func TestDracoIntegration(t *testing.T) {
	// 1. 加载一个包含Draco压缩的GLTF模型
	doc, err := gltf.Open("../../testdata/Draco/draco.gltf")
	require.NoError(t, err, "无法打开测试GLTF文件")
	require.NotNil(t, doc, "文档不应为nil")

	// 2. 验证初始状态 - 应该包含Draco扩展
	originalPrimitive := doc.Meshes[0].Primitives[0]
	_, hasDracoExt := originalPrimitive.Extensions[ExtensionName]
	assert.True(t, hasDracoExt, "初始模型应包含Draco扩展")
	assert.Contains(t, doc.ExtensionsUsed, ExtensionName, "ExtensionsUsed应包含Draco扩展")

	// 3. 记录解压前的访问器状态
	originalPositionAccessor := doc.Accessors[originalPrimitive.Attributes["POSITION"]]
	originalIndexAccessor := doc.Accessors[*originalPrimitive.Indices]

	// 解压前，这些访问器应该没有BufferView（因为数据在Draco压缩中）
	assert.Nil(t, originalPositionAccessor.BufferView, "解压前POSITION访问器不应有BufferView")
	assert.Nil(t, originalIndexAccessor.BufferView, "解压前索引访问器不应有BufferView")

	// 4. 执行解压缩
	err = DecodeAll(doc)
	require.NoError(t, err, "解压缩应成功")

	// 5. 验证解压后的状态
	decodedPrimitive := doc.Meshes[0].Primitives[0]
	_, hasDracoExt = decodedPrimitive.Extensions[ExtensionName]
	assert.False(t, hasDracoExt, "解压后应移除Draco扩展")
	assert.NotContains(t, doc.ExtensionsUsed, ExtensionName, "ExtensionsUsed应移除Draco扩展")

	// 6. 验证解压后的访问器状态
	decodedPositionAccessor := doc.Accessors[decodedPrimitive.Attributes["POSITION"]]
	decodedIndexAccessor := doc.Accessors[*decodedPrimitive.Indices]

	// 解压后，这些访问器应该有BufferView（因为数据已解压到缓冲区）
	assert.NotNil(t, decodedPositionAccessor.BufferView, "解压后POSITION访问器应有BufferView")
	assert.NotNil(t, decodedIndexAccessor.BufferView, "解压后索引访问器应有BufferView")

	// 验证数据计数
	assert.Equal(t, uint32(5025), decodedPositionAccessor.Count, "POSITION访问器计数应正确")
	assert.Equal(t, uint32(28800), decodedIndexAccessor.Count, "索引访问器计数应正确")

	// 7. 保存解压后的文档（模拟用户可能做的修改）
	// 这里我们直接进行重新压缩，但在实际应用中用户可能在这里修改数据

	// 8. 执行重新压缩
	options := map[string]interface{}{
		"quantization": map[string]int{
			"position": 14,
			"normal":   10,
			"texcoord": 12,
		},
	}
	err = EncodeAll(doc, options)
	require.NoError(t, err, "重新压缩应成功")

	// 9. 验证重新压缩后的状态
	encodedPrimitive := doc.Meshes[0].Primitives[0]
	_, hasDracoExt = encodedPrimitive.Extensions[ExtensionName]
	assert.True(t, hasDracoExt, "重新压缩后应包含Draco扩展")
	assert.Contains(t, doc.ExtensionsUsed, ExtensionName, "ExtensionsUsed应包含Draco扩展")

	// 10. 验证压缩后的访问器状态
	encodedPositionAccessor := doc.Accessors[encodedPrimitive.Attributes["POSITION"]]
	encodedIndexAccessor := doc.Accessors[*encodedPrimitive.Indices]

	// 压缩后，这些访问器应该没有BufferView（因为数据在Draco压缩中）
	assert.Nil(t, encodedPositionAccessor.BufferView, "压缩后POSITION访问器不应有BufferView")
	assert.Nil(t, encodedIndexAccessor.BufferView, "压缩后索引访问器不应有BufferView")

	// 11. 验证Draco扩展数据
	dracoExt, ok := encodedPrimitive.Extensions[ExtensionName].(*DracoExtension)
	require.True(t, ok, "Draco扩展应为正确类型")
	assert.NotNil(t, dracoExt.BufferView, "Draco扩展应有BufferView")
	assert.NotEmpty(t, dracoExt.Attributes, "Draco扩展应有属性映射")

	// 12. 最后再次解压以验证循环一致性
	err = DecodeAll(doc)
	require.NoError(t, err, "再次解压应成功")

	// 13. 验证最终状态与第一次解压后一致
	finalPrimitive := doc.Meshes[0].Primitives[0]
	finalPositionAccessor := doc.Accessors[finalPrimitive.Attributes["POSITION"]]
	finalIndexAccessor := doc.Accessors[*finalPrimitive.Indices]

	assert.NotNil(t, finalPositionAccessor.BufferView, "最终POSITION访问器应有BufferView")
	assert.NotNil(t, finalIndexAccessor.BufferView, "最终索引访问器应有BufferView")
	assert.Equal(t, decodedPositionAccessor.Count, finalPositionAccessor.Count, "POSITION访问器计数应一致")
	assert.Equal(t, decodedIndexAccessor.Count, finalIndexAccessor.Count, "索引访问器计数应一致")
}

// TestDracoEncodeDecodeCycle 测试编码-解码循环
func TestDracoEncodeDecodeCycle(t *testing.T) {
	// 1. 创建一个简单的测试文档
	doc := &gltf.Document{
		Asset: gltf.Asset{
			Version:   "2.0",
			Generator: "Draco Integration Test",
		},
		Buffers: []*gltf.Buffer{
			{
				ByteLength: 0,
				Data:       []byte{},
			},
		},
		BufferViews: []*gltf.BufferView{},
		Accessors:   []*gltf.Accessor{},
		Meshes: []*gltf.Mesh{
			{
				Name: "TestMesh",
				Primitives: []*gltf.Primitive{
					{
						Attributes: map[string]uint32{},
						Indices:    nil,
						Material:   nil,
					},
				},
			},
		},
	}

	// 2. 添加测试数据 - 简单的三角形
	// 位置数据 (3个顶点，每个顶点3个分量)
	positionData := []float32{
		0.0, 0.0, 0.0, // 顶点0
		1.0, 0.0, 0.0, // 顶点1
		0.0, 1.0, 0.0, // 顶点2
	}

	// 索引数据 (1个三角形)
	indexData := []uint16{0, 1, 2}

	// 3. 创建缓冲区和视图
	// 位置缓冲区
	posBuffer := make([]byte, len(positionData)*4) // float32占4字节
	for i, v := range positionData {
		bits := math.Float32bits(v)
		posBuffer[i*4] = byte(bits)
		posBuffer[i*4+1] = byte(bits >> 8)
		posBuffer[i*4+2] = byte(bits >> 16)
		posBuffer[i*4+3] = byte(bits >> 24)
	}

	// 索引缓冲区
	idxBuffer := make([]byte, len(indexData)*2) // uint16占2字节
	for i, v := range indexData {
		idxBuffer[i*2] = byte(v)
		idxBuffer[i*2+1] = byte(v >> 8)
	}

	// 4. 添加缓冲区数据到文档
	doc.Buffers[0].Data = append(doc.Buffers[0].Data, posBuffer...)
	doc.Buffers[0].Data = append(doc.Buffers[0].Data, idxBuffer...)
	doc.Buffers[0].ByteLength = uint32(len(doc.Buffers[0].Data))

	// 5. 创建缓冲区视图
	posView := &gltf.BufferView{
		Buffer:     0,
		ByteOffset: 0,
		ByteLength: uint32(len(posBuffer)),
		Target:     gltf.TargetArrayBuffer,
	}
	doc.BufferViews = append(doc.BufferViews, posView)

	idxView := &gltf.BufferView{
		Buffer:     0,
		ByteOffset: uint32(len(posBuffer)),
		ByteLength: uint32(len(idxBuffer)),
		Target:     gltf.TargetElementArrayBuffer,
	}
	doc.BufferViews = append(doc.BufferViews, idxView)

	// 6. 创建访问器
	posAccessor := &gltf.Accessor{
		BufferView:    gltf.Index(0),
		ByteOffset:    0,
		ComponentType: gltf.ComponentFloat,
		Count:         3,
		Type:          gltf.AccessorVec3,
		Max:           []float32{1.0, 1.0, 0.0},
		Min:           []float32{0.0, 0.0, 0.0},
	}
	doc.Accessors = append(doc.Accessors, posAccessor)

	idxAccessor := &gltf.Accessor{
		BufferView:    gltf.Index(1),
		ByteOffset:    0,
		ComponentType: gltf.ComponentUshort,
		Count:         3,
		Type:          gltf.AccessorScalar,
	}
	doc.Accessors = append(doc.Accessors, idxAccessor)

	// 7. 关联访问器到图元
	primitive := doc.Meshes[0].Primitives[0]
	primitive.Attributes["POSITION"] = 0
	primitive.Indices = gltf.Index(1)

	// 8. 执行编码
	err := EncodeAll(doc, nil)
	assert.NoError(t, err, "编码应成功")

	// 9. 验证编码结果
	encodedPrimitive := doc.Meshes[0].Primitives[0]
	_, hasDracoExt := encodedPrimitive.Extensions[ExtensionName]
	assert.True(t, hasDracoExt, "编码后应包含Draco扩展")

	// 10. 执行解码
	err = DecodeAll(doc)
	assert.NoError(t, err, "解码应成功")

	// 11. 验证解码结果
	decodedPrimitive := doc.Meshes[0].Primitives[0]
	_, hasDracoExt = decodedPrimitive.Extensions[ExtensionName]
	assert.False(t, hasDracoExt, "解码后应移除Draco扩展")

	// 12. 验证数据完整性
	finalPosAccessor := doc.Accessors[decodedPrimitive.Attributes["POSITION"]]
	finalIdxAccessor := doc.Accessors[*decodedPrimitive.Indices]

	assert.Equal(t, uint32(3), finalPosAccessor.Count, "POSITION访问器计数应正确")
	assert.Equal(t, uint32(3), finalIdxAccessor.Count, "索引访问器计数应正确")
}
