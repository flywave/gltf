package emissivestrength

import (
	"fmt"
	"log"

	"github.com/flywave/gltf"
)

func ExampleMaterialsEmissiveStrength() {
	// Create a new glTF document
	doc := gltf.NewDocument()

	// Create a material with emissive strength extension
	material := &gltf.Material{
		Name: "EmissiveStrengthMaterial",
		PBRMetallicRoughness: &gltf.PBRMetallicRoughness{
			BaseColorFactor: &[4]float32{1, 1, 1, 1},
			MetallicFactor:  gltf.Float(0),
			RoughnessFactor: gltf.Float(0.5),
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

	// Print information about the material
	fmt.Printf("Material: %s\n", material.Name)
	fmt.Printf("Emissive Strength: %.1f\n", *emissiveStrengthExt.EmissiveStrength)

	// Output:
	// Material: EmissiveStrengthMaterial
	// Emissive Strength: 5.0
}

func ExampleUnmarshal() {
	// JSON data representing a material with emissive strength extension
	jsonData := []byte(`{
		"emissiveStrength": 5.0
	}`)

	// Unmarshal the JSON data
	ext, err := Unmarshal(jsonData)
	if err != nil {
		log.Fatalf("Failed to unmarshal emissive strength extension: %v", err)
	}

	// Type assert to MaterialsEmissiveStrength
	emissiveStrengthExt, ok := ext.(*MaterialsEmissiveStrength)
	if !ok {
		log.Fatal("Failed to type assert to MaterialsEmissiveStrength")
	}

	fmt.Printf("Emissive Strength: %.1f\n", *emissiveStrengthExt.EmissiveStrength)

	// Output:
	// Emissive Strength: 5.0
}
