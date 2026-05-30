package lightspuntual

import (
	"encoding/json"
	"math"
	"reflect"
	"testing"

	"github.com/flywave/gltf"
	"github.com/go-test/deep"
	"github.com/stretchr/testify/require"
)

func TestLight_IntensityOrDefault(t *testing.T) {
	tests := []struct {
		name string
		l    *Light
		want float32
	}{
		{"empty", &Light{}, 1},
		{"other", &Light{Intensity: gltf.Float(0.5)}, 0.5},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.l.IntensityOrDefault(); got != tt.want {
				t.Errorf("Light.IntensityOrDefault() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLight_ColorOrDefault(t *testing.T) {
	tests := []struct {
		name string
		l    *Light
		want [3]float32
	}{
		{"empty", &Light{}, [3]float32{1, 1, 1}},
		{"other", &Light{Color: &[3]float32{0.8, 0.8, 0.8}}, [3]float32{0.8, 0.8, 0.8}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.l.ColorOrDefault(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Light.ColorOrDefault() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSpot_OuterConeAngleOrDefault(t *testing.T) {
	tests := []struct {
		name string
		s    *Spot
		want float32
	}{
		{"empty", &Spot{}, math.Pi / 4},
		{"other", &Spot{OuterConeAngle: gltf.Float(0.5)}, 0.5},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.s.OuterConeAngleOrDefault(); got != tt.want {
				t.Errorf("Spot.OuterConeAngleOrDefault() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLight_MarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		l       *Light
		want    []byte
		wantErr bool
	}{
		{
			"directional",
			&Light{
				Type:      "directional",
				Name:      "Sun",
				Color:     &[3]float32{1, 0.9, 0.7},
				Intensity: gltf.Float(3.0),
			},
			[]byte(`{"type":"directional","name":"Sun","color":[1,0.9,0.7],"intensity":3}`),
			false,
		},
		{
			"point",
			&Light{
				Type:      "point",
				Name:      "Point",
				Color:     &[3]float32{1, 0, 0},
				Intensity: gltf.Float(20.0),
			},
			[]byte(`{"type":"point","name":"Point","color":[1,0,0],"intensity":20}`),
			false,
		},
		{
			"spot",
			&Light{
				Type:      "spot",
				Name:      "Spot",
				Color:     &[3]float32{0.3, 0.7, 1},
				Intensity: gltf.Float(40.0),
				Range:     gltf.Float(10.0),
				Spot: &Spot{
					InnerConeAngle: 1.0,
					OuterConeAngle: gltf.Float(2.0),
				},
			},
			[]byte(`{"type":"spot","name":"Spot","color":[0.3,0.7,1],"intensity":40,"range":10,"spot":{"innerConeAngle":1,"outerConeAngle":2}}`),
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := json.Marshal(tt.l)
			if (err != nil) != tt.wantErr {
				t.Errorf("Light.MarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Light.MarshalJSON() = %v, want %v", string(got), string(tt.want))
			}
		})
	}
}

func TestSpot_MarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		s       *Spot
		want    []byte
		wantErr bool
	}{
		{
			"default",
			&Spot{},
			[]byte(`{}`),
			false,
		},
		{
			"full",
			&Spot{
				InnerConeAngle: 0.5,
				OuterConeAngle: gltf.Float(1.0),
			},
			[]byte(`{"innerConeAngle":0.5,"outerConeAngle":1}`),
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := json.Marshal(tt.s)
			if (err != nil) != tt.wantErr {
				t.Errorf("Spot.MarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Spot.MarshalJSON() = %v, want %v", string(got), string(tt.want))
			}
		})
	}
}

func TestSpot_UnmarshalJSON(t *testing.T) {
	type args struct {
		data []byte
	}
	tests := []struct {
		name    string
		s       *Spot
		args    args
		want    *Spot
		wantErr bool
	}{
		{
			"default",
			new(Spot),
			args{[]byte("{}")},
			&Spot{
				OuterConeAngle: gltf.Float(math.Pi / 4),
			},
			false,
		},
		{
			"onlyOuterConeAngle",
			new(Spot),
			args{[]byte(`{"outerConeAngle":1.0}`)},
			&Spot{
				InnerConeAngle: 0,
				OuterConeAngle: gltf.Float(1.0),
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.s.UnmarshalJSON(tt.args.data); (err != nil) != tt.wantErr {
				t.Errorf("Spot.UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(tt.s, tt.want) {
				t.Errorf("Spot.UnmarshalJSON() = %v, want %v", tt.s, tt.want)
			}
		})
	}
}

func TestLight_RoundTrip(t *testing.T) {
	orig := &Light{
		Type:      "spot",
		Name:      "Test",
		Color:     &[3]float32{0.5, 0.6, 0.7},
		Intensity: gltf.Float(2.5),
		Range:     gltf.Float(15.0),
		Spot: &Spot{
			InnerConeAngle: 0.3,
			OuterConeAngle: gltf.Float(0.8),
		},
	}
	data, err := json.Marshal(orig)
	require.NoError(t, err)
	got := new(Light)
	err = json.Unmarshal(data, got)
	require.NoError(t, err)
	require.Equal(t, orig, got)
}

func TestLight_UnmarshalJSON(t *testing.T) {
	type args struct {
		data []byte
	}
	tests := []struct {
		name    string
		l       *Light
		args    args
		want    *Light
		wantErr bool
	}{
		{"default", new(Light), args{[]byte("{}")}, &Light{
			Color: &[3]float32{1, 1, 1}, Intensity: gltf.Float(1), Range: gltf.Float(float32(math.Inf(0))),
		}, false},
		{"directional", new(Light), args{[]byte(`{
			"color": [1.0, 0.9, 0.7],
			"name": "Directional",
			"intensity": 3.0,
			"type": "directional"
		}`)}, &Light{
			Name: "Directional", Type: "directional", Color: &[3]float32{1, 0.9, 0.7}, Intensity: gltf.Float(3.0), Range: gltf.Float(float32(math.Inf(0))),
		}, false},
		{"point", new(Light), args{[]byte(`{
			"color": [1.0, 0.0, 0.0],
			"name": "Point",
			"intensity": 20.0,
			"type": "point"
		}`)}, &Light{
			Name: "Point", Type: "point", Color: &[3]float32{1, 0, 0}, Intensity: gltf.Float(20.0), Range: gltf.Float(float32(math.Inf(0))),
		}, false},
		{"nodefault", new(Light), args{[]byte(`{
			"color": [0.3, 0.7, 1.0],
			"name": "AAA",
			"intensity": 40.0,
			"type": "spot",
			"range": 10.0,
			"spot": {
			  "innerConeAngle": 1.0,
			  "outerConeAngle": 2.0
			}
		  }`)}, &Light{
			Name: "AAA", Type: "spot", Color: &[3]float32{0.3, 0.7, 1}, Intensity: gltf.Float(40), Range: gltf.Float(10),
			Spot: &Spot{
				InnerConeAngle: 1.0,
				OuterConeAngle: gltf.Float(2.0),
			},
		}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.l.UnmarshalJSON(tt.args.data); (err != nil) != tt.wantErr {
				t.Errorf("Light.UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(tt.l, tt.want) {
				t.Errorf("Light.UnmarshalJSON() = %+v, want %+v", tt.l, tt.want)
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
		{"error", args{[]byte(`{"light: 1}`)}, nil, true},
		{"index", args{[]byte(`{"light": 1}`)}, LightIndex(1), false},
		{"lights", args{[]byte(`{"lights": [
			{
			  "color": [1.0, 0.9, 0.7],
			  "name": "Directional",
			  "intensity": 3.0,
			  "type": "directional"
			},
			{
			  "color": [1.0, 0.0, 0.0],
			  "name": "Point",
			  "intensity": 20.0,
			  "type": "point"
			}
		  ]}`)}, Lights{
			{Color: &[3]float32{1, 0.9, 0.7}, Name: "Directional", Intensity: gltf.Float(3.0), Type: "directional", Range: gltf.Float(float32(math.Inf(0)))},
			{Color: &[3]float32{1, 0, 0}, Name: "Point", Intensity: gltf.Float(20.0), Type: "point", Range: gltf.Float(float32(math.Inf(0)))},
		}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Unmarshal(tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("Unmarshal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := deep.Equal(got, tt.want); diff != nil {
				t.Errorf("Unmarshal() = %v", diff)
			}
		})
	}
}
