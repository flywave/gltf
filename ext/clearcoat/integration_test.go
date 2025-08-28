package clearcoat

import (
	"testing"

	"github.com/flywave/gltf"
)

func TestClearcoatIntegration(t *testing.T) {
	// Create a new glTF document
	doc := gltf.NewDocument()

	// Create a material with clearcoat extension
	material := &gltf.Material{
		Name: "ClearcoatTestMaterial",
		PBRMetallicRoughness: &gltf.PBRMetallicRoughness{
			BaseColorFactor: &[4]float32{0.8, 0.8, 0.8, 1},
			MetallicFactor:  gltf.Float(0),
			RoughnessFactor: gltf.Float(0.2),
		},
		Extensions: make(gltf.Extensions),
	}

	// Create clearcoat extension data
	clearcoatExt := &MaterialsClearcoat{
		ClearcoatFactor:          gltf.Float(1.0),
		ClearcoatRoughnessFactor: gltf.Float(0.1),
	}

	// Add the extension to the material
	material.Extensions[ExtensionName] = clearcoatExt

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
		t.Error("Expected to find clearcoat extension in material")
	}

	// Verify that the retrieved extension has the correct type and values
	retrievedClearcoat, ok := retrievedExt.(*MaterialsClearcoat)
	if !ok {
		t.Error("Expected retrieved extension to be of type *MaterialsClearcoat")
	}

	if *retrievedClearcoat.ClearcoatFactor != 1.0 {
		t.Errorf("Expected clearcoat factor to be 1.0, got %f", *retrievedClearcoat.ClearcoatFactor)
	}

	if *retrievedClearcoat.ClearcoatRoughnessFactor != 0.1 {
		t.Errorf("Expected clearcoat roughness factor to be 0.1, got %f", *retrievedClearcoat.ClearcoatRoughnessFactor)
	}

	// Test with textures
	texture1 := &gltf.Texture{
		Sampler: gltf.Index(0),
		Source:  gltf.Index(0),
	}
	texture2 := &gltf.Texture{
		Sampler: gltf.Index(0),
		Source:  gltf.Index(1),
	}
	texture3 := &gltf.Texture{
		Sampler: gltf.Index(0),
		Source:  gltf.Index(2),
	}

	doc.Textures = append(doc.Textures, texture1, texture2, texture3)

	clearcoatExtWithTextures := &MaterialsClearcoat{
		ClearcoatFactor:           gltf.Float(0.8),
		ClearcoatTexture:          &gltf.TextureInfo{Index: *gltf.Index(0)},
		ClearcoatRoughnessFactor:  gltf.Float(0.2),
		ClearcoatRoughnessTexture: &gltf.TextureInfo{Index: *gltf.Index(1)},
		ClearcoatNormalTexture: &gltf.NormalTexture{
			Index: gltf.Index(2),
			Scale: gltf.Float(1.0),
		},
	}

	material2 := &gltf.Material{
		Name:       "ClearcoatTextureTestMaterial",
		Extensions: make(gltf.Extensions),
	}
	material2.Extensions[ExtensionName] = clearcoatExtWithTextures
	doc.Materials = append(doc.Materials, material2)

	// Verify textures are properly set
	retrievedExt2, ok := material2.Extensions[ExtensionName]
	if !ok {
		t.Error("Expected to find clearcoat extension in material2")
	}

	retrievedClearcoat2, ok := retrievedExt2.(*MaterialsClearcoat)
	if !ok {
		t.Error("Expected retrieved extension to be of type *MaterialsClearcoat")
	}

	if retrievedClearcoat2.ClearcoatTexture == nil {
		t.Error("Expected clearcoat texture to be set")
	}

	if retrievedClearcoat2.ClearcoatTexture.Index != 0 {
		t.Errorf("Expected clearcoat texture index to be 0, got %d", retrievedClearcoat2.ClearcoatTexture.Index)
	}

	if retrievedClearcoat2.ClearcoatRoughnessTexture == nil {
		t.Error("Expected clearcoat roughness texture to be set")
	}

	if retrievedClearcoat2.ClearcoatRoughnessTexture.Index != 1 {
		t.Errorf("Expected clearcoat roughness texture index to be 1, got %d", retrievedClearcoat2.ClearcoatRoughnessTexture.Index)
	}

	if retrievedClearcoat2.ClearcoatNormalTexture == nil {
		t.Error("Expected clearcoat normal texture to be set")
	}

	if *retrievedClearcoat2.ClearcoatNormalTexture.Index != 2 {
		t.Errorf("Expected clearcoat normal texture index to be 2, got %d", *retrievedClearcoat2.ClearcoatNormalTexture.Index)
	}

	if *retrievedClearcoat2.ClearcoatFactor != 0.8 {
		t.Errorf("Expected clearcoat factor to be 0.8, got %f", *retrievedClearcoat2.ClearcoatFactor)
	}

	if *retrievedClearcoat2.ClearcoatRoughnessFactor != 0.2 {
		t.Errorf("Expected clearcoat roughness factor to be 0.2, got %f", *retrievedClearcoat2.ClearcoatRoughnessFactor)
	}
}
