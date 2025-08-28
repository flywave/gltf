package agi

import (
	"encoding/json"
	"testing"
)

func TestAgiRootStkMetadata(t *testing.T) {
	// Test creating and validating AgiRootStkMetadata
	metadata := &AgiRootStkMetadata{
		SolarPanelGroups: []AgiStkSolarPanelGroup{
			{
				Name:       "test-group",
				Efficiency: 0.8,
			},
		},
	}

	// Test JSON marshaling and unmarshaling
	data, err := json.Marshal(metadata)
	if err != nil {
		t.Errorf("Failed to marshal AgiRootStkMetadata: %v", err)
	}

	unmarshaled, err := UnmarshalAgiRootStkMetadata(data)
	if err != nil {
		t.Errorf("Failed to unmarshal AgiRootStkMetadata: %v", err)
	}

	ext, ok := unmarshaled.(*AgiRootStkMetadata)
	if !ok {
		t.Error("Unmarshaled object is not of type *AgiRootStkMetadata")
	}

	if len(ext.SolarPanelGroups) != 1 {
		t.Errorf("Expected 1 solar panel group, got %d", len(ext.SolarPanelGroups))
	}

	if ext.SolarPanelGroups[0].Name != "test-group" {
		t.Errorf("Expected solar panel group name 'test-group', got '%s'", ext.SolarPanelGroups[0].Name)
	}

	if ext.SolarPanelGroups[0].Efficiency != 0.8 {
		t.Errorf("Expected efficiency 0.8, got %f", ext.SolarPanelGroups[0].Efficiency)
	}
}

func TestAgiNodeStkMetadata(t *testing.T) {
	// Test creating and validating AgiNodeStkMetadata
	name := "test-group"
	noObscuration := true
	metadata := &AgiNodeStkMetadata{
		SolarPanelGroupName: &name,
		NoObscuration:       &noObscuration,
	}

	// Test JSON marshaling and unmarshaling
	data, err := json.Marshal(metadata)
	if err != nil {
		t.Errorf("Failed to marshal AgiNodeStkMetadata: %v", err)
	}

	unmarshaled, err := UnmarshalAgiNodeStkMetadata(data)
	if err != nil {
		t.Errorf("Failed to unmarshal AgiNodeStkMetadata: %v", err)
	}

	ext, ok := unmarshaled.(*AgiNodeStkMetadata)
	if !ok {
		t.Error("Unmarshaled object is not of type *AgiNodeStkMetadata")
	}

	if ext.SolarPanelGroupName == nil {
		t.Error("Expected SolarPanelGroupName to be set")
	} else if *ext.SolarPanelGroupName != "test-group" {
		t.Errorf("Expected solar panel group name 'test-group', got '%s'", *ext.SolarPanelGroupName)
	}

	if ext.NoObscuration == nil {
		t.Error("Expected NoObscuration to be set")
	} else if !*ext.NoObscuration {
		t.Error("Expected NoObscuration to be true")
	}
}

func TestCreateSolarPanelGroup(t *testing.T) {
	root := &AgiRootStkMetadata{}
	group := root.CreateSolarPanelGroup("test-group")

	if group == nil {
		t.Error("CreateSolarPanelGroup returned nil")
	}

	if group.Name != "test-group" {
		t.Errorf("Expected group name 'test-group', got '%s'", group.Name)
	}

	if len(root.SolarPanelGroups) != 1 {
		t.Errorf("Expected 1 solar panel group in root, got %d", len(root.SolarPanelGroups))
	}

	if root.SolarPanelGroups[0].Name != "test-group" {
		t.Errorf("Expected group name 'test-group', got '%s'", root.SolarPanelGroups[0].Name)
	}
}

func TestSetEfficiency(t *testing.T) {
	group := &AgiStkSolarPanelGroup{
		Name: "test-group",
	}

	// Test valid efficiency
	err := group.SetEfficiency(0.8)
	if err != nil {
		t.Errorf("SetEfficiency failed with valid value: %v", err)
	}

	if group.Efficiency != 0.8 {
		t.Errorf("Expected efficiency 0.8, got %f", group.Efficiency)
	}

	// Test invalid efficiency (less than 0)
	err = group.SetEfficiency(-0.1)
	if err == nil {
		t.Error("SetEfficiency should have failed with negative value")
	}

	// Test invalid efficiency (greater than 1)
	err = group.SetEfficiency(1.1)
	if err == nil {
		t.Error("SetEfficiency should have failed with value greater than 1")
	}
}

func TestSetNoObscuration(t *testing.T) {
	metadata := &AgiNodeStkMetadata{}

	// Test setting to true
	metadata.SetNoObscuration(true)
	if metadata.NoObscuration == nil {
		t.Error("NoObscuration should not be nil after setting")
	} else if !*metadata.NoObscuration {
		t.Error("NoObscuration should be true")
	}

	// Test setting to false
	metadata.SetNoObscuration(false)
	if metadata.NoObscuration == nil {
		t.Error("NoObscuration should not be nil after setting")
	} else if *metadata.NoObscuration {
		t.Error("NoObscuration should be false")
	}
}

func TestGetNoObscuration(t *testing.T) {
	metadata := &AgiNodeStkMetadata{}

	// Test default value (should be false)
	if metadata.GetNoObscuration() {
		t.Error("Default value of NoObscuration should be false")
	}

	// Test when set to true
	trueValue := true
	metadata.NoObscuration = &trueValue
	if !metadata.GetNoObscuration() {
		t.Error("GetNoObscuration should return true when set to true")
	}

	// Test when set to false
	falseValue := false
	metadata.NoObscuration = &falseValue
	if metadata.GetNoObscuration() {
		t.Error("GetNoObscuration should return false when set to false")
	}
}

func TestValidateAgiStkSolarPanelGroup(t *testing.T) {
	// Test valid group
	group := &AgiStkSolarPanelGroup{
		Name:       "test-group",
		Efficiency: 0.8,
	}

	err := validateAgiStkSolarPanelGroup(group)
	if err != nil {
		t.Errorf("Valid group failed validation: %v", err)
	}

	// Test group with empty name
	invalidGroup := &AgiStkSolarPanelGroup{
		Name:       "",
		Efficiency: 0.8,
	}

	err = validateAgiStkSolarPanelGroup(invalidGroup)
	if err == nil {
		t.Error("Group with empty name should have failed validation")
	}

	// Test group with invalid efficiency (less than 0)
	invalidGroup = &AgiStkSolarPanelGroup{
		Name:       "test-group",
		Efficiency: -0.1,
	}

	err = validateAgiStkSolarPanelGroup(invalidGroup)
	if err == nil {
		t.Error("Group with negative efficiency should have failed validation")
	}

	// Test group with invalid efficiency (greater than 1)
	invalidGroup = &AgiStkSolarPanelGroup{
		Name:       "test-group",
		Efficiency: 1.1,
	}

	err = validateAgiStkSolarPanelGroup(invalidGroup)
	if err == nil {
		t.Error("Group with efficiency greater than 1 should have failed validation")
	}
}
