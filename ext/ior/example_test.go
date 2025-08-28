package ior

import (
	"fmt"
	"log"

	"github.com/flywave/gltf"
)

func ExampleMaterialsIOR() {
	// Create a new glTF document
	doc := gltf.NewDocument()

	// Create a material with ior extension
	material := &gltf.Material{
		Name: "IORMaterial",
		PBRMetallicRoughness: &gltf.PBRMetallicRoughness{
			BaseColorFactor: &[4]float32{1, 1, 1, 1},
			MetallicFactor:  gltf.Float(0),
			RoughnessFactor: gltf.Float(0.5),
		},
		Extensions: make(gltf.Extensions),
	}

	// Create ior extension data
	iorExt := &MaterialsIOR{
		IOR: gltf.Float(1.4),
	}

	// Add the extension to the material
	material.Extensions[ExtensionName] = iorExt

	// Add material to document
	doc.Materials = append(doc.Materials, material)

	// Add the extension to the document's required extensions
	doc.AddExtensionUsed(ExtensionName)

	// Print information about the material
	fmt.Printf("Material: %s\n", material.Name)
	fmt.Printf("IOR: %.2f\n", *iorExt.IOR)

	// Output:
	// Material: IORMaterial
	// IOR: 1.40
}

func ExampleUnmarshal() {
	// JSON data representing a material with ior extension
	jsonData := []byte(`{
		"ior": 1.33
	}`)

	// Unmarshal the JSON data
	ext, err := Unmarshal(jsonData)
	if err != nil {
		log.Fatalf("Failed to unmarshal ior extension: %v", err)
	}

	// Type assert to MaterialsIOR
	iorExt, ok := ext.(*MaterialsIOR)
	if !ok {
		log.Fatal("Failed to type assert to MaterialsIOR")
	}

	fmt.Printf("IOR: %.2f\n", *iorExt.IOR)

	// Output:
	// IOR: 1.33
}