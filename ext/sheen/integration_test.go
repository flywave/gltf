package sheen

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/flywave/gltf"
)

func TestIntegration(t *testing.T) {
	// Create a material with sheen extension
	material := &gltf.Material{
		Name: "SheenMaterial",
		PBRMetallicRoughness: &gltf.PBRMetallicRoughness{
			BaseColorFactor: &[4]float32{1, 1, 1, 1},
			MetallicFactor:  gltf.Float(0),
			RoughnessFactor: gltf.Float(0.5),
		},
		Extensions: make(gltf.Extensions),
	}

	// Create sheen extension data
	sheenExt := &MaterialsSheen{
		SheenColorFactor:     &[3]float32{0.9, 0.9, 0.9},
		SheenRoughnessFactor: gltf.Float(0.3),
	}

	// Add the extension to the material
	material.Extensions[ExtensionName] = sheenExt

	// Marshal the material to JSON
	data, err := json.Marshal(material)
	if err != nil {
		t.Fatalf("Failed to marshal material: %v", err)
	}

	// Unmarshal the material from JSON
	var unmarshaledMaterial gltf.Material
	err = json.Unmarshal(data, &unmarshaledMaterial)
	if err != nil {
		t.Fatalf("Failed to unmarshal material: %v", err)
	}

	// Get the extension from the unmarshaled material
	ext, ok := unmarshaledMaterial.Extensions[ExtensionName]
	if !ok {
		t.Fatal("Sheen extension not found in unmarshaled material")
	}

	// Type assert to MaterialsSheen
	unmarshaledSheenExt, ok := ext.(*MaterialsSheen)
	if !ok {
		t.Fatal("Failed to type assert to MaterialsSheen")
	}

	// Check that the values are correct
	if *unmarshaledSheenExt.SheenRoughnessFactor != *sheenExt.SheenRoughnessFactor {
		t.Errorf("SheenRoughnessFactor = %f, want %f", *unmarshaledSheenExt.SheenRoughnessFactor, *sheenExt.SheenRoughnessFactor)
	}

	if *unmarshaledSheenExt.SheenColorFactor != *sheenExt.SheenColorFactor {
		t.Errorf("SheenColorFactor = %v, want %v", *unmarshaledSheenExt.SheenColorFactor, *sheenExt.SheenColorFactor)
	}
}

func TestIntegrationWithTextures(t *testing.T) {
	// Create a material with sheen extension using textures
	material := &gltf.Material{
		Name: "TexturedSheenMaterial",
		PBRMetallicRoughness: &gltf.PBRMetallicRoughness{
			BaseColorFactor: &[4]float32{1, 1, 1, 1},
			MetallicFactor:  gltf.Float(0),
			RoughnessFactor: gltf.Float(0.3),
		},
		Extensions: make(gltf.Extensions),
	}

	// Create sheen extension data with textures
	sheenExt := &MaterialsSheen{
		SheenColorFactor:      &[3]float32{0.8, 0.8, 0.8},
		SheenColorTexture:     &gltf.TextureInfo{Index: 0},
		SheenRoughnessFactor:  gltf.Float(0.5),
		SheenRoughnessTexture: &gltf.TextureInfo{Index: 1},
	}

	// Add the extension to the material
	material.Extensions[ExtensionName] = sheenExt

	// Marshal the material to JSON
	data, err := json.Marshal(material)
	if err != nil {
		t.Fatalf("Failed to marshal material: %v", err)
	}

	// Unmarshal the material from JSON
	var unmarshaledMaterial gltf.Material
	err = json.Unmarshal(data, &unmarshaledMaterial)
	if err != nil {
		t.Fatalf("Failed to unmarshal material: %v", err)
	}

	// Get the extension from the unmarshaled material
	ext, ok := unmarshaledMaterial.Extensions[ExtensionName]
	if !ok {
		t.Fatal("Sheen extension not found in unmarshaled material")
	}

	// Type assert to MaterialsSheen
	unmarshaledSheenExt, ok := ext.(*MaterialsSheen)
	if !ok {
		t.Fatal("Failed to type assert to MaterialsSheen")
	}

	// Check that the values are correct
	if *unmarshaledSheenExt.SheenRoughnessFactor != *sheenExt.SheenRoughnessFactor {
		t.Errorf("SheenRoughnessFactor = %f, want %f", *unmarshaledSheenExt.SheenRoughnessFactor, *sheenExt.SheenRoughnessFactor)
	}

	if *unmarshaledSheenExt.SheenColorFactor != *sheenExt.SheenColorFactor {
		t.Errorf("SheenColorFactor = %v, want %v", *unmarshaledSheenExt.SheenColorFactor, *sheenExt.SheenColorFactor)
	}

	if unmarshaledSheenExt.SheenColorTexture.Index != sheenExt.SheenColorTexture.Index {
		t.Errorf("SheenColorTexture.Index = %d, want %d", unmarshaledSheenExt.SheenColorTexture.Index, sheenExt.SheenColorTexture.Index)
	}

	if unmarshaledSheenExt.SheenRoughnessTexture.Index != sheenExt.SheenRoughnessTexture.Index {
		t.Errorf("SheenRoughnessTexture.Index = %d, want %d", unmarshaledSheenExt.SheenRoughnessTexture.Index, sheenExt.SheenRoughnessTexture.Index)
	}
}

func TestIntegrationDefaultValues(t *testing.T) {
	// Create a material with sheen extension using default values
	material := &gltf.Material{
		Name: "DefaultSheenMaterial",
		PBRMetallicRoughness: &gltf.PBRMetallicRoughness{
			BaseColorFactor: &[4]float32{1, 1, 1, 1},
			MetallicFactor:  gltf.Float(0),
			RoughnessFactor: gltf.Float(0.5),
		},
		Extensions: make(gltf.Extensions),
	}

	// Create sheen extension data with default values
	sheenExt := &MaterialsSheen{}

	// Add the extension to the material
	material.Extensions[ExtensionName] = sheenExt

	// Marshal the material to JSON
	data, err := json.Marshal(material)
	if err != nil {
		t.Fatalf("Failed to marshal material: %v", err)
	}

	// Check that default values are not included in the JSON
	if bytes.Contains(data, []byte("sheenColorFactor")) {
		t.Error("Default sheenColorFactor should not be included in JSON")
	}

	if bytes.Contains(data, []byte("sheenRoughnessFactor")) {
		t.Error("Default sheenRoughnessFactor should not be included in JSON")
	}

	// Unmarshal the material from JSON
	var unmarshaledMaterial gltf.Material
	err = json.Unmarshal(data, &unmarshaledMaterial)
	if err != nil {
		t.Fatalf("Failed to unmarshal material: %v", err)
	}

	// Get the extension from the unmarshaled material
	ext, ok := unmarshaledMaterial.Extensions[ExtensionName]
	if !ok {
		t.Fatal("Sheen extension not found in unmarshaled material")
	}

	// Type assert to MaterialsSheen
	unmarshaledSheenExt, ok := ext.(*MaterialsSheen)
	if !ok {
		t.Fatal("Failed to type assert to MaterialsSheen")
	}

	// Check that the default values are correct
	if *unmarshaledSheenExt.SheenRoughnessFactor != 0.0 {
		t.Errorf("SheenRoughnessFactor = %f, want 0.0", *unmarshaledSheenExt.SheenRoughnessFactor)
	}

	if *unmarshaledSheenExt.SheenColorFactor != [3]float32{0.0, 0.0, 0.0} {
		t.Errorf("SheenColorFactor = %v, want [0.0, 0.0, 0.0]", *unmarshaledSheenExt.SheenColorFactor)
	}
}
