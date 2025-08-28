package iridescence

import (
	"fmt"
	"log"

	"github.com/flywave/gltf"
)

func ExampleMaterialsIridescence() {
	// Create a new glTF document
	doc := gltf.NewDocument()

	// Create a material with iridescence extension
	material := &gltf.Material{
		Name: "IridescenceMaterial",
		PBRMetallicRoughness: &gltf.PBRMetallicRoughness{
			BaseColorFactor: &[4]float32{1, 1, 1, 1},
			MetallicFactor:  gltf.Float(0),
			RoughnessFactor: gltf.Float(0.5),
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

	// Print information about the material
	fmt.Printf("Material: %s\n", material.Name)
	fmt.Printf("Iridescence Factor: %.1f\n", *iridescenceExt.IridescenceFactor)
	fmt.Printf("Iridescence IOR: %.1f\n", *iridescenceExt.IridescenceIor)
	fmt.Printf("Iridescence Thickness Maximum: %.1f\n", *iridescenceExt.IridescenceThicknessMaximum)

	// Output:
	// Material: IridescenceMaterial
	// Iridescence Factor: 1.0
	// Iridescence IOR: 1.3
	// Iridescence Thickness Maximum: 400.0
}

func ExampleMaterialsIridescence_withTextures() {
	// Create a new glTF document
	doc := gltf.NewDocument()

	// Create textures
	iridescenceTexture := &gltf.Texture{
		Sampler: gltf.Index(0),
		Source:  gltf.Index(0),
	}
	iridescenceThicknessTexture := &gltf.Texture{
		Sampler: gltf.Index(0),
		Source:  gltf.Index(1),
	}

	doc.Textures = append(doc.Textures, iridescenceTexture, iridescenceThicknessTexture)

	// Create a material with iridescence extension using textures
	material := &gltf.Material{
		Name: "TexturedIridescenceMaterial",
		PBRMetallicRoughness: &gltf.PBRMetallicRoughness{
			BaseColorFactor: &[4]float32{1, 1, 1, 1},
			MetallicFactor:  gltf.Float(0),
			RoughnessFactor: gltf.Float(0.3),
		},
		Extensions: make(gltf.Extensions),
	}

	// Create iridescence extension data with textures
	iridescenceExt := &MaterialsIridescence{
		IridescenceFactor:           gltf.Float(0.8),
		IridescenceTexture:          &gltf.TextureInfo{Index: *gltf.Index(0)},
		IridescenceIor:              gltf.Float(1.5),
		IridescenceThicknessMinimum: gltf.Float(200.0),
		IridescenceThicknessMaximum: gltf.Float(600.0),
		IridescenceThicknessTexture: &gltf.TextureInfo{Index: *gltf.Index(1)},
	}

	// Add the extension to the material
	material.Extensions[ExtensionName] = iridescenceExt

	// Add material to document
	doc.Materials = append(doc.Materials, material)

	// Add the extension to the document's required extensions
	doc.AddExtensionUsed(ExtensionName)

	// Print information about the material
	fmt.Printf("Material: %s\n", material.Name)
	fmt.Printf("Iridescence Factor: %.1f\n", *iridescenceExt.IridescenceFactor)
	fmt.Printf("Iridescence Texture Index: %d\n", iridescenceExt.IridescenceTexture.Index)
	fmt.Printf("Iridescence IOR: %.1f\n", *iridescenceExt.IridescenceIor)
	fmt.Printf("Iridescence Thickness Min: %.1f\n", *iridescenceExt.IridescenceThicknessMinimum)
	fmt.Printf("Iridescence Thickness Max: %.1f\n", *iridescenceExt.IridescenceThicknessMaximum)
	fmt.Printf("Iridescence Thickness Texture Index: %d\n", iridescenceExt.IridescenceThicknessTexture.Index)

	// Output:
	// Material: TexturedIridescenceMaterial
	// Iridescence Factor: 0.8
	// Iridescence Texture Index: 0
	// Iridescence IOR: 1.5
	// Iridescence Thickness Min: 200.0
	// Iridescence Thickness Max: 600.0
	// Iridescence Thickness Texture Index: 1
}

func ExampleUnmarshal() {
	// JSON data representing a material with iridescence extension
	jsonData := []byte(`{
		"iridescenceFactor": 1.0,
		"iridescenceIor": 1.3,
		"iridescenceThicknessMaximum": 400.0
	}`)

	// Unmarshal the JSON data
	ext, err := Unmarshal(jsonData)
	if err != nil {
		log.Fatalf("Failed to unmarshal iridescence extension: %v", err)
	}

	// Type assert to MaterialsIridescence
	iridescenceExt, ok := ext.(*MaterialsIridescence)
	if !ok {
		log.Fatal("Failed to type assert to MaterialsIridescence")
	}

	fmt.Printf("Iridescence Factor: %.1f\n", *iridescenceExt.IridescenceFactor)
	fmt.Printf("Iridescence IOR: %.1f\n", *iridescenceExt.IridescenceIor)
	fmt.Printf("Iridescence Thickness Maximum: %.1f\n", *iridescenceExt.IridescenceThicknessMaximum)

	// Output:
	// Iridescence Factor: 1.0
	// Iridescence IOR: 1.3
	// Iridescence Thickness Maximum: 400.0
}