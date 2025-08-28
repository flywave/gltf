package bim4d

import (
	"encoding/json"
	"testing"

	"github.com/flywave/gltf"
)

func TestUnmarshal(t *testing.T) {
	// Test unmarshaling with works
	data := []byte(`{
		"works": [
			{
				"id": "work1",
				"name": "Foundation Work",
				"description": "Laying the foundation",
				"status": "pending",
				"generateType": "auto",
				"workType": "schedule",
				"startTime": "2023-01-01T00:00:00Z",
				"endTime": "2023-01-10T00:00:00Z",
		"startValue": 0,
		"endValue": 100,
		"progressType": "percentage",
		"total": 100
			}
		],
		"version": "1.0"
	}`)

	ext, err := Unmarshal(data)
	if err != nil {
		t.Errorf("Unmarshal failed: %v", err)
	}

	env, ok := ext.(envelop)
	if !ok {
		t.Error("Unmarshaled object is not of type envelop")
	}

	if len(env.Works) != 1 {
		t.Errorf("Expected 1 work item, got %d", len(env.Works))
	}

	if env.Version != "1.0" {
		t.Errorf("Expected version 1.0, got %s", env.Version)
	}

	// Test unmarshaling with currentWorkId
	data = []byte(`{
		"currentWorkId": "work1",
		"version": "1.0"
	}`)

	ext, err = Unmarshal(data)
	if err != nil {
		t.Errorf("Unmarshal failed: %v", err)
	}

	env, ok = ext.(envelop)
	if !ok {
		t.Error("Unmarshaled object is not of type envelop")
	}

	if env.CurrentWorkID == nil {
		t.Error("Expected CurrentWorkID to be set")
	} else if *env.CurrentWorkID != "work1" {
		t.Errorf("Expected CurrentWorkID 'work1', got '%s'", *env.CurrentWorkID)
	}
}

