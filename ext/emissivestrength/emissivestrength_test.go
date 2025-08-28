package emissivestrength

import (
	"reflect"
	"testing"

	"github.com/flywave/gltf"
)

func TestMaterialsEmissiveStrength_UnmarshalJSON(t *testing.T) {
	type args struct {
		data []byte
	}
	tests := []struct {
		name    string
		m       *MaterialsEmissiveStrength
		args    args
		want    *MaterialsEmissiveStrength
		wantErr bool
	}{
		{
			"default",
			new(MaterialsEmissiveStrength),
			args{[]byte("{}")},
			&MaterialsEmissiveStrength{
				EmissiveStrength: gltf.Float(1.0),
			},
			false,
		},
		{
			"custom",
			new(MaterialsEmissiveStrength),
			args{[]byte(`{"emissiveStrength": 5.0}`)},
			&MaterialsEmissiveStrength{
				EmissiveStrength: gltf.Float(5.0),
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.m.UnmarshalJSON(tt.args.data); (err != nil) != tt.wantErr {
				t.Errorf("MaterialsEmissiveStrength.UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(tt.m, tt.want) {
				t.Errorf("MaterialsEmissiveStrength.UnmarshalJSON() = %v, want %v", tt.m, tt.want)
			}
		})
	}
}

func TestMaterialsEmissiveStrength_MarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		m       *MaterialsEmissiveStrength
		want    []byte
		wantErr bool
	}{
		{
			"default",
			&MaterialsEmissiveStrength{
				EmissiveStrength: gltf.Float(1.0),
			},
			[]byte(`{}`),
			false,
		},
		{
			"empty",
			&MaterialsEmissiveStrength{},
			[]byte(`{}`),
			false,
		},
		{
			"custom",
			&MaterialsEmissiveStrength{
				EmissiveStrength: gltf.Float(5.0),
			},
			[]byte(`{"emissiveStrength":5}`),
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.m.MarshalJSON()
			if (err != nil) != tt.wantErr {
				t.Errorf("MaterialsEmissiveStrength.MarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MaterialsEmissiveStrength.MarshalJSON() = %v, want %v", string(got), string(tt.want))
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
			&MaterialsEmissiveStrength{
				EmissiveStrength: gltf.Float(1.0),
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
