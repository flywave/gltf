package anisotropy

import (
	"fmt"
	"log"

	"github.com/flywave/gltf"
)

func ExampleMaterialsAnisotropy() {
	// Create a new glTF document
	doc := gltf.NewDocument()

	// Create a material with anisotropy extension
	material := &gltf.Material{
		Name: "AnisotropyMaterial",
		PBRMetallicRoughness: &gltf.PBRMetallicRoughness{
			BaseColorFactor: &[4]float32{1, 1, 1, 1},
			MetallicFactor:  gltf.Float(1),
			RoughnessFactor: gltf.Float(0.5),
		},
		Extensions: make(gltf.Extensions),
	}

	// Create anisotropy extension data
	anisotropyExt := &MaterialsAnisotropy{
		AnisotropyStrength: gltf.Float(0.6),
		AnisotropyRotation: gltf.Float(1.57),
	}

	// Add the extension to the material
	material.Extensions[ExtensionName] = anisotropyExt

	// Add material to document
	doc.Materials = append(doc.Materials, material)

	// Add the extension to the document's required extensions
	doc.AddExtensionUsed(ExtensionName)

	// Print information about the material
	fmt.Printf("Material: %s\n", material.Name)
	fmt.Printf("Anisotropy Strength: %.1f\n", *anisotropyExt.AnisotropyStrength)
	fmt.Printf("Anisotropy Rotation: %.2f\n", *anisotropyExt.AnisotropyRotation)

	// Output:
	// Material: AnisotropyMaterial
	// Anisotropy Strength: 0.6
	// Anisotropy Rotation: 1.57
}

func ExampleMaterialsAnisotropy_withTexture() {
	// Create a new glTF document
	doc := gltf.NewDocument()

	// Create texture
	anisotropyTexture := &gltf.Texture{
		Sampler: gltf.Index(0),
		Source:  gltf.Index(0),
	}

	doc.Textures = append(doc.Textures, anisotropyTexture)

	// Create a material with anisotropy extension using texture
	material := &gltf.Material{
		Name: "TexturedAnisotropyMaterial",
		PBRMetallicRoughness: &gltf.PBRMetallicRoughness{
			BaseColorFactor: &[4]float32{1, 1, 1, 1},
			MetallicFactor:  gltf.Float(1),
			RoughnessFactor: gltf.Float(0.3),
		},
		Extensions: make(gltf.Extensions),
	}

	// Create anisotropy extension data with texture
	anisotropyExt := &MaterialsAnisotropy{
		AnisotropyStrength: gltf.Float(0.8),
		AnisotropyRotation: gltf.Float(0.785),
		AnisotropyTexture: &gltf.TextureInfo{
			Index: *gltf.Index(0),
		},
	}

	// Add the extension to the material
	material.Extensions[ExtensionName] = anisotropyExt

	// Add material to document
	doc.Materials = append(doc.Materials, material)

	// Add the extension to the document's required extensions
	doc.AddExtensionUsed(ExtensionName)

	// Print information about the material
	fmt.Printf("Material: %s\n", material.Name)
	fmt.Printf("Anisotropy Strength: %.1f\n", *anisotropyExt.AnisotropyStrength)
	fmt.Printf("Anisotropy Rotation: %.3f\n", *anisotropyExt.AnisotropyRotation)
	fmt.Printf("Anisotropy Texture Index: %d\n", anisotropyExt.AnisotropyTexture.Index)

	// Output:
	// Material: TexturedAnisotropyMaterial
	// Anisotropy Strength: 0.8
	// Anisotropy Rotation: 0.785
	// Anisotropy Texture Index: 0
}

func ExampleUnmarshal() {
	// JSON data representing a material with anisotropy extension
	jsonData := []byte(`{
		"anisotropyStrength": 0.6,
		"anisotropyRotation": 1.57
	}`)

	// Unmarshal the JSON data
	ext, err := Unmarshal(jsonData)
	if err != nil {
		log.Fatalf("Failed to unmarshal anisotropy extension: %v", err)
	}

	// Type assert to MaterialsAnisotropy
	anisotropyExt, ok := ext.(*MaterialsAnisotropy)
	if !ok {
		log.Fatal("Failed to type assert to MaterialsAnisotropy")
	}

	fmt.Printf("Anisotropy Strength: %.1f\n", *anisotropyExt.AnisotropyStrength)
	fmt.Printf("Anisotropy Rotation: %.2f\n", *anisotropyExt.AnisotropyRotation)

	// Output:
	// Anisotropy Strength: 0.6
	// Anisotropy Rotation: 1.57
}
