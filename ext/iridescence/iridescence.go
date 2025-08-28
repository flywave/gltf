package iridescence

import (
	"bytes"
	"encoding/json"

	"github.com/flywave/gltf"
)

const (
	// ExtensionName defines the KHR_materials_iridescence unique key.
	ExtensionName = "KHR_materials_iridescence"
)

// Unmarshal decodes the json data into the correct type.
func Unmarshal(data []byte) (interface{}, error) {
	iridescence := new(MaterialsIridescence)
	err := json.Unmarshal(data, iridescence)
	return iridescence, err
}

func init() {
	gltf.RegisterExtension(ExtensionName, Unmarshal)
}

// MaterialsIridescence defines the iridescence material extension.
type MaterialsIridescence struct {
	IridescenceFactor           *float32          `json:"iridescenceFactor,omitempty" validate:"omitempty,gte=0,lte=1"`
	IridescenceTexture          *gltf.TextureInfo `json:"iridescenceTexture,omitempty"`
	IridescenceIor              *float32          `json:"iridescenceIor,omitempty" validate:"omitempty,gte=0"`
	IridescenceThicknessMinimum *float32          `json:"iridescenceThicknessMinimum,omitempty" validate:"omitempty,gte=0"`
	IridescenceThicknessMaximum *float32          `json:"iridescenceThicknessMaximum,omitempty" validate:"omitempty,gte=0"`
	IridescenceThicknessTexture *gltf.TextureInfo `json:"iridescenceThicknessTexture,omitempty"`
}

// UnmarshalJSON unmarshal the iridescence material with the correct default values.
func (m *MaterialsIridescence) UnmarshalJSON(data []byte) error {
	type alias MaterialsIridescence
	tmp := alias(MaterialsIridescence{
		IridescenceFactor:           gltf.Float(0.0),
		IridescenceIor:              gltf.Float(1.3),
		IridescenceThicknessMinimum: gltf.Float(100.0),
		IridescenceThicknessMaximum: gltf.Float(400.0),
	})
	err := json.Unmarshal(data, &tmp)
	if err == nil {
		*m = MaterialsIridescence(tmp)
	}
	return err
}

// MarshalJSON marshal the iridescence material with the correct default values.
func (m *MaterialsIridescence) MarshalJSON() ([]byte, error) {
	type alias MaterialsIridescence
	out, err := json.Marshal(&struct{ *alias }{alias: (*alias)(m)})
	if err == nil {
		if m.IridescenceFactor != nil && *m.IridescenceFactor == 0.0 {
			out = removeProperty([]byte(`"iridescenceFactor":0`), out)
		}
		if m.IridescenceIor != nil && *m.IridescenceIor == 1.3 {
			out = removeProperty([]byte(`"iridescenceIor":1.3`), out)
		}
		if m.IridescenceThicknessMinimum != nil && *m.IridescenceThicknessMinimum == 100.0 {
			out = removeProperty([]byte(`"iridescenceThicknessMinimum":100`), out)
		}
		if m.IridescenceThicknessMaximum != nil && *m.IridescenceThicknessMaximum == 400.0 {
			out = removeProperty([]byte(`"iridescenceThicknessMaximum":400`), out)
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
