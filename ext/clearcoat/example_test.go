package clearcoat

import (
	"fmt"
	"log"

	"github.com/flywave/gltf"
)

func ExampleMaterialsClearcoat() {
	// Create a new glTF document
	doc := gltf.NewDocument()

	// Create a material with clearcoat extension
	material := &gltf.Material{
		Name: "ClearcoatMaterial",
		PBRMetallicRoughness: &gltf.PBRMetallicRoughness{
			BaseColorFactor: &[4]float32{1, 1, 1, 1},
			MetallicFactor:  gltf.Float(0),
			RoughnessFactor: gltf.Float(0.5),
		},
		Extensions: make(gltf.Extensions),
	}

	// Create clearcoat extension data
	clearcoatExt := &MaterialsClearcoat{
		ClearcoatFactor:          gltf.Float(1.0),
		ClearcoatRoughnessFactor: gltf.Float(0.2),
	}

	// Add the extension to the material
	material.Extensions[ExtensionName] = clearcoatExt

	// Add material to document
	doc.Materials = append(doc.Materials, material)

	// Add the extension to the document's required extensions
	doc.AddExtensionUsed(ExtensionName)

	// Print information about the material
	fmt.Printf("Material: %s\n", material.Name)
	fmt.Printf("Clearcoat Factor: %.1f\n", *clearcoatExt.ClearcoatFactor)
	fmt.Printf("Clearcoat Roughness: %.1f\n", *clearcoatExt.ClearcoatRoughnessFactor)

	// Output:
	// Material: ClearcoatMaterial
	// Clearcoat Factor: 1.0
	// Clearcoat Roughness: 0.2
}

func ExampleMaterialsClearcoat_withTextures() {
	// Create a new glTF document
	doc := gltf.NewDocument()

	// Create textures
	clearcoatTexture := &gltf.Texture{
		Sampler: gltf.Index(0),
		Source:  gltf.Index(0),
	}
	clearcoatRoughnessTexture := &gltf.Texture{
		Sampler: gltf.Index(0),
		Source:  gltf.Index(1),
	}
	clearcoatNormalTexture := &gltf.Texture{
		Sampler: gltf.Index(0),
		Source:  gltf.Index(2),
	}

	doc.Textures = append(doc.Textures, clearcoatTexture, clearcoatRoughnessTexture, clearcoatNormalTexture)

	// Create a material with clearcoat extension using textures
	material := &gltf.Material{
		Name: "TexturedClearcoatMaterial",
		PBRMetallicRoughness: &gltf.PBRMetallicRoughness{
			BaseColorFactor: &[4]float32{1, 1, 1, 1},
			MetallicFactor:  gltf.Float(0),
			RoughnessFactor: gltf.Float(0.3),
		},
		Extensions: make(gltf.Extensions),
	}

	// Create clearcoat extension data with textures
	clearcoatExt := &MaterialsClearcoat{
		ClearcoatFactor:           gltf.Float(0.8),
		ClearcoatTexture:          &gltf.TextureInfo{Index: *gltf.Index(0)},
		ClearcoatRoughnessFactor:  gltf.Float(0.1),
		ClearcoatRoughnessTexture: &gltf.TextureInfo{Index: *gltf.Index(1)},
		ClearcoatNormalTexture: &gltf.NormalTexture{
			Index: gltf.Index(2),
			Scale: gltf.Float(1.0),
		},
	}

	// Add the extension to the material
	material.Extensions[ExtensionName] = clearcoatExt

	// Add material to document
	doc.Materials = append(doc.Materials, material)

	// Add the extension to the document's required extensions
	doc.AddExtensionUsed(ExtensionName)

	// Print information about the material
	fmt.Printf("Material: %s\n", material.Name)
	fmt.Printf("Clearcoat Factor: %.1f\n", *clearcoatExt.ClearcoatFactor)
	fmt.Printf("Clearcoat Texture Index: %d\n", clearcoatExt.ClearcoatTexture.Index)
	fmt.Printf("Clearcoat Roughness: %.1f\n", *clearcoatExt.ClearcoatRoughnessFactor)
	fmt.Printf("Clearcoat Roughness Texture Index: %d\n", clearcoatExt.ClearcoatRoughnessTexture.Index)
	fmt.Printf("Clearcoat Normal Texture Index: %d\n", *clearcoatExt.ClearcoatNormalTexture.Index)

	// Output:
	// Material: TexturedClearcoatMaterial
	// Clearcoat Factor: 0.8
	// Clearcoat Texture Index: 0
	// Clearcoat Roughness: 0.1
	// Clearcoat Roughness Texture Index: 1
	// Clearcoat Normal Texture Index: 2
}

func ExampleUnmarshal() {
	// JSON data representing a material with clearcoat extension
	jsonData := []byte(`{
		"clearcoatFactor": 1.0,
		"clearcoatRoughnessFactor": 0.1
	}`)

	// Unmarshal the JSON data
	ext, err := Unmarshal(jsonData)
	if err != nil {
		log.Fatalf("Failed to unmarshal clearcoat extension: %v", err)
	}

	// Type assert to MaterialsClearcoat
	clearcoatExt, ok := ext.(*MaterialsClearcoat)
	if !ok {
		log.Fatal("Failed to type assert to MaterialsClearcoat")
	}

	fmt.Printf("Clearcoat Factor: %.1f\n", *clearcoatExt.ClearcoatFactor)
	fmt.Printf("Clearcoat Roughness: %.1f\n", *clearcoatExt.ClearcoatRoughnessFactor)

	// Output:
	// Clearcoat Factor: 1.0
	// Clearcoat Roughness: 0.1
}
