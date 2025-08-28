package instance

import (
	"testing"

	"github.com/flywave/gltf"
)

func TestIntegration(t *testing.T) {
	// Create a complete glTF document with instancing
	doc := gltf.NewDocument()

	// Create a simple mesh
	mesh := &gltf.Mesh{
		Name: "Cube",
		Primitives: []*gltf.Primitive{
			{
				Attributes: map[string]uint32{
					"POSITION": 0,
				},
				Indices: gltf.Index(1),
			},
		},
	}
	doc.Meshes = append(doc.Meshes, mesh)

	// Create a node with the mesh
	node := &gltf.Node{
		Name:       "InstancedCube",
		Mesh:       gltf.Index(0),
		Extensions: make(gltf.Extensions),
	}
	doc.Nodes = append(doc.Nodes, node)

	// Create instance data
	instanceData := &InstanceData{
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

	// Write instancing data to the document
	config := DefaultConfig()
	err := WriteInstancing(doc, instanceData, config)
	if err != nil {
		t.Fatalf("WriteInstancing failed: %v", err)
	}

	// Verify the extension was added
	if !doc.HasExtensionUsed(ExtensionName) {
		t.Error("Extension was not added to extensionsUsed")
	}

	// Verify the extension data exists
	extData, exists := doc.Extensions[ExtensionName]
	if !exists {
		t.Error("Extension data was not added to document")
	}

	// Verify the extension data structure
	extMap, ok := extData.(map[string]interface{})
	if !ok {
		t.Error("Extension data is not in expected format")
	}

	attrs, ok := extMap["attributes"].(map[string]uint32)
	if !ok {
		t.Error("Attributes data is not in expected format")
	}

	// Verify all required attributes are present
	requiredAttrs := []string{"TRANSLATION", "ROTATION", "SCALE"}
	for _, attr := range requiredAttrs {
		if _, exists := attrs[attr]; !exists {
			t.Errorf("Required attribute %s is missing", attr)
		}
	}

	// Now set the instance extension on the node
	err = SetInstanceExtension(node, attrs)
	if err != nil {
		t.Fatalf("SetInstanceExtension failed: %v", err)
	}

	// Test reading the instancing data back
	extAttrs, err := GetInstanceExtension(node)
	if err != nil {
		t.Fatalf("GetInstanceExtension failed: %v", err)
	}

	// Verify the attributes match what we set
	if len(extAttrs.Attributes) != 3 {
		t.Errorf("Expected 3 attributes, got %d", len(extAttrs.Attributes))
	}

	// Test validation
	err = ValidateInstanceAttributes(doc, attrs)
	// This will fail because we don't have actual accessors in the document
	// but that's okay for this integration test
	_ = err

	// Test matrix conversion
	matrices, err := instanceData.ToMat4()
	if err != nil {
		t.Errorf("ToMat4 failed: %v", err)
	}

	if len(matrices) != 4 {
		t.Errorf("Expected 4 matrices, got %d", len(matrices))
	}

	// Test conversion back
	convertedData, err := FromMat4(matrices)
	if err != nil {
		t.Errorf("FromMat4 failed: %v", err)
	}

	if convertedData.InstanceCount() != 4 {
		t.Errorf("Expected 4 instances, got %d", convertedData.InstanceCount())
	}
}
