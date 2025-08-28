package bimdata

import (
	"encoding/json"
	"fmt"

	"github.com/flywave/gltf"
)

const (
	// BimDataExtensionName is the name of the GRIFFEL_bim_data extension
	BimDataExtensionName = "GRIFFEL_bim_data"
)

// BimData represents the GRIFFEL_bim_data glTF Node extension
type BimData struct {
	Properties []uint32                   `json:"properties,omitempty"`
	Type       *uint32                    `json:"type,omitempty"`
	BufferView *uint32                    `json:"bufferView,omitempty"`
	Extensions map[string]json.RawMessage `json:"extensions,omitempty"`
	Extras     json.RawMessage            `json:"extras,omitempty"`
}

// BimDataRoot represents the root GRIFFEL_bim_data extension
type BimDataRoot struct {
	PropertyNames  []string                   `json:"propertyNames"`
	PropertyValues []string                   `json:"propertyValues"`
	Properties     []BimProperty              `json:"properties"`
	Types          []BimType                  `json:"types"`
	NodeProperties []NodePropertyMapping      `json:"nodeProperties,omitempty"`
	Extensions     map[string]json.RawMessage `json:"extensions,omitempty"`
	Extras         json.RawMessage            `json:"extras,omitempty"`
}

// BimProperty represents a single property with name and value indices
type BimProperty struct {
	Name  uint32 `json:"name"`
	Value uint32 `json:"value"`
}

// BimType represents a type with common properties
type BimType struct {
	Properties []uint32 `json:"properties"`
}

// NodePropertyMapping maps node properties and types to nodes
type NodePropertyMapping struct {
	Node       uint32   `json:"node"`
	Properties []uint32 `json:"properties"`
	Type       *uint32  `json:"type,omitempty"`
}

// UnmarshalBimData unmarshals the GRIFFEL_bim_data extension data for node level
func UnmarshalBimData(data []byte) (interface{}, error) {
	var ext BimData
	if err := json.Unmarshal(data, &ext); err != nil {
		return nil, fmt.Errorf("GRIFFEL_bim_data node parsing failed: %w", err)
	}

	return &ext, nil
}

// UnmarshalBimDataRoot unmarshals the GRIFFEL_bim_data extension data for root level
func UnmarshalBimDataRoot(data []byte) (interface{}, error) {
	var ext BimDataRoot
	if err := json.Unmarshal(data, &ext); err != nil {
		return nil, fmt.Errorf("GRIFFEL_bim_data root parsing failed: %w", err)
	}

	return &ext, nil
}

// SetPropertyIndices sets the property indices for a node
func (e *BimData) SetPropertyIndices(indices []uint32) {
	e.Properties = indices
}

// SetTypeIndex sets the type index for a node
func (e *BimData) SetTypeIndex(index uint32) {
	e.Type = &index
}

// SetBufferViewIndex sets the buffer view index for a node
func (e *BimData) SetBufferViewIndex(index uint32) {
	e.BufferView = &index
}

// GetPropertyIndices returns the property indices for a node
func (e *BimData) GetPropertyIndices() []uint32 {
	return e.Properties
}

// GetTypeIndex returns the type index for a node
func (e *BimData) GetTypeIndex() *uint32 {
	return e.Type
}

// GetBufferViewIndex returns the buffer view index for a node
func (e *BimData) GetBufferViewIndex() *uint32 {
	return e.BufferView
}

// AddPropertyName adds a property name to the root extension
func (e *BimDataRoot) AddPropertyName(name string) {
	e.PropertyNames = append(e.PropertyNames, name)
}

// AddPropertyValue adds a property value to the root extension
func (e *BimDataRoot) AddPropertyValue(value string) {
	e.PropertyValues = append(e.PropertyValues, value)
}

// AddProperty adds a property to the root extension
func (e *BimDataRoot) AddProperty(nameIndex, valueIndex uint32) {
	e.Properties = append(e.Properties, BimProperty{Name: nameIndex, Value: valueIndex})
}

// AddType adds a type to the root extension
func (e *BimDataRoot) AddType(propertyIndices []uint32) {
	e.Types = append(e.Types, BimType{Properties: propertyIndices})
}

// AddNodePropertyMapping adds a node property mapping to the root extension
func (e *BimDataRoot) AddNodePropertyMapping(nodeIndex uint32, propertyIndices []uint32, typeIndex *uint32) {
	e.NodeProperties = append(e.NodeProperties, NodePropertyMapping{
		Node:       nodeIndex,
		Properties: propertyIndices,
		Type:       typeIndex,
	})
}

func init() {
	gltf.RegisterExtension(BimDataExtensionName, UnmarshalBimData)
	gltf.RegisterExtension(BimDataExtensionName, UnmarshalBimDataRoot)
}
