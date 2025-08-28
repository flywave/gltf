package geometry

import (
	"encoding/json"
	"fmt"

	"github.com/flywave/gltf"
)

const (
	// FbGeometryMetadataExtensionName is the name of the FB_geometry_metadata extension
	FbGeometryMetadataExtensionName = "FB_geometry_metadata"
)

// FbGeometryMetadata represents the FB_geometry_metadata glTF Scene extension
type FbGeometryMetadata struct {
	VertexCount    *float64                   `json:"vertexCount,omitempty"`
	PrimitiveCount *float64                   `json:"primitiveCount,omitempty"`
	SceneBounds    *SceneBounds               `json:"sceneBounds,omitempty"`
	Extensions     map[string]json.RawMessage `json:"extensions,omitempty"`
	Extras         json.RawMessage            `json:"extras,omitempty"`
}

// SceneBounds represents the minimum and maximum bounding box extent
type SceneBounds struct {
	Min        []float64                  `json:"min"`
	Max        []float64                  `json:"max"`
	Extensions map[string]json.RawMessage `json:"extensions,omitempty"`
	Extras     json.RawMessage            `json:"extras,omitempty"`
}

// UnmarshalFbGeometryMetadata unmarshals the FB_geometry_metadata extension data
func UnmarshalFbGeometryMetadata(data []byte) (interface{}, error) {
	var ext FbGeometryMetadata
	if err := json.Unmarshal(data, &ext); err != nil {
		return nil, fmt.Errorf("FB_geometry_metadata parsing failed: %w", err)
	}

	// Validate the extension data
	if err := validateFbGeometryMetadata(ext); err != nil {
		return nil, fmt.Errorf("FB_geometry_metadata validation failed: %w", err)
	}

	return &ext, nil
}

// validateFbGeometryMetadata validates the FB_geometry_metadata extension data
func validateFbGeometryMetadata(ext FbGeometryMetadata) error {
	// Validate scene bounds if present
	if ext.SceneBounds != nil {
		if err := validateSceneBounds(*ext.SceneBounds); err != nil {
			return fmt.Errorf("invalid scene bounds: %w", err)
		}
	}

	return nil
}

// validateSceneBounds validates the scene bounds
func validateSceneBounds(bounds SceneBounds) error {
	// Check that min and max are present
	if bounds.Min == nil {
		return fmt.Errorf("scene bounds min is required")
	}
	if bounds.Max == nil {
		return fmt.Errorf("scene bounds max is required")
	}

	// Check that min and max have the same number of elements
	if len(bounds.Min) != len(bounds.Max) {
		return fmt.Errorf("scene bounds min and max must have the same number of elements")
	}

	// Check that min and max have exactly 3 elements
	if len(bounds.Min) != 3 || len(bounds.Max) != 3 {
		return fmt.Errorf("scene bounds min and max must have exactly 3 elements")
	}

	return nil
}

// SetVertexCount sets the vertex count
func (f *FbGeometryMetadata) SetVertexCount(count float64) {
	f.VertexCount = &count
}

// SetPrimitiveCount sets the primitive count
func (f *FbGeometryMetadata) SetPrimitiveCount(count float64) {
	f.PrimitiveCount = &count
}

// SetSceneBounds sets the scene bounds
func (f *FbGeometryMetadata) SetSceneBounds(bounds SceneBounds) {
	f.SceneBounds = &bounds
}

// GetVertexCount returns the vertex count
func (f *FbGeometryMetadata) GetVertexCount() *float64 {
	return f.VertexCount
}

// GetPrimitiveCount returns the primitive count
func (f *FbGeometryMetadata) GetPrimitiveCount() *float64 {
	return f.PrimitiveCount
}

// GetSceneBounds returns the scene bounds
func (f *FbGeometryMetadata) GetSceneBounds() *SceneBounds {
	return f.SceneBounds
}

func init() {
	gltf.RegisterExtension(FbGeometryMetadataExtensionName, UnmarshalFbGeometryMetadata)
}
