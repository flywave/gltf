package tile3d

import (
	"encoding/json"
	"reflect"
	"testing"

	extgltf "github.com/flywave/gltf/ext/3dtile/gltf"
	"github.com/flywave/go3d/mat3"
	"github.com/flywave/go3d/mat4"
	"github.com/stretchr/testify/assert"
)

func TestCreateInlinePropertyTableProperty(t *testing.T) {
	// 测试允许内联的类型
	tests := []struct {
		name        string
		value       interface{}
		expectError bool
	}{
		{"string", "test", false},
		{"bool", true, false},
		{"float32", float32(1.0), false},
		{"float64", float64(1.0), false},
		{"int", 1, false},
		{"int32", int32(1), false},
		{"uint32", uint32(1), false},
		{"not allowed", []int{1, 2, 3}, true}, // 不允许的类型
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prop, err := CreateInlinePropertyTableProperty(tt.value)
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, prop)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, prop)
				// 验证Extras字段包含_inlineValue
				var extras map[string]interface{}
				err := json.Unmarshal(prop.Extras, &extras)
				assert.NoError(t, err)
				assert.Contains(t, extras, "_inlineValue")

				// 注意：JSON反序列化会将数字类型转换为float64
				if expected, ok := tt.value.(int); ok {
					assert.Equal(t, float64(expected), extras["_inlineValue"])
				} else if expected, ok := tt.value.(int32); ok {
					assert.Equal(t, float64(expected), extras["_inlineValue"])
				} else if expected, ok := tt.value.(uint32); ok {
					assert.Equal(t, float64(expected), extras["_inlineValue"])
				} else if expected, ok := tt.value.(float32); ok {
					assert.Equal(t, float64(expected), extras["_inlineValue"])
				} else {
					assert.Equal(t, tt.value, extras["_inlineValue"])
				}
			}
		})
	}
}

func TestPaddingByte(t *testing.T) {
	tests := []struct {
		size     int
		expected int
	}{
		{0, 0},
		{1, 3},
		{2, 2},
		{3, 1},
		{4, 0},
		{5, 3},
		{8, 0},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			result := PaddingByte(tt.size)
			assert.Len(t, result, tt.expected)
			// 验证所有填充字节都是0x00
			for _, b := range result {
				assert.Equal(t, byte(0x00), b)
			}
		})
	}
}

func TestConvertType(t *testing.T) {
	tests := []struct {
		name         string
		value        interface{}
		targetType   reflect.Type
		defaultValue interface{}
		expected     interface{}
	}{
		{"float32 to float64", float32(1.5), reflect.TypeOf(float64(0)), float64(0), float64(1.5)},
		{"int to float64", 5, reflect.TypeOf(float64(0)), float64(0), float64(5)},
		{"float64 to int", float64(3.7), reflect.TypeOf(int(0)), 0, 3},
		{"string to string", "test", reflect.TypeOf(""), "", "test"},
		{"nil value", nil, reflect.TypeOf(""), "default", "default"},
		{"unconvertible", "not a number", reflect.TypeOf(int(0)), -1, -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertType(tt.value, tt.targetType, tt.defaultValue)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestInferPropertyType(t *testing.T) {
	tests := []struct {
		name          string
		values        interface{}
		expectedType  extgltf.ClassPropertyType
		expectedComp  *extgltf.ClassPropertyComponentType
		expectedArray bool
		expectError   bool
	}{
		// 基本类型数组测试 (注意：这些测试的是数组元素的类型，不是数组本身)
		{"string", []string{"a", "b"}, extgltf.ClassPropertyTypeString, nil, false, false},
		{"bool", []bool{true, false}, extgltf.ClassPropertyTypeBoolean, nil, false, false},
		{"float32", []float32{1.0, 2.0}, extgltf.ClassPropertyTypeScalar, ptr(extgltf.ClassPropertyComponentTypeFloat32), false, false},
		{"float64", []float64{1.0, 2.0}, extgltf.ClassPropertyTypeScalar, ptr(extgltf.ClassPropertyComponentTypeFloat64), false, false},
		{"int", []int{1, 2}, extgltf.ClassPropertyTypeScalar, ptr(extgltf.ClassPropertyComponentTypeInt64), false, false},

		// 数组类型测试
		{"[]string", [][]string{{"a"}, {"b"}}, extgltf.ClassPropertyTypeString, nil, true, false},
		{"[]bool", [][]bool{{true}, {false}}, extgltf.ClassPropertyTypeBoolean, nil, true, false},

		// 向量和矩阵测试
		{"[][]float32 vec3", [][]float32{{1, 2, 3}, {4, 5, 6}}, extgltf.ClassPropertyTypeVec3, ptr(extgltf.ClassPropertyComponentTypeFloat32), false, false},
		{"[][]float32 vec4", [][]float32{{1, 2, 3, 4}, {5, 6, 7, 8}}, extgltf.ClassPropertyTypeVec4, ptr(extgltf.ClassPropertyComponentTypeFloat32), false, false},
		{"[]mat3.T", []mat3.T{{{1, 2, 3}, {4, 5, 6}, {7, 8, 9}}}, extgltf.ClassPropertyTypeMat3, ptr(extgltf.ClassPropertyComponentTypeFloat32), false, false},
		{"[]mat4.T", []mat4.T{{{1, 2, 3, 4}, {5, 6, 7, 8}, {9, 10, 11, 12}, {13, 14, 15, 16}}}, extgltf.ClassPropertyTypeMat4, ptr(extgltf.ClassPropertyComponentTypeFloat32), false, false},
		{"unsupported", [][]complex64{{1, 2}}, "", nil, false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			propType, compType, isArray, err := inferPropertyType(tt.values)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedType, propType)
				assert.Equal(t, tt.expectedArray, isArray)
				if tt.expectedComp != nil {
					assert.Equal(t, *tt.expectedComp, *compType)
				} else {
					assert.Nil(t, compType)
				}
			}
		})
	}
}