func TestCreateWorkItem(t *testing.T) {
	// Test creating work item with all fields
	info := map[string]interface{}{
		"id":           "work1",
		"name":         "Foundation Work",
		"description":  "Laying the foundation",
		"status":       Pending,
		"generateType": Auto,
		"workType":     Schedule,
		"startTime":    "2023-01-01T00:00:00Z",
		"endTime":      "2023-01-10T00:00:00Z",
		"startValue":   float32(0),
		"endValue":     float32(100),
		"progressType": Percentage,
		"total":        float32(100),
		"metadata": map[string]interface{}{
			"priority": "high",
		},
	}

	work := CreateWorkItem(info)

	if work.ID != "work1" {
		t.Errorf("Expected ID 'work1', got '%s'", work.ID)
	}

	if work.Name != "Foundation Work" {
		t.Errorf("Expected Name 'Foundation Work', got '%s'", work.Name)
	}

	if work.Description == nil {
		t.Error("Expected Description to be set")
	} else if *work.Description != "Laying the foundation" {
		t.Errorf("Expected Description 'Laying the foundation', got '%s'", *work.Description)
	}

	if work.Status != Pending {
		t.Errorf("Expected Status Pending, got '%s'", work.Status)
	}

	if work.GenerateType != Auto {
		t.Errorf("Expected GenerateType Auto, got '%s'", work.GenerateType)
	}

	if work.WorkType != Schedule {
		t.Errorf("Expected WorkType Schedule, got '%s'", work.WorkType)
	}

	if work.StartTime != "2023-01-01T00:00:00Z" {
		t.Errorf("Expected StartTime '2023-01-01T00:00:00Z', got '%s'", work.StartTime)
	}

	if work.EndTime != "2023-01-10T00:00:00Z" {
		t.Errorf("Expected EndTime '2023-01-10T00:00:00Z', got '%s'", work.EndTime)
	}

	if work.StartValue != 0 {
		t.Errorf("Expected StartValue 0, got %f", work.StartValue)
	}

	if work.EndValue != 100 {
		t.Errorf("Expected EndValue 100, got %f", work.EndValue)
	}

	if work.ProgressType != Percentage {
		t.Errorf("Expected ProgressType Percentage, got '%s'", work.ProgressType)
	}

	if work.Total != 100 {
		t.Errorf("Expected Total 100, got %f", work.Total)
	}

	if len(work.Metadata) != 1 {
		t.Errorf("Expected 1 metadata item, got %d", len(work.Metadata))
	}

	// Test creating work item with minimal fields
	info = map[string]interface{}{
		"id":        "work2",
		"name":      "Wall Construction",
		"startTime": "2023-01-11T00:00:00Z",
		"endTime":   "2023-01-20T00:00:00Z",
	}

	work = CreateWorkItem(info)

	if work.ID != "work2" {
		t.Errorf("Expected ID 'work2', got '%s'", work.ID)
	}

	if work.Name != "Wall Construction" {
		t.Errorf("Expected Name 'Wall Construction', got '%s'", work.Name)
	}

	// Check default values
	if work.Status != Pending {
		t.Errorf("Expected default Status Pending, got '%s'", work.Status)
	}

	if work.GenerateType != Auto {
		t.Errorf("Expected default GenerateType Auto, got '%s'", work.GenerateType)
	}

	if work.WorkType != Schedule {
		t.Errorf("Expected default WorkType Schedule, got '%s'", work.WorkType)
	}

	if work.ProgressType != Percentage {
		t.Errorf("Expected default ProgressType Percentage, got '%s'", work.ProgressType)
	}

	if work.Total != 100 {
		t.Errorf("Expected default Total 100, got %f", work.Total)
	}

	// Test with wrong data types (should be ignored)
	info = map[string]interface{}{
		"id":         "work3",
		"name":       "Roof Installation",
		"startTime":  "2023-01-21T00:00:00Z",
		"endTime":    "2023-01-30T00:00:00Z",
		"status":     "invalid_status", // Wrong type, should be ignored
		"startValue": "invalid_value",  // Wrong type, should be ignored
	}

	work = CreateWorkItem(info)

	if work.ID != "work3" {
		t.Errorf("Expected ID 'work3', got '%s'", work.ID)
	}

	if work.Name != "Roof Installation" {
		t.Errorf("Expected Name 'Roof Installation', got '%s'", work.Name)
	}

	// Check default values (should not be affected by wrong types)
	if work.Status != Pending {
		t.Errorf("Expected default Status Pending, got '%s'", work.Status)
	}

	if work.StartValue != 0 {
		t.Errorf("Expected default StartValue 0, got %f", work.StartValue)
	}
}

func TestWorkItem_DescriptionOrDefault(t *testing.T) {
	// Test with description set
	desc := "Laying the foundation"
	work := &WorkItem{
		Description: &desc,
	}

	result := work.DescriptionOrDefault()
	if result != desc {
		t.Errorf("Expected description '%s', got '%s'", desc, result)
	}

	// Test with nil description
	work = &WorkItem{
		Description: nil,
	}

	result = work.DescriptionOrDefault()
	if result != "" {
		t.Errorf("Expected empty string, got '%s'", result)
	}
}

func TestWorkItem_ScheduleStartOrDefault(t *testing.T) {
	startTime := "2023-01-01T00:00:00Z"

	// Test with schedule start set
	schedStart := "2023-01-02T00:00:00Z"
	work := &WorkItem{
		StartTime:     startTime,
		ScheduleStart: &schedStart,
	}

	result := work.ScheduleStartOrDefault()
	if result != schedStart {
		t.Errorf("Expected schedule start '%s', got '%s'", schedStart, result)
	}

	// Test with nil schedule start
	work = &WorkItem{
		StartTime:     startTime,
		ScheduleStart: nil,
	}

	result = work.ScheduleStartOrDefault()
	if result != startTime {
		t.Errorf("Expected start time '%s', got '%s'", startTime, result)
	}
}

func TestWorkItem_ScheduleEndOrDefault(t *testing.T) {
	endTime := "2023-01-10T00:00:00Z"

	// Test with schedule end set
	schedEnd := "2023-01-11T00:00:00Z"
	work := &WorkItem{
		EndTime:     endTime,
		ScheduleEnd: &schedEnd,
	}

	result := work.ScheduleEndOrDefault()
	if result != schedEnd {
		t.Errorf("Expected schedule end '%s', got '%s'", schedEnd, result)
	}

	// Test with nil schedule end
	work = &WorkItem{
		EndTime:     endTime,
		ScheduleEnd: nil,
	}

	result = work.ScheduleEndOrDefault()
	if result != endTime {
		t.Errorf("Expected end time '%s', got '%s'", endTime, result)
	}
}

