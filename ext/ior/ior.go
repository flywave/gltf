package ior

import (
	"bytes"
	"encoding/json"

	"github.com/flywave/gltf"
)

const (
	// ExtensionName defines the KHR_materials_ior unique key.
	ExtensionName = "KHR_materials_ior"
)

// Unmarshal decodes the json data into the correct type.
func Unmarshal(data []byte) (interface{}, error) {
	ior := new(MaterialsIOR)
	err := json.Unmarshal(data, ior)
	return ior, err
}

func init() {
	gltf.RegisterExtension(ExtensionName, Unmarshal)
}

// MaterialsIOR defines the index of refraction material extension.
type MaterialsIOR struct {
	IOR *float32 `json:"ior,omitempty" validate:"omitempty,gte=0"`
}

// UnmarshalJSON unmarshal the ior material with the correct default values.
func (m *MaterialsIOR) UnmarshalJSON(data []byte) error {
	type alias MaterialsIOR
	tmp := alias(MaterialsIOR{
		IOR: gltf.Float(1.5),
	})
	err := json.Unmarshal(data, &tmp)
	if err == nil {
		*m = MaterialsIOR(tmp)
	}
	return err
}

// MarshalJSON marshal the ior material with the correct default values.
func (m *MaterialsIOR) MarshalJSON() ([]byte, error) {
	type alias MaterialsIOR
	out, err := json.Marshal(&struct{ *alias }{alias: (*alias)(m)})
	if err == nil {
		if m.IOR != nil && *m.IOR == 1.5 {
			out = removeProperty([]byte(`"ior":1.5`), out)
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
