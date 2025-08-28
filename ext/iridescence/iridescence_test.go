package iridescence

import (
	"reflect"
	"testing"

	"github.com/flywave/gltf"
)

func TestMaterialsIridescence_UnmarshalJSON(t *testing.T) {
	type args struct {
		data []byte
	}
	tests := []struct {
		name    string
		m       *MaterialsIridescence
		args    args
		want    *MaterialsIridescence
		wantErr bool
	}{
		{
			"default",
			new(MaterialsIridescence),
			args{[]byte("{}")},
			&MaterialsIridescence{
				IridescenceFactor:           gltf.Float(0.0),
				IridescenceIor:              gltf.Float(1.3),
				IridescenceThicknessMinimum: gltf.Float(100.0),
				IridescenceThicknessMaximum: gltf.Float(400.0),
			},
			false,
		},
		{
			"custom",
			new(MaterialsIridescence),
			args{[]byte(`{"iridescenceFactor": 1.0, "iridescenceIor": 1.5, "iridescenceThicknessMaximum": 500.0}`)},
			&MaterialsIridescence{
				IridescenceFactor:           gltf.Float(1.0),
				IridescenceIor:              gltf.Float(1.5),
				IridescenceThicknessMinimum: gltf.Float(100.0),
				IridescenceThicknessMaximum: gltf.Float(500.0),
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.m.UnmarshalJSON(tt.args.data); (err != nil) != tt.wantErr {
				t.Errorf("MaterialsIridescence.UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(tt.m, tt.want) {
				t.Errorf("MaterialsIridescence.UnmarshalJSON() = %v, want %v", tt.m, tt.want)
			}
		})
	}
}

func TestMaterialsIridescence_MarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		m       *MaterialsIridescence
		want    []byte
		wantErr bool
	}{
		{
			"default",
			&MaterialsIridescence{
				IridescenceFactor:           gltf.Float(0.0),
				IridescenceIor:              gltf.Float(1.3),
				IridescenceThicknessMinimum: gltf.Float(100.0),
				IridescenceThicknessMaximum: gltf.Float(400.0),
			},
			[]byte(`{}`),
			false,
		},
		{
			"empty",
			&MaterialsIridescence{},
			[]byte(`{}`),
			false,
		},
		{
			"custom",
			&MaterialsIridescence{
				IridescenceFactor:           gltf.Float(1.0),
				IridescenceIor:              gltf.Float(1.5),
				IridescenceThicknessMinimum: gltf.Float(200.0),
				IridescenceThicknessMaximum: gltf.Float(500.0),
			},
			[]byte(`{"iridescenceFactor":1,"iridescenceIor":1.5,"iridescenceThicknessMinimum":200,"iridescenceThicknessMaximum":500}`),
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.m.MarshalJSON()
			if (err != nil) != tt.wantErr {
				t.Errorf("MaterialsIridescence.MarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MaterialsIridescence.MarshalJSON() = %v, want %v", string(got), string(tt.want))
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
			&MaterialsIridescence{
				IridescenceFactor:           gltf.Float(0.0),
				IridescenceIor:              gltf.Float(1.3),
				IridescenceThicknessMinimum: gltf.Float(100.0),
				IridescenceThicknessMaximum: gltf.Float(400.0),
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
