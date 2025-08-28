# EXT_texture_webp Extension for glTF

This package provides a Go implementation of the EXT_texture_webp extension for glTF models.

## Overview

The EXT_texture_webp extension allows glTF models to specify textures using the WebP image format. This extension is used in the Texture object to reference a WebP image in the images array.

## Installation

```bash
go get github.com/flywave/gltf/ext/webp
```

## Usage

### Creating EXT_texture_webp Extension

```go
import (
    "github.com/flywave/gltf"
    "github.com/flywave/gltf/ext/webp"
)

// Create a new glTF document
doc := gltf.NewDocument()

// Create an image with WebP format
image := &gltf.Image{
    URI: "texture.webp",
}
doc.Images = append(doc.Images, image)

// Create a texture
texture := &gltf.Texture{
    Source: gltf.Uint32(0), // Index of the image in the images array
    Extensions: make(gltf.Extensions),
}

// Create EXT_texture_webp extension
ext := &webp.ExtTextureWebp{}
ext.SetSource(0) // Index of the WebP image in the images array

// Add the extension to the texture
extData, err := json.Marshal(ext)
if err != nil {
    panic(err)
}

texture.Extensions[webp.TextureWebpExtensionName] = extData
doc.AddExtensionUsed(webp.TextureWebpExtensionName)

// Add texture to document
doc.Textures = append(doc.Textures, texture)
```

### Accessing EXT_texture_webp Extension

```go
// Assuming you have a glTF document with the extension
texture := doc.Textures[0]

// Get the extension data
extData, exists := texture.Extensions[webp.TextureWebpExtensionName]
if !exists {
    // Extension not present
    return
}

// Unmarshal the extension data
ext, err := webp.UnmarshalExtTextureWebp(extData.([]byte))
if err != nil {
    panic(err)
}

webpExt := ext.(*webp.ExtTextureWebp)

// Access the source
if webpExt.GetSource() != nil {
    fmt.Printf("WebP image source index: %d\n", *webpExt.GetSource())
}
```

## API Reference

### ExtTextureWebp

The main extension struct with the following field:
- `Source` - The index of the images node which points to a WebP image

### Methods

- `SetSource(source uint32)` - Sets the source image index
- `GetSource() *uint32` - Returns the source image index

## Testing

To run the tests:

```bash
go test ./ext/webp/... -v
```

## Implementation Status

This implementation fully supports the EXT_texture_webp extension as specified:

- [x] EXT_texture_webp texture extension
- [x] Source property
- [x] JSON serialization/deserialization
- [x] Complete unit test coverage

## References

- [EXT_texture_webp specification](https://github.com/KhronosGroup/glTF/tree/master/extensions/2.0/Vendor/EXT_texture_webp)