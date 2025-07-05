package bim4d

import (
	"encoding/json"
	"fmt"

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
	ID                 string                 `json:"id"`
	Name               string                 `json:"name"`
	Description        *string                `json:"description,omitempty"`
	Status             WorkStatus             `json:"status"`
	GenerateType       GenerateType           `json:"generateType"`
	WorkType           WorkType               `json:"workType"`
	StartTime          string                 `json:"startTime"`
	EndTime            string                 `json:"endTime"`
	ScheduleStart      *string                `json:"scheduleStart,omitempty"`
	ScheduleEnd        *string                `json:"scheduleEnd,omitempty"`
	StartValue         float32                `json:"startValue"`
	EndValue           float32                `json:"endValue"`
	ProgressType       ProgressType           `json:"progressType"`
	Total              float32                `json:"total"`
	Metadata           map[string]interface{} `json:"metadata,omitempty"`
	MetadataBufferView *uint32                `json:"metadataBufferView,omitempty"` // 指向二进制元数据
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

// DecodeBIM4dMetadata 处理二进制元数据解码
func DecodeBIM4dMetadata(doc *gltf.Document) error {
	for _, node := range doc.Nodes {
		ext, ok := node.Extensions[ExtensionName]
		if !ok {
			continue
		}

		env, ok := ext.(envelop)
		if !ok {
			return fmt.Errorf("invalid %s extension format", ExtensionName)
		}

		for _, work := range env.Works {
			if work.MetadataBufferView != nil {
				bvIndex := *work.MetadataBufferView
				if int(bvIndex) >= len(doc.BufferViews) {
					return fmt.Errorf("buffer view index out of range: %d", bvIndex)
				}

				bv := doc.BufferViews[bvIndex]
				if bv.Buffer >= uint32(len(doc.Buffers)) {
					return fmt.Errorf("buffer index out of range: %d", bv.Buffer)
				}

				buffer := doc.Buffers[bv.Buffer]
				data := buffer.Data[bv.ByteOffset : bv.ByteOffset+bv.ByteLength]

				var metadata map[string]interface{}
				if err := json.Unmarshal(data, &metadata); err != nil {
					return fmt.Errorf("failed to decode metadata: %w", err)
				}

				work.Metadata = metadata
				work.MetadataBufferView = nil
			}
		}
	}
	return nil
}

// EncodeBIM4dMetadata 处理二进制元数据编码
func EncodeBIM4dMetadata(doc *gltf.Document) error {
	for _, node := range doc.Nodes {
		ext, ok := node.Extensions[ExtensionName]
		if !ok {
			continue
		}

		env, ok := ext.(envelop)
		if !ok {
			return fmt.Errorf("invalid %s extension format", ExtensionName)
		}

		for _, work := range env.Works {
			if len(work.Metadata) > 0 {
				jsonData, err := json.Marshal(work.Metadata)
				if err != nil {
					return fmt.Errorf("failed to encode metadata: %w", err)
				}

				// 创建新缓冲区
				bufferIndex := uint32(len(doc.Buffers))
				doc.Buffers = append(doc.Buffers, &gltf.Buffer{
					ByteLength: uint32(len(jsonData)),
					Data:       jsonData,
				})

				// 创建缓冲视图
				bvIndex := uint32(len(doc.BufferViews))
				doc.BufferViews = append(doc.BufferViews, &gltf.BufferView{
					Buffer:     bufferIndex,
					ByteOffset: 0,
					ByteLength: uint32(len(jsonData)),
				})

				work.MetadataBufferView = &bvIndex
				work.Metadata = nil
			}
		}
	}
	return nil
}

// AddWorkItemToNode 添加工作项到节点
func AddWorkItemToNode(node *gltf.Node, work *WorkItem) {
	if node.Extensions == nil {
		node.Extensions = make(map[string]interface{})
	}

	ext, ok := node.Extensions[ExtensionName]
	if !ok {
		ext = envelop{
			Version: "1.0",
			Works:   []*WorkItem{work},
		}
		node.Extensions[ExtensionName] = ext
		return
	}

	env, ok := ext.(envelop)
	if !ok {
		return
	}

	env.Works = append(env.Works, work)
	node.Extensions[ExtensionName] = env
}

// GetWorkItemsFromNode 从节点获取工作项
func GetWorkItemsFromNode(node *gltf.Node) []*WorkItem {
	ext, ok := node.Extensions[ExtensionName]
	if !ok {
		return nil
	}

	env, ok := ext.(envelop)
	if !ok {
		return nil
	}

	return env.Works
}

// SetCurrentWorkItem 设置节点的当前工作项
func SetCurrentWorkItem(node *gltf.Node, workID string) {
	if node.Extensions == nil {
		node.Extensions = make(map[string]interface{})
	}

	ext, ok := node.Extensions[ExtensionName]
	if !ok {
		ext = envelop{
			Version:       "1.0",
			CurrentWorkID: &workID,
		}
		node.Extensions[ExtensionName] = ext
		return
	}

	env, ok := ext.(envelop)
	if !ok {
		return
	}

	env.CurrentWorkID = &workID
	node.Extensions[ExtensionName] = env
}

// GetCurrentWorkItem 获取节点的当前工作项
func GetCurrentWorkItem(node *gltf.Node) *WorkItem {
	ext, ok := node.Extensions[ExtensionName]
	if !ok {
		return nil
	}

	env, ok := ext.(envelop)
	if !ok {
		return nil
	}

	if env.CurrentWorkID == nil {
		return nil
	}

	for _, work := range env.Works {
		if work.ID == *env.CurrentWorkID {
			return work
		}
	}
	return nil
}

// CalculateWorkProgress 计算工作项进度
func CalculateWorkProgress(work *WorkItem) float32 {
	if work.ProgressType == Percentage {
		return work.StartValue
	}

	if work.Total == 0 {
		return 0
	}
	return (work.StartValue / work.Total) * 100
}

// ValidateWorkItem 验证工作项有效性
func ValidateWorkItem(work *WorkItem) error {
	if work.ID == "" {
		return fmt.Errorf("work item ID is required")
	}
	if work.Name == "" {
		return fmt.Errorf("work item name is required")
	}
	if work.StartTime == "" {
		return fmt.Errorf("start time is required")
	}
	if work.EndTime == "" {
		return fmt.Errorf("end time is required")
	}
	if work.StartValue < 0 {
		return fmt.Errorf("start value cannot be negative")
	}
	if work.EndValue < 0 {
		return fmt.Errorf("end value cannot be negative")
	}
	if work.Total <= 0 {
		return fmt.Errorf("total must be positive")
	}
	return nil
}