func TestWorkItem_UnmarshalJSON(t *testing.T) {
	// Test unmarshaling with all fields
	data := []byte(`{
		"id": "work1",
		"name": "Foundation Work",
		"description": "Laying the foundation",
		"status": "pending",
		"generateType": "auto",
		"workType": "schedule",
		"startTime": "2023-01-01T00:00:00Z",
		"endTime": "2023-01-10T00:00:00Z",
		"startValue": 0,
		"endValue": 100,
		"progressType": "percentage",
		"total": 100
	}`)

	var work WorkItem
	err := json.Unmarshal(data, &work)
	if err != nil {
		t.Errorf("UnmarshalJSON failed: %v", err)
	}

	if work.ID != "work1" {
		t.Errorf("Expected ID 'work1', got '%s'", work.ID)
	}

	if work.Name != "Foundation Work" {
		t.Errorf("Expected Name 'Foundation Work', got '%s'", work.Name)
	}

	if work.Description == nil {
		t.Error("Expected Description to be set")
	} else if *work.Description != "Laying the foundation" {
		t.Errorf("Expected Description 'Laying the foundation', got '%s'", *work.Description)
	}

	if work.Status != Pending {
		t.Errorf("Expected Status Pending, got '%s'", work.Status)
	}

	if work.GenerateType != Auto {
		t.Errorf("Expected GenerateType Auto, got '%s'", work.GenerateType)
	}

	if work.WorkType != Schedule {
		t.Errorf("Expected WorkType Schedule, got '%s'", work.WorkType)
	}

	if work.StartTime != "2023-01-01T00:00:00Z" {
		t.Errorf("Expected StartTime '2023-01-01T00:00:00Z', got '%s'", work.StartTime)
	}

	if work.EndTime != "2023-01-10T00:00:00Z" {
		t.Errorf("Expected EndTime '2023-01-10T00:00:00Z', got '%s'", work.EndTime)
	}

	if work.StartValue != 0 {
		t.Errorf("Expected StartValue 0, got %f", work.StartValue)
	}

	if work.EndValue != 100 {
		t.Errorf("Expected EndValue 100, got %f", work.EndValue)
	}

	if work.ProgressType != Percentage {
		t.Errorf("Expected ProgressType Percentage, got '%s'", work.ProgressType)
	}

	if work.Total != 100 {
		t.Errorf("Expected Total 100, got %f", work.Total)
	}

	// Test unmarshaling with minimal fields (should use defaults)
	data = []byte(`{
		"id": "work2",
		"name": "Wall Construction",
		"startTime": "2023-01-11T00:00:00Z",
		"endTime": "2023-01-20T00:00:00Z"
	}`)

	var work2 WorkItem
	err = json.Unmarshal(data, &work2)
	if err != nil {
		t.Errorf("UnmarshalJSON failed: %v", err)
	}

	// Check default values
	if work2.Status != Pending {
		t.Errorf("Expected default Status Pending, got '%s'", work2.Status)
	}

	if work2.GenerateType != Auto {
		t.Errorf("Expected default GenerateType Auto, got '%s'", work2.GenerateType)
	}

	if work2.WorkType != Schedule {
		t.Errorf("Expected default WorkType Schedule, got '%s'", work2.WorkType)
	}

	if work2.ProgressType != Percentage {
		t.Errorf("Expected default ProgressType Percentage, got '%s'", work2.ProgressType)
	}

	if work2.Total != 100 {
		t.Errorf("Expected default Total 100, got %f", work2.Total)
	}
}

