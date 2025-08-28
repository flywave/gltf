package webp

import (
	"encoding/json"
	"testing"
)

func TestExtTextureWebp(t *testing.T) {
	// Test creating and validating ExtTextureWebp
	source := uint32(0)
	ext := &ExtTextureWebp{
		Source: &source,
	}

	// Test JSON marshaling and unmarshaling
	data, err := json.Marshal(ext)
	if err != nil {
		t.Errorf("Failed to marshal ExtTextureWebp: %v", err)
	}

	unmarshaled, err := UnmarshalExtTextureWebp(data)
	if err != nil {
		t.Errorf("Failed to unmarshal ExtTextureWebp: %v", err)
	}

	extWebp, ok := unmarshaled.(*ExtTextureWebp)
	if !ok {
		t.Error("Unmarshaled object is not of type *ExtTextureWebp")
	}

	if extWebp.Source == nil {
		t.Error("Expected Source to be set")
	} else if *extWebp.Source != 0 {
		t.Errorf("Expected source 0, got %d", *extWebp.Source)
	}
}

func TestSettersAndGetters(t *testing.T) {
	ext := &ExtTextureWebp{}

	// Test SetSource and GetSource
	ext.SetSource(5)
	if ext.GetSource() == nil {
		t.Error("Expected Source to be set")
	} else if *ext.GetSource() != 5 {
		t.Errorf("Expected source 5, got %d", *ext.GetSource())
	}

	// Test with nil source
	ext.Source = nil
	if ext.GetSource() != nil {
		t.Error("Expected Source to be nil")
	}
}

func TestUnmarshalExtTextureWebpWithNilSource(t *testing.T) {
	// Test unmarshaling with nil source
	jsonData := `{}`

	_, err := UnmarshalExtTextureWebp([]byte(jsonData))
	if err != nil {
		t.Errorf("Failed to unmarshal ExtTextureWebp with nil source: %v", err)
	}
}

func TestUnmarshalExtTextureWebpWithInvalidData(t *testing.T) {
	// Test unmarshaling with invalid JSON
	invalidJson := `{invalid json}`

	_, err := UnmarshalExtTextureWebp([]byte(invalidJson))
	if err == nil {
		t.Error("UnmarshalExtTextureWebp should have failed with invalid JSON")
	}
}
