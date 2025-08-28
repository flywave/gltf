package specular

import (
	"reflect"
	"testing"

	"github.com/flywave/gltf"
)

func TestMaterialsSpecular_UnmarshalJSON(t *testing.T) {
	type args struct {
		data []byte
	}
	tests := []struct {
		name    string
		m       *MaterialsSpecular
		args    args
		want    *MaterialsSpecular
		wantErr bool
	}{
		{
			"default",
			new(MaterialsSpecular),
			args{[]byte("{}")},
			&MaterialsSpecular{
				SpecularFactor:      gltf.Float(1.0),
				SpecularColorFactor: &[3]float32{1, 1, 1},
			},
			false,
		},
		{
			"custom",
			new(MaterialsSpecular),
			args{[]byte(`{"specularFactor": 0.5, "specularColorFactor": [0.8, 0.6, 0.4]}`)},
			&MaterialsSpecular{
				SpecularFactor:      gltf.Float(0.5),
				SpecularColorFactor: &[3]float32{0.8, 0.6, 0.4},
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.m.UnmarshalJSON(tt.args.data); (err != nil) != tt.wantErr {
				t.Errorf("MaterialsSpecular.UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(tt.m, tt.want) {
				t.Errorf("MaterialsSpecular.UnmarshalJSON() = %v, want %v", tt.m, tt.want)
			}
		})
	}
}

func TestMaterialsSpecular_MarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		m       *MaterialsSpecular
		want    []byte
		wantErr bool
	}{
		{
			"default",
			&MaterialsSpecular{
				SpecularFactor:      gltf.Float(1.0),
				SpecularColorFactor: &[3]float32{1, 1, 1},
			},
			[]byte(`{}`),
			false,
		},
		{
			"empty",
			&MaterialsSpecular{},
			[]byte(`{}`),
			false,
		},
		{
			"custom",
			&MaterialsSpecular{
				SpecularFactor:      gltf.Float(0.5),
				SpecularColorFactor: &[3]float32{0.8, 0.6, 0.4},
			},
			[]byte(`{"specularFactor":0.5,"specularColorFactor":[0.8,0.6,0.4]}`),
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.m.MarshalJSON()
			if (err != nil) != tt.wantErr {
				t.Errorf("MaterialsSpecular.MarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MaterialsSpecular.MarshalJSON() = %v, want %v", string(got), string(tt.want))
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
			&MaterialsSpecular{
				SpecularFactor:      gltf.Float(1.0),
				SpecularColorFactor: &[3]float32{1, 1, 1},
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
