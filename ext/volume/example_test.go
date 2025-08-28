package volume

import (
	"fmt"
	"log"

	"github.com/flywave/gltf"
)

func ExampleMaterialsVolume() {
	// Create a new glTF document
	doc := gltf.NewDocument()

	// Create a material with volume extension
	material := &gltf.Material{
		Name: "VolumeMaterial",
		PBRMetallicRoughness: &gltf.PBRMetallicRoughness{
			BaseColorFactor: &[4]float32{1, 1, 1, 1},
			MetallicFactor:  gltf.Float(0),
			RoughnessFactor: gltf.Float(0.5),
		},
		Extensions: make(gltf.Extensions),
	}

	// Create volume extension data
	volumeExt := &MaterialsVolume{
		ThicknessFactor:     gltf.Float(1.0),
		AttenuationDistance: gltf.Float(0.006),
		AttenuationColor:    &[3]float32{0.5, 0.5, 0.5},
	}

	// Add the extension to the material
	material.Extensions[ExtensionName] = volumeExt

	// Add material to document
	doc.Materials = append(doc.Materials, material)

	// Add the extension to the document's required extensions
	doc.AddExtensionUsed(ExtensionName)

	// Print information about the material
	fmt.Printf("Material: %s\n", material.Name)
	fmt.Printf("Thickness Factor: %.1f\n", *volumeExt.ThicknessFactor)
	fmt.Printf("Attenuation Distance: %.3f\n", *volumeExt.AttenuationDistance)
	fmt.Printf("Attenuation Color: [%.1f, %.1f, %.1f]\n", volumeExt.AttenuationColor[0], volumeExt.AttenuationColor[1], volumeExt.AttenuationColor[2])

	// Output:
	// Material: VolumeMaterial
	// Thickness Factor: 1.0
	// Attenuation Distance: 0.006
	// Attenuation Color: [0.5, 0.5, 0.5]
}

func ExampleMaterialsVolume_withTexture() {
	// Create a new glTF document
	doc := gltf.NewDocument()

	// Create texture
	thicknessTexture := &gltf.Texture{
		Sampler: gltf.Index(0),
		Source:  gltf.Index(0),
	}

	doc.Textures = append(doc.Textures, thicknessTexture)

	// Create a material with volume extension using texture
	material := &gltf.Material{
		Name: "TexturedVolumeMaterial",
		PBRMetallicRoughness: &gltf.PBRMetallicRoughness{
			BaseColorFactor: &[4]float32{1, 1, 1, 1},
			MetallicFactor:  gltf.Float(0),
			RoughnessFactor: gltf.Float(0.3),
		},
		Extensions: make(gltf.Extensions),
	}

	// Create volume extension data with texture
	volumeExt := &MaterialsVolume{
		ThicknessFactor:     gltf.Float(1.0),
		ThicknessTexture:    &gltf.TextureInfo{Index: 0},
		AttenuationDistance: gltf.Float(0.006),
		AttenuationColor:    &[3]float32{0.5, 0.5, 0.5},
	}

	// Add the extension to the material
	material.Extensions[ExtensionName] = volumeExt

	// Add material to document
	doc.Materials = append(doc.Materials, material)

	// Add the extension to the document's required extensions
	doc.AddExtensionUsed(ExtensionName)

	// Print information about the material
	fmt.Printf("Material: %s\n", material.Name)
	fmt.Printf("Thickness Factor: %.1f\n", *volumeExt.ThicknessFactor)
	fmt.Printf("Thickness Texture Index: %d\n", volumeExt.ThicknessTexture.Index)
	fmt.Printf("Attenuation Distance: %.3f\n", *volumeExt.AttenuationDistance)
	fmt.Printf("Attenuation Color: [%.1f, %.1f, %.1f]\n", volumeExt.AttenuationColor[0], volumeExt.AttenuationColor[1], volumeExt.AttenuationColor[2])

	// Output:
	// Material: TexturedVolumeMaterial
	// Thickness Factor: 1.0
	// Thickness Texture Index: 0
	// Attenuation Distance: 0.006
	// Attenuation Color: [0.5, 0.5, 0.5]
}

func ExampleUnmarshal() {
	// JSON data representing a material with volume extension
	jsonData := []byte(`{
		"thicknessFactor": 1.0,
		"attenuationDistance": 0.006,
		"attenuationColor": [0.5, 0.5, 0.5]
	}`)

	// Unmarshal the JSON data
	ext, err := Unmarshal(jsonData)
	if err != nil {
		log.Fatalf("Failed to unmarshal volume extension: %v", err)
	}

	// Type assert to MaterialsVolume
	volumeExt, ok := ext.(*MaterialsVolume)
	if !ok {
		log.Fatal("Failed to type assert to MaterialsVolume")
	}

	fmt.Printf("Thickness Factor: %.1f\n", *volumeExt.ThicknessFactor)
	fmt.Printf("Attenuation Distance: %.3f\n", *volumeExt.AttenuationDistance)
	fmt.Printf("Attenuation Color: [%.1f, %.1f, %.1f]\n", volumeExt.AttenuationColor[0], volumeExt.AttenuationColor[1], volumeExt.AttenuationColor[2])

	// Output:
	// Thickness Factor: 1.0
	// Attenuation Distance: 0.006
	// Attenuation Color: [0.5, 0.5, 0.5]
}