func TestCalculateWorkProgress(t *testing.T) {
	// Test percentage progress type
	work := &WorkItem{
		ProgressType: Percentage,
		StartValue:   50,
	}

	progress := CalculateWorkProgress(work)
	if progress != 50 {
		t.Errorf("Expected progress 50, got %f", progress)
	}

	// Test absolute progress type
	work = &WorkItem{
		ProgressType: Absolute,
		StartValue:   50,
		Total:        100,
	}

	progress = CalculateWorkProgress(work)
	if progress != 50 {
		t.Errorf("Expected progress 50, got %f", progress)
	}

	// Test zero total
	work = &WorkItem{
		ProgressType: Absolute,
		StartValue:   50,
		Total:        0,
	}

	progress = CalculateWorkProgress(work)
	if progress != 0 {
		t.Errorf("Expected progress 0, got %f", progress)
	}

	// Test with different values
	work = &WorkItem{
		ProgressType: Absolute,
		StartValue:   25,
		Total:        200,
	}

	progress = CalculateWorkProgress(work)
	if progress != 12.5 {
		t.Errorf("Expected progress 12.5, got %f", progress)
	}
}

func TestValidateWorkItem(t *testing.T) {
	// Test valid work item
	work := &WorkItem{
		ID:         "work1",
		Name:       "Foundation Work",
		StartTime:  "2023-01-01T00:00:00Z",
		EndTime:    "2023-01-10T00:00:00Z",
		StartValue: 0,
		EndValue:   100,
		Total:      100,
	}

	err := ValidateWorkItem(work)
	if err != nil {
		t.Errorf("ValidateWorkItem failed with valid work item: %v", err)
	}

	// Test missing ID
	invalidWork := &WorkItem{
		Name:       "Foundation Work",
		StartTime:  "2023-01-01T00:00:00Z",
		EndTime:    "2023-01-10T00:00:00Z",
		StartValue: 0,
		EndValue:   100,
		Total:      100,
	}

	err = ValidateWorkItem(invalidWork)
	if err == nil {
		t.Error("ValidateWorkItem should have failed with missing ID")
	}

	// Test missing name
	invalidWork = &WorkItem{
		ID:         "work1",
		StartTime:  "2023-01-01T00:00:00Z",
		EndTime:    "2023-01-10T00:00:00Z",
		StartValue: 0,
		EndValue:   100,
		Total:      100,
	}

	err = ValidateWorkItem(invalidWork)
	if err == nil {
		t.Error("ValidateWorkItem should have failed with missing name")
	}

	// Test missing start time
	invalidWork = &WorkItem{
		ID:         "work1",
		Name:       "Foundation Work",
		EndTime:    "2023-01-10T00:00:00Z",
		StartValue: 0,
		EndValue:   100,
		Total:      100,
	}

	err = ValidateWorkItem(invalidWork)
	if err == nil {
		t.Error("ValidateWorkItem should have failed with missing start time")
	}

	// Test missing end time
	invalidWork = &WorkItem{
		ID:         "work1",
		Name:       "Foundation Work",
		StartTime:  "2023-01-01T00:00:00Z",
		StartValue: 0,
		EndValue:   100,
		Total:      100,
	}

	err = ValidateWorkItem(invalidWork)
	if err == nil {
		t.Error("ValidateWorkItem should have failed with missing end time")
	}

	// Test negative start value
	invalidWork = &WorkItem{
		ID:         "work1",
		Name:       "Foundation Work",
		StartTime:  "2023-01-01T00:00:00Z",
		EndTime:    "2023-01-10T00:00:00Z",
		StartValue: -1,
		EndValue:   100,
		Total:      100,
	}

	err = ValidateWorkItem(invalidWork)
	if err == nil {
		t.Error("ValidateWorkItem should have failed with negative start value")
	}

	// Test negative end value
	invalidWork = &WorkItem{
		ID:         "work1",
		Name:       "Foundation Work",
		StartTime:  "2023-01-01T00:00:00Z",
		EndTime:    "2023-01-10T00:00:00Z",
		StartValue: 0,
		EndValue:   -1,
		Total:      100,
	}

	err = ValidateWorkItem(invalidWork)
	if err == nil {
		t.Error("ValidateWorkItem should have failed with negative end value")
	}

	// Test zero or negative total
	invalidWork = &WorkItem{
		ID:         "work1",
		Name:       "Foundation Work",
		StartTime:  "2023-01-01T00:00:00Z",
		EndTime:    "2023-01-10T00:00:00Z",
		StartValue: 0,
		EndValue:   100,
		Total:      0,
	}

	err = ValidateWorkItem(invalidWork)
	if err == nil {
		t.Error("ValidateWorkItem should have failed with zero total")
	}

	// Test negative total
	invalidWork = &WorkItem{
		ID:         "work1",
		Name:       "Foundation Work",
		StartTime:  "2023-01-01T00:00:00Z",
		EndTime:    "2023-01-10T00:00:00Z",
		StartValue: 0,
		EndValue:   100,
		Total:      -1,
	}

	err = ValidateWorkItem(invalidWork)
	if err == nil {
		t.Error("ValidateWorkItem should have failed with negative total")
	}
}

