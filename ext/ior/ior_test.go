package ior

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/flywave/gltf"
	"github.com/stretchr/testify/require"
)

func TestMaterialsIOR_UnmarshalJSON(t *testing.T) {
	type args struct {
		data []byte
	}
	tests := []struct {
		name    string
		m       *MaterialsIOR
		args    args
		want    *MaterialsIOR
		wantErr bool
	}{
		{
			"default",
			new(MaterialsIOR),
			args{[]byte("{}")},
			&MaterialsIOR{
				IOR: gltf.Float(1.5),
			},
			false,
		},
		{
			"custom",
			new(MaterialsIOR),
			args{[]byte(`{"ior": 1.4}`)},
			&MaterialsIOR{
				IOR: gltf.Float(1.4),
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.m.UnmarshalJSON(tt.args.data); (err != nil) != tt.wantErr {
				t.Errorf("MaterialsIOR.UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(tt.m, tt.want) {
				t.Errorf("MaterialsIOR.UnmarshalJSON() = %v, want %v", tt.m, tt.want)
			}
		})
	}
}

func TestMaterialsIOR_MarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		m       *MaterialsIOR
		want    []byte
		wantErr bool
	}{
		{
			"default",
			&MaterialsIOR{
				IOR: gltf.Float(1.5),
			},
			[]byte(`{}`),
			false,
		},
		{
			"empty",
			&MaterialsIOR{},
			[]byte(`{}`),
			false,
		},
		{
			"custom",
			&MaterialsIOR{
				IOR: gltf.Float(1.4),
			},
			[]byte(`{"ior":1.4}`),
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.m.MarshalJSON()
			if (err != nil) != tt.wantErr {
				t.Errorf("MaterialsIOR.MarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MaterialsIOR.MarshalJSON() = %v, want %v", string(got), string(tt.want))
			}
		})
	}
}

func TestMaterialsIOR_RoundTrip(t *testing.T) {
	orig := &MaterialsIOR{
		IOR: gltf.Float(1.4),
	}
	data, err := json.Marshal(orig)
	require.NoError(t, err)
	got := new(MaterialsIOR)
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
			&MaterialsIOR{
				IOR: gltf.Float(1.5),
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
