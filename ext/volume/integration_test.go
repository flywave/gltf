package volume

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/flywave/gltf"
)

func TestIntegration(t *testing.T) {
	// Create a material with volume extension
	material := &gltf.Material{
		Name: "VolumeMaterial",
		PBRMetallicRoughness: &gltf.PBRMetallicRoughness{
			BaseColorFactor: &[4]float32{1, 1, 1, 1},
			MetallicFactor:  gltf.Float(0),
			RoughnessFactor: gltf.Float(0.5),
		},
		Extensions: make(gltf.Extensions),
	}

	// Create volume extension data
	volumeExt := &MaterialsVolume{
		ThicknessFactor:     gltf.Float(1.0),
		AttenuationDistance: gltf.Float(0.006),
		AttenuationColor:    &[3]float32{0.5, 0.5, 0.5},
	}

	// Add the extension to the material
	material.Extensions[ExtensionName] = volumeExt

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
		t.Fatal("Volume extension not found in unmarshaled material")
	}

	// Type assert to MaterialsVolume
	unmarshaledVolumeExt, ok := ext.(*MaterialsVolume)
	if !ok {
		t.Fatal("Failed to type assert to MaterialsVolume")
	}

	// Check that the values are correct
	if *unmarshaledVolumeExt.ThicknessFactor != *volumeExt.ThicknessFactor {
		t.Errorf("ThicknessFactor = %f, want %f", *unmarshaledVolumeExt.ThicknessFactor, *volumeExt.ThicknessFactor)
	}

	if *unmarshaledVolumeExt.AttenuationDistance != *volumeExt.AttenuationDistance {
		t.Errorf("AttenuationDistance = %f, want %f", *unmarshaledVolumeExt.AttenuationDistance, *volumeExt.AttenuationDistance)
	}

	if *unmarshaledVolumeExt.AttenuationColor != *volumeExt.AttenuationColor {
		t.Errorf("AttenuationColor = %v, want %v", *unmarshaledVolumeExt.AttenuationColor, *volumeExt.AttenuationColor)
	}
}

func TestIntegrationWithTextures(t *testing.T) {
	// Create a material with volume extension using textures
	material := &gltf.Material{
		Name: "TexturedVolumeMaterial",
		PBRMetallicRoughness: &gltf.PBRMetallicRoughness{
			BaseColorFactor: &[4]float32{1, 1, 1, 1},
			MetallicFactor:  gltf.Float(0),
			RoughnessFactor: gltf.Float(0.3),
		},
		Extensions: make(gltf.Extensions),
	}

	// Create volume extension data with textures
	volumeExt := &MaterialsVolume{
		ThicknessFactor:     gltf.Float(1.0),
		ThicknessTexture:    &gltf.TextureInfo{Index: 0},
		AttenuationDistance: gltf.Float(0.006),
		AttenuationColor:    &[3]float32{0.5, 0.5, 0.5},
	}

	// Add the extension to the material
	material.Extensions[ExtensionName] = volumeExt

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
		t.Fatal("Volume extension not found in unmarshaled material")
	}

	// Type assert to MaterialsVolume
	unmarshaledVolumeExt, ok := ext.(*MaterialsVolume)
	if !ok {
		t.Fatal("Failed to type assert to MaterialsVolume")
	}

	// Check that the values are correct
	if *unmarshaledVolumeExt.ThicknessFactor != *volumeExt.ThicknessFactor {
		t.Errorf("ThicknessFactor = %f, want %f", *unmarshaledVolumeExt.ThicknessFactor, *volumeExt.ThicknessFactor)
	}

	if *unmarshaledVolumeExt.AttenuationDistance != *volumeExt.AttenuationDistance {
		t.Errorf("AttenuationDistance = %f, want %f", *unmarshaledVolumeExt.AttenuationDistance, *volumeExt.AttenuationDistance)
	}

	if *unmarshaledVolumeExt.AttenuationColor != *volumeExt.AttenuationColor {
		t.Errorf("AttenuationColor = %v, want %v", *unmarshaledVolumeExt.AttenuationColor, *volumeExt.AttenuationColor)
	}

	if unmarshaledVolumeExt.ThicknessTexture.Index != volumeExt.ThicknessTexture.Index {
		t.Errorf("ThicknessTexture.Index = %d, want %d", unmarshaledVolumeExt.ThicknessTexture.Index, volumeExt.ThicknessTexture.Index)
	}
}

func TestIntegrationDefaultValues(t *testing.T) {
	// Create a material with volume extension using default values
	material := &gltf.Material{
		Name: "DefaultVolumeMaterial",
		PBRMetallicRoughness: &gltf.PBRMetallicRoughness{
			BaseColorFactor: &[4]float32{1, 1, 1, 1},
			MetallicFactor:  gltf.Float(0),
			RoughnessFactor: gltf.Float(0.5),
		},
		Extensions: make(gltf.Extensions),
	}

	// Create volume extension data with default values
	volumeExt := &MaterialsVolume{}

	// Add the extension to the material
	material.Extensions[ExtensionName] = volumeExt

	// Marshal the material to JSON
	data, err := json.Marshal(material)
	if err != nil {
		t.Fatalf("Failed to marshal material: %v", err)
	}

	// Check that default values are not included in the JSON
	if bytes.Contains(data, []byte("thicknessFactor")) {
		t.Error("Default thicknessFactor should not be included in JSON")
	}

	if bytes.Contains(data, []byte("attenuationDistance")) {
		t.Error("Default attenuationDistance should not be included in JSON")
	}

	if bytes.Contains(data, []byte("attenuationColor")) {
		t.Error("Default attenuationColor should not be included in JSON")
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
		t.Fatal("Volume extension not found in unmarshaled material")
	}

	// Type assert to MaterialsVolume
	unmarshaledVolumeExt, ok := ext.(*MaterialsVolume)
	if !ok {
		t.Fatal("Failed to type assert to MaterialsVolume")
	}

	// Check that the default values are correct
	if *unmarshaledVolumeExt.ThicknessFactor != 0.0 {
		t.Errorf("ThicknessFactor = %f, want 0.0", *unmarshaledVolumeExt.ThicknessFactor)
	}

	if *unmarshaledVolumeExt.AttenuationDistance != -1.0 {
		t.Errorf("AttenuationDistance = %f, want -1.0", *unmarshaledVolumeExt.AttenuationDistance)
	}

	if *unmarshaledVolumeExt.AttenuationColor != [3]float32{1.0, 1.0, 1.0} {
		t.Errorf("AttenuationColor = %v, want [1.0, 1.0, 1.0]", *unmarshaledVolumeExt.AttenuationColor)
	}
}