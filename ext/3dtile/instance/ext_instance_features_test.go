package instance

import (
	"testing"

	"github.com/flywave/gltf"
)

func TestInstanceFeatures(t *testing.T) {
	// Test creating instance feature ID with attribute
	attribute := uint32(0)
	featureID := CreateInstanceFeatureID(100, WithInstanceAttribute(attribute))
	if featureID.Attribute == nil || *featureID.Attribute != attribute {
		t.Error("Failed to set attribute for instance feature ID")
	}
	if featureID.FeatureCount != 100 {
		t.Error("Failed to set feature count for instance feature ID")
	}

	// Test creating instance feature ID with property table
	propertyTable := uint32(1)
	featureID = CreateInstanceFeatureID(200, WithInstancePropertyTable(propertyTable))
	if featureID.PropertyTable == nil || *featureID.PropertyTable != propertyTable {
		t.Error("Failed to set property table for instance feature ID")
	}
	if featureID.FeatureCount != 200 {
		t.Error("Failed to set feature count for instance feature ID")
	}

	// Test creating instance feature ID with label
	label := "test_label"
	featureID = CreateInstanceFeatureID(150, WithInstanceLabel(label))
	if featureID.Label == nil || *featureID.Label != label {
		t.Error("Failed to set label for instance feature ID")
	}
	if featureID.FeatureCount != 150 {
		t.Error("Failed to set feature count for instance feature ID")
	}

	// Test creating instance feature ID with null feature ID
	nullFeatureID := uint32(999)
	featureID = CreateInstanceFeatureID(120, WithInstanceNullFeatureID(nullFeatureID))
	if featureID.NullFeatureID == nil || *featureID.NullFeatureID != nullFeatureID {
		t.Error("Failed to set null feature ID for instance feature ID")
	}
	if featureID.FeatureCount != 120 {
		t.Error("Failed to set feature count for instance feature ID")
	}
}

func TestValidateInstanceFeatureID(t *testing.T) {
	// Test valid feature ID with attribute
	attribute := uint32(0)
	featureID := CreateInstanceFeatureID(100, WithInstanceAttribute(attribute))
	err := ValidateInstanceFeatureID(featureID)
	if err != nil {
		t.Errorf("Valid instance feature ID should pass validation: %v", err)
	}

	// Test valid feature ID with property table
	propertyTable := uint32(1)
	featureID = CreateInstanceFeatureID(100, WithInstancePropertyTable(propertyTable))
	err = ValidateInstanceFeatureID(featureID)
	if err != nil {
		t.Errorf("Valid instance feature ID should pass validation: %v", err)
	}

	// Test invalid feature ID with both attribute and property table
	featureID = CreateInstanceFeatureID(100, WithInstanceAttribute(0), WithInstancePropertyTable(1))
	err = ValidateInstanceFeatureID(featureID)
	if err == nil {
		t.Error("Instance feature ID with both attribute and property table should fail validation")
	}

	// Test invalid feature ID with neither attribute nor property table
	featureID = CreateInstanceFeatureID(100)
	err = ValidateInstanceFeatureID(featureID)
	if err == nil {
		t.Error("Instance feature ID with neither attribute nor property table should fail validation")
	}

	// Test invalid feature ID with zero feature count
	featureID = CreateInstanceFeatureID(0, WithInstanceAttribute(0))
	err = ValidateInstanceFeatureID(featureID)
	if err == nil {
		t.Error("Instance feature ID with zero feature count should fail validation")
	}
}

func TestSetInstanceFeatures(t *testing.T) {
	// Test setting instance features on a node
	node := &gltf.Node{
		Extensions: make(gltf.Extensions),
	}

	featureIDs := []MeshExtInstanceFeatureID{
		CreateInstanceFeatureID(100, WithInstanceAttribute(0)),
		CreateInstanceFeatureID(200, WithInstancePropertyTable(1)),
	}

	err := SetInstanceFeatures(node, featureIDs)
	if err != nil {
		t.Errorf("SetInstanceFeatures failed: %v", err)
	}

	// Test getting instance features from a node
	ext, err := GetInstanceFeatures(node)
	if err != nil {
		t.Errorf("GetInstanceFeatures failed: %v", err)
	}
	if ext == nil {
		t.Error("GetInstanceFeatures returned nil extension")
	}
	if len(ext.FeatureIDs) != 2 {
		t.Errorf("Expected 2 feature IDs, got %d", len(ext.FeatureIDs))
	}
}
