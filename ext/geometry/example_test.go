package geometry

import (
	"encoding/json"
	"fmt"

	"github.com/flywave/gltf"
)

func ExampleFbGeometryMetadata() {
	// Create a new glTF document
	doc := gltf.NewDocument()

	// Create a scene
	scene := &gltf.Scene{
		Name:       "SampleScene",
		Extensions: make(gltf.Extensions),
	}

	// Create FB_geometry_metadata extension
	metadata := &FbGeometryMetadata{}
	metadata.SetVertexCount(1000.0)
	metadata.SetPrimitiveCount(10.0)

	// Create scene bounds
	bounds := SceneBounds{
		Min: []float64{-10.0, -5.0, -2.0},
		Max: []float64{10.0, 5.0, 2.0},
	}
	metadata.SetSceneBounds(bounds)

	// Add the extension to the scene
	extData, err := json.Marshal(metadata)
	if err != nil {
		panic(err)
	}

	scene.Extensions[FbGeometryMetadataExtensionName] = extData
	doc.AddExtensionUsed(FbGeometryMetadataExtensionName)

	// Add scene to document
	doc.Scenes = append(doc.Scenes, scene)

	fmt.Printf("Scene vertex count: %.0f\n", *metadata.GetVertexCount())
	fmt.Printf("Scene primitive count: %.0f\n", *metadata.GetPrimitiveCount())
	fmt.Printf("Scene bounds min: [%.1f, %.1f, %.1f]\n", metadata.GetSceneBounds().Min[0], metadata.GetSceneBounds().Min[1], metadata.GetSceneBounds().Min[2])
	fmt.Printf("Scene bounds max: [%.1f, %.1f, %.1f]\n", metadata.GetSceneBounds().Max[0], metadata.GetSceneBounds().Max[1], metadata.GetSceneBounds().Max[2])

	// Output:
	// Scene vertex count: 1000
	// Scene primitive count: 10
	// Scene bounds min: [-10.0, -5.0, -2.0]
	// Scene bounds max: [10.0, 5.0, 2.0]
}
