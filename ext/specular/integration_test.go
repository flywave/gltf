package specular

import (
	"testing"

	"github.com/flywave/gltf"
)

func TestIntegration(t *testing.T) {
	// Create a complete glTF document with specular material
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

	// Create a material with specular extension
	material := &gltf.Material{
		Name: "SpecularMaterial",
		PBRMetallicRoughness: &gltf.PBRMetallicRoughness{
			BaseColorFactor: &[4]float32{1, 1, 1, 1},
			MetallicFactor:  gltf.Float(0),
			RoughnessFactor: gltf.Float(0.5),
		},
		Extensions: make(gltf.Extensions),
	}

	// Create specular extension data
	specularExt := &MaterialsSpecular{
		SpecularFactor:      gltf.Float(0.8),
		SpecularColorFactor: &[3]float32{0.9, 0.9, 0.9},
	}

	// Add the extension to the material
	material.Extensions[ExtensionName] = specularExt

	// Add material to document
	doc.Materials = append(doc.Materials, material)

	// Create a node with the mesh and material
	node := &gltf.Node{
		Name:       "SpecularCube",
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
		t.Fatal("Specular extension was not added to material")
	}

	// Type assert to MaterialsSpecular
	specular, ok := ext.(*MaterialsSpecular)
	if !ok {
		t.Fatal("Failed to type assert to MaterialsSpecular")
	}

	// Verify extension values
	if specular.SpecularFactor == nil || *specular.SpecularFactor != 0.8 {
		t.Errorf("Expected specular factor 0.8, got %v", specular.SpecularFactor)
	}

	if specular.SpecularColorFactor == nil || specular.SpecularColorFactor[0] != 0.9 {
		t.Errorf("Expected specular color factor [0.9, 0.9, 0.9], got %v", specular.SpecularColorFactor)
	}

	// Test JSON marshaling and unmarshaling
	jsonData, err := specular.MarshalJSON()
	if err != nil {
		t.Fatalf("Failed to marshal specular extension: %v", err)
	}

	// Unmarshal the JSON data
	unmarshaledExt, err := Unmarshal(jsonData)
	if err != nil {
		t.Fatalf("Failed to unmarshal specular extension: %v", err)
	}

	unmarshaledSpecular, ok := unmarshaledExt.(*MaterialsSpecular)
	if !ok {
		t.Fatal("Failed to type assert unmarshaled data to MaterialsSpecular")
	}

	// Verify unmarshaled values
	if unmarshaledSpecular.SpecularFactor == nil || *unmarshaledSpecular.SpecularFactor != 0.8 {
		t.Errorf("Expected unmarshaled specular factor 0.8, got %v", unmarshaledSpecular.SpecularFactor)
	}

	if unmarshaledSpecular.SpecularColorFactor == nil || unmarshaledSpecular.SpecularColorFactor[0] != 0.9 {
		t.Errorf("Expected unmarshaled specular color factor [0.9, 0.9, 0.9], got %v", unmarshaledSpecular.SpecularColorFactor)
	}

	// Test with textures
	texture := &gltf.Texture{
		Sampler: gltf.Index(0),
		Source:  gltf.Index(0),
	}
	doc.Textures = append(doc.Textures, texture)

	specularWithTextures := &MaterialsSpecular{
		SpecularFactor: gltf.Float(0.7),
		SpecularTexture: &gltf.TextureInfo{
			Index: *gltf.Index(0),
		},
		SpecularColorFactor: &[3]float32{0.8, 0.8, 0.8},
	}

	jsonDataWithTextures, err := specularWithTextures.MarshalJSON()
	if err != nil {
		t.Fatalf("Failed to marshal specular extension with textures: %v", err)
	}

	unmarshaledExtWithTextures, err := Unmarshal(jsonDataWithTextures)
	if err != nil {
		t.Fatalf("Failed to unmarshal specular extension with textures: %v", err)
	}

	unmarshaledSpecularWithTextures, ok := unmarshaledExtWithTextures.(*MaterialsSpecular)
	if !ok {
		t.Fatal("Failed to type assert unmarshaled data with textures to MaterialsSpecular")
	}

	if unmarshaledSpecularWithTextures.SpecularTexture == nil {
		t.Error("Expected specular texture to be preserved")
	} else if unmarshaledSpecularWithTextures.SpecularTexture.Index != 0 {
		t.Errorf("Expected specular texture index 0, got %d", unmarshaledSpecularWithTextures.SpecularTexture.Index)
	}
}
