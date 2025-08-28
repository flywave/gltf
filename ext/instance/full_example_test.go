package instance

import (
	"fmt"

	"github.com/flywave/gltf"
	"github.com/flywave/gltf/modeler"
)

func Example_fullUsage() {
	// Create a new glTF document
	doc := gltf.NewDocument()

	// Create a simple cube mesh
	positions := [][3]float32{
		{-0.5, -0.5, 0.5}, {0.5, -0.5, 0.5}, {0.5, 0.5, 0.5}, {-0.5, 0.5, 0.5},
		{-0.5, -0.5, -0.5}, {0.5, -0.5, -0.5}, {0.5, 0.5, -0.5}, {-0.5, 0.5, -0.5},
	}
	indices := []uint16{0, 1, 2, 0, 2, 3, 4, 6, 5, 4, 7, 6, 0, 4, 5, 0, 5, 1, 2, 6, 7, 2, 7, 3, 0, 3, 7, 0, 7, 4, 1, 5, 6, 1, 6, 2}

	// Write the mesh data to the document
	positionAccessor := modeler.WritePosition(doc, positions)
	indicesAccessor := modeler.WriteIndices(doc, indices)

	// Create the mesh
	mesh := &gltf.Mesh{
		Name: "Cube",
		Primitives: []*gltf.Primitive{
			{
				Attributes: map[string]uint32{
					gltf.POSITION: positionAccessor,
				},
				Indices: gltf.Index(indicesAccessor),
			},
		},
	}
	doc.Meshes = append(doc.Meshes, mesh)

	// Create a node with the mesh
	node := &gltf.Node{
		Name:       "InstancedCubes",
		Mesh:       gltf.Index(0),
		Extensions: make(gltf.Extensions),
	}
	doc.Nodes = append(doc.Nodes, node)

	// Create instance data for a 2x2 grid of cubes
	instanceData := &InstanceData{
		Translations: [][3]float32{
			{-1, -1, 0}, // Bottom-left
			{1, -1, 0},  // Bottom-right
			{-1, 1, 0},  // Top-left
			{1, 1, 0},   // Top-right
		},
		Rotations: [][4]float32{
			{0, 0, 0, 1}, // No rotation
			{0, 0, 0, 1},
			{0, 0, 0, 1},
			{0, 0, 0, 1},
		},
		Scales: [][3]float32{
			{1, 1, 1}, // Uniform scale
			{1, 1, 1},
			{1, 1, 1},
			{1, 1, 1},
		},
	}

	// Write instancing data to the document
	config := DefaultConfig()
	err := WriteInstancing(doc, instanceData, config)
	if err != nil {
		panic(err)
	}

	// Get the attributes that were created
	extData := doc.Extensions[ExtensionName].(map[string]interface{})
	attributes := extData["attributes"].(map[string]uint32)

	// Set the instance extension on the node
	err = SetInstanceExtension(node, attributes)
	if err != nil {
		panic(err)
	}

	// Add the extension to the document's required extensions
	doc.AddExtensionUsed(ExtensionName)

	// Print information about what we created
	fmt.Printf("Created a scene with %d instanced cubes\n", instanceData.InstanceCount())
	fmt.Printf("Mesh: %s\n", mesh.Name)
	fmt.Printf("Node: %s\n", node.Name)
	fmt.Println("Instance attributes:")
	for name, accessorIdx := range attributes {
		fmt.Printf("  %s: accessor %d\n", name, accessorIdx)
	}

	// Convert to matrices to show the transformation data
	matrices, err := instanceData.ToMat4()
	if err != nil {
		panic(err)
	}

	fmt.Printf("Converted to %d transformation matrices\n", len(matrices))
	fmt.Println("First instance translation:", instanceData.Translations[0])
	fmt.Println("Last instance translation:", instanceData.Translations[3])

	// Output:
	// Created a scene with 4 instanced cubes
	// Mesh: Cube
	// Node: InstancedCubes
	// Instance attributes:
	//   TRANSLATION: accessor 2
	//   ROTATION: accessor 3
	//   SCALE: accessor 4
	// Converted to 4 transformation matrices
	// First instance translation: [-1 -1 0]
	// Last instance translation: [1 1 0]
}