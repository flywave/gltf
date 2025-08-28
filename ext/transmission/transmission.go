package transmission

import (
	"bytes"
	"encoding/json"

	"github.com/flywave/gltf"
)

const (
	// ExtensionName defines the KHR_materials_transmission unique key.
	ExtensionName = "KHR_materials_transmission"
)

// Unmarshal decodes the json data into the correct type.
func Unmarshal(data []byte) (interface{}, error) {
	transmission := new(MaterialsTransmission)
	err := json.Unmarshal(data, transmission)
	return transmission, err
}

func init() {
	gltf.RegisterExtension(ExtensionName, Unmarshal)
}

// MaterialsTransmission defines the transmission material extension.
type MaterialsTransmission struct {
	TransmissionFactor  *float32          `json:"transmissionFactor,omitempty" validate:"omitempty,gte=0,lte=1"`
	TransmissionTexture *gltf.TextureInfo `json:"transmissionTexture,omitempty"`
}

// UnmarshalJSON unmarshal the transmission material with the correct default values.
func (m *MaterialsTransmission) UnmarshalJSON(data []byte) error {
	type alias MaterialsTransmission
	tmp := alias(MaterialsTransmission{
		TransmissionFactor: gltf.Float(0.0),
	})
	err := json.Unmarshal(data, &tmp)
	if err == nil {
		*m = MaterialsTransmission(tmp)
	}
	return err
}

// MarshalJSON marshal the transmission material with the correct default values.
func (m *MaterialsTransmission) MarshalJSON() ([]byte, error) {
	type alias MaterialsTransmission
	out, err := json.Marshal(&struct{ *alias }{alias: (*alias)(m)})
	if err == nil {
		if m.TransmissionFactor != nil && *m.TransmissionFactor == 0.0 {
			out = removeProperty([]byte(`"transmissionFactor":0`), out)
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
