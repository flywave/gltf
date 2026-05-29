package joint

import (
	"encoding/json"
	"testing"

	"github.com/flywave/gltf"
)

func TestJointDocumentMarshalUnmarshal(t *testing.T) {
	doc := JointDocument{
		Version: "1.0",
		Joints: []*Joint{
			{
				Id:         "j1",
				Name:       "test-joint",
				Type:       JointRevolute,
				Output:     OutputDynamic,
				ParentNode: 0,
				ChildNode:  1,
				Origin:     [3]float64{1, 0, 0},
				Axis:       [3]float64{0, 0, 1},
				Value:      floatPtr(0.5),
			},
		},
	}

	data, err := json.Marshal(doc)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var got JointDocument
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if got.Version != "1.0" {
		t.Errorf("Version: got %q, want %q", got.Version, "1.0")
	}
	if len(got.Joints) != 1 {
		t.Fatalf("Joints length: got %d, want 1", len(got.Joints))
	}
	if got.Joints[0].Id != "j1" {
		t.Errorf("Joint.Id: got %q, want %q", got.Joints[0].Id, "j1")
	}
	if got.Joints[0].Type != JointRevolute {
		t.Errorf("Joint.Type: got %q, want %q", got.Joints[0].Type, JointRevolute)
	}
	if got.Joints[0].Output != OutputDynamic {
		t.Errorf("Joint.Output: got %q, want %q", got.Joints[0].Output, OutputDynamic)
	}
	if got.Joints[0].ParentNode != 0 {
		t.Errorf("ParentNode: got %d, want 0", got.Joints[0].ParentNode)
	}
	if got.Joints[0].ChildNode != 1 {
		t.Errorf("ChildNode: got %d, want 1", got.Joints[0].ChildNode)
	}
	if got.Joints[0].Origin != [3]float64{1, 0, 0} {
		t.Errorf("Origin: got %v, want [1 0 0]", got.Joints[0].Origin)
	}
	if got.Joints[0].Axis != [3]float64{0, 0, 1} {
		t.Errorf("Axis: got %v, want [0 0 1]", got.Joints[0].Axis)
	}
	if got.Joints[0].Value == nil || *got.Joints[0].Value != 0.5 {
		t.Errorf("Value: got %v, want 0.5", got.Joints[0].Value)
	}
}

func TestJointTypeValues(t *testing.T) {
	tests := []struct {
		typ  JointType
		want string
	}{
		{JointFixed, "FIXED"},
		{JointRevolute, "REVOLUTE"},
		{JointPrismatic, "PRISMATIC"},
		{JointCylindrical, "CYLINDRICAL"},
		{JointPlanar, "PLANAR"},
		{JointSpherical, "SPHERICAL"},
		{JointUniversal, "UNIVERSAL"},
		{JointCurve, "CURVE"},
	}
	for _, tt := range tests {
		if string(tt.typ) != tt.want {
			t.Errorf("JointType %s: got %q, want %q", tt.typ, tt.typ, tt.want)
		}
	}
}

func TestOutputModeValues(t *testing.T) {
	tests := []struct {
		mode OutputMode
		want string
	}{
		{OutputStatic, "STATIC"},
		{OutputDynamic, "DYNAMIC"},
	}
	for _, tt := range tests {
		if string(tt.mode) != tt.want {
			t.Errorf("OutputMode %s: got %q, want %q", tt.mode, tt.mode, tt.want)
		}
	}
}

func TestPathTypeValues(t *testing.T) {
	tests := []struct {
		pt   PathType
		want string
	}{
		{PathLine, "LINE"},
		{PathArc, "ARC"},
		{PathBezier, "BEZIER"},
	}
	for _, tt := range tests {
		if string(tt.pt) != tt.want {
			t.Errorf("PathType %s: got %q, want %q", tt.pt, tt.pt, tt.want)
		}
	}
}

