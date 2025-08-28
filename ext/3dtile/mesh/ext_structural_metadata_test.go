package mesh

import (
	"testing"
)

func TestUnmarshalExtStructuralMetadata(t *testing.T) {
	// Test unmarshaling valid structural metadata data with property textures
	data := []byte(`{
		"propertyTextures": [0, 1],
		"propertyAttributes": [2, 3]
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
	if len(structuralMetadata.PropertyTextures) != 2 {
		t.Errorf("Expected 2 property textures, got %d", len(structuralMetadata.PropertyTextures))
	}
	if len(structuralMetadata.PropertyAttributes) != 2 {
		t.Errorf("Expected 2 property attributes, got %d", len(structuralMetadata.PropertyAttributes))
	}
}

func TestUnmarshalExtStructuralMetadataOnlyTextures(t *testing.T) {
	// Test unmarshaling valid structural metadata data with only property textures
	data := []byte(`{
		"propertyTextures": [0, 1]
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
	if len(structuralMetadata.PropertyTextures) != 2 {
		t.Errorf("Expected 2 property textures, got %d", len(structuralMetadata.PropertyTextures))
	}
	if len(structuralMetadata.PropertyAttributes) != 0 {
		t.Errorf("Expected 0 property attributes, got %d", len(structuralMetadata.PropertyAttributes))
	}
}

func TestUnmarshalExtStructuralMetadataOnlyAttributes(t *testing.T) {
	// Test unmarshaling valid structural metadata data with only property attributes
	data := []byte(`{
		"propertyAttributes": [2, 3]
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
	if len(structuralMetadata.PropertyTextures) != 0 {
		t.Errorf("Expected 0 property textures, got %d", len(structuralMetadata.PropertyTextures))
	}
	if len(structuralMetadata.PropertyAttributes) != 2 {
		t.Errorf("Expected 2 property attributes, got %d", len(structuralMetadata.PropertyAttributes))
	}
}

func TestUnmarshalExtStructuralMetadataInvalid(t *testing.T) {
	// Test unmarshaling invalid structural metadata data (missing both property textures and attributes)
	data := []byte(`{}`)

	ext, err := UnmarshalExtStructuralMetadata(data)
	if err == nil {
		t.Error("UnmarshalExtStructuralMetadata should fail for invalid data")
	}
	if ext != nil {
		t.Error("UnmarshalExtStructuralMetadata should return nil for invalid data")
	}
}

func TestExtStructuralMetadataMethods(t *testing.T) {
	ext := &ExtStructuralMetadata{}

	// Test AddPropertyTexture
	ext.AddPropertyTexture(0)
	ext.AddPropertyTexture(1)
	if len(ext.PropertyTextures) != 2 {
		t.Errorf("Expected 2 property textures, got %d", len(ext.PropertyTextures))
	}

	// Test AddPropertyAttribute
	ext.AddPropertyAttribute(2)
	ext.AddPropertyAttribute(3)
	if len(ext.PropertyAttributes) != 2 {
		t.Errorf("Expected 2 property attributes, got %d", len(ext.PropertyAttributes))
	}

	// Test SetPropertyTextures
	textures := []uint32{4, 5, 6}
	ext.SetPropertyTextures(textures)
	if len(ext.GetPropertyTextures()) != 3 {
		t.Errorf("Expected 3 property textures, got %d", len(ext.GetPropertyTextures()))
	}

	// Test SetPropertyAttributes
	attributes := []uint32{7, 8, 9}
	ext.SetPropertyAttributes(attributes)
	if len(ext.GetPropertyAttributes()) != 3 {
		t.Errorf("Expected 3 property attributes, got %d", len(ext.GetPropertyAttributes()))
	}

	// Test GetPropertyTextures
	gotTextures := ext.GetPropertyTextures()
	if len(gotTextures) != len(textures) {
		t.Errorf("Expected %d property textures, got %d", len(textures), len(gotTextures))
	}

	// Test GetPropertyAttributes
	gotAttributes := ext.GetPropertyAttributes()
	if len(gotAttributes) != len(attributes) {
		t.Errorf("Expected %d property attributes, got %d", len(attributes), len(gotAttributes))
	}
}