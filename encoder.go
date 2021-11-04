package gltf

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
)

// WriteHandler is the interface that wraps the Write method.
//
// WriteResource should behaves as io.Write in terms of reading the writing resource.
type WriteHandler interface {
	WriteResource(uri string, data []byte) error
}

// Save will save a document as a glTF with the specified by name.
func Save(doc *Document, name string) error {
	return save(doc, name, false)
}

// SaveBinary will save a document as a GLB file with the specified by name.
func SaveBinary(doc *Document, name string) error {
	return save(doc, name, true)
}

func save(doc *Document, name string, asBinary bool) error {
	f, err := os.Create(name)
	if err != nil {
		return err
	}
	e := NewEncoder(f).WithWriteHandler(&RelativeFileHandler{Dir: filepath.Dir(name)})
	e.AsBinary = asBinary
	if err := e.Encode(doc); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}

// An Encoder writes a GLTF to an output stream
// with relative external buffers support.
type Encoder struct {
	AsBinary     bool
	WriteHandler WriteHandler
	w            io.Writer
}

// NewEncoder returns a new encoder that writes to w as a normal glTF file.
func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{
		AsBinary:     true,
		WriteHandler: new(RelativeFileHandler),
		w:            w,
	}
}

// WithWriteHandler sets the WriteHandler.
func (e *Encoder) WithWriteHandler(h WriteHandler) *Encoder {
	e.WriteHandler = h
	return e
}

// Encode writes the encoding of doc to the stream.
func (e *Encoder) Encode(doc *Document) error {
	if doc.Asset.Version == "" {
		doc.Asset.Version = "2.0"
	}
	var err error
	var externalBufferIndex = 0
	if e.AsBinary {
		var hasBinChunk bool
		hasBinChunk, err = e.encodeBinary(doc)
		if hasBinChunk {
			externalBufferIndex = 1
		}
	} else {
		err = json.NewEncoder(e.w).Encode(doc)
	}
	if err != nil {
		return err
	}

	for i := externalBufferIndex; i < len(doc.Buffers); i++ {
		buffer := doc.Buffers[i]
		if len(buffer.Data) == 0 || buffer.IsEmbeddedResource() {
			continue
		}
		if err = e.encodeBuffer(buffer); err != nil {
			return err
		}
	}

	return err
}

func (e *Encoder) encodeBuffer(buffer *Buffer) error {
	if err := validateBufferURI(buffer.URI); err != nil {
		return err
	}

	return e.WriteHandler.WriteResource(buffer.URI, buffer.Data)
}

func (e *Encoder) encodeBinary(doc *Document) (bool, error) {
	jsonText, err := json.Marshal(doc)
	if err != nil {
		return false, err
	}
	jsonHeader := chunkHeader{
		Length: uint32(((len(jsonText) + 3) / 4) * 4),
		Type:   glbChunkJSON,
	}
	header := glbHeader{
		Magic:      glbHeaderMagic,
		Version:    2,
		Length:     12 + 8 + jsonHeader.Length, // 12-byte glb header + 8-byte json chunk header
		JSONHeader: jsonHeader,
	}
	headerPadding := make([]byte, header.JSONHeader.Length-uint32(len(jsonText)))
	for i := range headerPadding {
		headerPadding[i] = ' '
	}

	hasBinChunk := len(doc.Buffers) > 0 && doc.Buffers[0].URI == ""
	var binPaddedLength uint32
	if hasBinChunk {
		binPaddedLength = ((doc.Buffers[0].ByteLength + 3) / 4) * 4
		header.Length += uint32(8) + binPaddedLength
	}

	err = binary.Write(e.w, binary.LittleEndian, &header)
	if err != nil {
		return false, err
	}
	e.w.Write(jsonText)
	e.w.Write(headerPadding)

	if hasBinChunk {
		binBuffer := doc.Buffers[0]
		binPadding := make([]byte, binPaddedLength-binBuffer.ByteLength)
		for i := range binPadding {
			binPadding[i] = 0
		}
		binHeader := chunkHeader{Length: binPaddedLength, Type: glbChunkBIN}
		binary.Write(e.w, binary.LittleEndian, &binHeader)
		e.w.Write(binBuffer.Data)
		_, err = e.w.Write(binPadding)
	}

	return hasBinChunk, err
}

// UnmarshalJSON unmarshal the node with the correct default values.
func (n *Node) UnmarshalJSON(data []byte) error {
	type alias Node
	tmp := alias(Node{
		Matrix:   DefaultMatrix,
		Rotation: DefaultRotation,
		Scale:    DefaultScale,
	})
	err := json.Unmarshal(data, &tmp)
	if err == nil {
		*n = Node(tmp)
	}
	return err
}

// MarshalJSON marshal the node with the correct default values.
func (n *Node) MarshalJSON() ([]byte, error) {
	type alias Node
	tmp := &struct {
		Matrix      *[16]float32 `json:"matrix,omitempty"`                                          // A 4x4 transformation matrix stored in column-major order.
		Rotation    *[4]float32  `json:"rotation,omitempty" validate:"omitempty,dive,gte=-1,lte=1"` // The node's unit quaternion rotation in the order (x, y, z, w), where w is the scalar.
		Scale       *[3]float32  `json:"scale,omitempty"`
		Translation *[3]float32  `json:"translation,omitempty"`
		*alias
	}{
		alias: (*alias)(n),
	}
	if n.Matrix != DefaultMatrix && n.Matrix != emptyMatrix {
		tmp.Matrix = &n.Matrix
	}
	if n.Rotation != DefaultRotation && n.Rotation != emptyRotation {
		tmp.Rotation = &n.Rotation
	}
	if n.Scale != DefaultScale && n.Scale != emptyScale {
		tmp.Scale = &n.Scale
	}
	if n.Translation != DefaultTranslation {
		tmp.Translation = &n.Translation
	}
	return json.Marshal(tmp)
}

