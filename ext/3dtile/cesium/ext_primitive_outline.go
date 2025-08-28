package cesium

import (
	"encoding/json"
	"fmt"

	"github.com/flywave/gltf"
)

const (
	// ExtensionName is the name of the Cesium primitive outline extension
	ExtensionName = "CESIUM_primitive_outline"
)

func init() {
	gltf.RegisterExtension(ExtensionName, UnmarshalCesiumPrimitiveOutline)
}

// CesiumPrimitiveOutline represents the CESIUM_primitive_outline extension
type CesiumPrimitiveOutline struct {
	Indices    *uint32                    `json:"indices,omitempty"`
	Extensions map[string]json.RawMessage `json:"extensions,omitempty"`
	Extras     json.RawMessage            `json:"extras,omitempty"`
}

// UnmarshalCesiumPrimitiveOutline unmarshals the CESIUM_primitive_outline extension data
func UnmarshalCesiumPrimitiveOutline(data []byte) (interface{}, error) {
	var ext CesiumPrimitiveOutline
	if err := json.Unmarshal(data, &ext); err != nil {
		return nil, fmt.Errorf("CESIUM_primitive_outline parsing failed: %w", err)
	}
	return ext, nil
}

// SetCesiumOutline sets the Cesium outline vertex indices for a primitive
func SetCesiumOutline(primitive *gltf.Primitive, indices []uint32, accessorName string) error {
	if primitive.Extensions == nil {
		primitive.Extensions = make(gltf.Extensions)
	}

	// Create accessor for the indices
	// Note: In a full implementation, you would create a buffer view and accessor
	// For now, we'll just create a placeholder

	// Create the extension
	ext := CesiumPrimitiveOutline{}

	// Add the extension to the primitive
	extData, err := json.Marshal(ext)
	if err != nil {
		return fmt.Errorf("error marshaling CESIUM_primitive_outline extension: %w", err)
	}

	primitive.Extensions[ExtensionName] = extData
	// Note: ExtensionUsed should be added to the document, not the primitive
	return nil
}

// GetCesiumOutline gets the Cesium outline extension from a primitive
func GetCesiumOutline(primitive *gltf.Primitive) (*CesiumPrimitiveOutline, error) {
	if primitive.Extensions == nil {
		return nil, fmt.Errorf("no extensions found")
	}

	extData, exists := primitive.Extensions[ExtensionName]
	if !exists {
		return nil, fmt.Errorf("%s extension not found", ExtensionName)
	}

	// Type assertion
	extDataBytes, ok := extData.([]byte)
	if !ok {
		return nil, fmt.Errorf("extension data is not in expected format ([]byte)")
	}

	var ext CesiumPrimitiveOutline
	if err := json.Unmarshal(extDataBytes, &ext); err != nil {
		return nil, fmt.Errorf("error unmarshaling CESIUM_primitive_outline extension: %w", err)
	}

	return &ext, nil
}

// ValidateCesiumOutlineIndices validates that all indices are within the range of the mesh primitive indices
func ValidateCesiumOutlineIndices(outlineIndices []uint32, primitiveIndices []uint32) bool {
	if len(primitiveIndices) == 0 {
		return false
	}

	// Find max index in primitive indices
	maxIndex := uint32(0)
	for _, idx := range primitiveIndices {
		if idx > maxIndex {
			maxIndex = idx
		}
	}

	// Check that all outline indices are within range
	for _, idx := range outlineIndices {
		if idx > maxIndex {
			return false
		}
	}

	return true
}

// ValidateAccessor validates that the accessor meets the requirements for Cesium outline indices
func ValidateAccessor(accessor *gltf.Accessor) error {
	if accessor == nil {
		return fmt.Errorf("accessor is nil")
	}

	// Check component type - should be unsigned int
	if accessor.ComponentType != gltf.ComponentUint {
		return fmt.Errorf("accessor component type must be UNSIGNED_INT")
	}

	// Check dimensions - should be scalar
	if accessor.Type != gltf.AccessorScalar {
		return fmt.Errorf("accessor dimensions must be SCALAR")
	}

	// Check that it's not normalized
	if accessor.Normalized {
		return fmt.Errorf("accessor must not be normalized")
	}

	return nil
}
