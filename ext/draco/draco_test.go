package draco

import (
	"testing"

	"github.com/flywave/gltf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDracoExtensionUnmarshal(t *testing.T) {
	// 测试正常情况
	jsonData := []byte(`{
		"bufferView": 5,
		"attributes": {
			"POSITION": 0,
			"NORMAL": 1,
			"TEXCOORD_0": 2
		}
	}`)

	ext, err := Unmarshal(jsonData)
	require.NoError(t, err, "解组应成功")

	dracoExt, ok := ext.(*DracoExtension)
	require.True(t, ok, "应能转换为DracoExtension类型")
	assert.Equal(t, uint32(5), dracoExt.BufferView, "BufferView应正确解析")
	assert.Equal(t, 3, len(dracoExt.Attributes), "应有3个属性")
	assert.Equal(t, uint32(0), dracoExt.Attributes["POSITION"], "POSITION属性ID应为0")
	assert.Equal(t, uint32(1), dracoExt.Attributes["NORMAL"], "NORMAL属性ID应为1")
	assert.Equal(t, uint32(2), dracoExt.Attributes["TEXCOORD_0"], "TEXCOORD_0属性ID应为2")
}

func TestDracoExtensionUnmarshalInvalidJSON(t *testing.T) {
	// 测试无效JSON
	invalidJSON := []byte(`{invalid json}`)
	_, err := Unmarshal(invalidJSON)
	assert.Error(t, err, "应返回错误")
}

func TestDecodePrimitiveWithoutDracoExtension(t *testing.T) {
	// 测试没有Draco扩展的图元
	doc := &gltf.Document{}
	primitive := &gltf.Primitive{
		Extensions: make(map[string]interface{}),
	}

	err := decodePrimitive(doc, primitive)
	assert.NoError(t, err, "应无错误返回")
}

func TestDecodePrimitiveInvalidExtensionFormat(t *testing.T) {
	// 测试无效扩展格式
	doc := &gltf.Document{}
	primitive := &gltf.Primitive{
		Extensions: map[string]interface{}{
			ExtensionName: "invalid format",
		},
	}

	err := decodePrimitive(doc, primitive)
	assert.Error(t, err, "应返回错误")
}

func TestComponentsPerType(t *testing.T) {
	tests := []struct {
		accType    gltf.AccessorType
		components uint32
	}{
		{gltf.AccessorScalar, 1},
		{gltf.AccessorVec2, 2},
		{gltf.AccessorVec3, 3},
		{gltf.AccessorVec4, 4},
		{gltf.AccessorMat2, 0}, // 不支持的类型
		{gltf.AccessorMat3, 0}, // 不支持的类型
		{gltf.AccessorMat4, 0}, // 不支持的类型
	}

	for _, tt := range tests {
		result := componentsPerType(tt.accType)
		assert.Equal(t, tt.components, result, "组件数量应匹配")
	}
}

func TestGetAccessorType(t *testing.T) {
	tests := []struct {
		attrName     string
		expectedType gltf.AccessorType
		expectedComp gltf.ComponentType
	}{
		{"POSITION", gltf.AccessorVec3, gltf.ComponentFloat},
		{"NORMAL", gltf.AccessorVec3, gltf.ComponentFloat},
		{"TEXCOORD_0", gltf.AccessorVec2, gltf.ComponentFloat},
		{"TEXCOORD_1", gltf.AccessorVec2, gltf.ComponentFloat},
		{"COLOR_0", gltf.AccessorVec4, gltf.ComponentFloat},
		{"TANGENT", gltf.AccessorVec4, gltf.ComponentFloat},
		{"JOINTS_0", gltf.AccessorVec4, gltf.ComponentUshort},
		{"WEIGHTS_0", gltf.AccessorVec4, gltf.ComponentFloat},
		{"UNKNOWN", gltf.AccessorVec3, gltf.ComponentFloat}, // 默认值
	}

	for _, tt := range tests {
		accType, compType := getAccessorType(tt.attrName)
		assert.Equal(t, tt.expectedType, accType, "访问器类型应匹配")
		assert.Equal(t, tt.expectedComp, compType, "组件类型应匹配")
	}
}

func TestFloat32ToBytes(t *testing.T) {
	data := []float32{1.0, 2.0, 3.0, 4.0}
	bytes := float32ToBytes(data)
	assert.NotNil(t, bytes, "应返回字节数据")
	assert.Equal(t, 16, len(bytes), "应有16个字节(4个float32)")
}

