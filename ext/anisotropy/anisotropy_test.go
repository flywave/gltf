package anisotropy

import (
	"reflect"
	"testing"

	"github.com/flywave/gltf"
)

func TestMaterialsAnisotropy_UnmarshalJSON(t *testing.T) {
	type args struct {
		data []byte
	}
	tests := []struct {
		name    string
		m       *MaterialsAnisotropy
		args    args
		want    *MaterialsAnisotropy
		wantErr bool
	}{
		{
			"default",
			new(MaterialsAnisotropy),
			args{[]byte("{}")},
			&MaterialsAnisotropy{
				AnisotropyStrength: gltf.Float(0.0),
				AnisotropyRotation: gltf.Float(0.0),
			},
			false,
		},
		{
			"custom",
			new(MaterialsAnisotropy),
			args{[]byte(`{"anisotropyStrength": 0.6, "anisotropyRotation": 1.57}`)},
			&MaterialsAnisotropy{
				AnisotropyStrength: gltf.Float(0.6),
				AnisotropyRotation: gltf.Float(1.57),
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.m.UnmarshalJSON(tt.args.data); (err != nil) != tt.wantErr {
				t.Errorf("MaterialsAnisotropy.UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(tt.m, tt.want) {
				t.Errorf("MaterialsAnisotropy.UnmarshalJSON() = %v, want %v", tt.m, tt.want)
			}
		})
	}
}

func TestMaterialsAnisotropy_MarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		m       *MaterialsAnisotropy
		want    []byte
		wantErr bool
	}{
		{
			"default",
			&MaterialsAnisotropy{
				AnisotropyStrength: gltf.Float(0.0),
				AnisotropyRotation: gltf.Float(0.0),
			},
			[]byte(`{}`),
			false,
		},
		{
			"empty",
			&MaterialsAnisotropy{},
			[]byte(`{}`),
			false,
		},
		{
			"custom",
			&MaterialsAnisotropy{
				AnisotropyStrength: gltf.Float(0.6),
				AnisotropyRotation: gltf.Float(1.57),
			},
			[]byte(`{"anisotropyStrength":0.6,"anisotropyRotation":1.57}`),
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.m.MarshalJSON()
			if (err != nil) != tt.wantErr {
				t.Errorf("MaterialsAnisotropy.MarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MaterialsAnisotropy.MarshalJSON() = %v, want %v", string(got), string(tt.want))
			}
		})
	}
}

func TestUnmarshal(t *testing.T) {
	type args struct {
		data []byte
	}
	tests := []struct {
		name    string
		args    args
		want    interface{}
		wantErr bool
	}{
		{
			"base",
			args{[]byte("{}")},
			&MaterialsAnisotropy{
				AnisotropyStrength: gltf.Float(0.0),
				AnisotropyRotation: gltf.Float(0.0),
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Unmarshal(tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("Unmarshal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Unmarshal() = %v, want %v", got, tt.want)
			}
		})
	}
}
