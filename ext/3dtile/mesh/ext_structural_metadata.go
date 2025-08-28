package mesh

import (
	"encoding/json"
	"fmt"

	"github.com/flywave/gltf"
)

const (
	// StructuralMetadataExtensionName is the name of the EXT_structural_metadata extension
	StructuralMetadataExtensionName = "EXT_structural_metadata"
)

// ExtStructuralMetadata represents the EXT_structural_metadata glTF Mesh Primitive extension
type ExtStructuralMetadata struct {
	PropertyTextures   []uint32                   `json:"propertyTextures,omitempty"`
	PropertyAttributes []uint32                   `json:"propertyAttributes,omitempty"`
	Extensions         map[string]json.RawMessage `json:"extensions,omitempty"`
	Extras             json.RawMessage            `json:"extras,omitempty"`
}

// UnmarshalExtStructuralMetadata unmarshals the EXT_structural_metadata extension data
func UnmarshalExtStructuralMetadata(data []byte) (interface{}, error) {
	var ext ExtStructuralMetadata
	if err := json.Unmarshal(data, &ext); err != nil {
		return nil, fmt.Errorf("EXT_structural_metadata parsing failed: %w", err)
	}

	// Validate the extension data
	if err := validateExtStructuralMetadata(ext); err != nil {
		return nil, fmt.Errorf("EXT_structural_metadata validation failed: %w", err)
	}

	return ext, nil
}

// validateExtStructuralMetadata validates the EXT_structural_metadata extension data
func validateExtStructuralMetadata(ext ExtStructuralMetadata) error {
	// At least one of propertyTextures or propertyAttributes must be present
	if len(ext.PropertyTextures) == 0 && len(ext.PropertyAttributes) == 0 {
		return fmt.Errorf("at least one property texture or property attribute must be present")
	}

	return nil
}

// AddPropertyTexture adds a property texture to the extension
func (e *ExtStructuralMetadata) AddPropertyTexture(textureIndex uint32) {
	e.PropertyTextures = append(e.PropertyTextures, textureIndex)
}

// AddPropertyAttribute adds a property attribute to the extension
func (e *ExtStructuralMetadata) AddPropertyAttribute(attributeIndex uint32) {
	e.PropertyAttributes = append(e.PropertyAttributes, attributeIndex)
}

// SetPropertyTextures sets the property textures
func (e *ExtStructuralMetadata) SetPropertyTextures(textures []uint32) {
	e.PropertyTextures = textures
}

// SetPropertyAttributes sets the property attributes
func (e *ExtStructuralMetadata) SetPropertyAttributes(attributes []uint32) {
	e.PropertyAttributes = attributes
}

// GetPropertyTextures returns the property textures
func (e *ExtStructuralMetadata) GetPropertyTextures() []uint32 {
	return e.PropertyTextures
}

// GetPropertyAttributes returns the property attributes
func (e *ExtStructuralMetadata) GetPropertyAttributes() []uint32 {
	return e.PropertyAttributes
}

func init() {
	gltf.RegisterExtension(StructuralMetadataExtensionName, UnmarshalExtStructuralMetadata)
}
