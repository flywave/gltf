package instance

import (
	"encoding/json"
	"fmt"

	"github.com/flywave/gltf"
)

const (
	// ExtensionName is the name of the EXT_instance_features extension
	InstanceFeaturesExtensionName = "EXT_instance_features"
)

func init() {
	gltf.RegisterExtension(InstanceFeaturesExtensionName, UnmarshalInstanceFeatures)
}

// MeshExtInstanceFeatureID represents a feature ID for instance features
type MeshExtInstanceFeatureID struct {
	Attribute     *uint32                    `json:"attribute,omitempty"`
	FeatureCount  uint32                     `json:"featureCount"`
	Label         *string                    `json:"label,omitempty"`
	NullFeatureID *uint32                    `json:"nullFeatureId,omitempty"`
	PropertyTable *uint32                    `json:"propertyTable,omitempty"`
	Extensions    map[string]json.RawMessage `json:"extensions,omitempty"`
	Extras        json.RawMessage            `json:"extras,omitempty"`
}

// MeshExtInstanceFeatures represents the EXT_instance_features extension
type MeshExtInstanceFeatures struct {
	FeatureIDs []MeshExtInstanceFeatureID `json:"featureIds"`
	Extensions map[string]json.RawMessage `json:"extensions,omitempty"`
	Extras     json.RawMessage            `json:"extras,omitempty"`
}

// UnmarshalInstanceFeatures unmarshals the EXT_instance_features extension data
func UnmarshalInstanceFeatures(data []byte) (interface{}, error) {
	var ext MeshExtInstanceFeatures
	if err := json.Unmarshal(data, &ext); err != nil {
		return nil, fmt.Errorf("EXT_instance_features parsing failed: %w", err)
	}

	// Validate feature IDs
	for _, featureID := range ext.FeatureIDs {
		if err := validateInstanceFeatureID(featureID); err != nil {
			return nil, fmt.Errorf("invalid instance feature ID: %w", err)
		}
	}

	return ext, nil
}

// validateInstanceFeatureID validates an instance feature ID
func validateInstanceFeatureID(featureID MeshExtInstanceFeatureID) error {
	if featureID.FeatureCount == 0 {
		return fmt.Errorf("featureCount must be greater than 0")
	}

	// Check that exactly one of attribute, propertyTable is set
	referenceMethods := 0
	if featureID.Attribute != nil {
		referenceMethods++
	}
	if featureID.PropertyTable != nil {
		referenceMethods++
	}

	if referenceMethods != 1 {
		return fmt.Errorf("exactly one of attribute or propertyTable must be set")
	}

	return nil
}

// SetInstanceFeatures sets the instance features extension on a node
func SetInstanceFeatures(node *gltf.Node, featureIDs []MeshExtInstanceFeatureID) error {
	if node.Extensions == nil {
		node.Extensions = make(gltf.Extensions)
	}

	// Validate feature IDs
	for _, featureID := range featureIDs {
		if err := validateInstanceFeatureID(featureID); err != nil {
			return fmt.Errorf("invalid instance feature ID: %w", err)
		}
	}

	// Create the extension
	ext := MeshExtInstanceFeatures{
		FeatureIDs: featureIDs,
	}

	// Add the extension to the node
	extData, err := json.Marshal(ext)
	if err != nil {
		return fmt.Errorf("error marshaling EXT_instance_features extension: %w", err)
	}

	node.Extensions[InstanceFeaturesExtensionName] = extData
	// Note: ExtensionUsed should be added to the document, not the node
	return nil
}

// GetInstanceFeatures gets the instance features extension from a node
func GetInstanceFeatures(node *gltf.Node) (*MeshExtInstanceFeatures, error) {
	if node.Extensions == nil {
		return nil, fmt.Errorf("no extensions found")
	}

	extData, exists := node.Extensions[InstanceFeaturesExtensionName]
	if !exists {
		return nil, fmt.Errorf("%s extension not found", InstanceFeaturesExtensionName)
	}

	// Type assertion
	extDataBytes, ok := extData.([]byte)
	if !ok {
		return nil, fmt.Errorf("extension data is not in expected format ([]byte)")
	}

	var ext MeshExtInstanceFeatures
	if err := json.Unmarshal(extDataBytes, &ext); err != nil {
		return nil, fmt.Errorf("error unmarshaling EXT_instance_features extension: %w", err)
	}

	return &ext, nil
}

// CreateInstanceFeatureID creates a new instance feature ID
func CreateInstanceFeatureID(featureCount uint32, options ...InstanceFeatureIDOption) MeshExtInstanceFeatureID {
	featureID := MeshExtInstanceFeatureID{
		FeatureCount: featureCount,
	}

	// Apply options
	for _, option := range options {
		option(&featureID)
	}

	return featureID
}

// InstanceFeatureIDOption is a function that configures an instance feature ID
type InstanceFeatureIDOption func(*MeshExtInstanceFeatureID)

// WithInstanceAttribute sets the attribute index for the feature ID
func WithInstanceAttribute(attribute uint32) InstanceFeatureIDOption {
	return func(f *MeshExtInstanceFeatureID) {
		f.Attribute = &attribute
	}
}

// WithInstanceLabel sets the label for the feature ID
func WithInstanceLabel(label string) InstanceFeatureIDOption {
	return func(f *MeshExtInstanceFeatureID) {
		f.Label = &label
	}
}

// WithInstanceNullFeatureID sets the null feature ID
func WithInstanceNullFeatureID(nullID uint32) InstanceFeatureIDOption {
	return func(f *MeshExtInstanceFeatureID) {
		f.NullFeatureID = &nullID
	}
}

// WithInstancePropertyTable sets the property table index for the feature ID
func WithInstancePropertyTable(table uint32) InstanceFeatureIDOption {
	return func(f *MeshExtInstanceFeatureID) {
		f.PropertyTable = &table
	}
}

// ValidateInstanceFeatureID validates an instance feature ID object
func ValidateInstanceFeatureID(featureID MeshExtInstanceFeatureID) error {
	if featureID.FeatureCount == 0 {
		return fmt.Errorf("featureCount must be greater than zero")
	}

	// Check that exactly one reference method is set
	referenceMethods := 0
	if featureID.Attribute != nil {
		referenceMethods++
	}
	if featureID.PropertyTable != nil {
		referenceMethods++
	}

	if referenceMethods == 0 {
		return fmt.Errorf("at least one reference method (attribute or propertyTable) must be set")
	}
	if referenceMethods > 1 {
		return fmt.Errorf("only one reference method (attribute or propertyTable) can be set")
	}

	return nil
}
