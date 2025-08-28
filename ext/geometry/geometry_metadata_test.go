package geometry

import (
	"encoding/json"
	"testing"
)

func TestFbGeometryMetadata(t *testing.T) {
	// Test creating and validating FbGeometryMetadata
	metadata := &FbGeometryMetadata{
		VertexCount:    float64Ptr(1000.0),
		PrimitiveCount: float64Ptr(10.0),
		SceneBounds: &SceneBounds{
			Min: []float64{-1.0, -2.0, -3.0},
			Max: []float64{1.0, 2.0, 3.0},
		},
	}

	// Test JSON marshaling and unmarshaling
	data, err := json.Marshal(metadata)
	if err != nil {
		t.Errorf("Failed to marshal FbGeometryMetadata: %v", err)
	}

	unmarshaled, err := UnmarshalFbGeometryMetadata(data)
	if err != nil {
		t.Errorf("Failed to unmarshal FbGeometryMetadata: %v", err)
	}

	ext, ok := unmarshaled.(*FbGeometryMetadata)
	if !ok {
		t.Error("Unmarshaled object is not of type *FbGeometryMetadata")
	}

	if ext.VertexCount == nil {
		t.Error("Expected VertexCount to be set")
	} else if *ext.VertexCount != 1000.0 {
		t.Errorf("Expected vertex count 1000.0, got %f", *ext.VertexCount)
	}

	if ext.PrimitiveCount == nil {
		t.Error("Expected PrimitiveCount to be set")
	} else if *ext.PrimitiveCount != 10.0 {
		t.Errorf("Expected primitive count 10.0, got %f", *ext.PrimitiveCount)
	}

	if ext.SceneBounds == nil {
		t.Error("Expected SceneBounds to be set")
	} else {
		if len(ext.SceneBounds.Min) != 3 {
			t.Errorf("Expected scene bounds min to have 3 elements, got %d", len(ext.SceneBounds.Min))
		}
		if len(ext.SceneBounds.Max) != 3 {
			t.Errorf("Expected scene bounds max to have 3 elements, got %d", len(ext.SceneBounds.Max))
		}
	}
}

func TestSceneBoundsValidation(t *testing.T) {
	// Test valid scene bounds
	bounds := SceneBounds{
		Min: []float64{-1.0, -2.0, -3.0},
		Max: []float64{1.0, 2.0, 3.0},
	}

	err := validateSceneBounds(bounds)
	if err != nil {
		t.Errorf("Valid scene bounds failed validation: %v", err)
	}

	// Test scene bounds with missing min
	invalidBounds := SceneBounds{
		Max: []float64{1.0, 2.0, 3.0},
	}

	err = validateSceneBounds(invalidBounds)
	if err == nil {
		t.Error("Scene bounds with missing min should have failed validation")
	}

	// Test scene bounds with missing max
	invalidBounds = SceneBounds{
		Min: []float64{-1.0, -2.0, -3.0},
	}

	err = validateSceneBounds(invalidBounds)
	if err == nil {
		t.Error("Scene bounds with missing max should have failed validation")
	}

	// Test scene bounds with different element counts
	invalidBounds = SceneBounds{
		Min: []float64{-1.0, -2.0},
		Max: []float64{1.0, 2.0, 3.0},
	}

	err = validateSceneBounds(invalidBounds)
	if err == nil {
		t.Error("Scene bounds with different element counts should have failed validation")
	}

	// Test scene bounds with wrong element count
	invalidBounds = SceneBounds{
		Min: []float64{-1.0, -2.0, -3.0, -4.0},
		Max: []float64{1.0, 2.0, 3.0, 4.0},
	}

	err = validateSceneBounds(invalidBounds)
	if err == nil {
		t.Error("Scene bounds with wrong element count should have failed validation")
	}
}

func TestFbGeometryMetadataValidation(t *testing.T) {
	// Test valid metadata
	metadata := FbGeometryMetadata{
		VertexCount:    float64Ptr(1000.0),
		PrimitiveCount: float64Ptr(10.0),
		SceneBounds: &SceneBounds{
			Min: []float64{-1.0, -2.0, -3.0},
			Max: []float64{1.0, 2.0, 3.0},
		},
	}

	err := validateFbGeometryMetadata(metadata)
	if err != nil {
		t.Errorf("Valid metadata failed validation: %v", err)
	}

	// Test metadata with invalid scene bounds
	invalidMetadata := FbGeometryMetadata{
		VertexCount:    float64Ptr(1000.0),
		PrimitiveCount: float64Ptr(10.0),
		SceneBounds: &SceneBounds{
			Min: []float64{-1.0, -2.0},
			Max: []float64{1.0, 2.0, 3.0},
		},
	}

	err = validateFbGeometryMetadata(invalidMetadata)
	if err == nil {
		t.Error("Metadata with invalid scene bounds should have failed validation")
	}
}

func TestSettersAndGetters(t *testing.T) {
	metadata := &FbGeometryMetadata{}

	// Test SetVertexCount and GetVertexCount
	metadata.SetVertexCount(1000.0)
	if metadata.GetVertexCount() == nil {
		t.Error("Expected VertexCount to be set")
	} else if *metadata.GetVertexCount() != 1000.0 {
		t.Errorf("Expected vertex count 1000.0, got %f", *metadata.GetVertexCount())
	}

	// Test SetPrimitiveCount and GetPrimitiveCount
	metadata.SetPrimitiveCount(10.0)
	if metadata.GetPrimitiveCount() == nil {
		t.Error("Expected PrimitiveCount to be set")
	} else if *metadata.GetPrimitiveCount() != 10.0 {
		t.Errorf("Expected primitive count 10.0, got %f", *metadata.GetPrimitiveCount())
	}

	// Test SetSceneBounds and GetSceneBounds
	bounds := SceneBounds{
		Min: []float64{-1.0, -2.0, -3.0},
		Max: []float64{1.0, 2.0, 3.0},
	}
	metadata.SetSceneBounds(bounds)
	if metadata.GetSceneBounds() == nil {
		t.Error("Expected SceneBounds to be set")
	} else {
		if len(metadata.GetSceneBounds().Min) != 3 {
			t.Errorf("Expected scene bounds min to have 3 elements, got %d", len(metadata.GetSceneBounds().Min))
		}
		if len(metadata.GetSceneBounds().Max) != 3 {
			t.Errorf("Expected scene bounds max to have 3 elements, got %d", len(metadata.GetSceneBounds().Max))
		}
	}
}

func TestUnmarshalFbGeometryMetadataWithInvalidData(t *testing.T) {
	// Test unmarshaling with invalid scene bounds
	invalidJson := `{
		"vertexCount": 1000.0,
		"primitiveCount": 10.0,
		"sceneBounds": {
			"min": [-1.0, -2.0],
			"max": [1.0, 2.0, 3.0]
		}
	}`

	_, err := UnmarshalFbGeometryMetadata([]byte(invalidJson))
	if err == nil {
		t.Error("UnmarshalFbGeometryMetadata should have failed with invalid scene bounds")
	}
}

// Helper function for creating float64 pointers
func float64Ptr(f float64) *float64 {
	return &f
}