func TestJointWithAllFields(t *testing.T) {
	secAxis := [3]float64{0, 1, 0}
	joint := &Joint{
		Id:            "full",
		Name:          "full-joint",
		Type:          JointCylindrical,
		Output:        OutputDynamic,
		ParentNode:    2,
		ChildNode:     5,
		Origin:        [3]float64{0, 0, 0},
		Axis:          [3]float64{0, 1, 0},
		SecondaryAxis: &secAxis,
		Limits: &Limits{
			RotateY:    &LimitAxis{Min: floatPtr(0), Max: floatPtr(6.28)},
			TranslateY: &LimitAxis{Min: floatPtr(-10), Max: floatPtr(10)},
		},
		Value: floatPtr(1.57),
	}

	data, err := json.Marshal(joint)
	if err != nil {
		t.Fatalf("Marshal joint failed: %v", err)
	}

	var got Joint
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal joint failed: %v", err)
	}

	if got.Id != "full" {
		t.Errorf("Id: got %q", got.Id)
	}
	if got.Type != JointCylindrical {
		t.Errorf("Type: got %q", got.Type)
	}
	if got.SecondaryAxis == nil || *got.SecondaryAxis != secAxis {
		t.Errorf("SecondaryAxis: got %v", got.SecondaryAxis)
	}
	if got.Limits == nil {
		t.Fatal("Limits is nil")
	}
	if got.Limits.RotateY == nil || *got.Limits.RotateY.Min != 0 || *got.Limits.RotateY.Max != 6.28 {
		t.Errorf("RotateY limits: got %+v", got.Limits.RotateY)
	}
	if got.Limits.TranslateY == nil || *got.Limits.TranslateY.Min != -10 || *got.Limits.TranslateY.Max != 10 {
		t.Errorf("TranslateY limits: got %+v", got.Limits.TranslateY)
	}
	if got.Value == nil || *got.Value != 1.57 {
		t.Errorf("Value: got %v", got.Value)
	}
}

func TestJointCurvePath(t *testing.T) {
	joint := &Joint{
		Id:         "curve",
		Type:       JointCurve,
		Output:     OutputStatic,
		ParentNode: 0,
		ChildNode:  1,
		Path: &JointPath{
			Type:   PathBezier,
			Points: [][3]float64{{0, 0, 0}, {0, 10, 0}, {10, 10, 0}, {10, 0, 0}},
		},
	}

	data, err := json.Marshal(joint)
	if err != nil {
		t.Fatalf("Marshal curve joint failed: %v", err)
	}

	var got Joint
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal curve joint failed: %v", err)
	}

	if got.Path == nil {
		t.Fatal("Path is nil")
	}
	if got.Path.Type != PathBezier {
		t.Errorf("Path.Type: got %q, want %q", got.Path.Type, PathBezier)
	}
	if len(got.Path.Points) != 4 {
		t.Fatalf("Path.Points length: got %d, want 4", len(got.Path.Points))
	}
	if got.Path.Points[0] != [3]float64{0, 0, 0} {
		t.Errorf("Points[0]: got %v", got.Path.Points[0])
	}
	if got.Path.Points[3] != [3]float64{10, 0, 0} {
		t.Errorf("Points[3]: got %v", got.Path.Points[3])
	}
}

func TestJointPathTypes(t *testing.T) {
	points := [][3]float64{{0, 0, 0}, {10, 0, 0}}
	for _, pt := range []PathType{PathLine, PathArc, PathBezier} {
		joint := &Joint{
			Id:   "path-test",
			Type: JointCurve,
			Path: &JointPath{Type: pt, Points: points},
		}
		data, _ := json.Marshal(joint)
		var got Joint
		if err := json.Unmarshal(data, &got); err != nil {
			t.Errorf("Unmarshal PathType %s failed: %v", pt, err)
		}
		if got.Path.Type != pt {
			t.Errorf("Path.Type: got %q, want %q", got.Path.Type, pt)
		}
	}
}

func TestLimitsPartial(t *testing.T) {
	joint := &Joint{
		Id:   "partial-limits",
		Type: JointRevolute,
		Limits: &Limits{
			RotateX: &LimitAxis{Min: floatPtr(-3.14), Max: floatPtr(3.14)},
		},
	}

	data, err := json.Marshal(joint)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var got Joint
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if got.Limits == nil {
		t.Fatal("Limits is nil")
	}
	if got.Limits.RotateX == nil {
		t.Fatal("RotateX limit is nil")
	}
	if *got.Limits.RotateX.Min != -3.14 {
		t.Errorf("RotateX.Min: got %f", *got.Limits.RotateX.Min)
	}
	if got.Limits.RotateY != nil {
		t.Error("RotateY should be nil")
	}
}

