package splatting

import (
	"testing"

	"github.com/flywave/gltf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnmarshalGaussianSplatting(t *testing.T) {
	// 测试正常情况
	jsonData := []byte(`{}`)
	ext, err := UnmarshalGaussianSplatting(jsonData)
	require.NoError(t, err, "解组应成功")

	gs, ok := ext.(*GaussianSplatting)
	require.True(t, ok, "应能转换为GaussianSplatting类型")
	assert.NotNil(t, gs, "GaussianSplatting实例不应为nil")
}

func TestUnmarshalGaussianSplattingInvalidJSON(t *testing.T) {
	// 测试无效JSON
	invalidJSON := []byte(`{invalid json}`)
	_, err := UnmarshalGaussianSplatting(invalidJSON)
	assert.Error(t, err, "应返回错误")
}

func TestCreateGaussianPrimitive(t *testing.T) {
	doc := &gltf.Document{}
	attrs := map[string]uint32{
		"POSITION": 0,
		"COLOR_0":  1,
	}

	gs := CreateGaussianPrimitive(doc, attrs)
	assert.NotNil(t, gs, "应创建GaussianSplatting实例")

	// 验证扩展是否添加到ExtensionsUsed
	assert.Contains(t, doc.ExtensionsUsed, ExtensionName, "应添加扩展到ExtensionsUsed")
}

func TestValidateRotation(t *testing.T) {
	// 测试有效的单位四元数
	validRotations := []float32{1.0, 0.0, 0.0, 0.0} // 单位四元数
	err := ValidateRotation(validRotations)
	assert.NoError(t, err, "有效的单位四元数应通过验证")

	// 测试无效的四元数
	invalidRotations := []float32{2.0, 0.0, 0.0, 0.0} // 长度不为1
	err = ValidateRotation(invalidRotations)
	assert.Error(t, err, "无效的四元数应返回错误")
}

func TestClamp(t *testing.T) {
	tests := []struct {
		value    float32
		min      float32
		max      float32
		expected float32
	}{
		{0.5, 0.0, 1.0, 0.5},  // 在范围内
		{-1.0, 0.0, 1.0, 0.0}, // 低于最小值
		{2.0, 0.0, 1.0, 1.0},  // 高于最大值
	}

	for _, tt := range tests {
		result := clamp(tt.value, tt.min, tt.max)
		assert.Equal(t, tt.expected, result, "clamp结果应匹配")
	}
}

func TestComponentSize(t *testing.T) {
	tests := []struct {
		componentType gltf.ComponentType
		expectedSize  int
	}{
		{gltf.ComponentUbyte, 1},
		{gltf.ComponentByte, 1},
		{gltf.ComponentUshort, 2},
		{gltf.ComponentShort, 2},
		{gltf.ComponentUint, 4},
		{gltf.ComponentFloat, 4},
		{gltf.ComponentFloat, 4}, // 默认情况
	}

	for _, tt := range tests {
		result := componentSize(tt.componentType)
		assert.Equal(t, tt.expectedSize, result, "组件大小应匹配")
	}
}

func TestWireGaussianSplattingValidation(t *testing.T) {
	doc := &gltf.Document{}

	// 测试nil输入
	_, err := WireGaussianSplatting(doc, nil, false)
	assert.Error(t, err, "nil输入应返回错误")

	// 测试无效的位置数据
	invalidVertexData := &VertexData{
		Positions: []float32{1.0, 2.0}, // 不是3的倍数
		Colors:    []float32{1.0, 1.0, 1.0, 1.0},
		Scales:    []float32{1.0, 1.0, 1.0},
		Rotations: []float32{1.0, 0.0, 0.0, 0.0},
	}
	_, err = WireGaussianSplatting(doc, invalidVertexData, false)
	assert.Error(t, err, "无效位置数据应返回错误")

	// 测试不匹配的属性长度
	mismatchedVertexData := &VertexData{
		Positions: []float32{0.0, 0.0, 0.0, 1.0, 1.0, 1.0}, // 2个顶点
		Colors:    []float32{1.0, 1.0, 1.0, 1.0},           // 1个顶点
		Scales:    []float32{1.0, 1.0, 1.0},                // 1个顶点
		Rotations: []float32{1.0, 0.0, 0.0, 0.0},           // 1个顶点
	}
	_, err = WireGaussianSplatting(doc, mismatchedVertexData, false)
	assert.Error(t, err, "不匹配的属性长度应返回错误")
}

func TestProcessByteComponents(t *testing.T) {
	// 创建测试数据
	buffer := make([]byte, 12)
	for i := 0; i < 12; i++ {
		buffer[i] = byte(i * 10)
	}

	out := make([]float32, 8)
	processByteComponents(buffer, 0, 3, gltf.ComponentUbyte, true, 2, 4, out)

	// 验证结果
	expected := []float32{0.0, 10.0 / 255.0, 30.0 / 255.0, 40.0 / 255.0, 60.0 / 255.0, 70.0 / 255.0, 90.0 / 255.0, 100.0 / 255.0}
	for i, exp := range expected {
		assert.InDelta(t, exp, out[i], 1e-6, "字节组件处理结果应匹配")
	}
}

func TestProcessShortComponents(t *testing.T) {
	// 创建测试数据
	buffer := make([]byte, 16)
	for i := 0; i < 8; i++ {
		buffer[i*2] = byte(i * 10)
		buffer[i*2+1] = byte(i * 10 >> 8)
	}

	out := make([]float32, 4)
	processShortComponents(buffer, 0, 4, gltf.ComponentUshort, true, 1, 4, out)

	// 验证结果
	for i := 0; i < 4; i++ {
		expected := float32(i*10) / 65535.0
		assert.InDelta(t, expected, out[i], 1e-6, "短整型组件处理结果应匹配")
	}
}
