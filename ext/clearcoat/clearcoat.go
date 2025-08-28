package clearcoat

import (
	"bytes"
	"encoding/json"

	"github.com/flywave/gltf"
)

const (
	// ExtensionName defines the KHR_materials_clearcoat unique key.
	ExtensionName = "KHR_materials_clearcoat"
)

// Unmarshal decodes the json data into the correct type.
func Unmarshal(data []byte) (interface{}, error) {
	clearcoat := new(MaterialsClearcoat)
	err := json.Unmarshal(data, clearcoat)
	return clearcoat, err
}

func init() {
	gltf.RegisterExtension(ExtensionName, Unmarshal)
}

// MaterialsClearcoat defines the clearcoat material extension.
type MaterialsClearcoat struct {
	ClearcoatFactor           *float32            `json:"clearcoatFactor,omitempty" validate:"omitempty,gte=0,lte=1"`
	ClearcoatTexture          *gltf.TextureInfo   `json:"clearcoatTexture,omitempty"`
	ClearcoatRoughnessFactor  *float32            `json:"clearcoatRoughnessFactor,omitempty" validate:"omitempty,gte=0,lte=1"`
	ClearcoatRoughnessTexture *gltf.TextureInfo   `json:"clearcoatRoughnessTexture,omitempty"`
	ClearcoatNormalTexture    *gltf.NormalTexture `json:"clearcoatNormalTexture,omitempty"`
}

// UnmarshalJSON unmarshal the clearcoat material with the correct default values.
func (m *MaterialsClearcoat) UnmarshalJSON(data []byte) error {
	type alias MaterialsClearcoat
	tmp := alias(MaterialsClearcoat{
		ClearcoatFactor:          gltf.Float(0.0),
		ClearcoatRoughnessFactor: gltf.Float(0.0),
	})
	err := json.Unmarshal(data, &tmp)
	if err == nil {
		*m = MaterialsClearcoat(tmp)
	}
	return err
}

// MarshalJSON marshal the clearcoat material with the correct default values.
func (m *MaterialsClearcoat) MarshalJSON() ([]byte, error) {
	type alias MaterialsClearcoat
	out, err := json.Marshal(&struct{ *alias }{alias: (*alias)(m)})
	if err == nil {
		if m.ClearcoatFactor != nil && *m.ClearcoatFactor == 0.0 {
			out = removeProperty([]byte(`"clearcoatFactor":0`), out)
		}
		if m.ClearcoatRoughnessFactor != nil && *m.ClearcoatRoughnessFactor == 0.0 {
			out = removeProperty([]byte(`"clearcoatRoughnessFactor":0`), out)
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