func TestLimitsNil(t *testing.T) {
	joint := &Joint{
		Id:   "no-limits",
		Type: JointFixed,
	}

	data, err := json.Marshal(joint)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var got Joint
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if got.Limits != nil {
		t.Error("Limits should be nil")
	}
}

func TestJointSecondaryAxisOptional(t *testing.T) {
	joint := &Joint{
		Id:   "no-secondary",
		Type: JointRevolute,
		Axis: [3]float64{0, 0, 1},
	}

	data, err := json.Marshal(joint)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var got Joint
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if got.SecondaryAxis != nil {
		t.Error("SecondaryAxis should be nil when omitted")
	}
}

func TestUnmarshal(t *testing.T) {
	jsonData := `{
		"version": "1.0",
		"joints": [{
			"id": "u1",
			"type": "UNIVERSAL",
			"output": "STATIC",
			"parentNode": 0,
			"childNode": 2,
			"origin": [0,0,0],
			"axis": [1,0,0],
			"secondaryAxis": [0,1,0],
			"limits": {
				"rotateX": {"min": -1.57, "max": 1.57},
				"rotateY": {"min": -1.57, "max": 1.57}
			}
		}]
	}`

	got, err := Unmarshal([]byte(jsonData))
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	env, ok := got.(JointDocument)
	if !ok {
		t.Fatalf("Unmarshal returned %T, want JointDocument", got)
	}

	if len(env.Joints) != 1 {
		t.Fatalf("Joints length: got %d, want 1", len(env.Joints))
	}
	j := env.Joints[0]
	if j.Id != "u1" || j.Type != JointUniversal || j.Output != OutputStatic {
		t.Errorf("Joint fields mismatch: %+v", j)
	}
	if j.SecondaryAxis == nil || *j.SecondaryAxis != [3]float64{0, 1, 0} {
		t.Errorf("SecondaryAxis: got %v", j.SecondaryAxis)
	}
	if j.Limits == nil || j.Limits.RotateX == nil || *j.Limits.RotateX.Min != -1.57 {
		t.Errorf("Limits.RotateX: got %+v", j.Limits.RotateX)
	}
}

func TestUnmarshalEmptyJoints(t *testing.T) {
	jsonData := `{"version": "1.0"}`

	got, err := Unmarshal([]byte(jsonData))
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	env, ok := got.(JointDocument)
	if !ok {
		t.Fatalf("Unmarshal returned %T", env)
	}

	if env.Version != "1.0" {
		t.Errorf("Version: %q", env.Version)
	}
	if len(env.Joints) != 0 {
		t.Errorf("Expected 0 joints, got %d", len(env.Joints))
	}
}

func TestUnmarshalInvalidJSON(t *testing.T) {
	_, err := Unmarshal([]byte("{invalid"))
	if err == nil {
		t.Error("Expected error for invalid JSON")
	}
}

func TestGetJointsNil(t *testing.T) {
	doc := gltf.NewDocument()
	joints := GetJoints(doc)
	if joints != nil {
		t.Errorf("Expected nil, got %v", joints)
	}
}

func TestGetJointsEmpty(t *testing.T) {
	doc := gltf.NewDocument()
	doc.Extensions = make(gltf.Extensions)
	doc.Extensions[ExtensionName] = JointDocument{Version: "1.0"}

	joints := GetJoints(doc)
	if joints != nil {
		t.Errorf("Expected nil, got %d joints", len(joints))
	}
}

func TestAddJointToNewDoc(t *testing.T) {
	doc := gltf.NewDocument()
	j := &Joint{
		Id:         "new-joint",
		Type:       JointFixed,
		Output:     OutputStatic,
		ParentNode: 0,
		ChildNode:  1,
	}

	AddJoint(doc, j)

	if !doc.HasExtensionUsed(ExtensionName) {
		t.Errorf("Extension %s not registered as used", ExtensionName)
	}

	joints := GetJoints(doc)
	if joints == nil {
		t.Fatal("GetJoints returned nil")
	}
	if len(joints) != 1 {
		t.Fatalf("Expected 1 joint, got %d", len(joints))
	}
	if joints[0].Id != "new-joint" {
		t.Errorf("Joint.Id: got %q", joints[0].Id)
	}
}

