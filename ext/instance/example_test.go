package instance

import (
	"fmt"

	"github.com/flywave/gltf"
)

func ExampleSetInstanceExtension() {
	// Create a new glTF document
	doc := gltf.NewDocument()

	// Create a node with a mesh
	node := &gltf.Node{
		Name:       "InstancedNode",
		Mesh:       gltf.Index(0),
		Extensions: make(gltf.Extensions),
	}
	doc.Nodes = append(doc.Nodes, node)

	// Define instance attributes (these would reference actual accessor indices)
	attributes := map[string]uint32{
		"TRANSLATION": 0,
		"ROTATION":    1,
		"SCALE":       2,
	}

	// Set the instance extension on the node
	err := SetInstanceExtension(node, attributes)
	if err != nil {
		panic(err)
	}

	// Add the extension to the document's required extensions
	doc.AddExtensionUsed(ExtensionName)

	fmt.Printf("Node '%s' has instancing extension with %d attributes\n", node.Name, len(attributes))
	// Output: Node 'InstancedNode' has instancing extension with 3 attributes
}

func ExampleGetInstanceExtension() {
	// Create a node with instance extension
	node := &gltf.Node{
		Name:       "InstancedNode",
		Extensions: make(gltf.Extensions),
	}

	attributes := map[string]uint32{
		"TRANSLATION": 0,
		"ROTATION":    1,
		"SCALE":       2,
		"_ID":         3, // Custom attribute
	}

	// Set the extension
	err := SetInstanceExtension(node, attributes)
	if err != nil {
		panic(err)
	}

	// Get the extension
	ext, err := GetInstanceExtension(node)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Node '%s' has instancing extension with attributes:\n", node.Name)
	// Print attributes in a consistent order
	names := []string{"TRANSLATION", "ROTATION", "SCALE", "_ID"}
	for _, name := range names {
		if accessorIdx, exists := ext.Attributes[name]; exists {
			fmt.Printf("  %s: %d\n", name, accessorIdx)
		}
	}

	// Output:
	// Node 'InstancedNode' has instancing extension with attributes:
	//   TRANSLATION: 0
	//   ROTATION: 1
	//   SCALE: 2
	//   _ID: 3
}

func ExampleWriteInstancing() {
	// Create a new glTF document
	doc := gltf.NewDocument()

	// Create instance data
	data := &InstanceData{
		Translations: [][3]float32{
			{0, 0, 0},
			{1, 0, 0},
			{0, 1, 0},
			{1, 1, 0},
		},
		Rotations: [][4]float32{
			{0, 0, 0, 1},
			{0, 0, 0, 1},
			{0, 0, 0, 1},
			{0, 0, 0, 1},
		},
		Scales: [][3]float32{
			{1, 1, 1},
			{1, 1, 1},
			{1, 1, 1},
			{1, 1, 1},
		},
	}

	// Use default configuration
	config := DefaultConfig()

	// Write instancing data to the document
	err := WriteInstancing(doc, data, config)
	if err != nil {
		panic(err)
	}

	// The extension is now added to the document
	extData, exists := doc.Extensions[ExtensionName]
	if !exists {
		panic("Extension was not added to document")
	}

	// Parse the extension data to show what was created
	extMap, ok := extData.(map[string]interface{})
	if !ok {
		panic("Extension data is not in expected format")
	}

	attrs, ok := extMap["attributes"].(map[string]uint32)
	if !ok {
		panic("Attributes data is not in expected format")
	}

	fmt.Printf("Created instancing extension with %d attributes\n", len(attrs))
	// Print attributes in a consistent order
	names := []string{"TRANSLATION", "ROTATION", "SCALE"}
	for _, name := range names {
		if _, exists := attrs[name]; exists {
			fmt.Printf("  %s\n", name)
		}
	}

	// Output:
	// Created instancing extension with 3 attributes
	//   TRANSLATION
	//   ROTATION
	//   SCALE
}

func ExampleReadInstancing() {
	// In a real scenario, you would read a glTF document that contains instancing data
	// For this example, we'll just show the function signature and usage pattern

	// Read instancing data from a document for a specific node
	// data, err := ReadInstancing(doc, nodeIndex)
	// if err != nil {
	//     panic(err)
	// }
	//
	// fmt.Printf("Read %d instances\n", data.InstanceCount())
	// fmt.Printf("Translations: %v\n", data.Translations)
	// fmt.Printf("Rotations: %v\n", data.Rotations)
	// fmt.Printf("Scales: %v\n", data.Scales)

	fmt.Println("ReadInstancing function can be used to read instancing data from a glTF document")
	// Output: ReadInstancing function can be used to read instancing data from a glTF document
}

func ExampleInstanceData_ToMat4() {
	// Create instance data
	data := &InstanceData{
		Translations: [][3]float32{{1, 2, 3}},
		Rotations:    [][4]float32{{0, 0, 0, 1}}, // Identity quaternion
		Scales:       [][3]float32{{2, 2, 2}},
	}

	// Convert to matrices
	matrices, err := data.ToMat4()
	if err != nil {
		panic(err)
	}

	fmt.Printf("Converted %d instances to matrices\n", len(matrices))
	// In a real application, you would use these matrices for rendering
	_ = matrices // Suppress unused variable warning

	// Output: Converted 1 instances to matrices
}
