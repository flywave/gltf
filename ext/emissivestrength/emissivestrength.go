package emissivestrength

import (
	"bytes"
	"encoding/json"

	"github.com/flywave/gltf"
)

const (
	// ExtensionName defines the KHR_materials_emissive_strength unique key.
	ExtensionName = "KHR_materials_emissive_strength"
)

// Unmarshal decodes the json data into the correct type.
func Unmarshal(data []byte) (interface{}, error) {
	emissiveStrength := new(MaterialsEmissiveStrength)
	err := json.Unmarshal(data, emissiveStrength)
	return emissiveStrength, err
}

func init() {
	gltf.RegisterExtension(ExtensionName, Unmarshal)
}

// MaterialsEmissiveStrength defines the emissive strength material extension.
type MaterialsEmissiveStrength struct {
	EmissiveStrength *float32 `json:"emissiveStrength,omitempty" validate:"omitempty,gte=0"`
}

// UnmarshalJSON unmarshal the emissive strength material with the correct default values.
func (m *MaterialsEmissiveStrength) UnmarshalJSON(data []byte) error {
	type alias MaterialsEmissiveStrength
	tmp := alias(MaterialsEmissiveStrength{
		EmissiveStrength: gltf.Float(1.0),
	})
	err := json.Unmarshal(data, &tmp)
	if err == nil {
		*m = MaterialsEmissiveStrength(tmp)
	}
	return err
}

// MarshalJSON marshal the emissive strength material with the correct default values.
func (m *MaterialsEmissiveStrength) MarshalJSON() ([]byte, error) {
	type alias MaterialsEmissiveStrength
	out, err := json.Marshal(&struct{ *alias }{alias: (*alias)(m)})
	if err == nil {
		if m.EmissiveStrength != nil && *m.EmissiveStrength == 1.0 {
			out = removeProperty([]byte(`"emissiveStrength":1`), out)
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
