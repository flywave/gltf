package agi

import (
	"encoding/json"
	"testing"
)

func TestAgiRootArticulations(t *testing.T) {
	// Test creating and validating AgiRootArticulations
	articulations := &AgiRootArticulations{
		Articulations: []AgiArticulation{
			{
				Name: "test-articulation",
				Stages: []AgiArticulationStage{
					{
						Name:          "test-stage",
						TransformType: AgiArticulationTransformTypeXTranslate,
						MinimumValue:  -1.0,
						InitialValue:  0.0,
						MaximumValue:  1.0,
					},
				},
			},
		},
	}

	// Test JSON marshaling and unmarshaling
	data, err := json.Marshal(articulations)
	if err != nil {
		t.Errorf("Failed to marshal AgiRootArticulations: %v", err)
	}

	unmarshaled, err := UnmarshalAgiRootArticulations(data)
	if err != nil {
		t.Errorf("Failed to unmarshal AgiRootArticulations: %v", err)
	}

	ext, ok := unmarshaled.(*AgiRootArticulations)
	if !ok {
		t.Error("Unmarshaled object is not of type *AgiRootArticulations")
	}

	if len(ext.Articulations) != 1 {
		t.Errorf("Expected 1 articulation, got %d", len(ext.Articulations))
	}

	if ext.Articulations[0].Name != "test-articulation" {
		t.Errorf("Expected articulation name 'test-articulation', got '%s'", ext.Articulations[0].Name)
	}

	if len(ext.Articulations[0].Stages) != 1 {
		t.Errorf("Expected 1 stage, got %d", len(ext.Articulations[0].Stages))
	}
}

func TestAgiNodeArticulations(t *testing.T) {
	// Test creating and validating AgiNodeArticulations
	name := "test-articulation"
	isAttachPoint := true
	articulations := &AgiNodeArticulations{
		ArticulationName: &name,
		IsAttachPoint:    &isAttachPoint,
	}

	// Test JSON marshaling and unmarshaling
	data, err := json.Marshal(articulations)
	if err != nil {
		t.Errorf("Failed to marshal AgiNodeArticulations: %v", err)
	}

	unmarshaled, err := UnmarshalAgiNodeArticulations(data)
	if err != nil {
		t.Errorf("Failed to unmarshal AgiNodeArticulations: %v", err)
	}

	ext, ok := unmarshaled.(*AgiNodeArticulations)
	if !ok {
		t.Error("Unmarshaled object is not of type *AgiNodeArticulations")
	}

	if ext.ArticulationName == nil {
		t.Error("Expected ArticulationName to be set")
	} else if *ext.ArticulationName != "test-articulation" {
		t.Errorf("Expected articulation name 'test-articulation', got '%s'", *ext.ArticulationName)
	}

	if ext.IsAttachPoint == nil {
		t.Error("Expected IsAttachPoint to be set")
	} else if !*ext.IsAttachPoint {
		t.Error("Expected IsAttachPoint to be true")
	}
}

func TestCreateArticulation(t *testing.T) {
	root := &AgiRootArticulations{}
	articulation := root.CreateArticulation("test-articulation")

	if articulation == nil {
		t.Error("CreateArticulation returned nil")
	}

	if articulation.Name != "test-articulation" {
		t.Errorf("Expected articulation name 'test-articulation', got '%s'", articulation.Name)
	}

	if len(root.Articulations) != 1 {
		t.Errorf("Expected 1 articulation in root, got %d", len(root.Articulations))
	}

	if root.Articulations[0].Name != "test-articulation" {
		t.Errorf("Expected articulation name 'test-articulation', got '%s'", root.Articulations[0].Name)
	}
}

func TestCreateArticulationStage(t *testing.T) {
	articulation := &AgiArticulation{
		Name:   "test-articulation",
		Stages: []AgiArticulationStage{},
	}

	stage, err := articulation.CreateArticulationStage("test-stage", AgiArticulationTransformTypeXTranslate)
	if err != nil {
		t.Errorf("CreateArticulationStage failed: %v", err)
	}

	if stage == nil {
		t.Error("CreateArticulationStage returned nil")
	}

	if stage.Name != "test-stage" {
		t.Errorf("Expected stage name 'test-stage', got '%s'", stage.Name)
	}

	if stage.TransformType != AgiArticulationTransformTypeXTranslate {
		t.Errorf("Expected transform type '%s', got '%s'", AgiArticulationTransformTypeXTranslate, stage.TransformType)
	}

	if len(articulation.Stages) != 1 {
		t.Errorf("Expected 1 stage in articulation, got %d", len(articulation.Stages))
	}
}

func TestSetValues(t *testing.T) {
	stage := &AgiArticulationStage{
		Name:          "test-stage",
		TransformType: AgiArticulationTransformTypeXTranslate,
	}

	// Test valid values
	err := stage.SetValues(-1.0, 0.0, 1.0)
	if err != nil {
		t.Errorf("SetValues failed with valid values: %v", err)
	}

	if stage.MinimumValue != -1.0 {
		t.Errorf("Expected minimum value -1.0, got %f", stage.MinimumValue)
	}

	if stage.InitialValue != 0.0 {
		t.Errorf("Expected initial value 0.0, got %f", stage.InitialValue)
	}

	if stage.MaximumValue != 1.0 {
		t.Errorf("Expected maximum value 1.0, got %f", stage.MaximumValue)
	}

	// Test invalid values (min > initial)
	err = stage.SetValues(1.0, 0.0, 2.0)
	if err == nil {
		t.Error("SetValues should have failed with invalid values (min > initial)")
	}

	// Test invalid values (initial > max)
	err = stage.SetValues(-1.0, 2.0, 1.0)
	if err == nil {
		t.Error("SetValues should have failed with invalid values (initial > max)")
	}
}

