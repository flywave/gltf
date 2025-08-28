package webp

import (
	"encoding/json"
	"fmt"
	"image"

	fwebp "github.com/flywave/webp"

	"github.com/flywave/gltf"
)

const (
	// TextureWebpExtensionName is the name of the EXT_texture_webp extension
	TextureWebpExtensionName = "EXT_texture_webp"
)

// ExtTextureWebp represents the EXT_texture_webp glTF Texture extension
type ExtTextureWebp struct {
	Source     *uint32                    `json:"source,omitempty"`
	Extensions map[string]json.RawMessage `json:"extensions,omitempty"`
	Extras     json.RawMessage            `json:"extras,omitempty"`
}

// UnmarshalExtTextureWebp unmarshals the EXT_texture_webp extension data
func UnmarshalExtTextureWebp(data []byte) (interface{}, error) {
	var ext ExtTextureWebp
	if err := json.Unmarshal(data, &ext); err != nil {
		return nil, fmt.Errorf("EXT_texture_webp parsing failed: %w", err)
	}

	return &ext, nil
}

// SetSource sets the source image index
func (e *ExtTextureWebp) SetSource(source uint32) {
	e.Source = &source
}

// GetSource returns the source image index
func (e *ExtTextureWebp) GetSource() *uint32 {
	return e.Source
}

func init() {
	image.RegisterFormat("webp", "RIFF????WEBPVP8", fwebp.Decode, fwebp.DecodeConfig)
	gltf.RegisterExtension(TextureWebpExtensionName, UnmarshalExtTextureWebp)
}
