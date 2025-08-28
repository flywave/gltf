package ior

import (
	"testing"

	"github.com/flywave/gltf"
)

func TestIntegration(t *testing.T) {
	// Create a complete glTF document with ior material
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

	// Create a material with ior extension
	material := &gltf.Material{
		Name: "IORMaterial",
		PBRMetallicRoughness: &gltf.PBRMetallicRoughness{
			BaseColorFactor: &[4]float32{1, 1, 1, 1},
			MetallicFactor:  gltf.Float(0),
			RoughnessFactor: gltf.Float(0.5),
		},
		Extensions: make(gltf.Extensions),
	}

	// Create ior extension data
	iorExt := &MaterialsIOR{
		IOR: gltf.Float(1.4),
	}

	// Add the extension to the material
	material.Extensions[ExtensionName] = iorExt

	// Add material to document
	doc.Materials = append(doc.Materials, material)

	// Create a node with the mesh and material
	node := &gltf.Node{
		Name:       "IORCube",
		Mesh:       gltf.Index(0),
		Extensions: make(gltf.Extensions),
	}
	doc.Nodes = append(doc.Nodes, node)

	// Add the extension to the document's required extensions
	doc.AddExtensionUsed(ExtensionName)

	// Verify the extension was added
	if !doc.HasExtensionUsed(ExtensionName) {
		t.Error("Extension was not added to extensionsUsed")
	}

	// Verify the material has the extension
	if len(doc.Materials) != 1 {
		t.Fatalf("Expected 1 material, got %d", len(doc.Materials))
	}

	mat := doc.Materials[0]
	ext, exists := mat.Extensions[ExtensionName]
	if !exists {
		t.Fatal("IOR extension was not added to material")
	}

	// Type assert to MaterialsIOR
	ior, ok := ext.(*MaterialsIOR)
	if !ok {
		t.Fatal("Failed to type assert to MaterialsIOR")
	}

	// Verify extension values
	if ior.IOR == nil || *ior.IOR != 1.4 {
		t.Errorf("Expected ior 1.4, got %v", ior.IOR)
	}

	// Test JSON marshaling and unmarshaling
	jsonData, err := ior.MarshalJSON()
	if err != nil {
		t.Fatalf("Failed to marshal ior extension: %v", err)
	}

	// Unmarshal the JSON data
	unmarshaledExt, err := Unmarshal(jsonData)
	if err != nil {
		t.Fatalf("Failed to unmarshal ior extension: %v", err)
	}

	unmarshaledIOR, ok := unmarshaledExt.(*MaterialsIOR)
	if !ok {
		t.Fatal("Failed to type assert unmarshaled data to MaterialsIOR")
	}

	// Verify unmarshaled values
	if unmarshaledIOR.IOR == nil || *unmarshaledIOR.IOR != 1.4 {
		t.Errorf("Expected unmarshaled ior 1.4, got %v", unmarshaledIOR.IOR)
	}

	// Test with different IOR values
	iorWithDifferentValue := &MaterialsIOR{
		IOR: gltf.Float(2.42), // Diamond
	}

	jsonDataWithDifferentValue, err := iorWithDifferentValue.MarshalJSON()
	if err != nil {
		t.Fatalf("Failed to marshal ior extension with different value: %v", err)
	}

	unmarshaledExtWithDifferentValue, err := Unmarshal(jsonDataWithDifferentValue)
	if err != nil {
		t.Fatalf("Failed to unmarshal ior extension with different value: %v", err)
	}

	unmarshaledIORWithDifferentValue, ok := unmarshaledExtWithDifferentValue.(*MaterialsIOR)
	if !ok {
		t.Fatal("Failed to type assert unmarshaled data with different value to MaterialsIOR")
	}

	if unmarshaledIORWithDifferentValue.IOR == nil || *unmarshaledIORWithDifferentValue.IOR != 2.42 {
		t.Errorf("Expected unmarshaled ior 2.42, got %v", unmarshaledIORWithDifferentValue.IOR)
	}
}