package cesium

import (
	"encoding/binary"
	"encoding/json"
	"fmt"

	"github.com/flywave/gltf"
)

const (
	// ExtensionName is the name of the Cesium primitive outline extension
	ExtensionName = "CESIUM_primitive_outline"
)

func init() {
	gltf.RegisterExtension(ExtensionName, UnmarshalCesiumPrimitiveOutline)
}

// CesiumPrimitiveOutline represents the CESIUM_primitive_outline extension
type CesiumPrimitiveOutline struct {
	Indices    *uint32                    `json:"indices,omitempty"`
	Extensions map[string]json.RawMessage `json:"extensions,omitempty"`
	Extras     json.RawMessage            `json:"extras,omitempty"`
}

// UnmarshalCesiumPrimitiveOutline unmarshals the CESIUM_primitive_outline extension data
func UnmarshalCesiumPrimitiveOutline(data []byte) (interface{}, error) {
	var ext CesiumPrimitiveOutline
	if err := json.Unmarshal(data, &ext); err != nil {
		return nil, fmt.Errorf("CESIUM_primitive_outline parsing failed: %w", err)
	}
	return ext, nil
}

// SetCesiumOutline sets the Cesium outline vertex indices for a primitive.
// doc is the parent document; indices are triangle edge indices referencing the primitive's vertices.
func SetCesiumOutline(doc *gltf.Document, primitive *gltf.Primitive, indices []uint32) error {
	if len(indices) == 0 {
		return fmt.Errorf("outline indices must not be empty")
	}

	// Write indices as uint32 into a new buffer
	data := make([]byte, len(indices)*4)
	for i, idx := range indices {
		binary.LittleEndian.PutUint32(data[i*4:], idx)
	}

	buf := &gltf.Buffer{ByteLength: uint32(len(data)), Data: data}
	doc.Buffers = append(doc.Buffers, buf)
	bufIdx := uint32(len(doc.Buffers) - 1)

	bv := &gltf.BufferView{
		Buffer:     bufIdx,
		ByteOffset: 0,
		ByteLength: buf.ByteLength,
		Target:     gltf.TargetElementArrayBuffer,
	}
	doc.BufferViews = append(doc.BufferViews, bv)
	bvIdx := uint32(len(doc.BufferViews) - 1)

	acc := &gltf.Accessor{
		BufferView:    &bvIdx,
		ComponentType: gltf.ComponentUint,
		Count:         uint32(len(indices)),
		Type:          gltf.AccessorScalar,
	}
	doc.Accessors = append(doc.Accessors, acc)
	accIdx := uint32(len(doc.Accessors) - 1)

	if primitive.Extensions == nil {
		primitive.Extensions = make(gltf.Extensions)
	}
	primitive.Extensions[ExtensionName] = &CesiumPrimitiveOutline{
		Indices: &accIdx,
	}
	return nil
}

// GetCesiumOutline gets the Cesium outline extension from a primitive
func GetCesiumOutline(primitive *gltf.Primitive) (*CesiumPrimitiveOutline, error) {
	if primitive.Extensions == nil {
		return nil, fmt.Errorf("no extensions found")
	}

	extData, exists := primitive.Extensions[ExtensionName]
	if !exists {
		return nil, fmt.Errorf("%s extension not found", ExtensionName)
	}

	switch v := extData.(type) {
	case *CesiumPrimitiveOutline:
		return v, nil
	case CesiumPrimitiveOutline:
		return &v, nil
	case []byte:
		var ext CesiumPrimitiveOutline
		if err := json.Unmarshal(v, &ext); err != nil {
			return nil, fmt.Errorf("error unmarshaling CESIUM_primitive_outline extension: %w", err)
		}
		return &ext, nil
	default:
		return nil, fmt.Errorf("extension data is not in expected format")
	}
}

// ValidateCesiumOutlineIndices validates that all indices are within the range of the mesh primitive indices
func ValidateCesiumOutlineIndices(outlineIndices []uint32, primitiveIndices []uint32) bool {
	if len(primitiveIndices) == 0 {
		return false
	}

	// Find max index in primitive indices
	maxIndex := uint32(0)
	for _, idx := range primitiveIndices {
		if idx > maxIndex {
			maxIndex = idx
		}
	}

	// Check that all outline indices are within range
	for _, idx := range outlineIndices {
		if idx > maxIndex {
			return false
		}
	}

	return true
}

// ValidateAccessor validates that the accessor meets the requirements for Cesium outline indices
func ValidateAccessor(accessor *gltf.Accessor) error {
	if accessor == nil {
		return fmt.Errorf("accessor is nil")
	}

	// Check component type - should be unsigned int
	if accessor.ComponentType != gltf.ComponentUint {
		return fmt.Errorf("accessor component type must be UNSIGNED_INT")
	}

	// Check dimensions - should be scalar
	if accessor.Type != gltf.AccessorScalar {
		return fmt.Errorf("accessor dimensions must be SCALAR")
	}

	// Check that it's not normalized
	if accessor.Normalized {
		return fmt.Errorf("accessor must not be normalized")
	}

	return nil
}
