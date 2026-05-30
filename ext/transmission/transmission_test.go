package transmission

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/flywave/gltf"
	"github.com/stretchr/testify/require"
)

func TestMaterialsTransmission_UnmarshalJSON(t *testing.T) {
	type args struct {
		data []byte
	}
	tests := []struct {
		name    string
		m       *MaterialsTransmission
		args    args
		want    *MaterialsTransmission
		wantErr bool
	}{
		{
			"default",
			new(MaterialsTransmission),
			args{[]byte("{}")},
			&MaterialsTransmission{
				TransmissionFactor: gltf.Float(0.0),
			},
			false,
		},
		{
			"custom",
			new(MaterialsTransmission),
			args{[]byte(`{"transmissionFactor": 0.8}`)},
			&MaterialsTransmission{
				TransmissionFactor: gltf.Float(0.8),
			},
			false,
		},
		{
			"withTexture",
			new(MaterialsTransmission),
			args{[]byte(`{"transmissionFactor":0.5,"transmissionTexture":{"index":0,"texCoord":1}}`)},
			&MaterialsTransmission{
				TransmissionFactor:  gltf.Float(0.5),
				TransmissionTexture: &gltf.TextureInfo{Index: 0, TexCoord: 1},
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.m.UnmarshalJSON(tt.args.data); (err != nil) != tt.wantErr {
				t.Errorf("MaterialsTransmission.UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(tt.m, tt.want) {
				t.Errorf("MaterialsTransmission.UnmarshalJSON() = %v, want %v", tt.m, tt.want)
			}
		})
	}
}

func TestMaterialsTransmission_MarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		m       *MaterialsTransmission
		want    []byte
		wantErr bool
	}{
		{
			"default",
			&MaterialsTransmission{
				TransmissionFactor: gltf.Float(0.0),
			},
			[]byte(`{}`),
			false,
		},
		{
			"empty",
			&MaterialsTransmission{},
			[]byte(`{}`),
			false,
		},
		{
			"custom",
			&MaterialsTransmission{
				TransmissionFactor: gltf.Float(0.8),
			},
			[]byte(`{"transmissionFactor":0.8}`),
			false,
		},
		{
			"withTexture",
			&MaterialsTransmission{
				TransmissionFactor:  gltf.Float(0.5),
				TransmissionTexture: &gltf.TextureInfo{Index: 0, TexCoord: 1},
			},
			[]byte(`{"transmissionFactor":0.5,"transmissionTexture":{"index":0,"texCoord":1}}`),
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.m.MarshalJSON()
			if (err != nil) != tt.wantErr {
				t.Errorf("MaterialsTransmission.MarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MaterialsTransmission.MarshalJSON() = %v, want %v", string(got), string(tt.want))
			}
		})
	}
}

func TestMaterialsTransmission_RoundTrip(t *testing.T) {
	orig := &MaterialsTransmission{
		TransmissionFactor:  gltf.Float(0.5),
		TransmissionTexture: &gltf.TextureInfo{Index: 0, TexCoord: 1},
	}
	data, err := json.Marshal(orig)
	require.NoError(t, err)
	got := new(MaterialsTransmission)
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
			&MaterialsTransmission{
				TransmissionFactor: gltf.Float(0.0),
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
