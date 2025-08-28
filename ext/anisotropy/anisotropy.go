package anisotropy

import (
	"bytes"
	"encoding/json"

	"github.com/flywave/gltf"
)

const (
	// ExtensionName defines the KHR_materials_anisotropy unique key.
	ExtensionName = "KHR_materials_anisotropy"
)

// Unmarshal decodes the json data into the correct type.
func Unmarshal(data []byte) (interface{}, error) {
	anisotropy := new(MaterialsAnisotropy)
	err := json.Unmarshal(data, anisotropy)
	return anisotropy, err
}

func init() {
	gltf.RegisterExtension(ExtensionName, Unmarshal)
}

// MaterialsAnisotropy defines the anisotropy material extension.
type MaterialsAnisotropy struct {
	AnisotropyStrength *float32          `json:"anisotropyStrength,omitempty" validate:"omitempty,gte=0,lte=1"`
	AnisotropyRotation *float32          `json:"anisotropyRotation,omitempty"`
	AnisotropyTexture  *gltf.TextureInfo `json:"anisotropyTexture,omitempty"`
}

// UnmarshalJSON unmarshal the anisotropy material with the correct default values.
func (m *MaterialsAnisotropy) UnmarshalJSON(data []byte) error {
	type alias MaterialsAnisotropy
	tmp := alias(MaterialsAnisotropy{
		AnisotropyStrength: gltf.Float(0.0),
		AnisotropyRotation: gltf.Float(0.0),
	})
	err := json.Unmarshal(data, &tmp)
	if err == nil {
		*m = MaterialsAnisotropy(tmp)
	}
	return err
}

// MarshalJSON marshal the anisotropy material with the correct default values.
func (m *MaterialsAnisotropy) MarshalJSON() ([]byte, error) {
	type alias MaterialsAnisotropy
	out, err := json.Marshal(&struct{ *alias }{alias: (*alias)(m)})
	if err == nil {
		if m.AnisotropyStrength != nil && *m.AnisotropyStrength == 0.0 {
			out = removeProperty([]byte(`"anisotropyStrength":0`), out)
		}
		if m.AnisotropyRotation != nil && *m.AnisotropyRotation == 0.0 {
			out = removeProperty([]byte(`"anisotropyRotation":0`), out)
		}
		out = sanitizeJSON(out)
	}
	return out, err
}

func removeProperty(str []byte, b []byte) []byte {
	b = bytes.Replace(b, str, []byte(""), 1)
	return bytes.Replace(b, []byte(`,,`), []byte(","), 1)
}

func sanitizeJSON(b []byte) []byte {
	b = bytes.Replace(b, []byte(`{,`), []byte("{"), 1)
	return bytes.Replace(b, []byte(`,}`), []byte("}"), 1)
}
