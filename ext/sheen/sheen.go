package sheen

import (
	"bytes"
	"encoding/json"

	"github.com/flywave/gltf"
)

const (
	// ExtensionName defines the KHR_materials_sheen unique key.
	ExtensionName = "KHR_materials_sheen"
)

// Unmarshal decodes the json data into the correct type.
func Unmarshal(data []byte) (interface{}, error) {
	sheen := new(MaterialsSheen)
	err := json.Unmarshal(data, sheen)
	return sheen, err
}

func init() {
	gltf.RegisterExtension(ExtensionName, Unmarshal)
}

// MaterialsSheen defines the sheen material extension.
type MaterialsSheen struct {
	SheenColorFactor      *[3]float32       `json:"sheenColorFactor,omitempty"`
	SheenColorTexture     *gltf.TextureInfo `json:"sheenColorTexture,omitempty"`
	SheenRoughnessFactor  *float32          `json:"sheenRoughnessFactor,omitempty" validate:"omitempty,gte=0,lte=1"`
	SheenRoughnessTexture *gltf.TextureInfo `json:"sheenRoughnessTexture,omitempty"`
}

// UnmarshalJSON unmarshal the sheen material with the correct default values.
func (m *MaterialsSheen) UnmarshalJSON(data []byte) error {
	type alias MaterialsSheen
	tmp := alias(MaterialsSheen{
		SheenColorFactor:     &[3]float32{0.0, 0.0, 0.0},
		SheenRoughnessFactor: gltf.Float(0.0),
	})
	err := json.Unmarshal(data, &tmp)
	if err == nil {
		*m = MaterialsSheen(tmp)
	}
	return err
}

// MarshalJSON marshal the sheen material with the correct default values.
func (m *MaterialsSheen) MarshalJSON() ([]byte, error) {
	type alias MaterialsSheen
	out, err := json.Marshal(&struct{ *alias }{alias: (*alias)(m)})
	if err == nil {
		if m.SheenColorFactor != nil && *m.SheenColorFactor == [3]float32{0.0, 0.0, 0.0} {
			out = removeProperty([]byte(`"sheenColorFactor":[0,0,0]`), out)
		}
		if m.SheenRoughnessFactor != nil && *m.SheenRoughnessFactor == 0.0 {
			out = removeProperty([]byte(`"sheenRoughnessFactor":0`), out)
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
