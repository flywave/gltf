package texturebasisu

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExtTextureBasisu(t *testing.T) {
	// Test creating and validating ExtTextureBasisu
	ext := &ExtTextureBasisu{
		Source: 0,
	}

	// Test JSON marshaling and unmarshaling
	data, err := json.Marshal(ext)
	if err != nil {
		t.Errorf("Failed to marshal ExtTextureBasisu: %v", err)
	}

	unmarshaled, err := UnmarshalExtTextureBasisu(data)
	if err != nil {
		t.Errorf("Failed to unmarshal ExtTextureBasisu: %v", err)
	}

	extBasisu, ok := unmarshaled.(*ExtTextureBasisu)
	if !ok {
		t.Error("Unmarshaled object is not of type *ExtTextureBasisu")
	}

	if extBasisu.GetSource() != 0 {
		t.Errorf("Expected source 0, got %d", extBasisu.GetSource())
	}
}

func TestSettersAndGetters(t *testing.T) {
	ext := &ExtTextureBasisu{}

	// Test SetSource and GetSource
	ext.SetSource(5)
	if ext.GetSource() != 5 {
		t.Errorf("Expected source 5, got %d", ext.GetSource())
	}
}

func TestExtTextureBasisu_RoundTrip(t *testing.T) {
	orig := &ExtTextureBasisu{
		Source: 3,
	}
	data, err := json.Marshal(orig)
	require.NoError(t, err)
	got, err := UnmarshalExtTextureBasisu(data)
	require.NoError(t, err)
	require.Equal(t, orig, got)
}

func TestExtTextureBasisu_NonZeroSource(t *testing.T) {
	ext := &ExtTextureBasisu{
		Source: 7,
	}
	data, err := json.Marshal(ext)
	require.NoError(t, err)

	unmarshaled, err := UnmarshalExtTextureBasisu(data)
	require.NoError(t, err)

	extBasisu, ok := unmarshaled.(*ExtTextureBasisu)
	if !ok {
		t.Fatal("Unmarshaled object is not of type *ExtTextureBasisu")
	}
	if extBasisu.GetSource() != 7 {
		t.Errorf("Expected source 7, got %d", extBasisu.GetSource())
	}
}

func TestUnmarshalExtTextureBasisuWithInvalidData(t *testing.T) {
	// Test unmarshaling with invalid JSON
	invalidJson := `{invalid json}`

	_, err := UnmarshalExtTextureBasisu([]byte(invalidJson))
	if err == nil {
		t.Error("UnmarshalExtTextureBasisu should have failed with invalid JSON")
	}
}