func TestRackProps(t *testing.T) {
	// 测试基本功能
	props := []map[string]interface{}{
		{
			"name":  "item1",
			"value": 10,
			"flag":  true,
		},
		{
			"name":  "item2",
			"value": 20,
			"flag":  false,
		},
	}

	result := rackProps(props)

	// 验证结果
	assert.Contains(t, result, "name")
	assert.Contains(t, result, "value")
	assert.Contains(t, result, "flag")

	// 验证类型
	names, ok := result["name"].([]string)
	assert.True(t, ok)
	assert.Equal(t, []string{"item1", "item2"}, names)

	values, ok := result["value"].([]int)
	assert.True(t, ok)
	assert.Equal(t, []int{10, 20}, values)

	flags, ok := result["flag"].([]bool)
	assert.True(t, ok)
	assert.Equal(t, []bool{true, false}, flags)
}

func TestGetUnderlyingType(t *testing.T) {
	tests := []struct {
		name     string
		value    interface{}
		expected reflect.Type
	}{
		{"string", "test", reflect.TypeOf("")},
		{"int", 1, reflect.TypeOf(0)},
		{"float64", 1.5, reflect.TypeOf(float64(0))},
		{"bool", true, reflect.TypeOf(false)},
		{"nil", nil, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetUnderlyingType(tt.value)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCreateDefaultValue(t *testing.T) {
	tests := []struct {
		name     string
		typ      reflect.Type
		expected interface{}
	}{
		{"string", reflect.TypeOf(""), ""},
		{"int", reflect.TypeOf(0), 0},
		{"float64", reflect.TypeOf(float64(0)), float64(0)},
		{"bool", reflect.TypeOf(false), false},
		{"nil", nil, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CreateDefaultValue(tt.typ)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestUnmarshalExtension(t *testing.T) {
	// 创建一个测试结构体
	type TestExtension struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}

	// 创建原始数据
	original := map[string]interface{}{
		"name":  "test",
		"value": 42,
	}

	// 测试反序列化
	var target TestExtension
	err := unmarshalExtension(original, &target)
	assert.NoError(t, err)
	assert.Equal(t, "test", target.Name)
	assert.Equal(t, 42, target.Value)

	// 测试错误情况
	err = unmarshalExtension(make(chan int), &target) // 不可序列化的类型
	assert.Error(t, err)
}
