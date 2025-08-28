package bimdata

import (
	"encoding/json"
	"testing"
)

func TestBimData(t *testing.T) {
	// Test creating and validating BimData
	properties := []uint32{0, 1}
	typeIndex := uint32(0)
	ext := &BimData{
		Properties: properties,
		Type:       &typeIndex,
	}

	// Test JSON marshaling and unmarshaling
	data, err := json.Marshal(ext)
	if err != nil {
		t.Errorf("Failed to marshal BimData: %v", err)
	}

	unmarshaled, err := UnmarshalBimData(data)
	if err != nil {
		t.Errorf("Failed to unmarshal BimData: %v", err)
	}

	BimData, ok := unmarshaled.(*BimData)
	if !ok {
		t.Error("Unmarshaled object is not of type *BimData")
	}

	if len(BimData.GetPropertyIndices()) != 2 {
		t.Errorf("Expected 2 properties, got %d", len(BimData.GetPropertyIndices()))
	}

	if BimData.GetTypeIndex() == nil {
		t.Error("Expected Type to be set")
	} else if *BimData.GetTypeIndex() != 0 {
		t.Errorf("Expected type index 0, got %d", *BimData.GetTypeIndex())
	}
}

func TestBimDataRoot(t *testing.T) {
	// Test creating and validating BimDataRoot
	ext := &BimDataRoot{
		PropertyNames:  []string{"Height", "Width", "Material"},
		PropertyValues: []string{"2,1 m", "900 mm", "Timber", "2,4 m"},
		Properties: []BimProperty{
			{Name: 0, Value: 0},
			{Name: 1, Value: 1},
			{Name: 2, Value: 2},
			{Name: 0, Value: 3},
		},
		Types: []BimType{
			{Properties: []uint32{1, 2}},
		},
	}

	// Test JSON marshaling and unmarshaling
	data, err := json.Marshal(ext)
	if err != nil {
		t.Errorf("Failed to marshal BimDataRoot: %v", err)
	}

	unmarshaled, err := UnmarshalBimDataRoot(data)
	if err != nil {
		t.Errorf("Failed to unmarshal BimDataRoot: %v", err)
	}

	BimDataRoot, ok := unmarshaled.(*BimDataRoot)
	if !ok {
		t.Error("Unmarshaled object is not of type *BimDataRoot")
	}

	if len(BimDataRoot.PropertyNames) != 3 {
		t.Errorf("Expected 3 property names, got %d", len(BimDataRoot.PropertyNames))
	}

	if len(BimDataRoot.PropertyValues) != 4 {
		t.Errorf("Expected 4 property values, got %d", len(BimDataRoot.PropertyValues))
	}

	if len(BimDataRoot.Properties) != 4 {
		t.Errorf("Expected 4 properties, got %d", len(BimDataRoot.Properties))
	}

	if len(BimDataRoot.Types) != 1 {
		t.Errorf("Expected 1 type, got %d", len(BimDataRoot.Types))
	}
}

func TestSettersAndGetters(t *testing.T) {
	ext := &BimData{}

	// Test SetPropertyIndices and GetPropertyIndices
	properties := []uint32{0, 1, 2}
	ext.SetPropertyIndices(properties)
	if len(ext.GetPropertyIndices()) != 3 {
		t.Errorf("Expected 3 properties, got %d", len(ext.GetPropertyIndices()))
	}

	// Test SetTypeIndex and GetTypeIndex
	ext.SetTypeIndex(5)
	if ext.GetTypeIndex() == nil {
		t.Error("Expected Type to be set")
	} else if *ext.GetTypeIndex() != 5 {
		t.Errorf("Expected type index 5, got %d", *ext.GetTypeIndex())
	}

	// Test SetBufferViewIndex and GetBufferViewIndex
	ext.SetBufferViewIndex(10)
	if ext.GetBufferViewIndex() == nil {
		t.Error("Expected BufferView to be set")
	} else if *ext.GetBufferViewIndex() != 10 {
		t.Errorf("Expected buffer view index 10, got %d", *ext.GetBufferViewIndex())
	}
}

func TestBimDataRootMethods(t *testing.T) {
	ext := &BimDataRoot{}

	// Test AddPropertyName
	ext.AddPropertyName("Height")
	if len(ext.PropertyNames) != 1 {
		t.Errorf("Expected 1 property name, got %d", len(ext.PropertyNames))
	}

	// Test AddPropertyValue
	ext.AddPropertyValue("2,1 m")
	if len(ext.PropertyValues) != 1 {
		t.Errorf("Expected 1 property value, got %d", len(ext.PropertyValues))
	}

	// Test AddProperty
	ext.AddProperty(0, 0)
	if len(ext.Properties) != 1 {
		t.Errorf("Expected 1 property, got %d", len(ext.Properties))
	}

	// Test AddType
	ext.AddType([]uint32{0, 1})
	if len(ext.Types) != 1 {
		t.Errorf("Expected 1 type, got %d", len(ext.Types))
	}

	// Test AddNodePropertyMapping
	typeIndex := uint32(0)
	ext.AddNodePropertyMapping(0, []uint32{0, 1}, &typeIndex)
	if len(ext.NodeProperties) != 1 {
		t.Errorf("Expected 1 node property mapping, got %d", len(ext.NodeProperties))
	}
}

func TestUnmarshalBimDataWithInvalidData(t *testing.T) {
	// Test unmarshaling with invalid JSON
	invalidJson := `{invalid json}`

	_, err := UnmarshalBimData([]byte(invalidJson))
	if err == nil {
		t.Error("UnmarshalBimData should have failed with invalid JSON")
	}
}

func TestUnmarshalBimDataRootWithInvalidData(t *testing.T) {
	// Test unmarshaling with invalid JSON
	invalidJson := `{invalid json}`

	_, err := UnmarshalBimDataRoot([]byte(invalidJson))
	if err == nil {
		t.Error("UnmarshalBimDataRoot should have failed with invalid JSON")
	}
}
