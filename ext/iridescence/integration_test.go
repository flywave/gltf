package iridescence

import (
	"testing"

	"github.com/flywave/gltf"
)

func TestIridescenceIntegration(t *testing.T) {
	// Create a new glTF document
	doc := gltf.NewDocument()

	// Create a material with iridescence extension
	material := &gltf.Material{
		Name: "IridescenceTestMaterial",
		PBRMetallicRoughness: &gltf.PBRMetallicRoughness{
			BaseColorFactor: &[4]float32{0.8, 0.8, 0.8, 1},
			MetallicFactor:  gltf.Float(0),
			RoughnessFactor: gltf.Float(0.2),
		},
		Extensions: make(gltf.Extensions),
	}

	// Create iridescence extension data
	iridescenceExt := &MaterialsIridescence{
		IridescenceFactor:           gltf.Float(1.0),
		IridescenceIor:              gltf.Float(1.3),
		IridescenceThicknessMaximum: gltf.Float(400.0),
	}

	// Add the extension to the material
	material.Extensions[ExtensionName] = iridescenceExt

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
		t.Error("Expected to find iridescence extension in material")
	}

	// Verify that the retrieved extension has the correct type and values
	retrievedIridescence, ok := retrievedExt.(*MaterialsIridescence)
	if !ok {
		t.Error("Expected retrieved extension to be of type *MaterialsIridescence")
	}

	if *retrievedIridescence.IridescenceFactor != 1.0 {
		t.Errorf("Expected iridescence factor to be 1.0, got %f", *retrievedIridescence.IridescenceFactor)
	}

	if *retrievedIridescence.IridescenceIor != 1.3 {
		t.Errorf("Expected iridescence IOR to be 1.3, got %f", *retrievedIridescence.IridescenceIor)
	}

	if *retrievedIridescence.IridescenceThicknessMaximum != 400.0 {
		t.Errorf("Expected iridescence thickness maximum to be 400.0, got %f", *retrievedIridescence.IridescenceThicknessMaximum)
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

	doc.Textures = append(doc.Textures, texture1, texture2)

	iridescenceExtWithTextures := &MaterialsIridescence{
		IridescenceFactor:           gltf.Float(0.8),
		IridescenceTexture:          &gltf.TextureInfo{Index: *gltf.Index(0)},
		IridescenceIor:              gltf.Float(1.5),
		IridescenceThicknessMinimum: gltf.Float(200.0),
		IridescenceThicknessMaximum: gltf.Float(600.0),
		IridescenceThicknessTexture: &gltf.TextureInfo{Index: *gltf.Index(1)},
	}

	material2 := &gltf.Material{
		Name:       "IridescenceTextureTestMaterial",
		Extensions: make(gltf.Extensions),
	}
	material2.Extensions[ExtensionName] = iridescenceExtWithTextures
	doc.Materials = append(doc.Materials, material2)

	// Verify textures are properly set
	retrievedExt2, ok := material2.Extensions[ExtensionName]
	if !ok {
		t.Error("Expected to find iridescence extension in material2")
	}

	retrievedIridescence2, ok := retrievedExt2.(*MaterialsIridescence)
	if !ok {
		t.Error("Expected retrieved extension to be of type *MaterialsIridescence")
	}

	if retrievedIridescence2.IridescenceTexture == nil {
		t.Error("Expected iridescence texture to be set")
	}

	if retrievedIridescence2.IridescenceTexture.Index != 0 {
		t.Errorf("Expected iridescence texture index to be 0, got %d", retrievedIridescence2.IridescenceTexture.Index)
	}

	if retrievedIridescence2.IridescenceThicknessTexture == nil {
		t.Error("Expected iridescence thickness texture to be set")
	}

	if retrievedIridescence2.IridescenceThicknessTexture.Index != 1 {
		t.Errorf("Expected iridescence thickness texture index to be 1, got %d", retrievedIridescence2.IridescenceThicknessTexture.Index)
	}

	if *retrievedIridescence2.IridescenceFactor != 0.8 {
		t.Errorf("Expected iridescence factor to be 0.8, got %f", *retrievedIridescence2.IridescenceFactor)
	}

	if *retrievedIridescence2.IridescenceIor != 1.5 {
		t.Errorf("Expected iridescence IOR to be 1.5, got %f", *retrievedIridescence2.IridescenceIor)
	}

	if *retrievedIridescence2.IridescenceThicknessMinimum != 200.0 {
		t.Errorf("Expected iridescence thickness minimum to be 200.0, got %f", *retrievedIridescence2.IridescenceThicknessMinimum)
	}

	if *retrievedIridescence2.IridescenceThicknessMaximum != 600.0 {
		t.Errorf("Expected iridescence thickness maximum to be 600.0, got %f", *retrievedIridescence2.IridescenceThicknessMaximum)
	}
}
