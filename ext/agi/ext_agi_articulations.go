package agi

import (
	"encoding/json"
	"fmt"
)

const (
	// ArticulationsExtensionName is the name of the AGI_articulations extension
	ArticulationsExtensionName = "AGI_articulations"
)

// AgiRootArticulations represents the AGI_articulations extension at the root level
type AgiRootArticulations struct {
	Articulations []AgiArticulation          `json:"articulations,omitempty"`
	Extensions    map[string]json.RawMessage `json:"extensions,omitempty"`
	Extras        json.RawMessage            `json:"extras,omitempty"`
}

// AgiNodeArticulations represents the AGI_articulations extension at the node level
type AgiNodeArticulations struct {
	ArticulationName *string                    `json:"articulationName,omitempty"`
	IsAttachPoint    *bool                      `json:"isAttachPoint,omitempty"`
	Extensions       map[string]json.RawMessage `json:"extensions,omitempty"`
	Extras           json.RawMessage            `json:"extras,omitempty"`
}

// AgiArticulation represents an articulation
type AgiArticulation struct {
	Name           string                 `json:"name"`
	PointingVector *[3]float64            `json:"pointingVector,omitempty"`
	Stages         []AgiArticulationStage `json:"stages,omitempty"`
	LogicalIndex   int                    `json:"-"` // Not serialized
}

// AgiArticulationStage represents an articulation stage
type AgiArticulationStage struct {
	Name          string  `json:"name"`
	TransformType string  `json:"type"`
	MinimumValue  float64 `json:"minimumValue"`
	InitialValue  float64 `json:"initialValue"`
	MaximumValue  float64 `json:"maximumValue"`
	LogicalIndex  int     `json:"-"` // Not serialized
}

// AgiArticulationTransformType defines the transform types
const (
	AgiArticulationTransformTypeXTranslate   = "xTranslate"
	AgiArticulationTransformTypeYTranslate   = "yTranslate"
	AgiArticulationTransformTypeZTranslate   = "zTranslate"
	AgiArticulationTransformTypeXRotate      = "xRotate"
	AgiArticulationTransformTypeYRotate      = "yRotate"
	AgiArticulationTransformTypeZRotate      = "zRotate"
	AgiArticulationTransformTypeXScale       = "xScale"
	AgiArticulationTransformTypeYScale       = "yScale"
	AgiArticulationTransformTypeZScale       = "zScale"
	AgiArticulationTransformTypeUniformScale = "uniformScale"
)

// AgiRotationTypes defines the rotation transform types
var AgiRotationTypes = []string{
	AgiArticulationTransformTypeXRotate,
	AgiArticulationTransformTypeYRotate,
	AgiArticulationTransformTypeZRotate,
}

// UnmarshalAgiRootArticulations unmarshals the AGI_articulations extension data for root level
func UnmarshalAgiRootArticulations(data []byte) (interface{}, error) {
	var ext AgiRootArticulations
	if err := json.Unmarshal(data, &ext); err != nil {
		return nil, fmt.Errorf("AGI_articulations root parsing failed: %w", err)
	}

	// Validate articulations
	for i := range ext.Articulations {
		if err := validateAgiArticulation(&ext.Articulations[i]); err != nil {
			return nil, fmt.Errorf("invalid articulation: %w", err)
		}
	}

	return &ext, nil
}

// UnmarshalAgiNodeArticulations unmarshals the AGI_articulations extension data for node level
func UnmarshalAgiNodeArticulations(data []byte) (interface{}, error) {
	var ext AgiNodeArticulations
	if err := json.Unmarshal(data, &ext); err != nil {
		return nil, fmt.Errorf("AGI_articulations node parsing failed: %w", err)
	}

	return &ext, nil
}

// validateAgiArticulation validates an articulation
func validateAgiArticulation(articulation *AgiArticulation) error {
	if articulation.Name == "" {
		return fmt.Errorf("articulation name is required")
	}

	// Validate pointing vector if present
	if articulation.PointingVector != nil {
		// Check if it's a unit-length vector
		vec := *articulation.PointingVector
		magnitude := vec[0]*vec[0] + vec[1]*vec[1] + vec[2]*vec[2]
		if magnitude < 0.999 || magnitude > 1.001 {
			return fmt.Errorf("pointingVector must be a unit-length vector")
		}

		// Check that there are exactly 1 or 2 rotation stages
		rotationStages := 0
		for _, stage := range articulation.Stages {
			for _, rotationType := range AgiRotationTypes {
				if stage.TransformType == rotationType {
					rotationStages++
					break
				}
			}
		}
		if rotationStages != 1 && rotationStages != 2 {
			return fmt.Errorf("pointingVector requires exactly 1 or exactly 2 rotation stages")
		}
	}

	// Validate stages
	for i := range articulation.Stages {
		if err := validateAgiArticulationStage(&articulation.Stages[i]); err != nil {
			return fmt.Errorf("invalid articulation stage: %w", err)
		}
	}

	// Check that there are no more than 2 rotation stages when pointing vector is present
	if articulation.PointingVector != nil {
		rotationStages := 0
		for _, stage := range articulation.Stages {
			for _, rotationType := range AgiRotationTypes {
				if stage.TransformType == rotationType {
					rotationStages++
					break
				}
			}
		}
		if rotationStages > 2 {
			return fmt.Errorf("cannot add more than 2 rotation stages when a PointingVector is in use")
		}
	}

	return nil
}

