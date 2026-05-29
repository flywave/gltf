package joint

import (
	"encoding/json"

	"github.com/flywave/gltf"
)

const (
	ExtensionName = "FLYWAVE_joint_metadata"
)

func init() {
	gltf.RegisterExtension(ExtensionName, Unmarshal)
}

type JointType string

const (
	JointFixed       JointType = "FIXED"
	JointRevolute    JointType = "REVOLUTE"
	JointPrismatic   JointType = "PRISMATIC"
	JointCylindrical JointType = "CYLINDRICAL"
	JointPlanar      JointType = "PLANAR"
	JointSpherical   JointType = "SPHERICAL"
	JointUniversal   JointType = "UNIVERSAL"
	JointCurve       JointType = "CURVE"
)

type OutputMode string

const (
	OutputStatic  OutputMode = "STATIC"
	OutputDynamic OutputMode = "DYNAMIC"
)

type PathType string

const (
	PathLine   PathType = "LINE"
	PathArc    PathType = "ARC"
	PathBezier PathType = "BEZIER"
)

type LimitAxis struct {
	Min *float64 `json:"min,omitempty"`
	Max *float64 `json:"max,omitempty"`
}

type Limits struct {
	TranslateX *LimitAxis `json:"translateX,omitempty"`
	TranslateY *LimitAxis `json:"translateY,omitempty"`
	TranslateZ *LimitAxis `json:"translateZ,omitempty"`
	RotateX    *LimitAxis `json:"rotateX,omitempty"`
	RotateY    *LimitAxis `json:"rotateY,omitempty"`
	RotateZ    *LimitAxis `json:"rotateZ,omitempty"`
}

type JointPath struct {
	Type   PathType       `json:"type"`
	Points [][3]float64   `json:"points"`
}

type Joint struct {
	Id            string      `json:"id"`
	Name          string      `json:"name,omitempty"`
	Type          JointType   `json:"type"`
	Output        OutputMode  `json:"output"`
	ParentNode    uint32      `json:"parentNode"`
	ChildNode     uint32      `json:"childNode"`
	Origin        [3]float64  `json:"origin"`
	Axis          [3]float64  `json:"axis"`
	SecondaryAxis *[3]float64 `json:"secondaryAxis,omitempty"`
	Limits        *Limits     `json:"limits,omitempty"`
	Path          *JointPath  `json:"path,omitempty"`
	Value         *float64    `json:"value,omitempty"`
	Values        []float64   `json:"values,omitempty"`
}

type JointDocument struct {
	Version string `json:"version"`
	Joints  []*Joint `json:"joints,omitempty"`
}

func Unmarshal(data []byte) (interface{}, error) {
	env := JointDocument{Version: "1.0"}
	if err := json.Unmarshal(data, &env); err != nil {
		return nil, err
	}
	return env, nil
}

func GetJoints(doc *gltf.Document) []*Joint {
	ext, ok := doc.Extensions[ExtensionName]
	if !ok {
		return nil
	}
	env, ok := ext.(JointDocument)
	if !ok {
		return nil
	}
	return env.Joints
}

func AddJoint(doc *gltf.Document, joint *Joint) {
	ext, ok := doc.Extensions[ExtensionName]
	if !ok {
		if doc.Extensions == nil {
			doc.Extensions = make(gltf.Extensions)
		}
		doc.Extensions[ExtensionName] = JointDocument{
			Version: "1.0",
			Joints:  []*Joint{joint},
		}
		doc.AddExtensionUsed(ExtensionName)
		return
	}
	env, ok := ext.(JointDocument)
	if !ok {
		return
	}
	env.Joints = append(env.Joints, joint)
	doc.Extensions[ExtensionName] = env
}
