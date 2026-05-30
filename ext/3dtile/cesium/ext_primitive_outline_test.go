package cesium

import (
	"testing"

	"github.com/flywave/gltf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetCesiumOutline_CreatesAccessor(t *testing.T) {
	doc := &gltf.Document{}
	primitive := &gltf.Primitive{
		Extensions: make(gltf.Extensions),
	}
	indices := []uint32{0, 1, 2, 3, 4, 5}

	err := SetCesiumOutline(doc, primitive, indices)
	require.NoError(t, err)

	require.Len(t, doc.Accessors, 1)
	require.Len(t, doc.BufferViews, 1)
	require.Len(t, doc.Buffers, 1)

	acc := doc.Accessors[0]
	assert.Equal(t, gltf.ComponentUint, acc.ComponentType)
	assert.Equal(t, gltf.AccessorScalar, acc.Type)
	assert.Equal(t, uint32(6), acc.Count)
	assert.NotNil(t, acc.BufferView)

	// Verify the extension on the primitive
	ext, ok := primitive.Extensions[ExtensionName].(*CesiumPrimitiveOutline)
	require.True(t, ok)
	assert.NotNil(t, ext.Indices)
	assert.Equal(t, uint32(0), *ext.Indices)
}

func TestSetCesiumOutline_DataIntegrity(t *testing.T) {
	indices := []uint32{3, 1, 4, 1, 5, 9}
	doc := &gltf.Document{}
	primitive := &gltf.Primitive{Extensions: make(gltf.Extensions)}

	err := SetCesiumOutline(doc, primitive, indices)
	require.NoError(t, err)

	// Read back the data through the accessor
	acc := doc.Accessors[0]
	bv := doc.BufferViews[*acc.BufferView]
	buf := doc.Buffers[bv.Buffer]
	raw := buf.Data[bv.ByteOffset : bv.ByteOffset+bv.ByteLength]

	readBack := make([]uint32, len(indices))
	for i := range readBack {
		readBack[i] = uint32(raw[i*4]) | uint32(raw[i*4+1])<<8 |
			uint32(raw[i*4+2])<<16 | uint32(raw[i*4+3])<<24
	}
	assert.Equal(t, indices, readBack)
}

func TestSetCesiumOutline_EmptyIndices(t *testing.T) {
	doc := &gltf.Document{}
	err := SetCesiumOutline(doc, &gltf.Primitive{}, []uint32{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must not be empty")
}

func TestGetCesiumOutline(t *testing.T) {
	doc := &gltf.Document{}
	primitive := &gltf.Primitive{Extensions: make(gltf.Extensions)}
	err := SetCesiumOutline(doc, primitive, []uint32{0, 1, 2})
	require.NoError(t, err)

	ext, err := GetCesiumOutline(primitive)
	require.NoError(t, err)
	require.NotNil(t, ext)
	assert.NotNil(t, ext.Indices)
	assert.Equal(t, uint32(0), *ext.Indices)
}

func TestGetCesiumOutline_NotFound(t *testing.T) {
	primitive := &gltf.Primitive{Extensions: make(gltf.Extensions)}
	_, err := GetCesiumOutline(primitive)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestGetCesiumOutline_NoExtensions(t *testing.T) {
	primitive := &gltf.Primitive{}
	_, err := GetCesiumOutline(primitive)
	assert.Error(t, err)
}

func TestCesiumOutline_DocumentRoundTrip(t *testing.T) {
	doc := gltf.NewDocument()
	primitive := &gltf.Primitive{Extensions: make(gltf.Extensions)}
	indices := []uint32{0, 1, 2, 0, 2, 3}

	err := SetCesiumOutline(doc, primitive, indices)
	require.NoError(t, err)

	mesh := &gltf.Mesh{
		Name: "Test",
		Primitives: []*gltf.Primitive{primitive},
	}
	doc.Meshes = append(doc.Meshes, mesh)

	// Save and reload
	doc.ExtensionsUsed = append(doc.ExtensionsUsed, ExtensionName)

	tmp := t.TempDir() + "/outline.glb"
	err = gltf.SaveBinary(doc, tmp)
	require.NoError(t, err)

	reloaded, err := gltf.Open(tmp)
	require.NoError(t, err)

	require.Greater(t, len(reloaded.Meshes), 0)
	ext, err := GetCesiumOutline(reloaded.Meshes[0].Primitives[0])
	require.NoError(t, err)
	require.NotNil(t, ext.Indices)

	// Verify the accessor
	acc := reloaded.Accessors[*ext.Indices]
	assert.Equal(t, gltf.ComponentUint, acc.ComponentType)
	assert.Equal(t, uint32(6), acc.Count)
}

func TestValidateCesiumOutlineIndices(t *testing.T) {
	primitiveIndices := []uint32{0, 1, 2, 3, 4, 5, 6, 7}

	valid := ValidateCesiumOutlineIndices([]uint32{0, 1, 2, 3, 4, 5}, primitiveIndices)
	assert.True(t, valid)

	valid = ValidateCesiumOutlineIndices([]uint32{10, 11, 12}, primitiveIndices)
	assert.False(t, valid)

	// Empty outline indices should be valid (nothing to validate)
	valid = ValidateCesiumOutlineIndices([]uint32{}, primitiveIndices)
	assert.True(t, valid)
}

func TestValidateAccessor(t *testing.T) {
	err := ValidateAccessor(&gltf.Accessor{
		ComponentType: gltf.ComponentUint,
		Type:          gltf.AccessorScalar,
		Normalized:    false,
	})
	assert.NoError(t, err)

	err = ValidateAccessor(&gltf.Accessor{
		ComponentType: gltf.ComponentUshort,
		Type:          gltf.AccessorScalar,
	})
	assert.ErrorContains(t, err, "UNSIGNED_INT")

	err = ValidateAccessor(&gltf.Accessor{
		ComponentType: gltf.ComponentUint,
		Type:          gltf.AccessorVec2,
	})
	assert.ErrorContains(t, err, "SCALAR")

	err = ValidateAccessor(&gltf.Accessor{
		ComponentType: gltf.ComponentUint,
		Type:          gltf.AccessorScalar,
		Normalized:    true,
	})
	assert.ErrorContains(t, err, "normalized")

	err = ValidateAccessor(nil)
	assert.ErrorContains(t, err, "nil")
}
