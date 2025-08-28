package tile3d

import (
	"encoding/json"
	"testing"

	extgltf "github.com/flywave/gltf/ext/3dtile/gltf"

	"github.com/flywave/gltf"
	"github.com/stretchr/testify/assert"
)

func TestStructuralMetadataManager_AddPropertyTable(t *testing.T) {
	// 创建测试文档
	doc := &gltf.Document{
		Extensions: make(gltf.Extensions),
		Buffers:    []*gltf.Buffer{{}},
	}

	// 创建管理器
	manager := NewStructuralMetadataManager()

	// 测试数据
	properties := []PropertyData{
		{
			Name:          "testProperty",
			ElementType:   extgltf.ClassPropertyTypeScalar,
			ComponentType: extgltf.ClassPropertyComponentTypeFloat32,
			Values:        []float32{1.0, 2.0, 3.0},
		},
	}

	// 添加属性表
	tableIndex, err := manager.AddPropertyTable(doc, "TestClass", properties)
	assert.NoError(t, err)
	assert.Equal(t, 0, tableIndex)

	// 验证扩展已添加
	extData, exists := doc.Extensions[extgltf.ExtensionName]
	assert.True(t, exists)
	assert.NotNil(t, extData)

	// 验证属性表已创建
	extDataBytes, ok := extData.([]byte)
	assert.True(t, ok)

	var ext extgltf.ExtStructuralMetadata
	err = json.Unmarshal(extDataBytes, &ext)
	assert.NoError(t, err)
	assert.Len(t, ext.PropertyTables, 1)
	assert.Equal(t, "TestClass", ext.PropertyTables[0].Class)
	assert.Equal(t, uint32(3), ext.PropertyTables[0].Count)
}

func TestStructuralMetadataManager_DecodeProperty(t *testing.T) {
	// 创建测试文档
	doc := &gltf.Document{
		Extensions:  make(gltf.Extensions),
		Buffers:     []*gltf.Buffer{{}},
		BufferViews: []*gltf.BufferView{},
		Accessors:   []*gltf.Accessor{},
	}

	// 创建管理器
	manager := NewStructuralMetadataManager()

	// 测试数据
	properties := []PropertyData{
		{
			Name:          "testProperty",
			ElementType:   extgltf.ClassPropertyTypeScalar,
			ComponentType: extgltf.ClassPropertyComponentTypeFloat32,
			Values:        []float32{1.0, 2.0, 3.0},
		},
	}

	// 添加属性表
	_, err := manager.AddPropertyTable(doc, "TestClass", properties)
	assert.NoError(t, err)

	// 解码属性
	result, err := manager.DecodeProperty(doc, 0, "testProperty")
	assert.NoError(t, err)

	floatValues, ok := result.([]float32)
	assert.True(t, ok)
	assert.Len(t, floatValues, 3)
	assert.Equal(t, []float32{1.0, 2.0, 3.0}, floatValues)
}

func TestPropertyTableManager_CreatePropertyTable(t *testing.T) {
	// 创建测试文档
	doc := &gltf.Document{
		Buffers: []*gltf.Buffer{{}},
	}

	// 创建管理器
	manager := NewPropertyTableManager()

	// 测试数据
	properties := []PropertyData{
		{
			Name:          "testProperty",
			ElementType:   extgltf.ClassPropertyTypeScalar,
			ComponentType: extgltf.ClassPropertyComponentTypeFloat32,
			Values:        []float32{1.0, 2.0, 3.0},
		},
	}

	// 创建schema
	schema := &extgltf.Schema{
		Classes: map[string]extgltf.Class{
			"TestClass": {
				Properties: map[string]extgltf.ClassProperty{
					"testProperty": {
						Type:          extgltf.ClassPropertyTypeScalar,
						ComponentType: &[]extgltf.ClassPropertyComponentType{extgltf.ClassPropertyComponentTypeFloat32}[0],
					},
				},
			},
		},
	}

	// 创建属性表
	table, err := manager.createPropertyTable(doc, "TestClass", properties, schema)
	assert.NoError(t, err)
	assert.NotNil(t, table)
	assert.Equal(t, "TestClass", table.Class)
	assert.Equal(t, uint32(3), table.Count)
}
