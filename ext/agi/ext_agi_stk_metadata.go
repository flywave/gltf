package agi

import (
	"encoding/json"
	"fmt"
)

const (
	// StkMetadataExtensionName is the name of the AGI_stk_metadata extension
	StkMetadataExtensionName = "AGI_stk_metadata"
)

// AgiRootStkMetadata represents the AGI_stk_metadata extension at the root level
type AgiRootStkMetadata struct {
	SolarPanelGroups []AgiStkSolarPanelGroup    `json:"solarPanelGroups,omitempty"`
	Extensions       map[string]json.RawMessage `json:"extensions,omitempty"`
	Extras           json.RawMessage            `json:"extras,omitempty"`
}

// AgiNodeStkMetadata represents the AGI_stk_metadata extension at the node level
type AgiNodeStkMetadata struct {
	SolarPanelGroupName *string                    `json:"solarPanelGroupName,omitempty"`
	NoObscuration       *bool                      `json:"noObscuration,omitempty"`
	Extensions          map[string]json.RawMessage `json:"extensions,omitempty"`
	Extras              json.RawMessage            `json:"extras,omitempty"`
}

// AgiStkSolarPanelGroup represents a solar panel group
type AgiStkSolarPanelGroup struct {
	Name         string  `json:"name"`
	Efficiency   float64 `json:"efficiency"`
	LogicalIndex int     `json:"-"` // Not serialized
}

// UnmarshalAgiRootStkMetadata unmarshals the AGI_stk_metadata extension data for root level
func UnmarshalAgiRootStkMetadata(data []byte) (interface{}, error) {
	var ext AgiRootStkMetadata
	if err := json.Unmarshal(data, &ext); err != nil {
		return nil, fmt.Errorf("AGI_stk_metadata root parsing failed: %w", err)
	}

	// Validate solar panel groups
	for i := range ext.SolarPanelGroups {
		if err := validateAgiStkSolarPanelGroup(&ext.SolarPanelGroups[i]); err != nil {
			return nil, fmt.Errorf("invalid solar panel group: %w", err)
		}
	}

	return &ext, nil
}

// UnmarshalAgiNodeStkMetadata unmarshals the AGI_stk_metadata extension data for node level
func UnmarshalAgiNodeStkMetadata(data []byte) (interface{}, error) {
	var ext AgiNodeStkMetadata
	if err := json.Unmarshal(data, &ext); err != nil {
		return nil, fmt.Errorf("AGI_stk_metadata node parsing failed: %w", err)
	}

	return &ext, nil
}

// validateAgiStkSolarPanelGroup validates a solar panel group
func validateAgiStkSolarPanelGroup(group *AgiStkSolarPanelGroup) error {
	if group.Name == "" {
		return fmt.Errorf("solar panel group name is required")
	}

	if group.Efficiency < 0 || group.Efficiency > 1 {
		return fmt.Errorf("solar panel group efficiency must be between 0 and 1")
	}

	return nil
}

// CreateSolarPanelGroup creates a new solar panel group
func (a *AgiRootStkMetadata) CreateSolarPanelGroup(name string) *AgiStkSolarPanelGroup {
	group := AgiStkSolarPanelGroup{
		Name:         name,
		Efficiency:   0.0,
		LogicalIndex: len(a.SolarPanelGroups),
	}
	a.SolarPanelGroups = append(a.SolarPanelGroups, group)
	return &a.SolarPanelGroups[len(a.SolarPanelGroups)-1]
}

// SetEfficiency sets the efficiency of a solar panel group
func (g *AgiStkSolarPanelGroup) SetEfficiency(efficiency float64) error {
	if efficiency < 0 || efficiency > 1 {
		return fmt.Errorf("efficiency must be between 0 and 1")
	}
	g.Efficiency = efficiency
	return nil
}

// SetNoObscuration sets the no obscuration flag
func (n *AgiNodeStkMetadata) SetNoObscuration(noObscuration bool) {
	n.NoObscuration = &noObscuration
}

// GetNoObscuration gets the no obscuration flag with default value
func (n *AgiNodeStkMetadata) GetNoObscuration() bool {
	if n.NoObscuration == nil {
		return false // default value
	}
	return *n.NoObscuration
}

// Note: Extensions are registered through the RegisterExtensions function in register.go
// This avoids conflicts when multiple extensions share the same name
