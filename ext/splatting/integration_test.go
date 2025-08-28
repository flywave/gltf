package splatting

import (
	"testing"

	"github.com/flywave/gltf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGaussianSplattingIntegration 测试高斯泼溅扩展的完整集成流程
func TestGaussianSplattingIntegration(t *testing.T) {
	// 1. 创建测试顶点数据
	vertexData := &VertexData{
		Positions: []float32{
			0.0, 0.0, 0.0,
			1.0, 0.0, 0.0,
			0.0, 1.0, 0.0,
			1.0, 1.0, 0.0,
		},
		Colors: []float32{
			1.0, 0.0, 0.0, 1.0,
			0.0, 1.0, 0.0, 1.0,
			0.0, 0.0, 1.0, 1.0,
			1.0, 1.0, 1.0, 1.0,
		},
		Scales: []float32{
			0.1, 0.1, 0.1,
			0.2, 0.2, 0.2,
			0.3, 0.3, 0.3,
			0.4, 0.4, 0.4,
		},
		Rotations: []float32{
			1.0, 0.0, 0.0, 0.0,
			1.0, 0.0, 0.0, 0.0,
			1.0, 0.0, 0.0, 0.0,
			1.0, 0.0, 0.0, 0.0,
		},
	}

	// 2. 创建glTF文档
	doc := &gltf.Document{
		Asset: gltf.Asset{
			Version: "2.0",
		},
		Buffers:     []*gltf.Buffer{},
		BufferViews: []*gltf.BufferView{},
		Accessors:   []*gltf.Accessor{},
		Meshes:      []*gltf.Mesh{},
	}

	// 3. 测试不压缩的情况
	gs, err := WireGaussianSplatting(doc, vertexData, false)
	require.NoError(t, err, "不压缩的高斯泼溅连接应成功")
	assert.NotNil(t, gs, "应创建GaussianSplatting实例")

	// 4. 验证文档结构
	assert.Equal(t, 1, len(doc.Meshes), "应创建一个网格")
	assert.Equal(t, 1, len(doc.Meshes[0].Primitives), "应创建一个图元")

	primitive := doc.Meshes[0].Primitives[0]
	assert.Equal(t, gltf.PrimitivePoints, primitive.Mode, "图元模式应为点")

	// 5. 验证扩展
	ext, exists := primitive.Extensions[ExtensionName]
	assert.True(t, exists, "图元应包含KHR_gaussian_splatting扩展")
	_, ok := ext.(*GaussianSplatting)
	assert.True(t, ok, "扩展应为GaussianSplatting类型")

	// 6. 验证属性
	assert.Contains(t, primitive.Attributes, "POSITION", "应包含POSITION属性")
	assert.Contains(t, primitive.Attributes, "COLOR_0", "应包含COLOR_0属性")
	assert.Contains(t, primitive.Attributes, "_SCALE", "应包含_SCALE属性")
	assert.Contains(t, primitive.Attributes, "_ROTATION", "应包含_ROTATION属性")

	// 7. 验证访问器
	posAccessorIdx := primitive.Attributes["POSITION"]
	assert.Less(t, int(posAccessorIdx), len(doc.Accessors), "POSITION访问器索引应有效")

	colorAccessorIdx := primitive.Attributes["COLOR_0"]
	assert.Less(t, int(colorAccessorIdx), len(doc.Accessors), "COLOR_0访问器索引应有效")

	// 8. 测试读取功能
	readData, err := ReadGaussianSplatting(doc, primitive)
	require.NoError(t, err, "读取高斯泼溅数据应成功")
	assert.NotNil(t, readData, "应读取到顶点数据")

	// 9. 验证读取的数据
	assert.Equal(t, len(vertexData.Positions), len(readData.Positions), "位置数据长度应匹配")
	assert.Equal(t, len(vertexData.Colors), len(readData.Colors), "颜色数据长度应匹配")
	assert.Equal(t, len(vertexData.Scales), len(readData.Scales), "缩放数据长度应匹配")
	assert.Equal(t, len(vertexData.Rotations), len(readData.Rotations), "旋转数据长度应匹配")
}

// TestGaussianSplattingWithCompression 测试带压缩的高斯泼溅扩展
func TestGaussianSplattingWithCompression(t *testing.T) {
	// 1. 创建测试顶点数据
	vertexData := &VertexData{
		Positions: []float32{
			0.0, 0.0, 0.0,
			1.0, 0.0, 0.0,
			0.0, 1.0, 0.0,
			1.0, 1.0, 0.0,
		},
		Colors: []float32{
			1.0, 0.0, 0.0, 1.0,
			0.0, 1.0, 0.0, 1.0,
			0.0, 0.0, 1.0, 1.0,
			1.0, 1.0, 1.0, 1.0,
		},
		Scales: []float32{
			0.1, 0.1, 0.1,
			0.2, 0.2, 0.2,
			0.3, 0.3, 0.3,
			0.4, 0.4, 0.4,
		},
		Rotations: []float32{
			1.0, 0.0, 0.0, 0.0,
			1.0, 0.0, 0.0, 0.0,
			1.0, 0.0, 0.0, 0.0,
			1.0, 0.0, 0.0, 0.0,
		},
	}

	// 2. 创建glTF文档
	doc := &gltf.Document{
		Asset: gltf.Asset{
			Version: "2.0",
		},
		Buffers:     []*gltf.Buffer{},
		BufferViews: []*gltf.BufferView{},
		Accessors:   []*gltf.Accessor{},
		Meshes:      []*gltf.Mesh{},
	}

	// 3. 测试压缩的情况
	gs, err := WireGaussianSplatting(doc, vertexData, true)
	require.NoError(t, err, "压缩的高斯泼溅连接应成功")
	assert.NotNil(t, gs, "应创建GaussianSplatting实例")

	// 4. 验证扩展是否添加
	assert.Contains(t, doc.ExtensionsUsed, "EXT_meshopt_compression", "应添加meshopt压缩扩展")
	assert.Contains(t, doc.ExtensionsUsed, "KHR_mesh_quantization", "应添加网格量化扩展")
}

// TestReadGaussianSplattingFromFiles 测试从实际文件读取高斯泼溅数据
func TestReadGaussianSplattingFromFiles(t *testing.T) {
	// 测试 synthetic.gltf (使用内嵌数据)
	doc, err := gltf.Open("../../testdata/Splatting/synthetic.gltf")
	require.NoError(t, err, "应能打开synthetic.gltf文件")

	require.Greater(t, len(doc.Meshes), 0, "应包含网格")
	require.Greater(t, len(doc.Meshes[0].Primitives), 0, "应包含图元")

	primitive := doc.Meshes[0].Primitives[0]
	data, err := ReadGaussianSplatting(doc, primitive)
	require.NoError(t, err, "应能读取高斯泼溅数据")
	assert.NotNil(t, data, "应读取到数据")

	// 验证数据完整性
	vertexCount := len(data.Positions) / 3
	assert.Greater(t, vertexCount, 0, "应有顶点数据")
	assert.Equal(t, len(data.Colors)/4, vertexCount, "颜色数据长度应匹配")
	assert.Equal(t, len(data.Scales)/3, vertexCount, "缩放数据长度应匹配")
	assert.Equal(t, len(data.Rotations)/4, vertexCount, "旋转数据长度应匹配")

	// 验证具体数据
	assert.Equal(t, 5, vertexCount, "应有5个顶点")
	assert.Equal(t, 15, len(data.Positions), "位置数据应有15个元素")
	assert.Equal(t, 20, len(data.Colors), "颜色数据应有20个元素")
	assert.Equal(t, 15, len(data.Scales), "缩放数据应有15个元素")
	assert.Equal(t, 20, len(data.Rotations), "旋转数据应有20个元素")
}

// TestValidateRotationIntegration 测试旋转验证集成
func TestValidateRotationIntegration(t *testing.T) {
	// 创建有效的四元数数据
	validRotations := []float32{
		1.0, 0.0, 0.0, 0.0, // 单位四元数
		0.707, 0.0, 0.0, 0.707, // 90度旋转
	}

	// 验证应通过
	err := ValidateRotation(validRotations)
	assert.NoError(t, err, "有效的四元数应通过验证")

	// 创建无效的四元数数据
	invalidRotations := []float32{
		2.0, 0.0, 0.0, 0.0, // 长度不为1
	}

	// 验证应失败
	err = ValidateRotation(invalidRotations)
	assert.Error(t, err, "无效的四元数应返回错误")
}
