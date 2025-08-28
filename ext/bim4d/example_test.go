package bim4d

import (
	"fmt"
	"log"

	"github.com/flywave/gltf"
)

func ExampleCreateWorkItem() {
	// Create work item properties
	props := map[string]interface{}{
		"id":           "foundation-001",
		"name":         "浇筑地基",
		"description":  "浇筑建筑地基混凝土",
		"status":       InProgress,
		"generateType": Manual,
		"workType":     Schedule,
		"startTime":    "2023-01-01T08:00:00Z",
		"endTime":      "2023-01-05T17:00:00Z",
		"startValue":   float32(0),
		"endValue":     float32(100),
		"progressType": Percentage,
		"total":        float32(100),
		"metadata": map[string]interface{}{
			"contractor": "ABC建筑公司",
			"budget":     50000,
		},
	}

	// Create a work item
	work := CreateWorkItem(props)

	fmt.Printf("Work ID: %s\n", work.ID)
	fmt.Printf("Work Name: %s\n", work.Name)
	fmt.Printf("Status: %s\n", work.Status)
	fmt.Printf("Progress: %.1f%%\n", CalculateWorkProgress(work))
	fmt.Printf("Contractor: %s\n", work.Metadata["contractor"])

	// Output:
	// Work ID: foundation-001
	// Work Name: 浇筑地基
	// Status: in_progress
	// Progress: 0.0%
	// Contractor: ABC建筑公司
}

func ExampleAddWorkItemToNode() {
	// Create a glTF document
	doc := &gltf.Document{
		Nodes: []*gltf.Node{
			{Extensions: make(gltf.Extensions)},
		},
		Extensions: make(gltf.Extensions),
	}

	// Create a work item
	work := &WorkItem{
		ID:        "wall-001",
		Name:      "砌筑墙体",
		StartTime: "2023-01-10T08:00:00Z",
		EndTime:   "2023-01-15T17:00:00Z",
	}

	// Add work item to node
	AddWorkItemToNode(doc.Nodes[0], work)

	// Get work items from node
	works := GetWorkItemsFromNode(doc.Nodes[0])
	if len(works) > 0 {
		fmt.Printf("Node has %d work item(s)\n", len(works))
		fmt.Printf("First work item: %s\n", works[0].Name)
	}

	// Output:
	// Node has 1 work item(s)
	// First work item: 砌筑墙体
}

func ExampleWriteBatchModelBim4d() {
	// Create a glTF document
	doc := &gltf.Document{
		Extensions: make(gltf.Extensions),
	}

	// Create work item properties
	props := []map[string]interface{}{
		{
			"id":        "foundation-001",
			"name":      "浇筑地基",
			"startTime": "2023-01-01T08:00:00Z",
			"endTime":   "2023-01-05T17:00:00Z",
		},
		{
			"id":        "structure-001",
			"name":      "搭建结构",
			"startTime": "2023-01-06T08:00:00Z",
			"endTime":   "2023-01-15T17:00:00Z",
		},
		{
			"id":        "roof-001",
			"name":      "安装屋顶",
			"startTime": "2023-01-16T08:00:00Z",
			"endTime":   "2023-01-20T17:00:00Z",
		},
	}

	// Write batch model BIM4D data
	err := WriteBatchModelBim4d(doc, props)
	if err != nil {
		log.Fatalf("Failed to write BIM4D data: %v", err)
	}

	// Get all work items
	works := GetWorkItems(doc)
	fmt.Printf("Document has %d work item(s)\n", len(works))

	// Check that extension was added
	if doc.HasExtensionUsed(ExtensionName) {
		fmt.Println("BIM4D extension is used in the document")
	}

	// Output:
	// Document has 3 work item(s)
	// BIM4D extension is used in the document
}

func ExampleEncodeBIM4dMetadata() {
	// Create a glTF document
	doc := &gltf.Document{
		Buffers:     []*gltf.Buffer{},
		BufferViews: []*gltf.BufferView{},
		Nodes: []*gltf.Node{
			{Extensions: make(gltf.Extensions)},
		},
		Extensions: make(gltf.Extensions),
	}

	// Create a work item with metadata
	work := &WorkItem{
		ID:        "electrical-001",
		Name:      "电气安装",
		StartTime: "2023-02-01T08:00:00Z",
		EndTime:   "2023-02-10T17:00:00Z",
		Metadata: map[string]interface{}{
			"contractor": "XYZ电气公司",
			"supervisor": "张工",
			"materials":  []string{"电缆", "开关", "插座"},
		},
	}

	// Add work item to node
	AddWorkItemToNode(doc.Nodes[0], work)

	fmt.Printf("Before encoding - Metadata items: %d\n", len(work.Metadata))
	fmt.Printf("Before encoding - Has buffer view: %t\n", work.MetadataBufferView != nil)

	// Encode metadata
	err := EncodeBIM4dMetadata(doc)
	if err != nil {
		log.Fatalf("Failed to encode metadata: %v", err)
	}

	fmt.Printf("After encoding - Has buffer view: %t\n", work.MetadataBufferView != nil)

	// Decode metadata
	err = DecodeBIM4dMetadata(doc)
	if err != nil {
		log.Fatalf("Failed to decode metadata: %v", err)
	}

	// Get work items from node
	works := GetWorkItemsFromNode(doc.Nodes[0])
	if len(works) > 0 {
		fmt.Printf("After decoding - Metadata items: %d\n", len(works[0].Metadata))
		fmt.Printf("After decoding - Has buffer view: %t\n", works[0].MetadataBufferView != nil)
		fmt.Printf("Contractor: %s\n", works[0].Metadata["contractor"])
	}

	// Output:
	// Before encoding - Metadata items: 3
	// Before encoding - Has buffer view: false
	// After encoding - Has buffer view: true
	// After decoding - Metadata items: 3
	// After decoding - Has buffer view: false
	// Contractor: XYZ电气公司
}
