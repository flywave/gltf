package gltf

import (
	"encoding/json"
	"testing"
)

func TestUnmarshalExtStructuralMetadata_Valid(t *testing.T) {
	data := []byte(`{
		"schema": {
			"id": "test_schema",
			"classes": {
				"test_class": {
					"properties": {
						"p": {"type": "SCALAR", "componentType": "FLOAT32"}
					}
				}
			}
		},
		"propertyTables": [{
			"class": "test_class",
			"count": 100,
			"properties": {"p": {"values": 0}}
		}]
	}`)
	ext, err := UnmarshalExtStructuralMetadata(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	sm := ext.(ExtStructuralMetadata)
	if sm.Schema == nil {
		t.Fatal("schema should not be nil")
	}
	if sm.Schema.ID == nil || *sm.Schema.ID != "test_schema" {
		t.Error("id should be test_schema")
	}
	if len(sm.PropertyTables) != 1 {
		t.Errorf("expected 1 table, got %d", len(sm.PropertyTables))
	}
}

func TestUnmarshalExtStructuralMetadata_SchemaURI(t *testing.T) {
	data := []byte(`{"schemaUri": "https://example.com/schema.json"}`)
	ext, err := UnmarshalExtStructuralMetadata(data)
	if err != nil {
		t.Fatalf("schemaUri-only should be accepted: %v", err)
	}
	sm := ext.(ExtStructuralMetadata)
	if sm.SchemaURI == nil || *sm.SchemaURI != "https://example.com/schema.json" {
		t.Error("schemaUri mismatch")
	}
}

func TestUnmarshalExtStructuralMetadata_NoSchemaOrURI(t *testing.T) {
	data := []byte(`{"propertyTables": [{"class":"x","count":1,"properties":{"p":{"values":0}}}]}`)
	_, err := UnmarshalExtStructuralMetadata(data)
	if err == nil {
		t.Error("expected error for missing schema/schemaUri")
	}
}

func TestUnmarshalExtStructuralMetadata_UndefinedClass(t *testing.T) {
	data := []byte(`{
		"schema": {"classes": {"a":{"properties":{"p":{"type":"SCALAR"}}}}},
		"propertyTables": [{"class":"nonexistent","count":1,"properties":{"p":{"values":0}}}]
	}`)
	_, err := UnmarshalExtStructuralMetadata(data)
	if err == nil {
		t.Error("expected error for undefined class")
	}
}

func TestEnum_DefaultValueType(t *testing.T) {
	data := `{"values": [{"name":"A","value":0}]}`
	var e Enum
	if err := json.Unmarshal([]byte(data), &e); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if e.ValueType != EnumValueTypeUint16 {
		t.Errorf("default valueType should be UINT16, got %s", e.ValueType)
	}
}

func TestEnum_ExplicitValueType(t *testing.T) {
	data := `{"valueType":"INT32","values":[{"name":"A","value":0}]}`
	var e Enum
	if err := json.Unmarshal([]byte(data), &e); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if e.ValueType != EnumValueTypeInt32 {
		t.Errorf("expected INT32, got %s", e.ValueType)
	}
}

func TestPropertyTableProperty_DefaultOffsetTypes(t *testing.T) {
	data := `{"values": 0}`
	var p PropertyTableProperty
	if err := json.Unmarshal([]byte(data), &p); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if p.ArrayOffsetType != OffsetTypeUint32 {
		t.Errorf("default arrayOffsetType should be UINT32, got %s", p.ArrayOffsetType)
	}
	if p.StringOffsetType != OffsetTypeUint32 {
		t.Errorf("default stringOffsetType should be UINT32, got %s", p.StringOffsetType)
	}
}

func TestPropertyTableProperty_ExplicitOffsetTypes(t *testing.T) {
	data := `{"values":0,"arrayOffsetType":"UINT16","stringOffsetType":"UINT64"}`
	var p PropertyTableProperty
	if err := json.Unmarshal([]byte(data), &p); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if p.ArrayOffsetType != OffsetTypeUint16 {
		t.Errorf("expected UINT16, got %s", p.ArrayOffsetType)
	}
	if p.StringOffsetType != OffsetTypeUint64 {
		t.Errorf("expected UINT64, got %s", p.StringOffsetType)
	}
}

func TestPropertyTextureProperty_IndexAndChannels(t *testing.T) {
	data := `{"index":2,"channels":[0,1]}`
	var p PropertyTextureProperty
	if err := json.Unmarshal([]byte(data), &p); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if p.Index != 2 {
		t.Errorf("expected index 2, got %d", p.Index)
	}
	if len(p.Channels) != 2 || p.Channels[0] != 0 || p.Channels[1] != 1 {
		t.Errorf("channels mismatch: %v", p.Channels)
	}
}

func TestPropertyTexture_ExtensionsExtras(t *testing.T) {
	data := `{"class":"test","properties":{"p":{"index":0,"channels":[0]}},"extensions":{"EXT_test":{}},"extras":{"key":"val"}}`
	var pt PropertyTexture
	if err := json.Unmarshal([]byte(data), &pt); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if pt.Extensions == nil || len(pt.Extensions) != 1 {
		t.Error("extensions not preserved")
	}
	if pt.Extras == nil {
		t.Error("extras not preserved")
	}
}

func TestSchemaID_Optional(t *testing.T) {
	data := `{"classes":{}}`
	var s Schema
	if err := json.Unmarshal([]byte(data), &s); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if s.ID != nil {
		t.Error("id should be nil when omitted")
	}
}

func TestEnumValue_Int64(t *testing.T) {
	data := `{"name":"big","value":4294967295}` // max uint32
	var ev EnumValue
	if err := json.Unmarshal([]byte(data), &ev); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if ev.Value != 4294967295 {
		t.Errorf("expected 4294967295, got %d", ev.Value)
	}
}

func TestClass_PropertiesRequired(t *testing.T) {
	data := `{"name":"test","properties":{"p":{"type":"SCALAR"}}}`
	var c Class
	if err := json.Unmarshal([]byte(data), &c); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if len(c.Properties) != 1 {
		t.Error("properties not unmarshaled")
	}
}
