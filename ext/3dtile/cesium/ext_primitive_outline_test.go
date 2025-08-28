package cesium

import (
	"testing"

	"github.com/flywave/gltf"
)

func TestCesiumPrimitiveOutline(t *testing.T) {
	// Test creating and setting Cesium outline extension
	primitive := &gltf.Primitive{
		Extensions: make(gltf.Extensions),
	}

	// Test SetCesiumOutline
	indices := []uint32{0, 1, 2, 3, 4, 5}
	err := SetCesiumOutline(primitive, indices, "test outline")
	if err != nil {
		t.Errorf("SetCesiumOutline failed: %v", err)
	}

	// Test GetCesiumOutline
	ext, err := GetCesiumOutline(primitive)
	if err != nil {
		t.Errorf("GetCesiumOutline failed: %v", err)
	}
	if ext == nil {
		t.Error("GetCesiumOutline returned nil extension")
	}

	// Test ValidateCesiumOutlineIndices
	primitiveIndices := []uint32{0, 1, 2, 3, 4, 5, 6, 7}
	valid := ValidateCesiumOutlineIndices(indices, primitiveIndices)
	if !valid {
		t.Error("ValidateCesiumOutlineIndices should return true for valid indices")
	}

	// Test with invalid indices
	invalidIndices := []uint32{10, 11, 12}
	valid = ValidateCesiumOutlineIndices(invalidIndices, primitiveIndices)
	if valid {
		t.Error("ValidateCesiumOutlineIndices should return false for invalid indices")
	}
}

func TestValidateAccessor(t *testing.T) {
	// Test with valid accessor
	accessor := &gltf.Accessor{
		ComponentType: gltf.ComponentUint,
		Type:          gltf.AccessorScalar,
		Normalized:    false,
	}
	err := ValidateAccessor(accessor)
	if err != nil {
		t.Errorf("ValidateAccessor should pass for valid accessor: %v", err)
	}

	// Test with invalid component type
	accessor.ComponentType = gltf.ComponentUshort
	err = ValidateAccessor(accessor)
	if err == nil {
		t.Error("ValidateAccessor should fail for invalid component type")
	}

	// Test with invalid type
	accessor.ComponentType = gltf.ComponentUint
	accessor.Type = gltf.AccessorVec2
	err = ValidateAccessor(accessor)
	if err == nil {
		t.Error("ValidateAccessor should fail for invalid type")
	}

	// Test with normalized accessor
	accessor.Type = gltf.AccessorScalar
	accessor.Normalized = true
	err = ValidateAccessor(accessor)
	if err == nil {
		t.Error("ValidateAccessor should fail for normalized accessor")
	}
}
