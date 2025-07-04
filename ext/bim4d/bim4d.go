package bim4d

import (
	"encoding/json"

	"github.com/flywave/gltf"
)

const (
	// ExtensionName defines the EXT_bim4d_metadata unique key
	ExtensionName = "EXT_bim4d_metadata"
)

func init() {
	gltf.RegisterExtension(ExtensionName, Unmarshal)
}

type envelop struct {
	Works         []*WorkItem `json:"works,omitempty"`
	CurrentWorkID *string     `json:"currentWorkId,omitempty"`
	Version       string      `json:"version"`
}

// Unmarshal decodes the json data into the correct type
func Unmarshal(data []byte) (interface{}, error) {
	env := envelop{Version: "1.0"}
	if err := json.Unmarshal(data, &env); err != nil {
		return nil, err
	}
	return env, nil
}

// WorkStatus defines possible work statuses
type WorkStatus string

const (
	Pending    WorkStatus = "pending"
	InProgress WorkStatus = "in_progress"
	Completed  WorkStatus = "completed"
)

// GenerateType defines possible generation types
type GenerateType string

const (
	Auto   GenerateType = "auto"
	Manual GenerateType = "manual"
)

// ProgressType defines possible progress tracking types
type ProgressType string

const (
	Percentage ProgressType = "percentage"
	Absolute   ProgressType = "absolute"
)

// WorkType defines possible work types
type WorkType string

const (
	Schedule WorkType = "schedule"
	Plan     WorkType = "plan"
)

// WorkItem defines a BIM 4D work item
type WorkItem struct {
	ID            string                 `json:"id"`
	Name          string                 `json:"name"`
	Description   *string                `json:"description,omitempty"`
	Status        WorkStatus             `json:"status"`
	GenerateType  GenerateType           `json:"generateType"`
	WorkType      WorkType               `json:"workType"`
	StartTime     string                 `json:"startTime"`
	EndTime       string                 `json:"endTime"`
	ScheduleStart *string                `json:"scheduleStart,omitempty"`
	ScheduleEnd   *string                `json:"scheduleEnd,omitempty"`
	StartValue    float32                `json:"startValue"`
	EndValue      float32                `json:"endValue"`
	ProgressType  ProgressType           `json:"progressType"`
	Total         float32                `json:"total"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// CreateWorkItem creates a new WorkItem with default values
func CreateWorkItem(info map[string]interface{}) *WorkItem {
	item := &WorkItem{
		Status:       Pending,
		GenerateType: Auto,
		WorkType:     Schedule,
		ProgressType: Percentage,
		Total:        100,
		Metadata:     make(map[string]interface{}),
	}

	if id, ok := info["id"].(string); ok {
		item.ID = id
	}
	if name, ok := info["name"].(string); ok {
		item.Name = name
	}
	if desc, ok := info["description"].(string); ok {
		item.Description = &desc
	}
	if status, ok := info["status"].(WorkStatus); ok {
		item.Status = status
	}
	if genType, ok := info["generateType"].(GenerateType); ok {
		item.GenerateType = genType
	}
	if workType, ok := info["workType"].(WorkType); ok {
		item.WorkType = workType
	}
	if start, ok := info["startTime"].(string); ok {
		item.StartTime = start
	}
	if end, ok := info["endTime"].(string); ok {
		item.EndTime = end
	}
	if schedStart, ok := info["scheduleStart"].(string); ok {
		item.ScheduleStart = &schedStart
	}
	if schedEnd, ok := info["scheduleEnd"].(string); ok {
		item.ScheduleEnd = &schedEnd
	}
	if startVal, ok := info["startValue"].(float32); ok {
		item.StartValue = startVal
	}
	if endVal, ok := info["endValue"].(float32); ok {
		item.EndValue = endVal
	}
	if progType, ok := info["progressType"].(ProgressType); ok {
		item.ProgressType = progType
	}
	if total, ok := info["total"].(float32); ok {
		item.Total = total
	}
	if metadata, ok := info["metadata"].(map[string]interface{}); ok {
		item.Metadata = metadata
	}

	return item
}

// DescriptionOrDefault returns the description if set, or an empty string
func (w *WorkItem) DescriptionOrDefault() string {
	if w.Description == nil {
		return ""
	}
	return *w.Description
}

// ScheduleStartOrDefault returns schedule start if set, or startTime
func (w *WorkItem) ScheduleStartOrDefault() string {
	if w.ScheduleStart == nil {
		return w.StartTime
	}
	return *w.ScheduleStart
}

// ScheduleEndOrDefault returns schedule end if set, or endTime
func (w *WorkItem) ScheduleEndOrDefault() string {
	if w.ScheduleEnd == nil {
		return w.EndTime
	}
	return *w.ScheduleEnd
}

// UnmarshalJSON unmarshals the WorkItem with proper defaults
func (w *WorkItem) UnmarshalJSON(data []byte) error {
	type alias WorkItem
	tmp := alias{
		Status:       Pending,
		GenerateType: Auto,
		WorkType:     Schedule,
		ProgressType: Percentage,
		Total:        100,
	}

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*w = WorkItem(tmp)
	return nil
}
