package tile3d

import (
	"testing"

	"github.com/flywave/gltf"
	extmesh "github.com/flywave/gltf/ext/3dtile/mesh"
	"github.com/stretchr/testify/assert"
)

func TestMeshFeaturesManager_AddMeshFeatures(t *testing.T) {
	// 创建测试图元
	primitive := &gltf.Primitive{
		Extensions: make(gltf.Extensions),
	}

	// 创建管理器
	manager := NewMeshFeaturesManager()

	// 测试数据
	featureIDs := []extmesh.FeatureID{
		{
			FeatureCount: 10,
			Attribute:    gltf.Index(0),
		},
	}

	// 添加网格特征
	err := manager.AddMeshFeatures(primitive, featureIDs)
	assert.NoError(t, err)

	// 验证扩展已添加
	extData, exists := primitive.Extensions[extmesh.ExtensionName]
	assert.True(t, exists)
	assert.NotNil(t, extData)
}

func TestMeshFeaturesManager_CreateFeatureID(t *testing.T) {
	// 创建管理器
	manager := NewMeshFeaturesManager()

	// 创建特征ID
	featureID := manager.CreateFeatureID(10, WithLabel("testLabel"))

	// 验证特征ID
	assert.Equal(t, uint32(10), featureID.FeatureCount)
	assert.NotNil(t, featureID.Label)
	assert.Equal(t, "testLabel", *featureID.Label)
}

func TestMeshFeaturesManager_ValidateFeatureID(t *testing.T) {
	// 创建管理器
	manager := NewMeshFeaturesManager()

	// 测试有效的特征ID
	validFeatureID := extmesh.FeatureID{
		FeatureCount: 10,
		Attribute:    gltf.Index(0),
	}
	err := manager.ValidateFeatureID(validFeatureID)
	assert.NoError(t, err)

	// 测试无效的特征ID（没有引用方法）
	invalidFeatureID := extmesh.FeatureID{
		FeatureCount: 10,
	}
	err = manager.ValidateFeatureID(invalidFeatureID)
	assert.Error(t, err)

	// 测试无效的特征ID（多个引用方法）
	multipleFeatureID := extmesh.FeatureID{
		FeatureCount:  10,
		Attribute:     gltf.Index(0),
		PropertyTable: gltf.Index(0),
	}
	err = manager.ValidateFeatureID(multipleFeatureID)
	assert.Error(t, err)
}

func TestUnmarshalMeshFeatures(t *testing.T) {
	// 测试有效的网格特征数据
	validData := `{"featureIds": [{"featureCount": 10, "attribute": 0}]}`
	result, err := UnmarshalMeshFeatures([]byte(validData))
	assert.NoError(t, err)
	assert.NotNil(t, result)

	ext, ok := result.(extmesh.ExtMeshFeatures)
	assert.True(t, ok)
	assert.Len(t, ext.FeatureIDs, 1)
	assert.Equal(t, uint32(10), ext.FeatureIDs[0].FeatureCount)
	assert.Equal(t, uint32(0), *ext.FeatureIDs[0].Attribute)

	// 测试无效的网格特征数据（缺少引用方法）
	invalidData := `{"featureIds": [{"featureCount": 10}]}`
	_, err = UnmarshalMeshFeatures([]byte(invalidData))
	assert.Error(t, err)
}