func TestAddMultipleJoints(t *testing.T) {
	doc := gltf.NewDocument()

	AddJoint(doc, &Joint{Id: "a", Type: JointRevolute, ParentNode: 0, ChildNode: 1})
	AddJoint(doc, &Joint{Id: "b", Type: JointPrismatic, ParentNode: 1, ChildNode: 2})

	joints := GetJoints(doc)
	if len(joints) != 2 {
		t.Fatalf("Expected 2 joints, got %d", len(joints))
	}
	if joints[0].Id != "a" || joints[1].Id != "b" {
		t.Errorf("Joint order wrong: %v", joints)
	}
}

func TestAddJointToDocWithExistingExtension(t *testing.T) {
	doc := gltf.NewDocument()
	doc.Extensions = make(gltf.Extensions)
	doc.Extensions[ExtensionName] = JointDocument{
		Version: "1.0",
		Joints: []*Joint{
			{Id: "existing", Type: JointFixed, ParentNode: 0, ChildNode: 1},
		},
	}

	AddJoint(doc, &Joint{Id: "added", Type: JointRevolute, ParentNode: 1, ChildNode: 2})

	joints := GetJoints(doc)
	if len(joints) != 2 {
		t.Fatalf("Expected 2 joints, got %d", len(joints))
	}
	if joints[0].Id != "existing" || joints[1].Id != "added" {
		t.Errorf("Expected existing then added, got: %q, %q", joints[0].Id, joints[1].Id)
	}
}

func TestRoundTripThroughGLTFDoc(t *testing.T) {
	doc := gltf.NewDocument()
	doc.Nodes = []*gltf.Node{
		{Name: "parent", Children: []uint32{1}},
		{Name: "child"},
	}

	AddJoint(doc, &Joint{
		Id:         "rj1",
		Name:       "elbow",
		Type:       JointRevolute,
		Output:     OutputDynamic,
		ParentNode: 0,
		ChildNode:  1,
		Origin:     [3]float64{0.5, 0, 0},
		Axis:       [3]float64{0, 0, 1},
		Limits: &Limits{
			RotateZ: &LimitAxis{Min: floatPtr(-3.14), Max: floatPtr(3.14)},
		},
		Value: floatPtr(0),
	})

	data, err := json.Marshal(doc)
	if err != nil {
		t.Fatalf("Marshal document failed: %v", err)
	}

	var doc2 gltf.Document
	if err := json.Unmarshal(data, &doc2); err != nil {
		t.Fatalf("Unmarshal document failed: %v", err)
	}

	if !doc2.HasExtensionUsed(ExtensionName) {
		t.Error("Extension name not in ExtensionsUsed after round-trip")
	}

	joints := GetJoints(&doc2)
	if joints == nil || len(joints) != 1 {
		t.Fatalf("Expected 1 joint after round-trip, got %v", lenPtr(joints))
	}

	j := joints[0]
	if j.Id != "rj1" || j.Type != JointRevolute || j.Output != OutputDynamic {
		t.Errorf("Joint fields after round-trip: %+v", j)
	}
	if j.Origin != [3]float64{0.5, 0, 0} {
		t.Errorf("Origin after round-trip: %v", j.Origin)
	}
	if j.Value == nil || *j.Value != 0 {
		t.Errorf("Value after round-trip: %v", j.Value)
	}
}

func TestJointPathRoundTrip(t *testing.T) {
	doc := gltf.NewDocument()
	AddJoint(doc, &Joint{
		Id:         "curve1",
		Type:       JointCurve,
		Output:     OutputStatic,
		ParentNode: 0,
		ChildNode:  1,
		Origin:     [3]float64{0, 0, 0},
		Path: &JointPath{
			Type:   PathArc,
			Points: [][3]float64{{1, 0, 0}, {0, 1, 0}, {-1, 0, 0}},
		},
	})

	data, _ := json.Marshal(doc)
	var doc2 gltf.Document
	json.Unmarshal(data, &doc2)

	joints := GetJoints(&doc2)
	if len(joints) != 1 || joints[0].Path == nil {
		t.Fatal("Path lost during round-trip")
	}
	if joints[0].Path.Type != PathArc {
		t.Errorf("Path.Type: got %q", joints[0].Path.Type)
	}
	if len(joints[0].Path.Points) != 3 {
		t.Errorf("Path.Points: got %d points", len(joints[0].Path.Points))
	}
}

