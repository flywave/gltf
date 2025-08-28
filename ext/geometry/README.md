# FB_geometry_metadata Extension for glTF

This package provides a Go implementation of the FB_geometry_metadata extension for glTF models.

## Overview

The FB_geometry_metadata extension provides geometry metadata for glTF scenes, including:
- Vertex count
- Primitive count
- Scene bounds (bounding box)

## Installation

```bash
go get github.com/flywave/gltf/ext/geometry
```

## Usage

### Creating FB_geometry_metadata Extension

```go
import (
    "github.com/flywave/gltf"
    "github.com/flywave/gltf/ext/geometry"
)

// Create a new glTF document
doc := gltf.NewDocument()

// Create a scene
scene := &gltf.Scene{
    Name:       "SampleScene",
    Extensions: make(gltf.Extensions),
}

// Create FB_geometry_metadata extension
metadata := &geometry.FbGeometryMetadata{}
metadata.SetVertexCount(1000.0)
metadata.SetPrimitiveCount(10.0)

// Create scene bounds
bounds := geometry.SceneBounds{
    Min: []float64{-10.0, -5.0, -2.0},
    Max: []float64{10.0, 5.0, 2.0},
}
metadata.SetSceneBounds(bounds)

// Add the extension to the scene
extData, err := json.Marshal(metadata)
if err != nil {
    panic(err)
}

scene.Extensions[geometry.FbGeometryMetadataExtensionName] = extData
doc.AddExtensionUsed(geometry.FbGeometryMetadataExtensionName)

// Add scene to document
doc.Scenes = append(doc.Scenes, scene)
```

### Accessing FB_geometry_metadata Extension

```go
// Assuming you have a glTF document with the extension
scene := doc.Scenes[0]

// Get the extension data
extData, exists := scene.Extensions[geometry.FbGeometryMetadataExtensionName]
if !exists {
    // Extension not present
    return
}

// Unmarshal the extension data
metadata, err := geometry.UnmarshalFbGeometryMetadata(extData.([]byte))
if err != nil {
    panic(err)
}

fbMetadata := metadata.(*geometry.FbGeometryMetadata)

// Access the metadata
if fbMetadata.GetVertexCount() != nil {
    fmt.Printf("Vertex count: %f\n", *fbMetadata.GetVertexCount())
}

if fbMetadata.GetPrimitiveCount() != nil {
    fmt.Printf("Primitive count: %f\n", *fbMetadata.GetPrimitiveCount())
}

if fbMetadata.GetSceneBounds() != nil {
    bounds := fbMetadata.GetSceneBounds()
    fmt.Printf("Scene bounds min: %v\n", bounds.Min)
    fmt.Printf("Scene bounds max: %v\n", bounds.Max)
}
```

## API Reference

### FbGeometryMetadata

The main extension struct with the following fields:
- `VertexCount` - The number of distinct vertices recursively contained in this scene
- `PrimitiveCount` - The number of distinct primitives recursively contained in this scene
- `SceneBounds` - The bounding box of this scene, in static geometry scene-space coordinates

### SceneBounds

Represents the minimum and maximum bounding box extent:
- `Min` - The bounding box corner with the numerically lowest scene-space coordinates
- `Max` - The bounding box corner with the numerically highest scene-space coordinates

### Methods

- `SetVertexCount(count float64)` - Sets the vertex count
- `SetPrimitiveCount(count float64)` - Sets the primitive count
- `SetSceneBounds(bounds SceneBounds)` - Sets the scene bounds
- `GetVertexCount() *float64` - Returns the vertex count
- `GetPrimitiveCount() *float64` - Returns the primitive count
- `GetSceneBounds() *SceneBounds` - Returns the scene bounds

## Testing

To run the tests:

```bash
go test ./ext/geometry/... -v
```

## Implementation Status

This implementation fully supports the FB_geometry_metadata extension as specified:

- [x] FB_geometry_metadata scene extension
- [x] Vertex count property
- [x] Primitive count property
- [x] Scene bounds property
- [x] JSON serialization/deserialization
- [x] Data validation
- [x] Complete unit test coverage

## References

- [FB_geometry_metadata specification](https://github.com/KhronosGroup/glTF/tree/master/extensions/2.0/Vendor/FB_geometry_metadata)