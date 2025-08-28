package volume

import (
	"bytes"
	"encoding/json"

	"github.com/flywave/gltf"
)

const (
	// ExtensionName defines the KHR_materials_volume unique key.
	ExtensionName = "KHR_materials_volume"
)

// Unmarshal decodes the json data into the correct type.
func Unmarshal(data []byte) (interface{}, error) {
	volume := new(MaterialsVolume)
	err := json.Unmarshal(data, volume)
	return volume, err
}

func init() {
	gltf.RegisterExtension(ExtensionName, Unmarshal)
}

// MaterialsVolume defines the volume material extension.
type MaterialsVolume struct {
	ThicknessFactor     *float32          `json:"thicknessFactor,omitempty" validate:"omitempty,gte=0"`
	ThicknessTexture    *gltf.TextureInfo `json:"thicknessTexture,omitempty"`
	AttenuationDistance *float32          `json:"attenuationDistance,omitempty" validate:"omitempty,gt=0"`
	AttenuationColor    *[3]float32       `json:"attenuationColor,omitempty"`
}

// UnmarshalJSON unmarshal the volume material with the correct default values.
func (m *MaterialsVolume) UnmarshalJSON(data []byte) error {
	type alias MaterialsVolume
	tmp := alias(MaterialsVolume{
		ThicknessFactor:     gltf.Float(0.0),
		AttenuationDistance: gltf.Float(-1.0), // -1 represents positive infinity
		AttenuationColor:    &[3]float32{1.0, 1.0, 1.0},
	})
	err := json.Unmarshal(data, &tmp)
	if err == nil {
		*m = MaterialsVolume(tmp)
	}
	return err
}

// MarshalJSON marshal the volume material with the correct default values.
func (m *MaterialsVolume) MarshalJSON() ([]byte, error) {
	type alias MaterialsVolume
	out, err := json.Marshal(&struct{ *alias }{alias: (*alias)(m)})
	if err == nil {
		if m.ThicknessFactor != nil && *m.ThicknessFactor == 0.0 {
			out = removeProperty([]byte(`"thicknessFactor":0`), out)
		}
		if m.AttenuationDistance != nil && *m.AttenuationDistance == -1.0 {
			out = removeProperty([]byte(`"attenuationDistance":-1`), out)
		}
		if m.AttenuationColor != nil && *m.AttenuationColor == [3]float32{1.0, 1.0, 1.0} {
			out = removeProperty([]byte(`"attenuationColor":[1,1,1]`), out)
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