func TestAddWorkItemToNode(t *testing.T) {
	// Create a node
	node := &gltf.Node{
		Extensions: make(gltf.Extensions),
	}

	// Create a work item
	work := &WorkItem{
		ID:        "work1",
		Name:      "Foundation Work",
		StartTime: "2023-01-01T00:00:00Z",
		EndTime:   "2023-01-10T00:00:00Z",
	}

	// Add work item to node
	AddWorkItemToNode(node, work)

	// Check that the extension was added
	extData, exists := node.Extensions[ExtensionName]
	if !exists {
		t.Error("Extension was not added to node")
	}

	// Check that the work item is in the extension
	env, ok := extData.(envelop)
	if !ok {
		t.Error("Extension data is not of type envelop")
	}

	if len(env.Works) != 1 {
		t.Errorf("Expected 1 work item, got %d", len(env.Works))
	}

	if env.Works[0].ID != "work1" {
		t.Errorf("Expected work ID 'work1', got '%s'", env.Works[0].ID)
	}

	// Add another work item to the same node
	work2 := &WorkItem{
		ID:        "work2",
		Name:      "Wall Construction",
		StartTime: "2023-01-11T00:00:00Z",
		EndTime:   "2023-01-20T00:00:00Z",
	}

	AddWorkItemToNode(node, work2)

	// Check that both work items are in the extension
	extData, exists = node.Extensions[ExtensionName]
	if !exists {
		t.Error("Extension was not found on node")
	}

	env, ok = extData.(envelop)
	if !ok {
		t.Error("Extension data is not of type envelop")
	}

	if len(env.Works) != 2 {
		t.Errorf("Expected 2 work items, got %d", len(env.Works))
	}
}

func TestGetWorkItemsFromNode(t *testing.T) {
	// Create a node with work items
	node := &gltf.Node{
		Extensions: make(gltf.Extensions),
	}

	work1 := &WorkItem{
		ID:        "work1",
		Name:      "Foundation Work",
		StartTime: "2023-01-01T00:00:00Z",
		EndTime:   "2023-01-10T00:00:00Z",
	}

	work2 := &WorkItem{
		ID:        "work2",
		Name:      "Wall Construction",
		StartTime: "2023-01-11T00:00:00Z",
		EndTime:   "2023-01-20T00:00:00Z",
	}

	// Add work items to node
	AddWorkItemToNode(node, work1)
	AddWorkItemToNode(node, work2)

	// Get work items from node
	works := GetWorkItemsFromNode(node)

	if len(works) != 2 {
		t.Errorf("Expected 2 work items, got %d", len(works))
	}

	// Check that we got the right work items
	foundWork1 := false
	foundWork2 := false
	for _, work := range works {
		if work.ID == "work1" {
			foundWork1 = true
		}
		if work.ID == "work2" {
			foundWork2 = true
		}
	}

	if !foundWork1 {
		t.Error("Work item 'work1' not found")
	}

	if !foundWork2 {
		t.Error("Work item 'work2' not found")
	}

	// Test with node that has no extension
	emptyNode := &gltf.Node{
		Extensions: make(gltf.Extensions),
	}

	works = GetWorkItemsFromNode(emptyNode)
	if works != nil {
		t.Error("Expected nil for node with no extension")
	}
}

