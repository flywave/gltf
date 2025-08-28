package anisotropy

import (
	"testing"

	"github.com/flywave/gltf"
)

func TestAnisotropyIntegration(t *testing.T) {
	// Create a new glTF document
	doc := gltf.NewDocument()

	// Create a material with anisotropy extension
	material := &gltf.Material{
		Name: "AnisotropyTestMaterial",
		PBRMetallicRoughness: &gltf.PBRMetallicRoughness{
			BaseColorFactor: &[4]float32{0.8, 0.8, 0.8, 1},
			MetallicFactor:  gltf.Float(1),
			RoughnessFactor: gltf.Float(0.2),
		},
		Extensions: make(gltf.Extensions),
	}

	// Create anisotropy extension data
	anisotropyExt := &MaterialsAnisotropy{
		AnisotropyStrength: gltf.Float(0.7),
		AnisotropyRotation: gltf.Float(0.5),
	}

	// Add the extension to the material
	material.Extensions[ExtensionName] = anisotropyExt

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
		t.Error("Expected to find anisotropy extension in material")
	}

	// Verify that the retrieved extension has the correct type and values
	retrievedAnisotropy, ok := retrievedExt.(*MaterialsAnisotropy)
	if !ok {
		t.Error("Expected retrieved extension to be of type *MaterialsAnisotropy")
	}

	if *retrievedAnisotropy.AnisotropyStrength != 0.7 {
		t.Errorf("Expected anisotropy strength to be 0.7, got %f", *retrievedAnisotropy.AnisotropyStrength)
	}

	if *retrievedAnisotropy.AnisotropyRotation != 0.5 {
		t.Errorf("Expected anisotropy rotation to be 0.5, got %f", *retrievedAnisotropy.AnisotropyRotation)
	}

	// Test with texture
	texture := &gltf.Texture{
		Sampler: gltf.Index(0),
		Source:  gltf.Index(0),
	}
	doc.Textures = append(doc.Textures, texture)

	anisotropyExtWithTexture := &MaterialsAnisotropy{
		AnisotropyStrength: gltf.Float(0.9),
		AnisotropyTexture: &gltf.TextureInfo{
			Index: *gltf.Index(0),
		},
	}

	material2 := &gltf.Material{
		Name:       "AnisotropyTextureTestMaterial",
		Extensions: make(gltf.Extensions),
	}
	material2.Extensions[ExtensionName] = anisotropyExtWithTexture
	doc.Materials = append(doc.Materials, material2)

	// Verify texture is properly set
	retrievedExt2, ok := material2.Extensions[ExtensionName]
	if !ok {
		t.Error("Expected to find anisotropy extension in material2")
	}

	retrievedAnisotropy2, ok := retrievedExt2.(*MaterialsAnisotropy)
	if !ok {
		t.Error("Expected retrieved extension to be of type *MaterialsAnisotropy")
	}

	if retrievedAnisotropy2.AnisotropyTexture == nil {
		t.Error("Expected anisotropy texture to be set")
	}

	if retrievedAnisotropy2.AnisotropyTexture.Index != 0 {
		t.Errorf("Expected anisotropy texture index to be 0, got %d", retrievedAnisotropy2.AnisotropyTexture.Index)
	}

	if *retrievedAnisotropy2.AnisotropyStrength != 0.9 {
		t.Errorf("Expected anisotropy strength to be 0.9, got %f", *retrievedAnisotropy2.AnisotropyStrength)
	}
}
