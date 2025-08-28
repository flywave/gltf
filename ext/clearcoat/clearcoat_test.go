package clearcoat

import (
	"reflect"
	"testing"

	"github.com/flywave/gltf"
)

func TestMaterialsClearcoat_UnmarshalJSON(t *testing.T) {
	type args struct {
		data []byte
	}
	tests := []struct {
		name    string
		m       *MaterialsClearcoat
		args    args
		want    *MaterialsClearcoat
		wantErr bool
	}{
		{
			"default",
			new(MaterialsClearcoat),
			args{[]byte("{}")},
			&MaterialsClearcoat{
				ClearcoatFactor:          gltf.Float(0.0),
				ClearcoatRoughnessFactor: gltf.Float(0.0),
			},
			false,
		},
		{
			"custom",
			new(MaterialsClearcoat),
			args{[]byte(`{"clearcoatFactor": 1.0, "clearcoatRoughnessFactor": 0.5}`)},
			&MaterialsClearcoat{
				ClearcoatFactor:          gltf.Float(1.0),
				ClearcoatRoughnessFactor: gltf.Float(0.5),
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.m.UnmarshalJSON(tt.args.data); (err != nil) != tt.wantErr {
				t.Errorf("MaterialsClearcoat.UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(tt.m, tt.want) {
				t.Errorf("MaterialsClearcoat.UnmarshalJSON() = %v, want %v", tt.m, tt.want)
			}
		})
	}
}

func TestMaterialsClearcoat_MarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		m       *MaterialsClearcoat
		want    []byte
		wantErr bool
	}{
		{
			"default",
			&MaterialsClearcoat{
				ClearcoatFactor:          gltf.Float(0.0),
				ClearcoatRoughnessFactor: gltf.Float(0.0),
			},
			[]byte(`{}`),
			false,
		},
		{
			"empty",
			&MaterialsClearcoat{},
			[]byte(`{}`),
			false,
		},
		{
			"custom",
			&MaterialsClearcoat{
				ClearcoatFactor:          gltf.Float(1.0),
				ClearcoatRoughnessFactor: gltf.Float(0.5),
			},
			[]byte(`{"clearcoatFactor":1,"clearcoatRoughnessFactor":0.5}`),
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.m.MarshalJSON()
			if (err != nil) != tt.wantErr {
				t.Errorf("MaterialsClearcoat.MarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MaterialsClearcoat.MarshalJSON() = %v, want %v", string(got), string(tt.want))
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
			&MaterialsClearcoat{
				ClearcoatFactor:          gltf.Float(0.0),
				ClearcoatRoughnessFactor: gltf.Float(0.0),
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