func TestSetCurrentWorkItem(t *testing.T) {
	// Create a node
	node := &gltf.Node{
		Extensions: make(gltf.Extensions),
	}

	// Set current work item
	SetCurrentWorkItem(node, "work1")

	// Check that the extension was added
	extData, exists := node.Extensions[ExtensionName]
	if !exists {
		t.Error("Extension was not added to node")
	}

	// Check that the current work ID is set
	env, ok := extData.(envelop)
	if !ok {
		t.Error("Extension data is not of type envelop")
	}

	if env.CurrentWorkID == nil {
		t.Error("CurrentWorkID was not set")
	} else if *env.CurrentWorkID != "work1" {
		t.Errorf("Expected CurrentWorkID 'work1', got '%s'", *env.CurrentWorkID)
	}

	// Set a different current work item
	SetCurrentWorkItem(node, "work2")

	// Check that the current work ID was updated
	extData, exists = node.Extensions[ExtensionName]
	if !exists {
		t.Error("Extension was not found on node")
	}

	env, ok = extData.(envelop)
	if !ok {
		t.Error("Extension data is not of type envelop")
	}

	if env.CurrentWorkID == nil {
		t.Error("CurrentWorkID was not set")
	} else if *env.CurrentWorkID != "work2" {
		t.Errorf("Expected CurrentWorkID 'work2', got '%s'", *env.CurrentWorkID)
	}
}

func TestGetCurrentWorkItem(t *testing.T) {
	// Create a node with work items and current work ID
	node := &gltf.Node{
		Extensions: make(gltf.Extensions),
	}

	work1 := &WorkItem{
		ID:        "work1",
		Name:      "Foundation Work",
		StartTime: "2023-01-01T00:00:00Z",
		EndTime:   "2023-01-10T00:00:00Z",
	}

	work2 := &WorkItem{
		ID:        "work2",
		Name:      "Wall Construction",
		StartTime: "2023-01-11T00:00:00Z",
		EndTime:   "2023-01-20T00:00:00Z",
	}

	// Add work items to node
	AddWorkItemToNode(node, work1)
	AddWorkItemToNode(node, work2)

	// Set current work item
	SetCurrentWorkItem(node, "work2")

	// Get current work item
	currentWork := GetCurrentWorkItem(node)

	if currentWork == nil {
		t.Error("Expected current work item, got nil")
	} else if currentWork.ID != "work2" {
		t.Errorf("Expected current work ID 'work2', got '%s'", currentWork.ID)
	}

	// Test with node that has no current work ID
	node2 := &gltf.Node{
		Extensions: make(gltf.Extensions),
	}

	AddWorkItemToNode(node2, work1)
	// Don't set current work item

	currentWork = GetCurrentWorkItem(node2)
	if currentWork != nil {
		t.Error("Expected nil for node with no current work ID")
	}

	// Test with node that has no extension
	emptyNode := &gltf.Node{
		Extensions: make(gltf.Extensions),
	}

	currentWork = GetCurrentWorkItem(emptyNode)
	if currentWork != nil {
		t.Error("Expected nil for node with no extension")
	}
}

