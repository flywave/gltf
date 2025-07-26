package draco

import (
	"testing"

	"github.com/flywave/gltf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadDraco(t *testing.T) {
	// 读取 GLTF 文件
	doc, err := gltf.Open("../../testdata/Draco/draco.gltf")
	require.NoError(t, err, "无法打开 GLTF 文件")

	// 解码前验证扩展存在
	primitive := doc.Meshes[0].Primitives[0]
	_, extExists := primitive.Extensions[ExtensionName]
	assert.True(t, extExists, "解码前应包含Draco扩展")

	// 执行解码
	err = DecodeAll(doc)
	require.NoError(t, err, "解码失败")

	// 验证扩展已移除
	_, extExists = primitive.Extensions[ExtensionName]
	assert.False(t, extExists, "解码后应移除Draco扩展")
	assert.NotContains(t, doc.ExtensionsUsed, ExtensionName, "应从ExtensionsUsed中移除扩展")

	// 验证访问器数量（原始4个访问器 + 解码后新增的属性访问器）
	assert.GreaterOrEqual(t, len(doc.Accessors), 4, "访问器数量应不少于4个")

	// 验证属性存在性
	assert.Contains(t, primitive.Attributes, "POSITION", "应包含POSITION属性")
	assert.Contains(t, primitive.Attributes, "NORMAL", "应包含NORMAL属性")
	assert.Contains(t, primitive.Attributes, "TEXCOORD_0", "应包含TEXCOORD_0属性")

	// 验证POSITION访问器
	posAccessor := doc.Accessors[primitive.Attributes["POSITION"]]
	assert.Equal(t, gltf.ComponentFloat, posAccessor.ComponentType, "POSITION组件类型应为float")
	assert.Equal(t, gltf.AccessorVec3, posAccessor.Type, "POSITION类型应为VEC3")
	assert.Equal(t, uint32(5025), posAccessor.Count, "POSITION访问器计数应为5025")

	// 验证索引访问器
	assert.NotNil(t, primitive.Indices, "索引访问器不应为nil")
	indexAccessor := doc.Accessors[*primitive.Indices]
	assert.Equal(t, gltf.ComponentUshort, indexAccessor.ComponentType, "索引组件类型应为ushort")
	assert.Equal(t, uint32(28800), indexAccessor.Count, "索引访问器计数应为28800")

	// 验证解码后缓冲区
	posAccessorIdx := primitive.Attributes["POSITION"]
	posAccessor = doc.Accessors[posAccessorIdx]
	assert.NotNil(t, posAccessor.BufferView, "POSITION访问器应引用缓冲区视图")

	bufView := doc.BufferViews[*posAccessor.BufferView]
	decodedBuffer := doc.Buffers[bufView.Buffer]
	assert.Greater(t, decodedBuffer.ByteLength, uint32(0), "解码后缓冲区应设置为0")

	// 执行重新编码
	err = EncodeAll(doc, nil)
	require.NoError(t, err, "编码失败")

	// 验证编码后扩展存在
	primitive = doc.Meshes[0].Primitives[0]
	extData, extExists := primitive.Extensions[ExtensionName]
	assert.True(t, extExists, "编码后应包含Draco扩展")
	assert.Contains(t, doc.ExtensionsUsed, ExtensionName, "ExtensionsUsed应包含Draco扩展")

	// 验证Draco扩展数据结构
	dracoExt, ok := extData.(*DracoExtension)
	require.True(t, ok, "Draco扩展格式错误")
	assert.Equal(t, dracoExt.BufferView, uint32(0), "BufferView索引应有效")
	assert.NotEmpty(t, dracoExt.Attributes, "属性映射不应为空")

	// 验证压缩数据缓冲区
	bufferView := doc.BufferViews[dracoExt.BufferView]
	assert.NotNil(t, bufferView, "缓冲区视图不应为nil")
	assert.Less(t, int(bufferView.Buffer), len(doc.Buffers), "缓冲区索引有效")
	buffer := doc.Buffers[bufferView.Buffer]
	assert.NotNil(t, buffer, "缓冲区不应为nil")
	assert.Greater(t, buffer.ByteLength, uint32(0), "缓冲区长度应大于0")

	// 验证原始访问器已被置空
	for _, attrAccessorIdx := range primitive.Attributes {
		attrAccessor := doc.Accessors[attrAccessorIdx]
		assert.Nil(t, attrAccessor.BufferView, "属性访问器的BufferView应被置空")
		assert.Equal(t, uint32(0), attrAccessor.ByteOffset, "属性访问器的ByteOffset应重置为0")
	}

	// 验证索引访问器被置空
	if primitive.Indices != nil {
		indexAccessor := doc.Accessors[*primitive.Indices]
		assert.Nil(t, indexAccessor.BufferView, "索引访问器的BufferView应被置空")
		assert.Equal(t, uint32(0), indexAccessor.ByteOffset, "索引访问器的ByteOffset应重置为0")
	}

	// 验证解码和编码循环后的数据一致性
	assert.Equal(t, doc.Meshes[0].Primitives[0].Attributes, doc.Meshes[0].Primitives[0].Attributes, "属性映射应保持一致")

	err = DecodeAll(doc)
	require.NoError(t, err, "重新解码失败")
}
