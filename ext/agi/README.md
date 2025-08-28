# AGI Extensions for glTF

This package provides Go implementations of the AGI (Analytical Graphics, Inc.) extensions for glTF models.

## Extensions Included

1. **AGI_articulations** - Defines articulations (movable parts) of a model
2. **AGI_stk_metadata** - Provides metadata for use with STK (Systems Tool Kit)

## Installation

```bash
go get github.com/flywave/gltf/ext/agi
```

## Usage

### Register Extensions

Before using the extensions, you need to register them:

```go
import "github.com/flywave/gltf/ext/agi"

// Register the AGI extensions
agi.RegisterExtensions()
```

### AGI_articulations

#### Root Level Extension

The root level extension defines the articulations available in the model:

```go
// Create root articulations extension
rootArticulations := &agi.AgiRootArticulations{}

// Create an articulation
articulation := rootArticulations.CreateArticulation("SampleArticulation")

// Set a pointing vector (unit vector)
vector := &[3]float64{1.0, 0.0, 0.0}
articulation.SetPointingVector(vector)

// Create an articulation stage
stage, err := articulation.CreateArticulationStage("RotationStage", agi.AgiArticulationTransformTypeXRotate)
if err != nil {
    panic(err)
}

// Set stage values
stage.SetValues(-180.0, 0.0, 180.0)

// Add the extension to the document
extData, err := json.Marshal(rootArticulations)
if err != nil {
    panic(err)
}

doc.Extensions[agi.ArticulationsExtensionName] = extData
doc.AddExtensionUsed(agi.ArticulationsExtensionName)
```

#### Node Level Extension

The node level extension references articulations defined at the root level:

```go
// Create node articulations extension
nodeArticulations := &agi.AgiNodeArticulations{
    ArticulationName: stringPtr("SampleArticulation"),
    IsAttachPoint:    boolPtr(true),
}

// Add the extension to the node
extData, err := json.Marshal(nodeArticulations)
if err != nil {
    panic(err)
}

node.Extensions[agi.ArticulationsExtensionName] = extData
```

#### Available Transform Types

The following transform types are available for articulation stages:

- `xTranslate` - Translation along the node's X axis
- `yTranslate` - Translation along the node's Y axis
- `zTranslate` - Translation along the node's Z axis
- `xRotate` - Rotation in degrees about the node's +X axis
- `yRotate` - Rotation in degrees about the node's +Y axis
- `zRotate` - Rotation in degrees about the node's +Z axis
- `xScale` - Scaling factor along the node's X axis
- `yScale` - Scaling factor along the node's Y axis
- `zScale` - Scaling factor along the node's Z axis
- `uniformScale` - Uniform scaling factor to apply to the node

### AGI_stk_metadata

#### Root Level Extension

The root level extension defines solar panel groups:

```go
// Create root STK metadata extension
rootStkMetadata := &agi.AgiRootStkMetadata{}

// Create a solar panel group
group := rootStkMetadata.CreateSolarPanelGroup("SampleSolarPanelGroup")
group.SetEfficiency(0.85)

// Add the extension to the document
extData, err := json.Marshal(rootStkMetadata)
if err != nil {
    panic(err)
}

doc.Extensions[agi.StkMetadataExtensionName] = extData
doc.AddExtensionUsed(agi.StkMetadataExtensionName)
```

#### Node Level Extension

The node level extension references solar panel groups defined at the root level:

```go
// Create node STK metadata extension
nodeStkMetadata := &agi.AgiNodeStkMetadata{
    SolarPanelGroupName: stringPtr("SampleSolarPanelGroup"),
}

// Set no obscuration flag
nodeStkMetadata.SetNoObscuration(true)

// Add the extension to the node
extData, err := json.Marshal(nodeStkMetadata)
if err != nil {
    panic(err)
}

node.Extensions[agi.StkMetadataExtensionName] = extData
```

## Testing

To run the tests:

```bash
go test ./ext/agi/... -v
```

## Implementation Status

This implementation fully supports the AGI extensions as specified:

- [x] AGI_articulations root extension
- [x] AGI_articulations node extension
- [x] Articulation stages with all transform types
- [x] Pointing vectors
- [x] Attach points
- [x] AGI_stk_metadata root extension
- [x] AGI_stk_metadata node extension
- [x] Solar panel groups
- [x] No obscuration flag

## References

- [AGI_articulations specification](https://github.com/KhronosGroup/glTF/tree/master/extensions/2.0/Vendor/AGI_articulations)
- [AGI_stk_metadata specification](https://github.com/KhronosGroup/glTF/tree/master/extensions/2.0/Vendor/AGI_stk_metadata)