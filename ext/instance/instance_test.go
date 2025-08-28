package instance

import (
	"testing"

	"github.com/flywave/gltf"
)

func TestUnmarshal(t *testing.T) {
	// Test unmarshaling with full extension object
	data := []byte(`{
		"attributes": {
			"TRANSLATION": 0,
			"ROTATION": 1,
			"SCALE": 2
		}
	}`)

	ext, err := Unmarshal(data)
	if err != nil {
		t.Errorf("Unmarshal failed: %v", err)
	}

	attrs, ok := ext.(*InstanceAttributes)
	if !ok {
		t.Error("Unmarshaled object is not of type *InstanceAttributes")
	}

	if len(attrs.Attributes) != 3 {
		t.Errorf("Expected 3 attributes, got %d", len(attrs.Attributes))
	}

	// Test unmarshaling with direct attributes object
	data = []byte(`{
		"TRANSLATION": 0,
		"ROTATION": 1,
		"SCALE": 2
	}`)

	ext, err = Unmarshal(data)
	if err != nil {
		t.Errorf("Unmarshal failed: %v", err)
	}

	attrs, ok = ext.(*InstanceAttributes)
	if !ok {
		t.Error("Unmarshaled object is not of type *InstanceAttributes")
	}

	if len(attrs.Attributes) != 3 {
		t.Errorf("Expected 3 attributes, got %d", len(attrs.Attributes))
	}

	// Test unmarshaling with invalid data
	data = []byte(`{ "invalid": "data" }`)
	ext, err = Unmarshal(data)
	if err == nil {
		t.Error("Unmarshal should have failed with invalid data")
	}

	_ = ext
}

func TestInstanceData_InstanceCount(t *testing.T) {
	data := &InstanceData{
		Translations: [][3]float32{{1, 2, 3}, {4, 5, 6}},
		Rotations:    [][4]float32{{0, 0, 0, 1}, {0, 0, 1, 0}},
		Scales:       [][3]float32{{1, 1, 1}},
	}

	count := data.InstanceCount()
	if count != 2 {
		t.Errorf("Expected instance count 2, got %d", count)
	}

	// Test with empty data
	emptyData := &InstanceData{}
	count = emptyData.InstanceCount()
	if count != 0 {
		t.Errorf("Expected instance count 0, got %d", count)
	}
}

func TestInstanceData_ToMat4(t *testing.T) {
	data := &InstanceData{
		Translations: [][3]float32{{1, 2, 3}},
		Rotations:    [][4]float32{{0, 0, 0, 1}},
		Scales:       [][3]float32{{2, 2, 2}},
	}

	matrices, err := data.ToMat4()
	if err != nil {
		t.Errorf("ToMat4 failed: %v", err)
	}

	if len(matrices) != 1 {
		t.Errorf("Expected 1 matrix, got %d", len(matrices))
	}

	// Test with empty data
	emptyData := &InstanceData{}
	_, err = emptyData.ToMat4()
	if err == nil {
		t.Error("ToMat4 should have failed with empty data")
	}
}

func TestFromMat4(t *testing.T) {
	// TODO: Implement this test when we have a way to create test matrices
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()
	if config.TranslationType != gltf.ComponentFloat {
		t.Errorf("Expected TranslationType to be ComponentFloat, got %v", config.TranslationType)
	}
	if config.RotationType != gltf.ComponentFloat {
		t.Errorf("Expected RotationType to be ComponentFloat, got %v", config.RotationType)
	}
	if config.ScaleType != gltf.ComponentFloat {
		t.Errorf("Expected ScaleType to be ComponentFloat, got %v", config.ScaleType)
	}
	if config.Normalized != false {
		t.Errorf("Expected Normalized to be false, got %v", config.Normalized)
	}
}