func TestAllJointTypesSerialization(t *testing.T) {
	types := []JointType{
		JointFixed, JointRevolute, JointPrismatic, JointCylindrical,
		JointPlanar, JointSpherical, JointUniversal, JointCurve,
	}
	for _, typ := range types {
		j := &Joint{
			Id:         string(typ),
			Type:       typ,
			Output:     OutputStatic,
			ParentNode: 0,
			ChildNode:  1,
			Origin:     [3]float64{0, 0, 0},
			Axis:       [3]float64{1, 0, 0},
		}
		data, err := json.Marshal(j)
		if err != nil {
			t.Errorf("Marshal JointType %s failed: %v", typ, err)
			continue
		}
		var got Joint
		if err := json.Unmarshal(data, &got); err != nil {
			t.Errorf("Unmarshal JointType %s failed: %v", typ, err)
			continue
		}
		if got.Type != typ {
			t.Errorf("JointType %s: got %q after round-trip", typ, got.Type)
		}
	}
}

func TestLimitAxisNilMinMax(t *testing.T) {
	axis := &LimitAxis{}
	data, err := json.Marshal(axis)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}
	if string(data) != "{}" {
		t.Errorf("Expected empty object, got: %s", data)
	}
}

func TestLimitAxisOnlyMin(t *testing.T) {
	axis := &LimitAxis{Min: floatPtr(-5)}
	data, err := json.Marshal(axis)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}
	var got LimitAxis
	json.Unmarshal(data, &got)
	if got.Min == nil || *got.Min != -5 {
		t.Errorf("Min: got %v", got.Min)
	}
	if got.Max != nil {
		t.Error("Max should be nil")
	}
}

func TestLimitAxisOnlyMax(t *testing.T) {
	axis := &LimitAxis{Max: floatPtr(100)}
	data, err := json.Marshal(axis)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}
	var got LimitAxis
	json.Unmarshal(data, &got)
	if got.Min != nil {
		t.Error("Min should be nil")
	}
	if got.Max == nil || *got.Max != 100 {
		t.Errorf("Max: got %v", got.Max)
	}
}

func TestJointPathNil(t *testing.T) {
	joint := &Joint{
		Id:   "no-path",
		Type: JointCurve,
	}
	data, _ := json.Marshal(joint)
	var got Joint
	json.Unmarshal(data, &got)
	if got.Path != nil {
		t.Error("Path should be nil for curve joint without path")
	}
}

func TestJointValueNil(t *testing.T) {
	joint := &Joint{
		Id:   "no-value",
		Type: JointFixed,
	}
	data, _ := json.Marshal(joint)
	var got Joint
	json.Unmarshal(data, &got)
	if got.Value != nil {
		t.Error("Value should be nil")
	}
}

func TestJointNameOmitEmpty(t *testing.T) {
	joint := &Joint{
		Id:   "no-name",
		Type: JointFixed,
	}
	data, _ := json.Marshal(joint)
	if containsKey(data, "name") {
		t.Error("Name should be omitted when empty")
	}
}

func TestJointNodeIndexes(t *testing.T) {
	joint := &Joint{
		Id:         "index-test",
		Type:       JointRevolute,
		ParentNode: 10,
		ChildNode:  20,
	}
	data, _ := json.Marshal(joint)
	var got Joint
	json.Unmarshal(data, &got)
	if got.ParentNode != 10 {
		t.Errorf("ParentNode: got %d, want 10", got.ParentNode)
	}
	if got.ChildNode != 20 {
		t.Errorf("ChildNode: got %d, want 20", got.ChildNode)
	}
}

func TestJointSecondaryAxisRoundTrip(t *testing.T) {
	sec := [3]float64{0.707, 0.707, 0}
	joint := &Joint{
		Id:            "sec-axis",
		Type:          JointUniversal,
		SecondaryAxis: &sec,
	}
	data, _ := json.Marshal(joint)
	var got Joint
	json.Unmarshal(data, &got)
	if got.SecondaryAxis == nil {
		t.Fatal("SecondaryAxis is nil")
	}
	if *got.SecondaryAxis != sec {
		t.Errorf("SecondaryAxis: got %v, want %v", *got.SecondaryAxis, sec)
	}
}

func floatPtr(v float64) *float64 {
	return &v
}

func containsKey(data []byte, key string) bool {
	var raw map[string]interface{}
	json.Unmarshal(data, &raw)
	_, ok := raw[key]
	return ok
}

func lenPtr(joints []*Joint) int {
	if joints == nil {
		return -1
	}
	return len(joints)
}
