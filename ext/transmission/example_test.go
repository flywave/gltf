package transmission

import (
	"fmt"
	"log"

	"github.com/flywave/gltf"
)

func ExampleMaterialsTransmission() {
	// Create a new glTF document
	doc := gltf.NewDocument()

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

	// Add material to document
	doc.Materials = append(doc.Materials, material)

	// Add the extension to the document's required extensions
	doc.AddExtensionUsed(ExtensionName)

	// Print information about the material
	fmt.Printf("Material: %s\n", material.Name)
	fmt.Printf("Transmission Factor: %.1f\n", *transmissionExt.TransmissionFactor)

	// Output:
	// Material: TransmissionMaterial
	// Transmission Factor: 0.8
}

func ExampleMaterialsTransmission_withTexture() {
	// Create a new glTF document
	doc := gltf.NewDocument()

	// Create texture
	transmissionTexture := &gltf.Texture{
		Sampler: gltf.Index(0),
		Source:  gltf.Index(0),
	}

	doc.Textures = append(doc.Textures, transmissionTexture)

	// Create a material with transmission extension using texture
	material := &gltf.Material{
		Name: "TexturedTransmissionMaterial",
		PBRMetallicRoughness: &gltf.PBRMetallicRoughness{
			BaseColorFactor: &[4]float32{1, 1, 1, 1},
			MetallicFactor:  gltf.Float(0),
			RoughnessFactor: gltf.Float(0.3),
		},
		Extensions: make(gltf.Extensions),
	}

	// Create transmission extension data with texture
	transmissionExt := &MaterialsTransmission{
		TransmissionFactor: gltf.Float(0.6),
		TransmissionTexture: &gltf.TextureInfo{
			Index: 0,
		},
	}

	// Add the extension to the material
	material.Extensions[ExtensionName] = transmissionExt

	// Add material to document
	doc.Materials = append(doc.Materials, material)

	// Add the extension to the document's required extensions
	doc.AddExtensionUsed(ExtensionName)

	// Print information about the material
	fmt.Printf("Material: %s\n", material.Name)
	fmt.Printf("Transmission Factor: %.1f\n", *transmissionExt.TransmissionFactor)
	fmt.Printf("Transmission Texture Index: %d\n", transmissionExt.TransmissionTexture.Index)

	// Output:
	// Material: TexturedTransmissionMaterial
	// Transmission Factor: 0.6
	// Transmission Texture Index: 0
}

func ExampleUnmarshal() {
	// JSON data representing a material with transmission extension
	jsonData := []byte(`{
		"transmissionFactor": 0.8
	}`)

	// Unmarshal the JSON data
	ext, err := Unmarshal(jsonData)
	if err != nil {
		log.Fatalf("Failed to unmarshal transmission extension: %v", err)
	}

	// Type assert to MaterialsTransmission
	transmissionExt, ok := ext.(*MaterialsTransmission)
	if !ok {
		log.Fatal("Failed to type assert to MaterialsTransmission")
	}

	fmt.Printf("Transmission Factor: %.1f\n", *transmissionExt.TransmissionFactor)

	// Output:
	// Transmission Factor: 0.8
}