func TestSetInstanceExtension(t *testing.T) {
	// Create a node
	node := &gltf.Node{
		Extensions: make(gltf.Extensions),
	}

	// Set instance extension
	attributes := map[string]uint32{
		"TRANSLATION": 0,
		"ROTATION":    1,
		"SCALE":       2,
	}

	err := SetInstanceExtension(node, attributes)
	if err != nil {
		t.Errorf("SetInstanceExtension failed: %v", err)
	}

	// Check that the extension was set
	extData, exists := node.Extensions[ExtensionName]
	if !exists {
		t.Error("Instance extension was not set on the node")
	}

	// Check that the extension data is in the correct format
	extDataBytes, ok := extData.([]byte)
	if !ok {
		t.Error("Extension data is not in expected format ([]byte)")
	}

	// Try to unmarshal the extension data
	ext, err := Unmarshal(extDataBytes)
	if err != nil {
		t.Errorf("Failed to unmarshal extension data: %v", err)
	}

	attrs, ok := ext.(*InstanceAttributes)
	if !ok {
		t.Error("Unmarshaled object is not of type *InstanceAttributes")
	}

	if len(attrs.Attributes) != 3 {
		t.Errorf("Expected 3 attributes, got %d", len(attrs.Attributes))
	}
}

func TestGetInstanceExtension(t *testing.T) {
	// Create a node with instance extension
	node := &gltf.Node{
		Extensions: make(gltf.Extensions),
	}

	attributes := map[string]uint32{
		"TRANSLATION": 0,
		"ROTATION":    1,
		"SCALE":       2,
	}

	// Set the extension
	err := SetInstanceExtension(node, attributes)
	if err != nil {
		t.Errorf("SetInstanceExtension failed: %v", err)
	}

	// Get the extension
	ext, err := GetInstanceExtension(node)
	if err != nil {
		t.Errorf("GetInstanceExtension failed: %v", err)
	}

	if ext == nil {
		t.Error("GetInstanceExtension returned nil")
		return
	}

	if len(ext.Attributes) != 3 {
		t.Errorf("Expected 3 attributes, got %d", len(ext.Attributes))
	}

	// Test with node that doesn't have the extension
	emptyNode := &gltf.Node{
		Extensions: make(gltf.Extensions),
	}

	_, err = GetInstanceExtension(emptyNode)
	if err == nil {
		t.Error("GetInstanceExtension should have failed with node that doesn't have the extension")
	}
}

func TestValidateInstanceAttributes(t *testing.T) {
	// Create a document with some accessors
	doc := &gltf.Document{
		Accessors: []*gltf.Accessor{
			{
				Type:          gltf.AccessorVec3,
				ComponentType: gltf.ComponentFloat,
				Count:         10,
			},
			{
				Type:          gltf.AccessorVec4,
				ComponentType: gltf.ComponentFloat,
				Count:         10,
			},
			{
				Type:          gltf.AccessorVec3,
				ComponentType: gltf.ComponentFloat,
				Count:         10,
			},
		},
	}

	// Valid attributes
	attributes := map[string]uint32{
		"TRANSLATION": 0,
		"ROTATION":    1,
		"SCALE":       2,
	}

	err := ValidateInstanceAttributes(doc, attributes)
	if err != nil {
		t.Errorf("ValidateInstanceAttributes failed with valid attributes: %v", err)
	}

	// Invalid attribute count (mismatch)
	doc.Accessors[1].Count = 5
	err = ValidateInstanceAttributes(doc, attributes)
	if err == nil {
		t.Error("ValidateInstanceAttributes should have failed with mismatched counts")
	}

	// Reset count for next test
	doc.Accessors[1].Count = 10

	// Invalid TRANSLATION type
	doc.Accessors[0].Type = gltf.AccessorVec4
	err = ValidateInstanceAttributes(doc, attributes)
	if err == nil {
		t.Error("ValidateInstanceAttributes should have failed with invalid TRANSLATION type")
	}

	// Reset type for next test
	doc.Accessors[0].Type = gltf.AccessorVec3

	// Invalid ROTATION component type
	doc.Accessors[1].ComponentType = gltf.ComponentUint
	err = ValidateInstanceAttributes(doc, attributes)
	if err == nil {
		t.Error("ValidateInstanceAttributes should have failed with invalid ROTATION component type")
	}

	// Reset component type for next test
	doc.Accessors[1].ComponentType = gltf.ComponentFloat

	// Valid custom attribute (with underscore prefix)
	attributes["_ID"] = 0
	err = ValidateInstanceAttributes(doc, attributes)
	if err != nil {
		t.Errorf("ValidateInstanceAttributes failed with valid custom attribute: %v", err)
	}

	// Invalid custom attribute (without underscore prefix)
	attributes["ID"] = 0
	err = ValidateInstanceAttributes(doc, attributes)
	if err == nil {
		t.Error("ValidateInstanceAttributes should have failed with invalid custom attribute name")
	}
}
