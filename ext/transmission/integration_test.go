package transmission

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/flywave/gltf"
)

func TestIntegration(t *testing.T) {
	// Create a material with transmission extension
	material := &gltf.Material{
		Name: "TransmissionMaterial",
		PBRMetallicRoughness: &gltf.PBRMetallicRoughness{
			BaseColorFactor: &[4]float32{1, 1, 1, 1},
			MetallicFactor:  gltf.Float(0),
			RoughnessFactor: gltf.Float(0.5),
		},
		Extensions: make(gltf.Extensions),
	}

	// Create transmission extension data
	transmissionExt := &MaterialsTransmission{
		TransmissionFactor: gltf.Float(0.8),
	}

	// Add the extension to the material
	material.Extensions[ExtensionName] = transmissionExt

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
		t.Fatal("Transmission extension not found in unmarshaled material")
	}

	// Type assert to MaterialsTransmission
	unmarshaledTransmissionExt, ok := ext.(*MaterialsTransmission)
	if !ok {
		t.Fatal("Failed to type assert to MaterialsTransmission")
	}

	// Check that the values are correct
	if *unmarshaledTransmissionExt.TransmissionFactor != *transmissionExt.TransmissionFactor {
		t.Errorf("TransmissionFactor = %f, want %f", *unmarshaledTransmissionExt.TransmissionFactor, *transmissionExt.TransmissionFactor)
	}
}

func TestIntegrationWithTextures(t *testing.T) {
	// Create a material with transmission extension using textures
	material := &gltf.Material{
		Name: "TexturedTransmissionMaterial",
		PBRMetallicRoughness: &gltf.PBRMetallicRoughness{
			BaseColorFactor: &[4]float32{1, 1, 1, 1},
			MetallicFactor:  gltf.Float(0),
			RoughnessFactor: gltf.Float(0.3),
		},
		Extensions: make(gltf.Extensions),
	}

	// Create transmission extension data with textures
	transmissionExt := &MaterialsTransmission{
		TransmissionFactor:  gltf.Float(0.6),
		TransmissionTexture: &gltf.TextureInfo{Index: 0},
	}

	// Add the extension to the material
	material.Extensions[ExtensionName] = transmissionExt

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
		t.Fatal("Transmission extension not found in unmarshaled material")
	}

	// Type assert to MaterialsTransmission
	unmarshaledTransmissionExt, ok := ext.(*MaterialsTransmission)
	if !ok {
		t.Fatal("Failed to type assert to MaterialsTransmission")
	}

	// Check that the values are correct
	if *unmarshaledTransmissionExt.TransmissionFactor != *transmissionExt.TransmissionFactor {
		t.Errorf("TransmissionFactor = %f, want %f", *unmarshaledTransmissionExt.TransmissionFactor, *transmissionExt.TransmissionFactor)
	}

	if unmarshaledTransmissionExt.TransmissionTexture.Index != transmissionExt.TransmissionTexture.Index {
		t.Errorf("TransmissionTexture.Index = %d, want %d", unmarshaledTransmissionExt.TransmissionTexture.Index, transmissionExt.TransmissionTexture.Index)
	}
}

func TestIntegrationDefaultValues(t *testing.T) {
	// Create a material with transmission extension using default values
	material := &gltf.Material{
		Name: "DefaultTransmissionMaterial",
		PBRMetallicRoughness: &gltf.PBRMetallicRoughness{
			BaseColorFactor: &[4]float32{1, 1, 1, 1},
			MetallicFactor:  gltf.Float(0),
			RoughnessFactor: gltf.Float(0.5),
		},
		Extensions: make(gltf.Extensions),
	}

	// Create transmission extension data with default values
	transmissionExt := &MaterialsTransmission{}

	// Add the extension to the material
	material.Extensions[ExtensionName] = transmissionExt

	// Marshal the material to JSON
	data, err := json.Marshal(material)
	if err != nil {
		t.Fatalf("Failed to marshal material: %v", err)
	}

	// Check that default values are not included in the JSON
	if bytes.Contains(data, []byte("transmissionFactor")) {
		t.Error("Default transmissionFactor should not be included in JSON")
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
		t.Fatal("Transmission extension not found in unmarshaled material")
	}

	// Type assert to MaterialsTransmission
	unmarshaledTransmissionExt, ok := ext.(*MaterialsTransmission)
	if !ok {
		t.Fatal("Failed to type assert to MaterialsTransmission")
	}

	// Check that the default values are correct
	if *unmarshaledTransmissionExt.TransmissionFactor != 0.0 {
		t.Errorf("TransmissionFactor = %f, want 0.0", *unmarshaledTransmissionExt.TransmissionFactor)
	}
}