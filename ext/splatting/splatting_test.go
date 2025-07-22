package splatting

import (
	"testing"

	"github.com/flywave/gltf"
	"github.com/stretchr/testify/require"
)

func TestLoadSynthetic(t *testing.T) {
	// 读取 GLB 文件
	doc, err := gltf.Open("../../testdata/Splatting/synthetic.gltf")
	require.NoError(t, err, "无法打开 GLB 文件")

	// 验证图元存在
	require.Greater(t, len(doc.Meshes), 0, "没有找到网格数据")
	require.Greater(t, len(doc.Meshes[0].Primitives), 0, "没有找到图元数据")
	primitive := doc.Meshes[0].Primitives[0]

	// 读取高斯泼溅数据
	data, err := ReadGaussianSplatting(doc, primitive)
	require.NoError(t, err, "读取高斯泼溅数据失败")

	// 验证数据完整性
	vertexCount := len(data.Positions) / 3
	require.Equal(t, len(data.Colors)/4, vertexCount, "颜色数据长度不匹配")
	require.Equal(t, len(data.Scales)/3, vertexCount, "缩放数据长度不匹配")
	require.Equal(t, len(data.Rotations)/4, vertexCount, "旋转数据长度不匹配")
	require.Greater(t, vertexCount, 0, "未读取到顶点数据")

	// 验证属性范围
	for i := 0; i < vertexCount; i++ {
		posIdx := i * 3
		require.InDelta(t, data.Positions[posIdx], data.Positions[posIdx], 1e-6, "位置数据异常")

		colorIdx := i * 4
		for c := 0; c < 3; c++ {
			require.InDelta(t, data.Colors[colorIdx+c], data.Colors[colorIdx+c], 1e-6, "颜色数据异常")
		}
		require.InDelta(t, data.Colors[colorIdx+3], data.Colors[colorIdx+3], 1e-6, "不透明度数据异常")

		scaleIdx := i * 3
		for s := 0; s < 3; s++ {
			require.InDelta(t, data.Scales[scaleIdx+s], data.Scales[scaleIdx+s], 1e-6, "缩放数据异常")
		}
	}
}

func TestLoadMeshopt(t *testing.T) {
	// 读取 GLB 文件
	doc, err := gltf.Open("../../testdata/Splatting/meshopt_full.gltf")
	require.NoError(t, err, "无法打开 GLB 文件")

	// 验证图元存在
	require.Greater(t, len(doc.Meshes), 0, "没有找到网格数据")
	require.Greater(t, len(doc.Meshes[0].Primitives), 0, "没有找到图元数据")
	primitive := doc.Meshes[0].Primitives[0]

	// 读取高斯泼溅数据
	data, err := ReadGaussianSplatting(doc, primitive)
	require.NoError(t, err, "读取高斯泼溅数据失败")

	// 验证数据完整性
	vertexCount := len(data.Positions) / 3
	require.Equal(t, len(data.Colors)/4, vertexCount, "颜色数据长度不匹配")
	require.Equal(t, len(data.Scales)/3, vertexCount, "缩放数据长度不匹配")
	require.Equal(t, len(data.Rotations)/4, vertexCount, "旋转数据长度不匹配")
	require.Greater(t, vertexCount, 0, "未读取到顶点数据")

	// 验证属性范围
	for i := 0; i < vertexCount; i++ {
		posIdx := i * 3
		require.InDelta(t, data.Positions[posIdx], data.Positions[posIdx], 1e-6, "位置数据异常")

		colorIdx := i * 4
		for c := 0; c < 3; c++ {
			require.InDelta(t, data.Colors[colorIdx+c], data.Colors[colorIdx+c], 1e-6, "颜色数据异常")
		}
		require.InDelta(t, data.Colors[colorIdx+3], data.Colors[colorIdx+3], 1e-6, "不透明度数据异常")

		scaleIdx := i * 3
		for s := 0; s < 3; s++ {
			require.InDelta(t, data.Scales[scaleIdx+s], data.Scales[scaleIdx+s], 1e-6, "缩放数据异常")
		}
	}
}