func TestSetPointingVector(t *testing.T) {
	articulation := &AgiArticulation{
		Name: "test-articulation",
		Stages: []AgiArticulationStage{
			{
				Name:          "rotation-stage",
				TransformType: AgiArticulationTransformTypeXRotate,
				MinimumValue:  -180.0,
				InitialValue:  0.0,
				MaximumValue:  180.0,
			},
		},
	}

	// Test setting a valid unit vector
	vector := &[3]float64{1.0, 0.0, 0.0}
	err := articulation.SetPointingVector(vector)
	if err != nil {
		t.Errorf("SetPointingVector failed with valid unit vector: %v", err)
	}

	if articulation.PointingVector == nil {
		t.Error("PointingVector should not be nil after setting")
	} else {
		if (*articulation.PointingVector)[0] != 1.0 ||
			(*articulation.PointingVector)[1] != 0.0 ||
			(*articulation.PointingVector)[2] != 0.0 {
			t.Error("PointingVector values do not match")
		}
	}

	// Test setting nil vector (turning off pointing)
	err = articulation.SetPointingVector(nil)
	if err != nil {
		t.Errorf("SetPointingVector failed with nil vector: %v", err)
	}

	if articulation.PointingVector != nil {
		t.Error("PointingVector should be nil after setting to nil")
	}

	// Test setting invalid vector (not unit length)
	invalidVector := &[3]float64{1.0, 1.0, 1.0}
	err = articulation.SetPointingVector(invalidVector)
	if err == nil {
		t.Error("SetPointingVector should have failed with invalid vector")
	}
}

func TestValidateAgiArticulation(t *testing.T) {
	// Test valid articulation
	articulation := &AgiArticulation{
		Name: "test-articulation",
		Stages: []AgiArticulationStage{
			{
				Name:          "test-stage",
				TransformType: AgiArticulationTransformTypeXTranslate,
				MinimumValue:  -1.0,
				InitialValue:  0.0,
				MaximumValue:  1.0,
			},
		},
	}

	err := validateAgiArticulation(articulation)
	if err != nil {
		t.Errorf("Valid articulation failed validation: %v", err)
	}

	// Test articulation with empty name
	invalidArticulation := &AgiArticulation{
		Name: "",
		Stages: []AgiArticulationStage{
			{
				Name:          "test-stage",
				TransformType: AgiArticulationTransformTypeXTranslate,
				MinimumValue:  -1.0,
				InitialValue:  0.0,
				MaximumValue:  1.0,
			},
		},
	}

	err = validateAgiArticulation(invalidArticulation)
	if err == nil {
		t.Error("Articulation with empty name should have failed validation")
	}

	// Test articulation with invalid stage
	invalidArticulation = &AgiArticulation{
		Name: "test-articulation",
		Stages: []AgiArticulationStage{
			{
				Name:          "", // Empty name
				TransformType: AgiArticulationTransformTypeXTranslate,
				MinimumValue:  -1.0,
				InitialValue:  0.0,
				MaximumValue:  1.0,
			},
		},
	}

	err = validateAgiArticulation(invalidArticulation)
	if err == nil {
		t.Error("Articulation with invalid stage should have failed validation")
	}
}

func TestValidateAgiArticulationStage(t *testing.T) {
	// Test valid stage
	stage := &AgiArticulationStage{
		Name:          "test-stage",
		TransformType: AgiArticulationTransformTypeXTranslate,
		MinimumValue:  -1.0,
		InitialValue:  0.0,
		MaximumValue:  1.0,
	}

	err := validateAgiArticulationStage(stage)
	if err != nil {
		t.Errorf("Valid stage failed validation: %v", err)
	}

	// Test stage with empty name
	invalidStage := &AgiArticulationStage{
		Name:          "",
		TransformType: AgiArticulationTransformTypeXTranslate,
		MinimumValue:  -1.0,
		InitialValue:  0.0,
		MaximumValue:  1.0,
	}

	err = validateAgiArticulationStage(invalidStage)
	if err == nil {
		t.Error("Stage with empty name should have failed validation")
	}

	// Test stage with invalid transform type
	invalidStage = &AgiArticulationStage{
		Name:          "test-stage",
		TransformType: "invalid-type",
		MinimumValue:  -1.0,
		InitialValue:  0.0,
		MaximumValue:  1.0,
	}

	err = validateAgiArticulationStage(invalidStage)
	if err == nil {
		t.Error("Stage with invalid transform type should have failed validation")
	}

	// Test stage with invalid values
	invalidStage = &AgiArticulationStage{
		Name:          "test-stage",
		TransformType: AgiArticulationTransformTypeXTranslate,
		MinimumValue:  1.0, // min > initial
		InitialValue:  0.0,
		MaximumValue:  2.0,
	}

	err = validateAgiArticulationStage(invalidStage)
	if err == nil {
		t.Error("Stage with invalid values should have failed validation")
	}
}