func TestIndicesToBytes(t *testing.T) {
	indices := []uint32{0, 1, 2, 3, 4, 5}

	// 测试ubyte类型
	bytes := indicesToBytes(indices, gltf.ComponentUbyte)
	assert.NotNil(t, bytes, "应返回字节数据")
	assert.Equal(t, 6, len(bytes), "应有6个字节")

	// 测试ushort类型
	bytes = indicesToBytes(indices, gltf.ComponentUshort)
	assert.NotNil(t, bytes, "应返回字节数据")
	assert.Equal(t, 12, len(bytes), "应有12个字节")

	// 测试uint类型
	bytes = indicesToBytes(indices, gltf.ComponentUint)
	assert.NotNil(t, bytes, "应返回字节数据")
	assert.Equal(t, 24, len(bytes), "应有24个字节")
}

func TestDracoAttributeType(t *testing.T) {
	tests := []struct {
		attrName string
		expected int32
	}{
		{"POSITION", 0},   // GAT_POSITION
		{"NORMAL", 1},     // GAT_NORMAL
		{"TEXCOORD_0", 3}, // GAT_TEX_COORD
		{"COLOR_0", 2},    // GAT_COLOR
		{"UNKNOWN", 4},    // GAT_GENERIC
	}

	for _, tt := range tests {
		result := dracoAttributeType(tt.attrName)
		assert.Equal(t, tt.expected, int32(result), "属性类型应匹配")
	}
}

func TestPaddingBytes(t *testing.T) {
	tests := []struct {
		size     int
		expected int
	}{
		{0, 0}, // 0 % 4 = 0
		{1, 3}, // 1 % 4 = 1, 4-1 = 3
		{2, 2}, // 2 % 4 = 2, 4-2 = 2
		{3, 1}, // 3 % 4 = 3, 4-3 = 1
		{4, 0}, // 4 % 4 = 0
		{5, 3}, // 5 % 4 = 1, 4-1 = 3
	}

	for _, tt := range tests {
		result := paddingBytes(tt.size)
		assert.Equal(t, tt.expected, result, "填充字节数应匹配")
	}
}

func TestCalculateMinMax(t *testing.T) {
	// 测试VEC3数据
	data := []float32{
		1.0, 2.0, 3.0, // 第一个点
		4.0, 5.0, 6.0, // 第二个点
		-1.0, -2.0, -3.0, // 第三个点
		7.0, 8.0, 9.0, // 第四个点
	}

	min, max := calculateMinMax(data, 3)
	assert.Equal(t, []float32{-1.0, -2.0, -3.0}, min, "最小值应正确")
	assert.Equal(t, []float32{7.0, 8.0, 9.0}, max, "最大值应正确")

	// 测试VEC2数据
	data2 := []float32{
		1.0, 2.0, // 第一个点
		4.0, 5.0, // 第二个点
		-1.0, -2.0, // 第三个点
		7.0, 8.0, // 第四个点
	}

	min2, max2 := calculateMinMax(data2, 2)
	assert.Equal(t, []float32{-1.0, -2.0}, min2, "最小值应正确")
	assert.Equal(t, []float32{7.0, 8.0}, max2, "最大值应正确")
}

func TestEncodePrimitiveMissingData(t *testing.T) {
	doc := &gltf.Document{}
	primitive := &gltf.Primitive{}

	// 测试缺少索引
	err := encodePrimitive(doc, nil, primitive, nil)
	assert.Error(t, err, "应返回错误")
	assert.Contains(t, err.Error(), "图元缺少索引", "错误信息应包含索引缺失")

	// 测试缺少位置属性
	primitive.Indices = gltf.Index(0)
	err = encodePrimitive(doc, nil, primitive, nil)
	assert.Error(t, err, "应返回错误")
	assert.Contains(t, err.Error(), "缺少位置属性", "错误信息应包含位置属性缺失")
}

func TestCleanUpUnusedResources(t *testing.T) {
	// 创建测试文档
	doc := &gltf.Document{
		Buffers: []*gltf.Buffer{
			{Data: []byte("buffer1")},
			{Data: []byte("buffer2")},
		},
		BufferViews: []*gltf.BufferView{
			{Buffer: 0, ByteLength: 10},
			{Buffer: 1, ByteLength: 10},
		},
		Accessors: []*gltf.Accessor{
			{BufferView: gltf.Index(0)},
			{BufferView: nil},
		},
		Meshes: []*gltf.Mesh{
			{
				Primitives: []*gltf.Primitive{
					{
						Extensions: map[string]interface{}{
							ExtensionName: &DracoExtension{
								BufferView: 1,
							},
						},
					},
				},
			},
		},
	}

	// 执行清理
	cleanUpUnusedResources(doc)

	// 验证结果
	assert.Equal(t, 2, len(doc.Buffers), "应保留所有缓冲区")
	assert.Equal(t, 2, len(doc.BufferViews), "应保留所有缓冲区视图")
}

func TestBug(t *testing.T) {
	doc, err := gltf.Open("../../testdata/Draco/debug.glb")
	require.NoError(t, err)

	err = EncodeAll(doc, nil)
	assert.NoError(t, err, "编码应成功")
	gltf.SaveBinary(doc, "./bug.glb")
}
