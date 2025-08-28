package mesh

import (
	"testing"
)

func TestMeshFeatures(t *testing.T) {
	// Test creating feature ID with attribute
	attribute := uint32(0)
	featureID := FeatureID{
		FeatureCount: 100,
		Attribute:    &attribute,
	}
	err := ValidateFeatureID(featureID)
	if err != nil {
		t.Errorf("Valid feature ID should pass validation: %v", err)
	}

	// Test creating feature ID with texture
	texture := FeatureIDTexture{
		Channels: []uint32{0},
		Index:    0,
	}
	featureID = FeatureID{
		FeatureCount: 100,
		Texture:      &texture,
	}
	err = ValidateFeatureID(featureID)
	if err != nil {
		t.Errorf("Valid feature ID should pass validation: %v", err)
	}

	// Test creating feature ID with property table
	propertyTable := uint32(1)
	featureID = FeatureID{
		FeatureCount:  100,
		PropertyTable: &propertyTable,
	}
	err = ValidateFeatureID(featureID)
	if err != nil {
		t.Errorf("Valid feature ID should pass validation: %v", err)
	}
}

func TestValidateFeatureID(t *testing.T) {
	// Test valid feature ID with attribute
	attribute := uint32(0)
	featureID := FeatureID{
		FeatureCount: 100,
		Attribute:    &attribute,
	}
	err := ValidateFeatureID(featureID)
	if err != nil {
		t.Errorf("Valid feature ID should pass validation: %v", err)
	}

	// Test valid feature ID with texture
	texture := FeatureIDTexture{
		Channels: []uint32{0},
		Index:    0,
	}
	featureID = FeatureID{
		FeatureCount: 100,
		Texture:      &texture,
	}
	err = ValidateFeatureID(featureID)
	if err != nil {
		t.Errorf("Valid feature ID should pass validation: %v", err)
	}

	// Test valid feature ID with property table
	propertyTable := uint32(1)
	featureID = FeatureID{
		FeatureCount:  100,
		PropertyTable: &propertyTable,
	}
	err = ValidateFeatureID(featureID)
	if err != nil {
		t.Errorf("Valid feature ID should pass validation: %v", err)
	}

	// Test invalid feature ID with multiple reference methods
	featureID = FeatureID{
		FeatureCount:  100,
		Attribute:     &attribute,
		PropertyTable: &propertyTable,
	}
	err = ValidateFeatureID(featureID)
	if err == nil {
		t.Error("Feature ID with multiple reference methods should fail validation")
	}

	// Test invalid feature ID with no reference methods
	featureID = FeatureID{
		FeatureCount: 100,
	}
	err = ValidateFeatureID(featureID)
	if err == nil {
		t.Error("Feature ID with no reference methods should fail validation")
	}

	// Test invalid feature ID with zero feature count
	featureID = FeatureID{
		FeatureCount: 0,
		Attribute:    &attribute,
	}
	err = ValidateFeatureID(featureID)
	if err == nil {
		t.Error("Feature ID with zero feature count should fail validation")
	}

	// Test invalid feature ID with empty texture channels
	texture = FeatureIDTexture{
		Channels: []uint32{},
		Index:    0,
	}
	featureID = FeatureID{
		FeatureCount: 100,
		Texture:      &texture,
	}
	err = ValidateFeatureID(featureID)
	if err == nil {
		t.Error("Feature ID with empty texture channels should fail validation")
	}

	// Test invalid feature ID with invalid texture channel
	texture = FeatureIDTexture{
		Channels: []uint32{5}, // Invalid channel (should be 0-3)
		Index:    0,
	}
	featureID = FeatureID{
		FeatureCount: 100,
		Texture:      &texture,
	}
	err = ValidateFeatureID(featureID)
	if err == nil {
		t.Error("Feature ID with invalid texture channel should fail validation")
	}
}

func TestUnmarshalMeshFeatures(t *testing.T) {
	// Test unmarshaling valid mesh features data
	data := []byte(`{
		"featureIds": [
			{
				"featureCount": 100,
				"attribute": 0
			},
			{
				"featureCount": 200,
				"propertyTable": 1
			}
		]
	}`)

	ext, err := UnmarshalMeshFeatures(data)
	if err != nil {
		t.Errorf("UnmarshalMeshFeatures failed: %v", err)
	}
	if ext == nil {
		t.Error("UnmarshalMeshFeatures returned nil extension")
	}

	meshFeatures, ok := ext.(ExtMeshFeatures)
	if !ok {
		t.Error("Unmarshaled extension is not of type ExtMeshFeatures")
	}
	if len(meshFeatures.FeatureIDs) != 2 {
		t.Errorf("Expected 2 feature IDs, got %d", len(meshFeatures.FeatureIDs))
	}
}

func TestDefaultChannels(t *testing.T) {
	channels := DefaultChannels()
	if len(channels) != 1 || channels[0] != 0 {
		t.Error("DefaultChannels should return [0]")
	}
}

func TestIsDefaultChannels(t *testing.T) {
	// Test with default channels
	channels := []uint32{0}
	if !IsDefaultChannels(channels) {
		t.Error("IsDefaultChannels should return true for [0]")
	}

	// Test with non-default channels
	channels = []uint32{1}
	if IsDefaultChannels(channels) {
		t.Error("IsDefaultChannels should return false for [1]")
	}

	// Test with multiple channels
	channels = []uint32{0, 1}
	if IsDefaultChannels(channels) {
		t.Error("IsDefaultChannels should return false for [0, 1]")
	}
}