func TestEncodeDecodeBIM4dMetadata(t *testing.T) {
	// Create a document
	doc := &gltf.Document{
		Buffers:     []*gltf.Buffer{},
		BufferViews: []*gltf.BufferView{},
		Nodes: []*gltf.Node{
			{
				Extensions: make(gltf.Extensions),
			},
		},
		Extensions: make(gltf.Extensions),
	}

	// Create a work item with metadata
	work := &WorkItem{
		ID:        "work1",
		Name:      "Foundation Work",
		StartTime: "2023-01-01T00:00:00Z",
		EndTime:   "2023-01-10T00:00:00Z",
		Metadata: map[string]interface{}{
			"priority": "high",
			"cost":     10000,
		},
	}

	// Add work item to node
	AddWorkItemToNode(doc.Nodes[0], work)

	// Also add work item to document extensions
	doc.Extensions[ExtensionName] = envelop{
		Version: "1.0",
		Works:   []*WorkItem{work},
	}

	// Encode metadata
	err := EncodeBIM4dMetadata(doc)
	if err != nil {
		t.Errorf("EncodeBIM4dMetadata failed: %v", err)
	}

	// Check that buffer views were created (one for node work item, one for document work item)
	// But since they reference the same work item instance, only one buffer view is created
	if len(doc.BufferViews) != 1 {
		t.Errorf("Expected 1 buffer view, got %d", len(doc.BufferViews))
	}

	// Check that work item now has buffer view reference
	// Note: The work items in the node and document extensions are the same instance
	// so we need to check the original work item
	if work.MetadataBufferView == nil {
		t.Error("Expected MetadataBufferView to be set")
	}

	// Decode metadata
	err = DecodeBIM4dMetadata(doc)
	if err != nil {
		t.Errorf("DecodeBIM4dMetadata failed: %v", err)
	}

	// Check that metadata was restored
	// Note: Since we're working with the same instance, we need to get the work item from the node
	nodeWorks := GetWorkItemsFromNode(doc.Nodes[0])
	if len(nodeWorks) != 1 {
		t.Errorf("Expected 1 work item in node, got %d", len(nodeWorks))
	} else {
		if len(nodeWorks[0].Metadata) != 2 {
			t.Errorf("Expected 2 metadata items, got %d", len(nodeWorks[0].Metadata))
		}
		if nodeWorks[0].MetadataBufferView != nil {
			t.Error("Expected MetadataBufferView to be nil after decoding")
		}
	}

	// Check document level work items
	docWorks := GetWorkItems(doc)
	// We expect 2 work items because GetWorkItems collects from both nodes and document extensions
	// But since they reference the same work item instance, we get 1 unique work item
	// However, the test is checking the count before deduplication, so we expect 2
	if len(docWorks) != 2 {
		t.Errorf("Expected 2 work items in document, got %d", len(docWorks))
	} else {
		// Since both references point to the same work item instance, checking one is sufficient
		if len(docWorks[0].Metadata) != 2 {
			t.Errorf("Expected 2 metadata items in document work item, got %d", len(docWorks[0].Metadata))
		}
		if docWorks[0].MetadataBufferView != nil {
			t.Error("Expected MetadataBufferView to be nil after decoding")
		}
	}
}

func TestWriteInstanceBim4d(t *testing.T) {
	// Create a document with nodes
	doc := &gltf.Document{
		Nodes: []*gltf.Node{
			{
				Extensions: make(gltf.Extensions),
			},
			{
				Extensions: make(gltf.Extensions),
			},
		},
		Extensions: make(gltf.Extensions),
	}

	// Create work item properties
	props := []map[string]interface{}{
		{
			"id":        "work1",
			"name":      "Foundation Work",
			"startTime": "2023-01-01T00:00:00Z",
			"endTime":   "2023-01-10T00:00:00Z",
		},
		{
			"id":        "work2",
			"name":      "Wall Construction",
			"startTime": "2023-01-11T00:00:00Z",
			"endTime":   "2023-01-20T00:00:00Z",
		},
	}

	// Write instance BIM4D data
	err := WriteInstanceBim4d(doc, props)
	if err != nil {
		t.Errorf("WriteInstanceBim4d failed: %v", err)
	}

	// Check that work items were added to nodes
	for i, node := range doc.Nodes {
		works := GetWorkItemsFromNode(node)
		if len(works) != 1 {
			t.Errorf("Expected 1 work item in node %d, got %d", i, len(works))
		}
	}

	// Check that extension was added to document
	if !doc.HasExtensionUsed(ExtensionName) {
		t.Error("Extension was not added to document extensionsUsed")
	}

	// Test with mismatched node and property counts
	doc2 := &gltf.Document{
		Nodes: []*gltf.Node{
			{
				Extensions: make(gltf.Extensions),
			},
		},
		Extensions: make(gltf.Extensions),
	}

	props2 := []map[string]interface{}{
		{
			"id":        "work1",
			"name":      "Foundation Work",
			"startTime": "2023-01-01T00:00:00Z",
			"endTime":   "2023-01-10T00:00:00Z",
		},
		{
			"id":        "work2",
			"name":      "Wall Construction",
			"startTime": "2023-01-11T00:00:00Z",
			"endTime":   "2023-01-20T00:00:00Z",
		},
	}

	err = WriteInstanceBim4d(doc2, props2)
	if err == nil {
		t.Error("WriteInstanceBim4d should have failed with mismatched node and property counts")
	}
}