// MarshalJSON marshal the camera with the correct default values.
func (c *Camera) MarshalJSON() ([]byte, error) {
	type alias Camera
	if c.Perspective != nil {
		return json.Marshal(&struct {
			Type string `json:"type"`
			*alias
		}{
			Type:  "perspective",
			alias: (*alias)(c),
		})
	} else if c.Orthographic != nil {
		return json.Marshal(&struct {
			Type string `json:"type"`
			*alias
		}{
			Type:  "orthographic",
			alias: (*alias)(c),
		})
	}
	return nil, errors.New("gltf: camera must defined either the perspective or orthographic property")
}

// UnmarshalJSON unmarshal the material with the correct default values.
func (m *Material) UnmarshalJSON(data []byte) error {
	type alias Material
	tmp := alias(Material{AlphaCutoff: Float(0.5)})
	err := json.Unmarshal(data, &tmp)
	if err == nil {
		*m = Material(tmp)
	}
	return err
}

// MarshalJSON marshal the material with the correct default values.
func (m *Material) MarshalJSON() ([]byte, error) {
	type alias Material
	tmp := &struct {
		EmissiveFactor *[3]float32 `json:"emissiveFactor,omitempty" validate:"dive,gte=0,lte=1"`
		AlphaCutoff    *float32    `json:"alphaCutoff,omitempty" validate:"omitempty,gte=0"`
		*alias
	}{
		alias: (*alias)(m),
	}
	if m.AlphaCutoff != nil && *m.AlphaCutoff != 0.5 {
		tmp.AlphaCutoff = m.AlphaCutoff
	}
	if m.EmissiveFactor != [3]float32{0, 0, 0} {
		tmp.EmissiveFactor = &m.EmissiveFactor
	}
	return json.Marshal(tmp)
}

// UnmarshalJSON unmarshal the texture info with the correct default values.
func (n *NormalTexture) UnmarshalJSON(data []byte) error {
	type alias NormalTexture
	tmp := alias(NormalTexture{Scale: Float(1)})
	err := json.Unmarshal(data, &tmp)
	if err == nil {
		*n = NormalTexture(tmp)
	}
	return err
}

// MarshalJSON marshal the texture info with the correct default values.
func (n *NormalTexture) MarshalJSON() ([]byte, error) {
	type alias NormalTexture
	if n.Scale != nil && *n.Scale == 1 {
		return json.Marshal(&struct {
			Scale float32 `json:"scale,omitempty"`
			*alias
		}{
			Scale: 0,
			alias: (*alias)(n),
		})
	}
	return json.Marshal(&struct{ *alias }{alias: (*alias)(n)})
}

// UnmarshalJSON unmarshal the texture info with the correct default values.
func (o *OcclusionTexture) UnmarshalJSON(data []byte) error {
	type alias OcclusionTexture
	tmp := alias(OcclusionTexture{Strength: Float(1)})
	err := json.Unmarshal(data, &tmp)
	if err == nil {
		*o = OcclusionTexture(tmp)
	}
	return err
}

// MarshalJSON marshal the texture info with the correct default values.
func (o *OcclusionTexture) MarshalJSON() ([]byte, error) {
	type alias OcclusionTexture
	if o.Strength != nil && *o.Strength == 1 {
		return json.Marshal(&struct {
			Strength float32 `json:"strength,omitempty"`
			*alias
		}{
			Strength: 0,
			alias:    (*alias)(o),
		})
	}
	return json.Marshal(&struct{ *alias }{alias: (*alias)(o)})
}

// UnmarshalJSON unmarshal the pbr with the correct default values.
func (p *PBRMetallicRoughness) UnmarshalJSON(data []byte) error {
	type alias PBRMetallicRoughness
	tmp := alias(PBRMetallicRoughness{BaseColorFactor: &[4]float32{1, 1, 1, 1}, MetallicFactor: Float(1), RoughnessFactor: Float(1)})
	err := json.Unmarshal(data, &tmp)
	if err == nil {
		*p = PBRMetallicRoughness(tmp)
	}
	return err
}

// MarshalJSON marshal the pbr with the correct default values.
func (p *PBRMetallicRoughness) MarshalJSON() ([]byte, error) {
	type alias PBRMetallicRoughness
	tmp := &struct {
		alias
	}{
		alias: (alias)(*p),
	}
	if p.MetallicFactor != nil && *p.MetallicFactor == 1 {
		tmp.MetallicFactor = nil
	}
	if p.RoughnessFactor != nil && *p.RoughnessFactor == 1 {
		tmp.RoughnessFactor = nil
	}
	if p.BaseColorFactor != nil && *p.BaseColorFactor == [4]float32{1, 1, 1, 1} {
		tmp.BaseColorFactor = nil
	}
	return json.Marshal(tmp)
}

// UnmarshalJSON unmarshal the extensions with the supported extensions initialized.
func (ext *Extensions) UnmarshalJSON(data []byte) error {
	if len(*ext) == 0 {
		*ext = make(Extensions)
	}
	var raw map[string]json.RawMessage
	err := json.Unmarshal(data, &raw)
	if err == nil {
		for key, value := range raw {
			if extFactory, ok := queryExtension(key); ok {
				n, err := extFactory(value)
				if err != nil {
					(*ext)[key] = value
				} else {
					(*ext)[key] = n
				}
			} else {
				(*ext)[key] = value
			}
		}
	}

	return err
}
