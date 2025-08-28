package specular

import (
	"fmt"
	"log"

	"github.com/flywave/gltf"
)

func ExampleMaterialsSpecular() {
	// Create a new glTF document
	doc := gltf.NewDocument()

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

	// Add the extension to the document's required extensions
	doc.AddExtensionUsed(ExtensionName)

	// Print information about the material
	fmt.Printf("Material: %s\n", material.Name)
	fmt.Printf("Specular Factor: %.1f\n", *specularExt.SpecularFactor)
	fmt.Printf("Specular Color: [%.1f, %.1f, %.1f]\n",
		specularExt.SpecularColorFactor[0],
		specularExt.SpecularColorFactor[1],
		specularExt.SpecularColorFactor[2])

	// Output:
	// Material: SpecularMaterial
	// Specular Factor: 0.8
	// Specular Color: [0.9, 0.9, 0.9]
}

func ExampleMaterialsSpecular_withTextures() {
	// Create a new glTF document
	doc := gltf.NewDocument()

	// Create textures
	specularTexture := &gltf.Texture{
		Sampler: gltf.Index(0),
		Source:  gltf.Index(0),
	}
	specularColorTexture := &gltf.Texture{
		Sampler: gltf.Index(0),
		Source:  gltf.Index(1),
	}

	doc.Textures = append(doc.Textures, specularTexture, specularColorTexture)

	// Create a material with specular extension using textures
	material := &gltf.Material{
		Name: "TexturedSpecularMaterial",
		PBRMetallicRoughness: &gltf.PBRMetallicRoughness{
			BaseColorFactor: &[4]float32{1, 1, 1, 1},
			MetallicFactor:  gltf.Float(0),
			RoughnessFactor: gltf.Float(0.3),
		},
		Extensions: make(gltf.Extensions),
	}

	// Create specular extension data with textures
	specularExt := &MaterialsSpecular{
		SpecularFactor: gltf.Float(0.7),
		SpecularTexture: &gltf.TextureInfo{
			Index: *gltf.Index(0),
		},
		SpecularColorFactor: &[3]float32{0.8, 0.8, 0.8},
		SpecularColorTexture: &gltf.TextureInfo{
			Index: *gltf.Index(1),
		},
	}

	// Add the extension to the material
	material.Extensions[ExtensionName] = specularExt

	// Add material to document
	doc.Materials = append(doc.Materials, material)

	// Add the extension to the document's required extensions
	doc.AddExtensionUsed(ExtensionName)

	// Print information about the material
	fmt.Printf("Material: %s\n", material.Name)
	fmt.Printf("Specular Factor: %.1f\n", *specularExt.SpecularFactor)
	fmt.Printf("Specular Texture Index: %d\n", specularExt.SpecularTexture.Index)
	fmt.Printf("Specular Color: [%.1f, %.1f, %.1f]\n",
		specularExt.SpecularColorFactor[0],
		specularExt.SpecularColorFactor[1],
		specularExt.SpecularColorFactor[2])
	fmt.Printf("Specular Color Texture Index: %d\n", specularExt.SpecularColorTexture.Index)

	// Output:
	// Material: TexturedSpecularMaterial
	// Specular Factor: 0.7
	// Specular Texture Index: 0
	// Specular Color: [0.8, 0.8, 0.8]
	// Specular Color Texture Index: 1
}

func ExampleUnmarshal() {
	// JSON data representing a material with specular extension
	jsonData := []byte(`{
		"specularFactor": 0.5,
		"specularColorFactor": [0.7, 0.7, 0.7]
	}`)

	// Unmarshal the JSON data
	ext, err := Unmarshal(jsonData)
	if err != nil {
		log.Fatalf("Failed to unmarshal specular extension: %v", err)
	}

	// Type assert to MaterialsSpecular
	specularExt, ok := ext.(*MaterialsSpecular)
	if !ok {
		log.Fatal("Failed to type assert to MaterialsSpecular")
	}

	fmt.Printf("Specular Factor: %.1f\n", *specularExt.SpecularFactor)
	fmt.Printf("Specular Color: [%.1f, %.1f, %.1f]\n",
		specularExt.SpecularColorFactor[0],
		specularExt.SpecularColorFactor[1],
		specularExt.SpecularColorFactor[2])

	// Output:
	// Specular Factor: 0.5
	// Specular Color: [0.7, 0.7, 0.7]
}