// validateAgiArticulationStage validates an articulation stage
func validateAgiArticulationStage(stage *AgiArticulationStage) error {
	if stage.Name == "" {
		return fmt.Errorf("stage name is required")
	}

	if stage.TransformType == "" {
		return fmt.Errorf("stage transform type is required")
	}

	// Validate transform type
	validTypes := []string{
		AgiArticulationTransformTypeXTranslate, AgiArticulationTransformTypeYTranslate, AgiArticulationTransformTypeZTranslate,
		AgiArticulationTransformTypeXRotate, AgiArticulationTransformTypeYRotate, AgiArticulationTransformTypeZRotate,
		AgiArticulationTransformTypeXScale, AgiArticulationTransformTypeYScale, AgiArticulationTransformTypeZScale,
		AgiArticulationTransformTypeUniformScale,
	}
	valid := false
	for _, validType := range validTypes {
		if stage.TransformType == validType {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("invalid transform type: %s", stage.TransformType)
	}

	// Validate values
	if stage.MinimumValue > stage.InitialValue {
		return fmt.Errorf("minimum value must be less than or equal to initial value")
	}
	if stage.InitialValue > stage.MaximumValue {
		return fmt.Errorf("initial value must be less than or equal to maximum value")
	}

	return nil
}

// CreateArticulation creates a new articulation
func (a *AgiRootArticulations) CreateArticulation(name string) *AgiArticulation {
	articulation := AgiArticulation{
		Name:         name,
		Stages:       []AgiArticulationStage{},
		LogicalIndex: len(a.Articulations),
	}
	a.Articulations = append(a.Articulations, articulation)
	return &a.Articulations[len(a.Articulations)-1]
}

// CreateArticulationStage creates a new articulation stage
func (a *AgiArticulation) CreateArticulationStage(name string, transformType string) (*AgiArticulationStage, error) {
	// Check if adding a rotation stage would exceed the limit when pointing vector is present
	if a.PointingVector != nil {
		isRotationType := false
		for _, rotationType := range AgiRotationTypes {
			if transformType == rotationType {
				isRotationType = true
				break
			}
		}
		if isRotationType {
			rotationStages := 0
			for _, stage := range a.Stages {
				for _, rotationType := range AgiRotationTypes {
					if stage.TransformType == rotationType {
						rotationStages++
						break
					}
				}
			}
			if rotationStages >= 2 {
				return nil, fmt.Errorf("cannot add more than 2 rotation stages when a PointingVector is in use")
			}
		}
	}

	stage := AgiArticulationStage{
		Name:          name,
		TransformType: transformType,
		LogicalIndex:  len(a.Stages),
	}
	a.Stages = append(a.Stages, stage)
	return &a.Stages[len(a.Stages)-1], nil
}

// SetValues sets the values for an articulation stage
func (s *AgiArticulationStage) SetValues(minValue, initial, maxValue float64) error {
	if minValue > initial {
		return fmt.Errorf("minimum value must be less than or equal to initial value")
	}
	if initial > maxValue {
		return fmt.Errorf("initial value must be less than or equal to maximum value")
	}

	s.MinimumValue = minValue
	s.InitialValue = initial
	s.MaximumValue = maxValue
	return nil
}

// SetPointingVector sets the pointing vector for an articulation
func (a *AgiArticulation) SetPointingVector(vector *[3]float64) error {
	if vector == nil {
		// Pointing is turned off
		a.PointingVector = nil
		return nil
	}

	// Check if it's a unit-length vector
	magnitude := (*vector)[0]*(*vector)[0] + (*vector)[1]*(*vector)[1] + (*vector)[2]*(*vector)[2]
	if magnitude < 0.999 || magnitude > 1.001 {
		return fmt.Errorf("pointingVector must be a unit-length vector")
	}

	// Check that there are exactly 1 or 2 rotation stages
	rotationStages := 0
	for _, stage := range a.Stages {
		for _, rotationType := range AgiRotationTypes {
			if stage.TransformType == rotationType {
				rotationStages++
				break
			}
		}
	}
	if rotationStages != 1 && rotationStages != 2 {
		return fmt.Errorf("pointingVector requires exactly 1 or exactly 2 rotation stages")
	}

	a.PointingVector = vector
	return nil
}

// Note: Extensions are registered through the RegisterExtensions function in register.go
// This avoids conflicts when multiple extensions share the same name
