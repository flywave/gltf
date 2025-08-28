package mesh

import (
	"encoding/json"
	"fmt"

	"github.com/flywave/gltf"
)

const ExtensionName = "EXT_mesh_features"

func init() {
	gltf.RegisterExtension(ExtensionName, UnmarshalMeshFeatures)
}

// ExtMeshFeatures represents the EXT_mesh_features glTF Mesh Primitive extension
type ExtMeshFeatures struct {
	FeatureIDs []FeatureID                `json:"featureIds"`
	Extensions map[string]json.RawMessage `json:"extensions,omitempty"`
	Extras     json.RawMessage            `json:"extras,omitempty"`
}

// FeatureID represents a feature ID set
type FeatureID struct {
	FeatureCount  uint32                     `json:"featureCount"`
	NullFeatureID *uint32                    `json:"nullFeatureId,omitempty"`
	Label         *string                    `json:"label,omitempty"`
	Attribute     *uint32                    `json:"attribute,omitempty"`
	Texture       *FeatureIDTexture          `json:"texture,omitempty"`
	PropertyTable *uint32                    `json:"propertyTable,omitempty"`
	Extensions    map[string]json.RawMessage `json:"extensions,omitempty"`
	Extras        json.RawMessage            `json:"extras,omitempty"`
}

// FeatureIDTexture represents a texture containing feature IDs
type FeatureIDTexture struct {
	Channels   []uint32                   `json:"channels,omitempty"`
	Index      uint32                     `json:"index"`
	TexCoord   uint32                     `json:"texCoord,omitempty"`
	Extensions map[string]json.RawMessage `json:"extensions,omitempty"`
	Extras     json.RawMessage            `json:"extras,omitempty"`
}

// UnmarshalMeshFeatures unmarshals the EXT_mesh_features extension data
func UnmarshalMeshFeatures(data []byte) (interface{}, error) {
	var ext ExtMeshFeatures
	if err := json.Unmarshal(data, &ext); err != nil {
		return nil, fmt.Errorf("EXT_mesh_features parsing failed: %w", err)
	}

	// Validate feature IDs
	for _, featureID := range ext.FeatureIDs {
		if err := validateFeatureID(featureID); err != nil {
			return nil, fmt.Errorf("invalid feature ID: %w", err)
		}
	}

	return ext, nil
}

// validateFeatureID validates a feature ID
func validateFeatureID(featureID FeatureID) error {
	if featureID.FeatureCount == 0 {
		return fmt.Errorf("featureCount must be greater than 0")
	}

	// Check that exactly one of attribute, texture, propertyTable is set
	referenceMethods := 0
	if featureID.Attribute != nil {
		referenceMethods++
	}
	if featureID.Texture != nil {
		referenceMethods++
	}
	if featureID.PropertyTable != nil {
		referenceMethods++
	}

	if referenceMethods != 1 {
		return fmt.Errorf("exactly one of attribute, texture, or propertyTable must be set")
	}

	return nil
}

// DefaultChannels returns the default channels value
func DefaultChannels() []uint32 {
	return []uint32{0}
}

// IsDefaultChannels checks if channels has the default value
func IsDefaultChannels(channels []uint32) bool {
	return len(channels) == 1 && channels[0] == 0
}

// IsDefault checks if a value equals its type's default value
func IsDefault[T comparable](value T) bool {
	var zero T
	return value == zero
}

// ValidateFeatureID validates a feature ID object
func ValidateFeatureID(featureID FeatureID) error {
	if featureID.FeatureCount == 0 {
		return fmt.Errorf("featureCount must be greater than zero")
	}

	// Check that exactly one reference method is set
	referenceMethods := 0
	if featureID.Attribute != nil {
		referenceMethods++
	}
	if featureID.Texture != nil {
		referenceMethods++
	}
	if featureID.PropertyTable != nil {
		referenceMethods++
	}

	if referenceMethods == 0 {
		return fmt.Errorf("at least one reference method (attribute, texture, or propertyTable) must be set")
	}
	if referenceMethods > 1 {
		return fmt.Errorf("only one reference method (attribute, texture, or propertyTable) can be set")
	}

	// Validate texture configuration
	if featureID.Texture != nil {
		if len(featureID.Texture.Channels) == 0 {
			return fmt.Errorf("texture channels must not be empty")
		}
		for _, channel := range featureID.Texture.Channels {
			if channel > 3 {
				return fmt.Errorf("texture channel must be between 0 and 3")
			}
		}
	}

	return nil
}
