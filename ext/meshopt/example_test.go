package meshopt

import (
	"fmt"
	"log"

	"github.com/flywave/gltf"
)

// Example_basic_usage 演示meshopt扩展的基本用法
func Example_basic_usage() {
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

	// 尝试解码所有压缩数据（在这个简单例子中不会有任何效果）
	err := DecodeAll(doc)
	if err != nil {
		log.Printf("解码失败: %v", err)
		return
	}

	fmt.Println("meshopt扩展已正确集成")
	// Output: meshopt扩展已正确集成
}

// Example_extension_unmarshal 演示扩展的反序列化
func Example_extension_unmarshal() {
	// 示例JSON数据
	jsonData := []byte(`{
		"buffer": 1,
		"byteOffset": 100,
		"byteLength": 200,
		"byteStride": 12,
		"count": 50,
		"mode": "ATTRIBUTES",
		"filter": "OCTAHEDRAL"
	}`)

	// 反序列化扩展数据
	ext, err := Unmarshal(jsonData)
	if err != nil {
		log.Printf("反序列化失败: %v", err)
		return
	}

	// 类型断言
	compressionExt, ok := ext.(*CompressionExtension)
	if !ok {
		log.Printf("类型转换失败")
		return
	}

	// 输出扩展信息
	fmt.Printf("Buffer: %d\n", compressionExt.Buffer)
	fmt.Printf("ByteOffset: %d\n", compressionExt.ByteOffset)
	fmt.Printf("ByteLength: %d\n", compressionExt.ByteLength)
	fmt.Printf("ByteStride: %d\n", compressionExt.ByteStride)
	fmt.Printf("Count: %d\n", compressionExt.Count)
	fmt.Printf("Mode: %s\n", compressionExt.Mode)
	fmt.Printf("Filter: %s\n", compressionExt.Filter)

	// Output:
	// Buffer: 1
	// ByteOffset: 100
	// ByteLength: 200
	// ByteStride: 12
	// Count: 50
	// Mode: ATTRIBUTES
	// Filter: OCTAHEDRAL
}

// Example_encode_decode 演示编码和解码过程
func Example_encode_decode() {
	// 创建一些测试数据（模拟顶点位置数据）
	// 使用重复的数据模式，这样更容易看到压缩效果
	data := make([]byte, 120) // 10个顶点，每个顶点3个float32（12字节）
	// 创建一些重复的模式，模拟实际的网格数据
	for i := 0; i < 10; i++ {
		// 每个顶点3个float32值（12字节）
		base := i * 12
		// X坐标
		data[base] = byte(i % 256)
		data[base+1] = 0
		data[base+2] = 0
		data[base+3] = 0x3F // 0.5的float32表示的一部分
		// Y坐标
		data[base+4] = byte((i * 2) % 256)
		data[base+5] = 0
		data[base+6] = 0
		data[base+7] = 0x40 // 1.0的float32表示的一部分
		// Z坐标
		data[base+8] = byte((i * 3) % 256)
		data[base+9] = 0
		data[base+10] = 0
		data[base+11] = 0x40 // 1.5的float32表示的一部分
	}

	// 编码数据
	compressedData, ext, err := MeshoptEncode(data, 10, 12, ModeAttributes, FilterNone)
	if err != nil {
		log.Printf("编码失败: %v", err)
		return
	}

	fmt.Printf("原始数据大小: %d 字节\n", len(data))
	fmt.Printf("压缩数据大小: %d 字节\n", len(compressedData))
	fmt.Printf("压缩扩展信息 - Count: %d, ByteStride: %d, Mode: %s\n",
		ext.Count, ext.ByteStride, ext.Mode)

	// 解码数据
	decompressedData, err := MeshoptDecode(ext.Count, ext.ByteStride, compressedData, ext.Mode, ext.Filter)
	if err != nil {
		log.Printf("解码失败: %v", err)
		return
	}

	fmt.Printf("解压数据大小: %d 字节\n", len(decompressedData))

	// 验证数据一致性
	if len(data) == len(decompressedData) {
		fmt.Println("数据大小一致")
	} else {
		fmt.Println("数据大小不一致")
	}

	// Output:
	// 原始数据大小: 120 字节
	// 压缩数据大小: 65 字节
	// 压缩扩展信息 - Count: 10, ByteStride: 12, Mode: ATTRIBUTES
	// 解压数据大小: 120 字节
	// 数据大小一致
}
