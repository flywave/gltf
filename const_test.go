package gltf

import (
	"encoding/json"
	"testing"
)

func TestAccessorType_JSON(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    AccessorType
		wantErr bool
	}{
		{"scalar", `"SCALAR"`, AccessorScalar, false},
		{"vec2", `"VEC2"`, AccessorVec2, false},
		{"vec3", `"VEC3"`, AccessorVec3, false},
		{"vec4", `"VEC4"`, AccessorVec4, false},
		{"mat2", `"MAT2"`, AccessorMat2, false},
		{"mat3", `"MAT3"`, AccessorMat3, false},
		{"mat4", `"MAT4"`, AccessorMat4, false},
		{"unknown", `"CUSTOM"`, 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got AccessorType
			err := json.Unmarshal([]byte(tt.input), &got)
			if (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}
			if got != tt.want {
				t.Errorf("UnmarshalJSON() = %v, want %v", got, tt.want)
			}
			// Round-trip: use &got because MarshalJSON has pointer receiver
			out, err := json.Marshal(&got)
			if err != nil {
				t.Errorf("MarshalJSON() error = %v", err)
				return
			}
			if string(out) != tt.input {
				t.Errorf("MarshalJSON() = %s, want %s", string(out), tt.input)
			}
		})
	}
}

func TestComponentType_JSON(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    ComponentType
		wantErr bool
	}{
		{"float", `5126`, ComponentFloat, false},
		{"byte", `5120`, ComponentByte, false},
		{"ubyte", `5121`, ComponentUbyte, false},
		{"short", `5122`, ComponentShort, false},
		{"ushort", `5123`, ComponentUshort, false},
		{"uint", `5125`, ComponentUint, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got ComponentType
			err := json.Unmarshal([]byte(tt.input), &got)
			if err != nil {
				t.Errorf("UnmarshalJSON() error = %v", err)
				return
			}
			if got != tt.want {
				t.Errorf("UnmarshalJSON() = %v, want %v", got, tt.want)
			}
			out, err := json.Marshal(&got)
			if err != nil {
				t.Errorf("MarshalJSON() error = %v", err)
				return
			}
			if string(out) != tt.input {
				t.Errorf("MarshalJSON() = %s, want %s", string(out), tt.input)
			}
		})
	}
}

func TestComponentType_ByteSize(t *testing.T) {
	tests := []struct {
		typ ComponentType
		n   uint32
	}{
		{ComponentByte, 1},
		{ComponentUbyte, 1},
		{ComponentShort, 2},
		{ComponentUshort, 2},
		{ComponentUint, 4},
		{ComponentFloat, 4},
	}
	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			if got := tt.typ.ByteSize(); got != tt.n {
				t.Errorf("ByteSize() = %d, want %d", got, tt.n)
			}
		})
	}
}

func TestAccessorType_Components(t *testing.T) {
	tests := []struct {
		typ AccessorType
		n   uint32
	}{
		{AccessorScalar, 1},
		{AccessorVec2, 2},
		{AccessorVec3, 3},
		{AccessorVec4, 4},
		{AccessorMat2, 4},
		{AccessorMat3, 9},
		{AccessorMat4, 16},
	}
	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			if got := tt.typ.Components(); got != tt.n {
				t.Errorf("Components() = %d, want %d", got, tt.n)
			}
		})
	}
}

func TestPrimitiveMode_JSON(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  PrimitiveMode
	}{
		{"points", `0`, PrimitivePoints},
		{"lines", `1`, PrimitiveLines},
		{"line_loop", `2`, PrimitiveLineLoop},
		{"line_strip", `3`, PrimitiveLineStrip},
		{"triangles", `4`, PrimitiveTriangles},
		{"triangle_strip", `5`, PrimitiveTriangleStrip},
		{"triangle_fan", `6`, PrimitiveTriangleFan},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got PrimitiveMode
			err := json.Unmarshal([]byte(tt.input), &got)
			if err != nil {
				t.Errorf("UnmarshalJSON() error = %v", err)
				return
			}
			if got != tt.want {
				t.Errorf("UnmarshalJSON() = %v, want %v", got, tt.want)
			}
			out, err := json.Marshal(&got)
			if err != nil {
				t.Errorf("MarshalJSON() error = %v", err)
				return
			}
			if string(out) != tt.input {
				t.Errorf("MarshalJSON() = %s, want %s", string(out), tt.input)
			}
		})
	}
}

func TestPrimitiveMode_Unmarshal_Invalid(t *testing.T) {
	var m PrimitiveMode
	err := json.Unmarshal([]byte(`255`), &m)
	// UnmarshalJSON silently maps unknown values to zero (PrimitiveTriangles)
	// instead of returning an error; verify the codec handles this gracefully
	if err != nil {
		t.Errorf("UnmarshalJSON() should not error on unknown value, got %v", err)
	}
}

func TestAlphaMode_JSON(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  AlphaMode
	}{
		{"opaque", `"OPAQUE"`, AlphaOpaque},
		{"mask", `"MASK"`, AlphaMask},
		{"blend", `"BLEND"`, AlphaBlend},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got AlphaMode
			err := json.Unmarshal([]byte(tt.input), &got)
			if err != nil {
				t.Errorf("UnmarshalJSON() error = %v", err)
				return
			}
			if got != tt.want {
				t.Errorf("UnmarshalJSON() = %v, want %v", got, tt.want)
			}
			out, err := json.Marshal(&got)
			if err != nil {
				t.Errorf("MarshalJSON() error = %v", err)
				return
			}
			if string(out) != tt.input {
				t.Errorf("MarshalJSON() = %s, want %s", string(out), tt.input)
			}
		})
	}
}

