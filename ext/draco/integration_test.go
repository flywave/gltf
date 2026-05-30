package draco

import (
	"testing"

	"github.com/flywave/gltf"
	"github.com/flywave/gltf/modeler"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDracoIntegration(t *testing.T) {
	doc, err := gltf.Open("../../testdata/Draco/draco.gltf")
	require.NoError(t, err)
	require.NotNil(t, doc)

	originalPrimitive := doc.Meshes[0].Primitives[0]
	_, hasDracoExt := originalPrimitive.Extensions[ExtensionName]
	assert.True(t, hasDracoExt)
	assert.Contains(t, doc.ExtensionsUsed, ExtensionName)

	originalPositionAccessor := doc.Accessors[originalPrimitive.Attributes["POSITION"]]
	originalIndexAccessor := doc.Accessors[*originalPrimitive.Indices]

	assert.Nil(t, originalPositionAccessor.BufferView)
	assert.Nil(t, originalIndexAccessor.BufferView)

	err = DecodeAll(doc)
	require.NoError(t, err)

	decodedPrimitive := doc.Meshes[0].Primitives[0]
	_, hasDracoExt = decodedPrimitive.Extensions[ExtensionName]
	assert.False(t, hasDracoExt)
	assert.NotContains(t, doc.ExtensionsUsed, ExtensionName)

	decodedPositionAccessor := doc.Accessors[decodedPrimitive.Attributes["POSITION"]]
	decodedIndexAccessor := doc.Accessors[*decodedPrimitive.Indices]

	assert.NotNil(t, decodedPositionAccessor.BufferView)
	assert.NotNil(t, decodedIndexAccessor.BufferView)

	assert.Equal(t, uint32(5025), decodedPositionAccessor.Count)
	assert.Equal(t, uint32(28800), decodedIndexAccessor.Count)

	options := map[string]interface{}{
		"quantization": map[string]int{
			"position": 14,
			"normal":   10,
			"texcoord": 12,
		},
	}
	err = EncodeAll(doc, options)
	require.NoError(t, err)

	encodedPrimitive := doc.Meshes[0].Primitives[0]
	_, hasDracoExt = encodedPrimitive.Extensions[ExtensionName]
	assert.True(t, hasDracoExt)
	assert.Contains(t, doc.ExtensionsUsed, ExtensionName)

	encodedPositionAccessor := doc.Accessors[encodedPrimitive.Attributes["POSITION"]]
	encodedIndexAccessor := doc.Accessors[*encodedPrimitive.Indices]

	assert.Nil(t, encodedPositionAccessor.BufferView)
	assert.Nil(t, encodedIndexAccessor.BufferView)

	dracoExt, ok := encodedPrimitive.Extensions[ExtensionName].(*DracoExtension)
	require.True(t, ok)
	assert.NotNil(t, dracoExt.BufferView)
	assert.NotEmpty(t, dracoExt.Attributes)

	err = DecodeAll(doc)
	require.NoError(t, err)

	finalPrimitive := doc.Meshes[0].Primitives[0]
	finalPositionAccessor := doc.Accessors[finalPrimitive.Attributes["POSITION"]]
	finalIndexAccessor := doc.Accessors[*finalPrimitive.Indices]

	assert.NotNil(t, finalPositionAccessor.BufferView)
	assert.NotNil(t, finalIndexAccessor.BufferView)
	assert.Equal(t, decodedPositionAccessor.Count, finalPositionAccessor.Count)
	assert.Equal(t, decodedIndexAccessor.Count, finalIndexAccessor.Count)
}

func TestDracoEncodeDecodeCycle(t *testing.T) {
	doc, origPosAcc, origIdxAcc := buildTriangleDoc(t)

	err := EncodeAll(doc, nil)
	require.NoError(t, err)

	encodedPrimitive := doc.Meshes[0].Primitives[0]
	_, hasDracoExt := encodedPrimitive.Extensions[ExtensionName]
	assert.True(t, hasDracoExt)

	err = DecodeAll(doc)
	require.NoError(t, err)

	decodedPrimitive := doc.Meshes[0].Primitives[0]
	_, hasDracoExt = decodedPrimitive.Extensions[ExtensionName]
	assert.False(t, hasDracoExt)

	finalPosAccessor := doc.Accessors[decodedPrimitive.Attributes["POSITION"]]
	finalIdxAccessor := doc.Accessors[*decodedPrimitive.Indices]

	assert.Equal(t, origPosAcc.Count, finalPosAccessor.Count)
	assert.Equal(t, origIdxAcc.Count, finalIdxAccessor.Count)
}

func TestDracoEncodeDecode_QuadSharedVertices(t *testing.T) {
	doc, origPosAcc, origIdxAcc := buildQuadDoc(t)

	err := EncodeAll(doc, nil)
	require.NoError(t, err)
	err = DecodeAll(doc)
	require.NoError(t, err)

	p := doc.Meshes[0].Primitives[0]
	finalPosAcc := doc.Accessors[p.Attributes["POSITION"]]
	finalIdxAcc := doc.Accessors[*p.Indices]

	assert.Equal(t, origPosAcc.Count, finalPosAcc.Count, "quad vertex count")
	assert.Equal(t, origIdxAcc.Count, finalIdxAcc.Count, "quad index count")
}

func TestDracoEncodeDecode_MultiAttribute(t *testing.T) {
	doc, origPosAcc, origNrmAcc, origTexAcc, origIdxAcc := buildMultiAttrTriangleDoc(t)

	err := EncodeAll(doc, nil)
	require.NoError(t, err)
	err = DecodeAll(doc)
	require.NoError(t, err)

	p := doc.Meshes[0].Primitives[0]

	finalPosAcc := doc.Accessors[p.Attributes["POSITION"]]
	finalNrmAcc := doc.Accessors[p.Attributes["NORMAL"]]
	finalTexAcc := doc.Accessors[p.Attributes["TEXCOORD_0"]]
	finalIdxAcc := doc.Accessors[*p.Indices]

	assert.Equal(t, origPosAcc.Count, finalPosAcc.Count, "POSITION count")
	assert.Equal(t, origNrmAcc.Count, finalNrmAcc.Count, "NORMAL count")
	assert.Equal(t, origTexAcc.Count, finalTexAcc.Count, "TEXCOORD_0 count")
	assert.Equal(t, origIdxAcc.Count, finalIdxAcc.Count, "INDEX count")
}

func TestDracoDataIntegrity(t *testing.T) {
	expectedPositions := [][3]float32{
		{0.0, 0.0, 0.0},
		{1.0, 0.0, 0.0},
		{0.0, 1.0, 0.0},
	}

	doc := buildDocWithPositions(t, expectedPositions)

	err := EncodeAll(doc, nil)
	require.NoError(t, err)
	err = DecodeAll(doc)
	require.NoError(t, err)

	finalPosAcc := doc.Accessors[doc.Meshes[0].Primitives[0].Attributes["POSITION"]]
	assert.Equal(t, uint32(len(expectedPositions)), finalPosAcc.Count)

	readPositions, err := modeler.ReadPosition(doc, finalPosAcc, nil)
	require.NoError(t, err)

	for i := range expectedPositions {
		assert.InDelta(t, expectedPositions[i][0], readPositions[i][0], 1e-4,
			"vertex %d X mismatch", i)
		assert.InDelta(t, expectedPositions[i][1], readPositions[i][1], 1e-4,
			"vertex %d Y mismatch", i)
		assert.InDelta(t, expectedPositions[i][2], readPositions[i][2], 1e-4,
			"vertex %d Z mismatch", i)
	}
}

func TestDracoDataIntegrity_QuadWithAllAttributes(t *testing.T) {
	// 使用不对称顶点确保Draco重排后可以唯一识别
	positions := [][3]float32{
		{0, 0, 0},
		{2, 0, 0},
		{2, 1, 0},
		{0, 1, 0},
	}
	normals := [][3]float32{
		{0, 0, 1},
		{0, 0, 1},
		{0, 0, 1},
		{0, 0, 1},
	}
	texcoords := [][2]float32{
		{0, 0},
		{2, 0},
		{2, 2},
		{0, 2},
	}
	indices := []uint16{0, 1, 2, 0, 2, 3}

	doc := gltf.NewDocument()
	posAcc := modeler.WritePosition(doc, positions)
	nrmAcc := modeler.WriteNormal(doc, normals)
	texAcc := modeler.WriteTextureCoord(doc, texcoords)
	idxAcc := modeler.WriteIndices(doc, indices)

	mesh := &gltf.Mesh{Name: "Quad",
		Primitives: []*gltf.Primitive{{
			Attributes: map[string]uint32{
				gltf.POSITION:   posAcc,
				gltf.NORMAL:     nrmAcc,
				gltf.TEXCOORD_0: texAcc,
			},
			Indices: gltf.Index(idxAcc),
		}},
	}
	doc.Meshes = append(doc.Meshes, mesh)

	err := EncodeAll(doc, nil)
	require.NoError(t, err)

	err = DecodeAll(doc)
	require.NoError(t, err)

	p := doc.Meshes[0].Primitives[0]

	finalPosAcc := doc.Accessors[p.Attributes["POSITION"]]
	finalNrmAcc := doc.Accessors[p.Attributes["NORMAL"]]
	finalTexAcc := doc.Accessors[p.Attributes["TEXCOORD_0"]]
	finalIdxAcc := doc.Accessors[*p.Indices]

	assert.Equal(t, uint32(4), finalPosAcc.Count, "vertex count should be 4")
	assert.Equal(t, uint32(4), finalNrmAcc.Count, "normal count should be 4")
	assert.Equal(t, uint32(4), finalTexAcc.Count, "texcoord count should be 4")
	assert.Equal(t, uint32(6), finalIdxAcc.Count, "index count should be 6")

	// 读取位置数据 (Draco 可能重排顶点, 按位置值排序后比较)
	readPos, err := modeler.ReadPosition(doc, finalPosAcc, nil)
	require.NoError(t, err)
	assert.Equal(t, len(positions), len(readPos))

	// 构建原始顶点位置集合并验证每个原始顶点都能在解码结果中找到
	findInDelta := func(target [3]float32, list [][3]float32, eps float32) bool {
		for _, p := range list {
			d := float32(0)
			for j := 0; j < 3; j++ {
				diff := p[j] - target[j]
				if diff < 0 {
					diff = -diff
				}
				d += diff
			}
			if d < eps {
				return true
			}
		}
		return false
	}

	for i, pos := range positions {
		assert.True(t, findInDelta(pos, readPos, 1e-3),
			"original vertex %d %v not found in decoded positions", i, pos)
	}

	// 验证索引计数
	assert.Equal(t, uint32(6), finalIdxAcc.Count, "index count")
}

// --- 辅助函数 ---

func buildTriangleDoc(t *testing.T) (*gltf.Document, *gltf.Accessor, *gltf.Accessor) {
	t.Helper()
	positions := [][3]float32{{0, 0, 0}, {1, 0, 0}, {0, 1, 0}}
	indices := []uint16{0, 1, 2}
	doc := gltf.NewDocument()
	posAcc := modeler.WritePosition(doc, positions)
	idxAcc := modeler.WriteIndices(doc, indices)
	mesh := &gltf.Mesh{Name: "Triangle",
		Primitives: []*gltf.Primitive{{
			Attributes: map[string]uint32{gltf.POSITION: posAcc},
			Indices:    gltf.Index(idxAcc),
		}},
	}
	doc.Meshes = append(doc.Meshes, mesh)
	return doc, doc.Accessors[posAcc], doc.Accessors[idxAcc]
}

func buildQuadDoc(t *testing.T) (*gltf.Document, *gltf.Accessor, *gltf.Accessor) {
	t.Helper()
	positions := [][3]float32{{-0.5, -0.5, 0}, {0.5, -0.5, 0}, {0.5, 0.5, 0}, {-0.5, 0.5, 0}}
	indices := []uint16{0, 1, 2, 0, 2, 3}
	doc := gltf.NewDocument()
	posAcc := modeler.WritePosition(doc, positions)
	idxAcc := modeler.WriteIndices(doc, indices)
	mesh := &gltf.Mesh{Name: "Quad",
		Primitives: []*gltf.Primitive{{
			Attributes: map[string]uint32{gltf.POSITION: posAcc},
			Indices:    gltf.Index(idxAcc),
		}},
	}
	doc.Meshes = append(doc.Meshes, mesh)
	return doc, doc.Accessors[posAcc], doc.Accessors[idxAcc]
}

func buildMultiAttrTriangleDoc(t *testing.T) (*gltf.Document, *gltf.Accessor, *gltf.Accessor, *gltf.Accessor, *gltf.Accessor) {
	t.Helper()
	positions := [][3]float32{{0, 0, 0}, {1, 0, 0}, {0, 1, 0}}
	normals := [][3]float32{{0, 0, 1}, {0, 0, 1}, {0, 0, 1}}
	texcoords := [][2]float32{{0, 0}, {1, 0}, {0, 1}}
	indices := []uint16{0, 1, 2}

	doc := gltf.NewDocument()
	posAcc := modeler.WritePosition(doc, positions)
	nrmAcc := modeler.WriteNormal(doc, normals)
	texAcc := modeler.WriteTextureCoord(doc, texcoords)
	idxAcc := modeler.WriteIndices(doc, indices)

	mesh := &gltf.Mesh{Name: "Triangle",
		Primitives: []*gltf.Primitive{{
			Attributes: map[string]uint32{
				gltf.POSITION:   posAcc,
				gltf.NORMAL:     nrmAcc,
				gltf.TEXCOORD_0: texAcc,
			},
			Indices: gltf.Index(idxAcc),
		}},
	}
	doc.Meshes = append(doc.Meshes, mesh)
	return doc, doc.Accessors[posAcc], doc.Accessors[nrmAcc], doc.Accessors[texAcc], doc.Accessors[idxAcc]
}

func buildDocWithPositions(t *testing.T, positions [][3]float32) *gltf.Document {
	t.Helper()
	indices := make([]uint16, len(positions))
	for i := range indices {
		indices[i] = uint16(i)
	}
	doc := gltf.NewDocument()
	posAcc := modeler.WritePosition(doc, positions)
	idxAcc := modeler.WriteIndices(doc, indices)
	mesh := &gltf.Mesh{Name: "Points",
		Primitives: []*gltf.Primitive{{
			Attributes: map[string]uint32{gltf.POSITION: posAcc},
			Indices:    gltf.Index(idxAcc),
		}},
	}
	doc.Meshes = append(doc.Meshes, mesh)
	return doc
}
