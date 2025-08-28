package gltf

import (
	"testing"
)

func TestUnmarshalExtStructuralMetadata(t *testing.T) {
	// Test unmarshaling valid structural metadata data
	data := []byte(`{
		"schema": {
			"id": "test_schema",
			"classes": {
				"test_class": {
					"properties": {
						"test_property": {
							"type": "SCALAR",
							"componentType": "FLOAT32"
						}
					}
				}
			}
		},
		"propertyTables": [
			{
				"class": "test_class",
				"count": 100,
				"properties": {
					"test_property": {
						"values": 0
					}
				}
			}
		]
	}`)

	ext, err := UnmarshalExtStructuralMetadata(data)
	if err != nil {
		t.Errorf("UnmarshalExtStructuralMetadata failed: %v", err)
	}
	if ext == nil {
		t.Error("UnmarshalExtStructuralMetadata returned nil extension")
	}

	structuralMetadata, ok := ext.(ExtStructuralMetadata)
	if !ok {
		t.Error("Unmarshaled extension is not of type ExtStructuralMetadata")
	}
	if structuralMetadata.Schema == nil {
		t.Error("Schema should not be nil")
	}
	if len(structuralMetadata.PropertyTables) != 1 {
		t.Errorf("Expected 1 property table, got %d", len(structuralMetadata.PropertyTables))
	}
}

func TestUnmarshalExtStructuralMetadataInvalid(t *testing.T) {
	// Test unmarshaling invalid structural metadata data (missing schema)
	data := []byte(`{
		"propertyTables": [
			{
				"class": "test_class",
				"count": 100,
				"properties": {
					"test_property": {
						"values": 0
					}
				}
			}
		]
	}`)

	ext, err := UnmarshalExtStructuralMetadata(data)
	if err == nil {
		t.Error("UnmarshalExtStructuralMetadata should fail for invalid data")
	}
	if ext != nil {
		t.Error("UnmarshalExtStructuralMetadata should return nil for invalid data")
	}
}