func TestMagFilter_JSON(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  MagFilter
	}{
		{"nearest", `9728`, MagNearest},
		{"linear", `9729`, MagLinear},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got MagFilter
			err := json.Unmarshal([]byte(tt.input), &got)
			if err != nil {
				t.Errorf("UnmarshalJSON() error = %v", err)
				return
			}
			if got != tt.want {
				t.Errorf("UnmarshalJSON() = %v, want %v", got, tt.want)
			}
			out, err := json.Marshal(&got)
			if err != nil {
				t.Errorf("MarshalJSON() error = %v", err)
				return
			}
			if string(out) != tt.input {
				t.Errorf("MarshalJSON() = %s, want %s", string(out), tt.input)
			}
		})
	}
}

func TestMinFilter_JSON(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  MinFilter
	}{
		{"nearest", `9728`, MinNearest},
		{"linear", `9729`, MinLinear},
		{"nearest_mipmap_nearest", `9984`, MinNearestMipMapNearest},
		{"linear_mipmap_nearest", `9985`, MinLinearMipMapNearest},
		{"nearest_mipmap_linear", `9986`, MinNearestMipMapLinear},
		{"linear_mipmap_linear", `9987`, MinLinearMipMapLinear},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got MinFilter
			err := json.Unmarshal([]byte(tt.input), &got)
			if err != nil {
				t.Errorf("UnmarshalJSON() error = %v", err)
				return
			}
			if got != tt.want {
				t.Errorf("UnmarshalJSON() = %v, want %v", got, tt.want)
			}
			out, err := json.Marshal(&got)
			if err != nil {
				t.Errorf("MarshalJSON() error = %v", err)
				return
			}
			if string(out) != tt.input {
				t.Errorf("MarshalJSON() = %s, want %s", string(out), tt.input)
			}
		})
	}
}

func TestWrappingMode_JSON(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  WrappingMode
	}{
		{"clamp_to_edge", `33071`, WrapClampToEdge},
		{"mirrored_repeat", `33648`, WrapMirroredRepeat},
		{"repeat", `10497`, WrapRepeat},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got WrappingMode
			err := json.Unmarshal([]byte(tt.input), &got)
			if err != nil {
				t.Errorf("UnmarshalJSON() error = %v", err)
				return
			}
			if got != tt.want {
				t.Errorf("UnmarshalJSON() = %v, want %v", got, tt.want)
			}
			out, err := json.Marshal(&got)
			if err != nil {
				t.Errorf("MarshalJSON() error = %v", err)
				return
			}
			if string(out) != tt.input {
				t.Errorf("MarshalJSON() = %s, want %s", string(out), tt.input)
			}
		})
	}
}

func TestInterpolation_JSON(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  Interpolation
	}{
		{"linear", `"LINEAR"`, InterpolationLinear},
		{"step", `"STEP"`, InterpolationStep},
		{"cubic_spline", `"CUBICSPLINE"`, InterpolationCubicSpline},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got Interpolation
			err := json.Unmarshal([]byte(tt.input), &got)
			if err != nil {
				t.Errorf("UnmarshalJSON() error = %v", err)
				return
			}
			if got != tt.want {
				t.Errorf("UnmarshalJSON() = %v, want %v", got, tt.want)
			}
			out, err := json.Marshal(&got)
			if err != nil {
				t.Errorf("MarshalJSON() error = %v", err)
				return
			}
			if string(out) != tt.input {
				t.Errorf("MarshalJSON() = %s, want %s", string(out), tt.input)
			}
		})
	}
}

func TestTRSProperty_JSON(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  TRSProperty
	}{
		{"translation", `"translation"`, TRSTranslation},
		{"rotation", `"rotation"`, TRSRotation},
		{"scale", `"scale"`, TRSScale},
		{"weights", `"weights"`, TRSWeights},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got TRSProperty
			err := json.Unmarshal([]byte(tt.input), &got)
			if err != nil {
				t.Errorf("UnmarshalJSON() error = %v", err)
				return
			}
			if got != tt.want {
				t.Errorf("UnmarshalJSON() = %v, want %v", got, tt.want)
			}
			out, err := json.Marshal(&got)
			if err != nil {
				t.Errorf("MarshalJSON() error = %v", err)
				return
			}
			if string(out) != tt.input {
				t.Errorf("MarshalJSON() = %s, want %s", string(out), tt.input)
			}
		})
	}
}

func TestTarget_Constants(t *testing.T) {
	if TargetArrayBuffer != 34962 {
		t.Errorf("TargetArrayBuffer = %d, want 34962", TargetArrayBuffer)
	}
	if TargetElementArrayBuffer != 34963 {
		t.Errorf("TargetElementArrayBuffer = %d, want 34963", TargetElementArrayBuffer)
	}
	if TargetNone != 0 {
		t.Errorf("TargetNone = %d, want 0", TargetNone)
	}
}
