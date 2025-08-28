package specular

import (
	"bytes"
	"encoding/json"

	"github.com/flywave/gltf"
)

const (
	// ExtensionName defines the KHR_materials_specular unique key.
	ExtensionName = "KHR_materials_specular"
)

// Unmarshal decodes the json data into the correct type.
func Unmarshal(data []byte) (interface{}, error) {
	specular := new(MaterialsSpecular)
	err := json.Unmarshal(data, specular)
	return specular, err
}

func init() {
	gltf.RegisterExtension(ExtensionName, Unmarshal)
}

// MaterialsSpecular defines the specular material extension.
type MaterialsSpecular struct {
	SpecularFactor       *float32          `json:"specularFactor,omitempty" validate:"omitempty,gte=0"`
	SpecularTexture      *gltf.TextureInfo `json:"specularTexture,omitempty"`
	SpecularColorFactor  *[3]float32       `json:"specularColorFactor,omitempty" validate:"omitempty,dive,gte=0,lte=1"`
	SpecularColorTexture *gltf.TextureInfo `json:"specularColorTexture,omitempty"`
}

// UnmarshalJSON unmarshal the specular material with the correct default values.
func (m *MaterialsSpecular) UnmarshalJSON(data []byte) error {
	type alias MaterialsSpecular
	tmp := alias(MaterialsSpecular{
		SpecularFactor:      gltf.Float(1.0),
		SpecularColorFactor: &[3]float32{1, 1, 1},
	})
	err := json.Unmarshal(data, &tmp)
	if err == nil {
		*m = MaterialsSpecular(tmp)
	}
	return err
}

// MarshalJSON marshal the specular material with the correct default values.
func (m *MaterialsSpecular) MarshalJSON() ([]byte, error) {
	type alias MaterialsSpecular
	out, err := json.Marshal(&struct{ *alias }{alias: (*alias)(m)})
	if err == nil {
		if m.SpecularFactor != nil && *m.SpecularFactor == 1.0 {
			out = removeProperty([]byte(`"specularFactor":1`), out)
		}
		if m.SpecularColorFactor != nil && *m.SpecularColorFactor == [3]float32{1, 1, 1} {
			out = removeProperty([]byte(`"specularColorFactor":[1,1,1]`), out)
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
