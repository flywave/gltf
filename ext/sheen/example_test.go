package sheen

import (
	"fmt"
	"log"

	"github.com/flywave/gltf"
)

func ExampleMaterialsSheen() {
	// Create a new glTF document
	doc := gltf.NewDocument()

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

	// Add material to document
	doc.Materials = append(doc.Materials, material)

	// Add the extension to the document's required extensions
	doc.AddExtensionUsed(ExtensionName)

	// Print information about the material
	fmt.Printf("Material: %s\n", material.Name)
	fmt.Printf("Sheen Color: [%.1f, %.1f, %.1f]\n", sheenExt.SheenColorFactor[0], sheenExt.SheenColorFactor[1], sheenExt.SheenColorFactor[2])
	fmt.Printf("Sheen Roughness: %.1f\n", *sheenExt.SheenRoughnessFactor)

	// Output:
	// Material: SheenMaterial
	// Sheen Color: [0.9, 0.9, 0.9]
	// Sheen Roughness: 0.3
}

func ExampleMaterialsSheen_withTexture() {
	// Create a new glTF document
	doc := gltf.NewDocument()

	// Create textures
	sheenColorTexture := &gltf.Texture{
		Sampler: gltf.Index(0),
		Source:  gltf.Index(0),
	}

	sheenRoughnessTexture := &gltf.Texture{
		Sampler: gltf.Index(1),
		Source:  gltf.Index(1),
	}

	doc.Textures = append(doc.Textures, sheenColorTexture, sheenRoughnessTexture)

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

	// Add material to document
	doc.Materials = append(doc.Materials, material)

	// Add the extension to the document's required extensions
	doc.AddExtensionUsed(ExtensionName)

	// Print information about the material
	fmt.Printf("Material: %s\n", material.Name)
	fmt.Printf("Sheen Color: [%.1f, %.1f, %.1f]\n", sheenExt.SheenColorFactor[0], sheenExt.SheenColorFactor[1], sheenExt.SheenColorFactor[2])
	fmt.Printf("Sheen Color Texture Index: %d\n", sheenExt.SheenColorTexture.Index)
	fmt.Printf("Sheen Roughness: %.1f\n", *sheenExt.SheenRoughnessFactor)
	fmt.Printf("Sheen Roughness Texture Index: %d\n", sheenExt.SheenRoughnessTexture.Index)

	// Output:
	// Material: TexturedSheenMaterial
	// Sheen Color: [0.8, 0.8, 0.8]
	// Sheen Color Texture Index: 0
	// Sheen Roughness: 0.5
	// Sheen Roughness Texture Index: 1
}

func ExampleUnmarshal() {
	// JSON data representing a material with sheen extension
	jsonData := []byte(`{
		"sheenColorFactor": [0.9, 0.9, 0.9],
		"sheenRoughnessFactor": 0.3
	}`)

	// Unmarshal the JSON data
	ext, err := Unmarshal(jsonData)
	if err != nil {
		log.Fatalf("Failed to unmarshal sheen extension: %v", err)
	}

	// Type assert to MaterialsSheen
	sheenExt, ok := ext.(*MaterialsSheen)
	if !ok {
		log.Fatal("Failed to type assert to MaterialsSheen")
	}

	fmt.Printf("Sheen Color: [%.1f, %.1f, %.1f]\n", sheenExt.SheenColorFactor[0], sheenExt.SheenColorFactor[1], sheenExt.SheenColorFactor[2])
	fmt.Printf("Sheen Roughness: %.1f\n", *sheenExt.SheenRoughnessFactor)

	// Output:
	// Sheen Color: [0.9, 0.9, 0.9]
	// Sheen Roughness: 0.3
}
