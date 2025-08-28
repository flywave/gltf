package texturebasisu

import (
	"encoding/json"
	"fmt"

	"github.com/flywave/gltf"
)

const (
	// TextureBasisuExtensionName is the name of the KHR_texture_basisu extension
	TextureBasisuExtensionName = "KHR_texture_basisu"
)

// ExtTextureBasisu represents the KHR_texture_basisu glTF Texture extension
type ExtTextureBasisu struct {
	Source     uint32                     `json:"source"`
	Extensions map[string]json.RawMessage `json:"extensions,omitempty"`
	Extras     json.RawMessage            `json:"extras,omitempty"`
}

// UnmarshalExtTextureBasisu unmarshals the KHR_texture_basisu extension data
func UnmarshalExtTextureBasisu(data []byte) (interface{}, error) {
	var ext ExtTextureBasisu
	if err := json.Unmarshal(data, &ext); err != nil {
		return nil, fmt.Errorf("KHR_texture_basisu parsing failed: %w", err)
	}

	return &ext, nil
}

// SetSource sets the source image index
func (e *ExtTextureBasisu) SetSource(source uint32) {
	e.Source = source
}

// GetSource returns the source image index
func (e *ExtTextureBasisu) GetSource() uint32 {
	return e.Source
}

func init() {
	gltf.RegisterExtension(TextureBasisuExtensionName, UnmarshalExtTextureBasisu)
}