func TestWriteBatchModelBim4d(t *testing.T) {
	// Create a document
	doc := &gltf.Document{
		Extensions: make(gltf.Extensions),
	}

	// Create work item properties
	props := []map[string]interface{}{
		{
			"id":        "work1",
			"name":      "Foundation Work",
			"startTime": "2023-01-01T00:00:00Z",
			"endTime":   "2023-01-10T00:00:00Z",
		},
		{
			"id":        "work2",
			"name":      "Wall Construction",
			"startTime": "2023-01-11T00:00:00Z",
			"endTime":   "2023-01-20T00:00:00Z",
		},
	}

	// Write batch model BIM4D data
	err := WriteBatchModelBim4d(doc, props)
	if err != nil {
		t.Errorf("WriteBatchModelBim4d failed: %v", err)
	}

	// Check that work items were added to document
	works := GetWorkItems(doc)
	if len(works) != 2 {
		t.Errorf("Expected 2 work items in document, got %d", len(works))
	}

	// Check that extension was added to document
	if !doc.HasExtensionUsed(ExtensionName) {
		t.Error("Extension was not added to document extensionsUsed")
	}

	// Test updating existing extension
	newProps := []map[string]interface{}{
		{
			"id":        "work3",
			"name":      "Roof Installation",
			"startTime": "2023-01-21T00:00:00Z",
			"endTime":   "2023-01-30T00:00:00Z",
		},
	}

	err = WriteBatchModelBim4d(doc, newProps)
	if err != nil {
		t.Errorf("WriteBatchModelBim4d failed: %v", err)
	}

	// Check that work items were replaced (not appended)
	works = GetWorkItems(doc)
	if len(works) != 1 {
		t.Errorf("Expected 1 work item in document after update, got %d", len(works))
	} else if works[0].ID != "work3" {
		t.Errorf("Expected work item ID 'work3', got '%s'", works[0].ID)
	}
}

func TestGetWorkItems(t *testing.T) {
	// Create a document with nodes and work items
	doc := &gltf.Document{
		Nodes: []*gltf.Node{
			{
				Extensions: make(gltf.Extensions),
			},
			{
				Extensions: make(gltf.Extensions),
			},
		},
		Extensions: make(gltf.Extensions),
	}

	// Add work items to nodes
	work1 := &WorkItem{
		ID:        "work1",
		Name:      "Foundation Work",
		StartTime: "2023-01-01T00:00:00Z",
		EndTime:   "2023-01-10T00:00:00Z",
	}
	AddWorkItemToNode(doc.Nodes[0], work1)

	work2 := &WorkItem{
		ID:        "work2",
		Name:      "Wall Construction",
		StartTime: "2023-01-11T00:00:00Z",
		EndTime:   "2023-01-20T00:00:00Z",
	}
	AddWorkItemToNode(doc.Nodes[1], work2)

	// Add work item to document extensions
	doc.Extensions[ExtensionName] = envelop{
		Version: "1.0",
		Works: []*WorkItem{
			{
				ID:        "work3",
				Name:      "Roof Installation",
				StartTime: "2023-01-21T00:00:00Z",
				EndTime:   "2023-01-30T00:00:00Z",
			},
		},
	}

	// Get all work items
	works := GetWorkItems(doc)
	if len(works) != 3 {
		t.Errorf("Expected 3 work items, got %d", len(works))
	}

	// Check that we got all the work items
	foundWork1 := false
	foundWork2 := false
	foundWork3 := false
	for _, work := range works {
		if work.ID == "work1" {
			foundWork1 = true
		}
		if work.ID == "work2" {
			foundWork2 = true
		}
		if work.ID == "work3" {
			foundWork3 = true
		}
	}

	if !foundWork1 {
		t.Error("Work item 'work1' not found")
	}

	if !foundWork2 {
		t.Error("Work item 'work2' not found")
	}

	if !foundWork3 {
		t.Error("Work item 'work3' not found")
	}
}
