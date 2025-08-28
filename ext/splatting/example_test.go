package splatting

import (
	"fmt"
	"os"

	"github.com/flywave/gltf"
)

// ExampleWireGaussianSplatting 演示如何创建一个高斯泼溅模型
func ExampleWireGaussianSplatting() {
	// 创建示例顶点数据
	vertexData := &VertexData{
		Positions: []float32{
			// 顶点1
			0.0, 0.0, 0.0,
			// 顶点2
			1.0, 0.0, 0.0,
			// 顶点3
			0.0, 1.0, 0.0,
		},
		Colors: []float32{
			// 顶点1 RGBA
			1.0, 0.0, 0.0, 1.0,
			// 顶点2 RGBA
			0.0, 1.0, 0.0, 1.0,
			// 顶点3 RGBA
			0.0, 0.0, 1.0, 1.0,
		},
		Scales: []float32{
			// 顶点1 XYZ
			0.1, 0.1, 0.1,
			// 顶点2 XYZ
			0.2, 0.2, 0.2,
			// 顶点3 XYZ
			0.3, 0.3, 0.3,
		},
		Rotations: []float32{
			// 顶点1 四元数 XYZW
			1.0, 0.0, 0.0, 0.0,
			// 顶点2 四元数 XYZW
			1.0, 0.0, 0.0, 0.0,
			// 顶点3 四元数 XYZW
			1.0, 0.0, 0.0, 0.0,
		},
	}

	// 创建glTF文档
	doc := &gltf.Document{
		Asset: gltf.Asset{
			Version: "2.0",
		},
	}

	// 连接高斯泼溅数据到glTF文档
	_, err := WireGaussianSplatting(doc, vertexData, false)
	if err != nil {
		fmt.Printf("创建高斯泼溅模型失败: %v\n", err)
		return
	}

	// 保存到文件
	err = gltf.Save(doc, "example.gltf")
	if err != nil {
		fmt.Printf("保存文件失败: %v\n", err)
		return
	}

	fmt.Println("成功创建高斯泼溅模型并保存为example.gltf")

	// 清理示例文件
	_ = os.Remove("example.gltf")
	_ = os.Remove("example.bin")

	// Output: 成功创建高斯泼溅模型并保存为example.gltf
}

// ExampleReadGaussianSplatting 演示如何从glTF文件读取高斯泼溅数据
func ExampleReadGaussianSplatting() {
	// 打开现有的glTF文件（这里使用测试数据）
	doc, err := gltf.Open("../../testdata/Splatting/synthetic.gltf")
	if err != nil {
		fmt.Printf("打开文件失败: %v\n", err)
		return
	}

	// 获取第一个网格的第一个图元
	if len(doc.Meshes) == 0 || len(doc.Meshes[0].Primitives) == 0 {
		fmt.Println("文件中没有网格或图元数据")
		return
	}

	primitive := doc.Meshes[0].Primitives[0]

	// 读取高斯泼溅数据
	data, err := ReadGaussianSplatting(doc, primitive)
	if err != nil {
		fmt.Printf("读取高斯泼溅数据失败: %v\n", err)
		return
	}

	// 输出数据信息
	vertexCount := len(data.Positions) / 3
	fmt.Printf("读取到 %d 个顶点的高斯泼溅数据\n", vertexCount)
	fmt.Printf("位置数据长度: %d\n", len(data.Positions))
	fmt.Printf("颜色数据长度: %d\n", len(data.Colors))
	fmt.Printf("缩放数据长度: %d\n", len(data.Scales))
	fmt.Printf("旋转数据长度: %d\n", len(data.Rotations))

	// Output:
	// 读取到 5 个顶点的高斯泼溅数据
	// 位置数据长度: 15
	// 颜色数据长度: 20
	// 缩放数据长度: 15
	// 旋转数据长度: 20
}
