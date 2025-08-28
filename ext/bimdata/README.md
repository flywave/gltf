# GRIFFEL_bim_data Extension for glTF

This package provides a Go implementation of the GRIFFEL_bim_data extension for glTF models.

## Overview

The GRIFFEL_bim_data extension allows keeping domain specific data in relation to geometry. It's applicable when glTF standard `extras` property is not enough to store huge amount of metadata or when it's insufficient from the file size or data complexity perspective.

The extension was originally developed for architecture, engineering and construction (AEC) industry to keep Building Information Model (BIM) data, but it's suitable for any other industry which operates huge amounts of information.

## Installation

```bash
go get github.com/flywave/gltf/ext/bimdata
```

## Usage

### Creating GRIFFEL_bim_data Extension

```go
import (
    "github.com/flywave/gltf"
    "github.com/flywave/gltf/ext/bimdata"
)

// Create a new glTF document
doc := gltf.NewDocument()

// Create nodes
node1 := &gltf.Node{
    Name:       "Door 1",
    Extensions: make(gltf.Extensions),
}
doc.Nodes = append(doc.Nodes, node1)

// Create BIM data extension for nodes
nodeExt := &bimdata.ExtBimData{}
nodeExt.SetPropertyIndices([]uint32{0})
nodeExt.SetTypeIndex(0)

// Add the extension to the node
extData, err := json.Marshal(nodeExt)
if err != nil {
    panic(err)
}

node1.Extensions[bimdata.BimDataExtensionName] = extData

// Create root BIM data extension
rootExt := &bimdata.ExtBimDataRoot{
    PropertyNames:  []string{"Height", "Width", "Material"},
    PropertyValues: []string{"2,1 m", "900 mm", "Timber", "2,4 m"},
    Properties: []bimdata.BimProperty{
        {Name: 0, Value: 0},
        {Name: 1, Value: 1},
        {Name: 2, Value: 2},
        {Name: 0, Value: 3},
    },
    Types: []bimdata.BimType{
        {Properties: []uint32{1, 2}},
    },
}

// Add the root extension to the document
rootExtData, err := json.Marshal(rootExt)
if err != nil {
    panic(err)
}

if doc.Extensions == nil {
    doc.Extensions = make(gltf.Extensions)
}
doc.Extensions[bimdata.BimDataExtensionName] = rootExtData
doc.AddExtensionUsed(bimdata.BimDataExtensionName)
```

### Accessing GRIFFEL_bim_data Extension

```go
// Assuming you have a glTF document with the extension
node := doc.Nodes[0]

// Get the node extension data
nodeExtData, exists := node.Extensions[bimdata.BimDataExtensionName]
if !exists {
    // Extension not present
    return
}

// Unmarshal the node extension data
nodeExt, err := bimdata.UnmarshalExtBimData(nodeExtData.([]byte))
if err != nil {
    panic(err)
}

bimNodeExt := nodeExt.(*bimdata.ExtBimData)

// Access the node properties
fmt.Printf("Node properties: %v\n", bimNodeExt.GetPropertyIndices())
if bimNodeExt.GetTypeIndex() != nil {
    fmt.Printf("Node type: %d\n", *bimNodeExt.GetTypeIndex())
}

// Get the root extension data
rootExtData, exists := doc.Extensions[bimdata.BimDataExtensionName]
if !exists {
    // Extension not present
    return
}

// Unmarshal the root extension data
rootExt, err := bimdata.UnmarshalExtBimDataRoot(rootExtData.([]byte))
if err != nil {
    panic(err)
}

bimRootExt := rootExt.(*bimdata.ExtBimDataRoot)

// Access the root properties
fmt.Printf("Property names: %v\n", bimRootExt.PropertyNames)
fmt.Printf("Property values: %v\n", bimRootExt.PropertyValues)
fmt.Printf("Properties: %v\n", bimRootExt.Properties)
fmt.Printf("Types: %v\n", bimRootExt.Types)
```

## API Reference

### ExtBimData

The node-level extension struct with the following fields:
- `Properties` - Array of property indices
- `Type` - Type index
- `BufferView` - Buffer view index for external metadata

### ExtBimDataRoot

The root-level extension struct with the following fields:
- `PropertyNames` - Array of unique property names
- `PropertyValues` - Array of unique property values
- `Properties` - Array of properties (name-value pairs)
- `Types` - Array of types (collections of properties)
- `NodeProperties` - Mapping of node properties and types to nodes

### BimProperty

Represents a single property with name and value indices:
- `Name` - Index of the property name in the PropertyNames array
- `Value` - Index of the property value in the PropertyValues array

### BimType

Represents a type with common properties:
- `Properties` - Array of property indices

### NodePropertyMapping

Maps node properties and types to nodes:
- `Node` - Node index
- `Properties` - Array of property indices
- `Type` - Type index

### Methods

#### ExtBimData
- `SetPropertyIndices(indices []uint32)` - Sets the property indices for a node
- `SetTypeIndex(index uint32)` - Sets the type index for a node
- `SetBufferViewIndex(index uint32)` - Sets the buffer view index for a node
- `GetPropertyIndices() []uint32` - Returns the property indices for a node
- `GetTypeIndex() *uint32` - Returns the type index for a node
- `GetBufferViewIndex() *uint32` - Returns the buffer view index for a node

#### ExtBimDataRoot
- `AddPropertyName(name string)` - Adds a property name to the root extension
- `AddPropertyValue(value string)` - Adds a property value to the root extension
- `AddProperty(nameIndex, valueIndex uint32)` - Adds a property to the root extension
- `AddType(propertyIndices []uint32)` - Adds a type to the root extension
- `AddNodePropertyMapping(nodeIndex uint32, propertyIndices []uint32, typeIndex *uint32)` - Adds a node property mapping to the root extension

## Testing

To run the tests:

```bash
go test ./ext/bimdata/... -v
```

## Implementation Status

This implementation fully supports the GRIFFEL_bim_data extension as specified:

- [x] GRIFFEL_bim_data node extension
- [x] GRIFFEL_bim_data root extension
- [x] Property names and values collections
- [x] Properties collection
- [x] Types collection
- [x] Node properties mapping
- [x] Buffer view support
- [x] JSON serialization/deserialization
- [x] Complete unit test coverage

## References

- [GRIFFEL_bim_data specification](https://github.com/KhronosGroup/glTF/tree/master/extensions/2.0/Vendor/GRIFFEL_bim_data)