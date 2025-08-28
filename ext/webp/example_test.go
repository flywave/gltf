package webp_test

import (
	"encoding/json"
	"fmt"

	"github.com/flywave/gltf"
	"github.com/flywave/gltf/ext/webp"
)

func ExampleExtTextureWebp() {
	// Create a new glTF document
	doc := gltf.NewDocument()

	// Create an image
	image := &gltf.Image{
		URI: "texture.webp",
	}
	doc.Images = append(doc.Images, image)

	// Create a texture
	texture := &gltf.Texture{
		Source:     uint32Ptr(0),
		Extensions: make(gltf.Extensions),
	}
	doc.Textures = append(doc.Textures, texture)

	// Create EXT_texture_webp extension
	ext := &webp.ExtTextureWebp{}
	ext.SetSource(0)

	// Add the extension to the texture
	extData, err := json.Marshal(ext)
	if err != nil {
		panic(err)
	}

	texture.Extensions[webp.TextureWebpExtensionName] = extData
	doc.AddExtensionUsed(webp.TextureWebpExtensionName)

	fmt.Printf("Texture source: %d\n", *ext.GetSource())

	// Output:
	// Texture source: 0
}

// Helper function to create uint32 pointer
func uint32Ptr(v uint32) *uint32 {
	return &v
}
