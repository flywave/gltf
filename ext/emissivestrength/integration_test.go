package emissivestrength

import (
	"testing"

	"github.com/flywave/gltf"
)

func TestEmissiveStrengthIntegration(t *testing.T) {
	// Create a new glTF document
	doc := gltf.NewDocument()

	// Create a material with emissive strength extension
	material := &gltf.Material{
		Name: "EmissiveStrengthTestMaterial",
		PBRMetallicRoughness: &gltf.PBRMetallicRoughness{
			BaseColorFactor: &[4]float32{0.8, 0.8, 0.8, 1},
			MetallicFactor:  gltf.Float(0),
			RoughnessFactor: gltf.Float(0.2),
		},
		EmissiveFactor: [3]float32{1, 1, 1},
		Extensions:     make(gltf.Extensions),
	}

	// Create emissive strength extension data
	emissiveStrengthExt := &MaterialsEmissiveStrength{
		EmissiveStrength: gltf.Float(5.0),
	}

	// Add the extension to the material
	material.Extensions[ExtensionName] = emissiveStrengthExt

	// Add material to document
	doc.Materials = append(doc.Materials, material)

	// Add the extension to the document's required extensions
	doc.AddExtensionUsed(ExtensionName)

	// Verify that the extension is properly registered
	if !doc.HasExtensionUsed(ExtensionName) {
		t.Errorf("Expected extension %s to be registered as used", ExtensionName)
	}

	// Verify that we can retrieve the extension from the material
	retrievedExt, ok := material.Extensions[ExtensionName]
	if !ok {
		t.Error("Expected to find emissive strength extension in material")
	}

	// Verify that the retrieved extension has the correct type and values
	retrievedEmissiveStrength, ok := retrievedExt.(*MaterialsEmissiveStrength)
	if !ok {
		t.Error("Expected retrieved extension to be of type *MaterialsEmissiveStrength")
	}

	if *retrievedEmissiveStrength.EmissiveStrength != 5.0 {
		t.Errorf("Expected emissive strength to be 5.0, got %f", *retrievedEmissiveStrength.EmissiveStrength)
	}

	// Test with default value
	material2 := &gltf.Material{
		Name:       "EmissiveStrengthDefaultMaterial",
		Extensions: make(gltf.Extensions),
	}

	defaultEmissiveStrengthExt := &MaterialsEmissiveStrength{}
	// Initialize with default values by calling UnmarshalJSON
	defaultEmissiveStrengthExt.UnmarshalJSON([]byte("{}"))
	material2.Extensions[ExtensionName] = defaultEmissiveStrengthExt
	doc.Materials = append(doc.Materials, material2)

	// Verify default value is set
	retrievedExt2, ok := material2.Extensions[ExtensionName]
	if !ok {
		t.Error("Expected to find emissive strength extension in material2")
	}

	retrievedEmissiveStrength2, ok := retrievedExt2.(*MaterialsEmissiveStrength)
	if !ok {
		t.Error("Expected retrieved extension to be of type *MaterialsEmissiveStrength")
	}

	if retrievedEmissiveStrength2.EmissiveStrength == nil || *retrievedEmissiveStrength2.EmissiveStrength != 1.0 {
		t.Errorf("Expected default emissive strength to be 1.0, got %v", retrievedEmissiveStrength2.EmissiveStrength)
	}
}
