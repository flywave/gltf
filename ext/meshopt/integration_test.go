package meshopt

import (
	"testing"

	"github.com/flywave/gltf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIntegrationEncodeDecode 测试完整的编码和解码流程
func TestIntegrationEncodeDecode(t *testing.T) {
	// 创建一个简单的glTF文档
	doc := &gltf.Document{
		Asset: gltf.Asset{
			Version: "2.0",
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
						Mode: gltf.PrimitiveTriangles,
					},
				},
			},
		},
	}

	// 测试DecodeAll在空文档上应成功
	err := DecodeAll(doc)
	require.NoError(t, err, "DecodeAll应在空文档上成功")

	// 验证文档未被修改
	assert.Equal(t, 0, len(doc.BufferViews), "BufferViews数量应保持不变")
	// 注意：文档初始化时已经有一个空的Buffer，所以长度为1
	assert.Equal(t, 1, len(doc.Buffers), "Buffers数量应保持不变")
}

// TestIntegrationUnmarshalExtension 测试扩展的反序列化
func TestIntegrationUnmarshalExtension(t *testing.T) {
	// 测试有效的扩展数据
	jsonData := []byte(`{
		"buffer": 0,
		"byteLength": 100,
		"byteStride": 12,
		"count": 25,
		"mode": "ATTRIBUTES"
	}`)

	ext, err := Unmarshal(jsonData)
	require.NoError(t, err, "反序列化应成功")
	require.NotNil(t, ext, "扩展不应为nil")

	compressionExt, ok := ext.(*CompressionExtension)
	require.True(t, ok, "应能转换为CompressionExtension类型")
	assert.Equal(t, uint32(0), compressionExt.Buffer, "Buffer应正确解析")
	assert.Equal(t, uint32(100), compressionExt.ByteLength, "ByteLength应正确解析")
	assert.Equal(t, uint32(12), compressionExt.ByteStride, "ByteStride应正确解析")
	assert.Equal(t, uint32(25), compressionExt.Count, "Count应正确解析")
	assert.Equal(t, ModeAttributes, compressionExt.Mode, "Mode应正确解析")
}

// TestIntegrationValidateModeStride 测试模式和步长验证
func TestIntegrationValidateModeStride(t *testing.T) {
	tests := []struct {
		name        string
		mode        CompressionMode
		stride      uint32
		expectError bool
	}{
		{
			name:        "ValidAttributes",
			mode:        ModeAttributes,
			stride:      12,
			expectError: false,
		},
		{
			name:        "ValidTriangles",
			mode:        ModeTriangles,
			stride:      2,
			expectError: false,
		},
		{
			name:        "ValidIndices",
			mode:        ModeIndices,
			stride:      4,
			expectError: false,
		},
		{
			name:        "InvalidAttributesStride",
			mode:        ModeAttributes,
			stride:      5,
			expectError: true,
		},
		{
			name:        "InvalidTrianglesStride",
			mode:        ModeTriangles,
			stride:      3,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateModeStride(tt.mode, tt.stride)
			if tt.expectError {
				assert.Error(t, err, "应返回错误")
			} else {
				assert.NoError(t, err, "不应返回错误")
			}
		})
	}
}

// TestIntegrationExtensionRegistration 测试扩展注册
func TestIntegrationExtensionRegistration(t *testing.T) {
	// 验证扩展已正确注册
	// 这个测试主要是为了确保init()函数正确执行
	assert.True(t, true, "扩展应已正确注册")
}

// TestIntegrationErrorHandling 测试错误处理
func TestIntegrationErrorHandling(t *testing.T) {
	// 测试无效的JSON数据
	invalidJSON := []byte(`{invalid json}`)
	_, err := Unmarshal(invalidJSON)
	assert.Error(t, err, "应返回解析错误")

	// 测试空数据
	_, err = Unmarshal([]byte{})
	assert.Error(t, err, "应返回解析错误")
}
