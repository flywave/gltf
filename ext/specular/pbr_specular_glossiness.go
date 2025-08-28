package specular

import (
	"encoding/json"

	"github.com/flywave/gltf"
)

const (
	// ExtensionName defines the PBRSpecularGlossiness unique key.
	PBRExtensionName = "KHR_materials_pbrSpecularGlossiness"
)

// Unmarshal decodes the json data into the correct type.
func UnmarshalPBR(data []byte) (interface{}, error) {
	pbr := new(PBRSpecularGlossiness)
	err := json.Unmarshal(data, pbr)
	return pbr, err
}

func init() {
	gltf.RegisterExtension(PBRExtensionName, UnmarshalPBR)
}

// PBRSpecularGlossiness defines a specular-glossiness material model.
type PBRSpecularGlossiness struct {
	DiffuseFactor             *[4]float32       `json:"diffuseFactor,omitempty" validate:"omitempty,dive,gte=0,lte=1"`
	DiffuseTexture            *gltf.TextureInfo `json:"diffuseTexture,omitempty"`
	SpecularFactor            *[3]float32       `json:"specularFactor,omitempty" validate:"omitempty,dive,gte=0,lte=1"`
	GlossinessFactor          *float32          `json:"glossinessFactor,omitempty" validate:"omitempty,gte=0,lte=1"`
	SpecularGlossinessTexture *gltf.TextureInfo `json:"specularGlossinessTexture,omitempty"`
}

// UnmarshalJSON unmarshal the pbr with the correct default values.
func (p *PBRSpecularGlossiness) UnmarshalJSON(data []byte) error {
	type alias PBRSpecularGlossiness
	tmp := alias(PBRSpecularGlossiness{DiffuseFactor: &[4]float32{1, 1, 1, 1}, SpecularFactor: &[3]float32{1, 1, 1}, GlossinessFactor: gltf.Float(1)})
	err := json.Unmarshal(data, &tmp)
	if err == nil {
		*p = PBRSpecularGlossiness(tmp)
	}
	return err
}

// MarshalJSON marshal the pbr with the correct default values.
func (p *PBRSpecularGlossiness) MarshalJSON() ([]byte, error) {
	type alias PBRSpecularGlossiness
	out, err := json.Marshal(&struct{ *alias }{alias: (*alias)(p)})
	if err == nil {
		if p.GlossinessFactor != nil && *p.GlossinessFactor == 1 {
			out = removeProperty([]byte(`"glossinessFactor":1`), out)
		}
		if p.DiffuseFactor != nil && *p.DiffuseFactor == [4]float32{1, 1, 1, 1} {
			out = removeProperty([]byte(`"diffuseFactor":[1,1,1,1]`), out)
		}
		if p.SpecularFactor != nil && *p.SpecularFactor == [3]float32{1, 1, 1} {
			out = removeProperty([]byte(`"specularFactor":[1,1,1]`), out)
		}
		out = sanitizeJSON(out)
	}
	return out, err
}
