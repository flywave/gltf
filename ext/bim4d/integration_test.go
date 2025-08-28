package bim4d

import (
	"testing"

	"github.com/flywave/gltf"
)

func TestIntegration(t *testing.T) {
	// Create a complete glTF document
	doc := gltf.NewDocument()

	// Create some nodes
	node1 := &gltf.Node{
		Name:       "Foundation",
		Extensions: make(gltf.Extensions),
	}
	node2 := &gltf.Node{
		Name:       "Walls",
		Extensions: make(gltf.Extensions),
	}
	doc.Nodes = append(doc.Nodes, node1, node2)

	// Create work item properties for instance-level BIM4D data
	instanceProps := []map[string]interface{}{
		{
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
		},
		{
			"id":           "wall-001",
			"name":         "砌筑墙体",
			"description":  "砌筑建筑墙体",
			"status":       Pending,
			"generateType": Auto,
			"workType":     Schedule,
			"startTime":    "2023-01-06T08:00:00Z",
			"endTime":      "2023-01-15T17:00:00Z",
			"startValue":   float32(0),
			"endValue":     float32(100),
			"progressType": Percentage,
			"total":        float32(100),
			"metadata": map[string]interface{}{
				"contractor": "XYZ建筑公司",
				"budget":     30000,
			},
		},
	}

	// Write instance-level BIM4D data
	err := WriteInstanceBim4d(doc, instanceProps)
	if err != nil {
		t.Fatalf("WriteInstanceBim4d failed: %v", err)
	}

	// Create batch model-level work item properties
	batchProps := []map[string]interface{}{
		{
			"id":           "roof-001",
			"name":         "安装屋顶",
			"description":  "安装建筑屋顶",
			"status":       Pending,
			"generateType": Manual,
			"workType":     Schedule,
			"startTime":    "2023-01-16T08:00:00Z",
			"endTime":      "2023-01-20T17:00:00Z",
			"startValue":   float32(0),
			"endValue":     float32(100),
			"progressType": Percentage,
			"total":        float32(100),
			"metadata": map[string]interface{}{
				"contractor": "DEF建筑公司",
				"budget":     20000,
			},
		},
		{
			"id":           "electrical-001",
			"name":         "电气安装",
			"description":  "安装电气系统",
			"status":       Pending,
			"generateType": Auto,
			"workType":     Plan,
			"startTime":    "2023-01-21T08:00:00Z",
			"endTime":      "2023-01-25T17:00:00Z",
			"startValue":   float32(0),
			"endValue":     float32(100),
			"progressType": Percentage,
			"total":        float32(100),
			"metadata": map[string]interface{}{
				"contractor": "GHI电气公司",
				"budget":     15000,
			},
		},
	}

	// Write batch model-level BIM4D data
	err = WriteBatchModelBim4d(doc, batchProps)
	if err != nil {
		t.Fatalf("WriteBatchModelBim4d failed: %v", err)
	}

	// Verify the extension was added
	if !doc.HasExtensionUsed(ExtensionName) {
		t.Error("Extension was not added to extensionsUsed")
	}

	// Verify instance-level work items
	for i, node := range doc.Nodes {
		works := GetWorkItemsFromNode(node)
		if len(works) != 1 {
			t.Errorf("Expected 1 work item in node %d, got %d", i, len(works))
		}
	}

	// Verify all work items
	allWorks := GetWorkItems(doc)
	if len(allWorks) != 4 {
		t.Errorf("Expected 4 total work items, got %d", len(allWorks))
	}

	// Verify work item details
	foundFoundation := false
	foundWall := false
	foundRoof := false
	foundElectrical := false

	for _, work := range allWorks {
		switch work.ID {
		case "foundation-001":
			foundFoundation = true
			if work.Name != "浇筑地基" {
				t.Errorf("Expected name '浇筑地基', got '%s'", work.Name)
			}
			if work.Status != InProgress {
				t.Errorf("Expected status 'in_progress', got '%s'", work.Status)
			}
			if work.Metadata["contractor"] != "ABC建筑公司" {
				t.Errorf("Expected contractor 'ABC建筑公司', got '%s'", work.Metadata["contractor"])
			}
		case "wall-001":
			foundWall = true
			if work.Name != "砌筑墙体" {
				t.Errorf("Expected name '砌筑墙体', got '%s'", work.Name)
			}
			if work.Status != Pending {
				t.Errorf("Expected status 'pending', got '%s'", work.Status)
			}
			if work.Metadata["contractor"] != "XYZ建筑公司" {
				t.Errorf("Expected contractor 'XYZ建筑公司', got '%s'", work.Metadata["contractor"])
			}
		case "roof-001":
			foundRoof = true
			if work.Name != "安装屋顶" {
				t.Errorf("Expected name '安装屋顶', got '%s'", work.Name)
			}
			if work.Status != Pending {
				t.Errorf("Expected status 'pending', got '%s'", work.Status)
			}
			if work.Metadata["contractor"] != "DEF建筑公司" {
				t.Errorf("Expected contractor 'DEF建筑公司', got '%s'", work.Metadata["contractor"])
			}
		case "electrical-001":
			foundElectrical = true
			if work.Name != "电气安装" {
				t.Errorf("Expected name '电气安装', got '%s'", work.Name)
			}
			if work.Status != Pending {
				t.Errorf("Expected status 'pending', got '%s'", work.Status)
			}
			if work.Metadata["contractor"] != "GHI电气公司" {
				t.Errorf("Expected contractor 'GHI电气公司', got '%s'", work.Metadata["contractor"])
			}
		}
	}

	if !foundFoundation {
		t.Error("Foundation work item not found")
	}
	if !foundWall {
		t.Error("Wall work item not found")
	}
	if !foundRoof {
		t.Error("Roof work item not found")
	}
	if !foundElectrical {
		t.Error("Electrical work item not found")
	}

	// Test encoding and decoding metadata
	err = EncodeBIM4dMetadata(doc)
	if err != nil {
		t.Fatalf("EncodeBIM4dMetadata failed: %v", err)
	}

	// Check that work items have buffer view references
	workItemsWithBufferView := 0
	for _, work := range allWorks {
		if work.MetadataBufferView != nil {
			workItemsWithBufferView++
		}
	}
	if workItemsWithBufferView == 0 {
		t.Error("No work items have buffer view references after encoding")
	}

	// Decode metadata
	err = DecodeBIM4dMetadata(doc)
	if err != nil {
		t.Fatalf("DecodeBIM4dMetadata failed: %v", err)
	}

	// Check that metadata was restored and buffer view references were cleared
	for _, work := range allWorks {
		if len(work.Metadata) == 0 {
			t.Error("Metadata was not restored after decoding")
		}
		if work.MetadataBufferView != nil {
			t.Error("Buffer view reference was not cleared after decoding")
		}
	}

	// Test progress calculation
	foundationWork := allWorks[0] // foundation-001
	progress := CalculateWorkProgress(foundationWork)
	if progress != 0 {
		t.Errorf("Expected progress 0, got %f", progress)
	}

	// Test validation
	err = ValidateWorkItem(foundationWork)
	if err != nil {
		t.Errorf("ValidateWorkItem failed with valid work item: %v", err)
	}

	// Test invalid work item validation
	invalidWork := &WorkItem{
		Name:      "Invalid Work",
		StartTime: "2023-01-01T08:00:00Z",
		EndTime:   "2023-01-05T17:00:00Z",
		Total:     100,
	}
	err = ValidateWorkItem(invalidWork)
	if err == nil {
		t.Error("ValidateWorkItem should have failed with missing ID")
	}
}