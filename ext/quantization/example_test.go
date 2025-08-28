package quantization

import (
	"encoding/binary"
	"fmt"
	"math"

	"github.com/flywave/gltf"
)

// Example_quantization demonstrates how to use the KHR_mesh_quantization extension
// to quantize mesh attributes for more efficient storage and transmission.
func Example_quantization() {
	// Create a simple glTF document with a mesh
	doc := &gltf.Document{
		Asset: gltf.Asset{
			Version:   "2.0",
			Generator: "Example Generator",
		},
		Buffers: []*gltf.Buffer{
			{
				ByteLength: 48, // 4 vertices * 3 components * 4 bytes per float
				Data:       make([]byte, 48),
			},
		},
		BufferViews: []*gltf.BufferView{
			{
				Buffer:     0,
				ByteOffset: 0,
				ByteLength: 48,
			},
		},
		Accessors: []*gltf.Accessor{
			{
				BufferView:    gltf.Index(0),
				ByteOffset:    0,
				ComponentType: gltf.ComponentFloat,
				Count:         4,
				Type:          gltf.AccessorVec3,
				Min:           []float32{0.0, 0.0, 0.0},
				Max:           []float32{1.0, 1.0, 1.0},
			},
		},
		Meshes: []*gltf.Mesh{
			{
				Name: "QuantizedMesh",
				Primitives: []*gltf.Primitive{
					{
						Attributes: map[string]uint32{
							"POSITION": 0,
						},
						Mode: gltf.PrimitiveTriangles,
					},
				},
			},
		},
	}

	// Fill buffer with some example data (4 vertices with positions)
	// In a real application, this would be your actual mesh data
	buffer := doc.Buffers[0].Data
	offset := 0
	for i := 0; i < 4; i++ {
		for j := 0; j < 3; j++ {
			value := float32(i*3 + j)
			binary.LittleEndian.PutUint32(buffer[offset:], math.Float32bits(value))
			offset += 4
		}
	}

	// Create a quantizer with default configuration
	// This will quantize positions to 12 bits
	quantizer := NewQuantizer(doc, nil)

	// Process the document to apply quantization
	if err := quantizer.Process(); err != nil {
		fmt.Printf("Error quantizing document: %v\n", err)
		return
	}

	// Check that the extension was added
	hasExtension := false
	for _, ext := range doc.ExtensionsUsed {
		if ext == ExtensionName {
			hasExtension = true
			break
		}
	}

	if hasExtension {
		fmt.Println("Document successfully quantized with KHR_mesh_quantization extension")
	} else {
		fmt.Println("Failed to add KHR_mesh_quantization extension")
	}

	// Create a dequantizer to reverse the process
	dequantizer := NewDequantizer(doc)

	// Process the document to remove quantization
	if err := dequantizer.Process(); err != nil {
		fmt.Printf("Error dequantizing document: %v\n", err)
		return
	}

	// Check that the extension was removed
	hasExtension = false
	for _, ext := range doc.ExtensionsUsed {
		if ext == ExtensionName {
			hasExtension = true
			break
		}
	}

	if !hasExtension {
		fmt.Println("Document successfully dequantized and extension removed")
	} else {
		fmt.Println("Failed to remove KHR_mesh_quantization extension")
	}

	// Output:
	// Document successfully quantized with KHR_mesh_quantization extension
	// Document successfully dequantized and extension removed
}
