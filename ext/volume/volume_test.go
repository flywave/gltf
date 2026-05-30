package volume

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/flywave/gltf"
	"github.com/stretchr/testify/require"
)

func TestMaterialsVolume_UnmarshalJSON(t *testing.T) {
	type args struct {
		data []byte
	}
	tests := []struct {
		name    string
		m       *MaterialsVolume
		args    args
		want    *MaterialsVolume
		wantErr bool
	}{
		{
			"default",
			new(MaterialsVolume),
			args{[]byte("{}")},
			&MaterialsVolume{
				ThicknessFactor:     gltf.Float(0.0),
				AttenuationDistance: gltf.Float(-1.0),
				AttenuationColor:    &[3]float32{1.0, 1.0, 1.0},
			},
			false,
		},
		{
			"custom",
			new(MaterialsVolume),
			args{[]byte(`{"thicknessFactor": 1.0, "attenuationDistance": 0.006, "attenuationColor": [0.5, 0.5, 0.5]}`)},
			&MaterialsVolume{
				ThicknessFactor:     gltf.Float(1.0),
				AttenuationDistance: gltf.Float(0.006),
				AttenuationColor:    &[3]float32{0.5, 0.5, 0.5},
			},
			false,
		},
		{
			"withTexture",
			new(MaterialsVolume),
			args{[]byte(`{"thicknessFactor":0.5,"thicknessTexture":{"index":0,"texCoord":1},"attenuationDistance":0.1,"attenuationColor":[0.8,0.8,0.8]}`)},
			&MaterialsVolume{
				ThicknessFactor:     gltf.Float(0.5),
				ThicknessTexture:    &gltf.TextureInfo{Index: 0, TexCoord: 1},
				AttenuationDistance: gltf.Float(0.1),
				AttenuationColor:    &[3]float32{0.8, 0.8, 0.8},
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.m.UnmarshalJSON(tt.args.data); (err != nil) != tt.wantErr {
				t.Errorf("MaterialsVolume.UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(tt.m, tt.want) {
				t.Errorf("MaterialsVolume.UnmarshalJSON() = %v, want %v", tt.m, tt.want)
			}
		})
	}
}

func TestMaterialsVolume_MarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		m       *MaterialsVolume
		want    []byte
		wantErr bool
	}{
		{
			"default",
			&MaterialsVolume{
				ThicknessFactor:     gltf.Float(0.0),
				AttenuationDistance: gltf.Float(-1.0),
				AttenuationColor:    &[3]float32{1.0, 1.0, 1.0},
			},
			[]byte(`{}`),
			false,
		},
		{
			"empty",
			&MaterialsVolume{},
			[]byte(`{}`),
			false,
		},
		{
			"custom",
			&MaterialsVolume{
				ThicknessFactor:     gltf.Float(1.0),
				AttenuationDistance: gltf.Float(0.006),
				AttenuationColor:    &[3]float32{0.5, 0.5, 0.5},
			},
			[]byte(`{"thicknessFactor":1,"attenuationDistance":0.006,"attenuationColor":[0.5,0.5,0.5]}`),
			false,
		},
		{
			"withTexture",
			&MaterialsVolume{
				ThicknessFactor:     gltf.Float(0.5),
				ThicknessTexture:    &gltf.TextureInfo{Index: 0, TexCoord: 1},
				AttenuationDistance: gltf.Float(0.1),
				AttenuationColor:    &[3]float32{0.8, 0.8, 0.8},
			},
			[]byte(`{"thicknessFactor":0.5,"thicknessTexture":{"index":0,"texCoord":1},"attenuationDistance":0.1,"attenuationColor":[0.8,0.8,0.8]}`),
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.m.MarshalJSON()
			if (err != nil) != tt.wantErr {
				t.Errorf("MaterialsVolume.MarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MaterialsVolume.MarshalJSON() = %v, want %v", string(got), string(tt.want))
			}
		})
	}
}

func TestMaterialsVolume_RoundTrip(t *testing.T) {
	orig := &MaterialsVolume{
		ThicknessFactor:     gltf.Float(0.5),
		ThicknessTexture:    &gltf.TextureInfo{Index: 0, TexCoord: 1},
		AttenuationDistance: gltf.Float(0.1),
		AttenuationColor:    &[3]float32{0.8, 0.8, 0.8},
	}
	data, err := json.Marshal(orig)
	require.NoError(t, err)
	got := new(MaterialsVolume)
	err = json.Unmarshal(data, got)
	require.NoError(t, err)
	require.Equal(t, orig, got)
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
			&MaterialsVolume{
				ThicknessFactor:     gltf.Float(0.0),
				AttenuationDistance: gltf.Float(-1.0),
				AttenuationColor:    &[3]float32{1.0, 1.0, 1.0},
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
