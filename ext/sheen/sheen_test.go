package sheen

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/flywave/gltf"
	"github.com/stretchr/testify/require"
)

func TestMaterialsSheen_UnmarshalJSON(t *testing.T) {
	type args struct {
		data []byte
	}
	tests := []struct {
		name    string
		m       *MaterialsSheen
		args    args
		want    *MaterialsSheen
		wantErr bool
	}{
		{
			"default",
			new(MaterialsSheen),
			args{[]byte("{}")},
			&MaterialsSheen{
				SheenColorFactor:     &[3]float32{0.0, 0.0, 0.0},
				SheenRoughnessFactor: gltf.Float(0.0),
			},
			false,
		},
		{
			"custom",
			new(MaterialsSheen),
			args{[]byte(`{"sheenColorFactor": [0.9, 0.9, 0.9], "sheenRoughnessFactor": 0.5}`)},
			&MaterialsSheen{
				SheenColorFactor:     &[3]float32{0.9, 0.9, 0.9},
				SheenRoughnessFactor: gltf.Float(0.5),
			},
			false,
		},
		{
			"withTextures",
			new(MaterialsSheen),
			args{[]byte(`{"sheenColorFactor":[0.5,0.5,0.5],"sheenRoughnessFactor":0.3,"sheenColorTexture":{"index":0,"texCoord":1},"sheenRoughnessTexture":{"index":1}}`)},
			&MaterialsSheen{
				SheenColorFactor:      &[3]float32{0.5, 0.5, 0.5},
				SheenRoughnessFactor:  gltf.Float(0.3),
				SheenColorTexture:     &gltf.TextureInfo{Index: 0, TexCoord: 1},
				SheenRoughnessTexture: &gltf.TextureInfo{Index: 1},
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.m.UnmarshalJSON(tt.args.data); (err != nil) != tt.wantErr {
				t.Errorf("MaterialsSheen.UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(tt.m, tt.want) {
				t.Errorf("MaterialsSheen.UnmarshalJSON() = %v, want %v", tt.m, tt.want)
			}
		})
	}
}

func TestMaterialsSheen_MarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		m       *MaterialsSheen
		want    []byte
		wantErr bool
	}{
		{
			"default",
			&MaterialsSheen{
				SheenColorFactor:     &[3]float32{0.0, 0.0, 0.0},
				SheenRoughnessFactor: gltf.Float(0.0),
			},
			[]byte(`{}`),
			false,
		},
		{
			"empty",
			&MaterialsSheen{},
			[]byte(`{}`),
			false,
		},
		{
			"custom",
			&MaterialsSheen{
				SheenColorFactor:     &[3]float32{0.9, 0.9, 0.9},
				SheenRoughnessFactor: gltf.Float(0.5),
			},
			[]byte(`{"sheenColorFactor":[0.9,0.9,0.9],"sheenRoughnessFactor":0.5}`),
			false,
		},
		{
			"withTextures",
			&MaterialsSheen{
				SheenColorFactor:      &[3]float32{0.5, 0.5, 0.5},
				SheenRoughnessFactor:  gltf.Float(0.3),
				SheenColorTexture:     &gltf.TextureInfo{Index: 0, TexCoord: 1},
				SheenRoughnessTexture: &gltf.TextureInfo{Index: 1},
			},
			[]byte(`{"sheenColorFactor":[0.5,0.5,0.5],"sheenColorTexture":{"index":0,"texCoord":1},"sheenRoughnessFactor":0.3,"sheenRoughnessTexture":{"index":1}}`),
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.m.MarshalJSON()
			if (err != nil) != tt.wantErr {
				t.Errorf("MaterialsSheen.MarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MaterialsSheen.MarshalJSON() = %v, want %v", string(got), string(tt.want))
			}
		})
	}
}

func TestMaterialsSheen_RoundTrip(t *testing.T) {
	orig := &MaterialsSheen{
		SheenColorFactor:      &[3]float32{0.5, 0.5, 0.5},
		SheenRoughnessFactor:  gltf.Float(0.3),
		SheenColorTexture:     &gltf.TextureInfo{Index: 0, TexCoord: 1},
		SheenRoughnessTexture: &gltf.TextureInfo{Index: 1},
	}
	data, err := json.Marshal(orig)
	require.NoError(t, err)
	got := new(MaterialsSheen)
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
			&MaterialsSheen{
				SheenColorFactor:     &[3]float32{0.0, 0.0, 0.0},
				SheenRoughnessFactor: gltf.Float(0.0),
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
