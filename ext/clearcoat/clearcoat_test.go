package clearcoat

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/flywave/gltf"
	"github.com/stretchr/testify/require"
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
		{
			"withTextures",
			new(MaterialsClearcoat),
			args{[]byte(`{"clearcoatFactor":0.8,"clearcoatRoughnessFactor":0.3,"clearcoatTexture":{"index":0,"texCoord":1},"clearcoatRoughnessTexture":{"index":1},"clearcoatNormalTexture":{"index":2,"scale":0.5,"texCoord":1}}`)},
			&MaterialsClearcoat{
				ClearcoatFactor:          gltf.Float(0.8),
				ClearcoatRoughnessFactor: gltf.Float(0.3),
				ClearcoatTexture:         &gltf.TextureInfo{Index: 0, TexCoord: 1},
				ClearcoatRoughnessTexture: &gltf.TextureInfo{Index: 1},
				ClearcoatNormalTexture: &gltf.NormalTexture{
					Index:   gltf.Index(2),
					Scale:   gltf.Float(0.5),
					TexCoord: 1,
				},
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
		{
			"withTextures",
			&MaterialsClearcoat{
				ClearcoatFactor:          gltf.Float(0.8),
				ClearcoatRoughnessFactor: gltf.Float(0.3),
				ClearcoatTexture:         &gltf.TextureInfo{Index: 0, TexCoord: 1},
				ClearcoatRoughnessTexture: &gltf.TextureInfo{Index: 1},
				ClearcoatNormalTexture: &gltf.NormalTexture{
					Index:   gltf.Index(2),
					Scale:   gltf.Float(0.5),
					TexCoord: 1,
				},
			},
			[]byte(`{"clearcoatFactor":0.8,"clearcoatTexture":{"index":0,"texCoord":1},"clearcoatRoughnessFactor":0.3,"clearcoatRoughnessTexture":{"index":1},"clearcoatNormalTexture":{"index":2,"texCoord":1,"scale":0.5}}`),
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

func TestMaterialsClearcoat_RoundTrip(t *testing.T) {
	orig := &MaterialsClearcoat{
		ClearcoatFactor:          gltf.Float(0.8),
		ClearcoatRoughnessFactor: gltf.Float(0.3),
		ClearcoatTexture:         &gltf.TextureInfo{Index: 0, TexCoord: 1},
		ClearcoatRoughnessTexture: &gltf.TextureInfo{Index: 1},
		ClearcoatNormalTexture: &gltf.NormalTexture{
			Index:   gltf.Index(2),
			Scale:   gltf.Float(0.5),
			TexCoord: 1,
		},
	}
	data, err := json.Marshal(orig)
	require.NoError(t, err)
	got := new(MaterialsClearcoat)
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
