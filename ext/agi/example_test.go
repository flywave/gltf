package agi

import (
	"encoding/json"
	"fmt"

	"github.com/flywave/gltf"
)

func ExampleRegisterExtensions() {
	// Register the AGI extensions
	RegisterExtensions()

	// Now you can use the extensions in your glTF documents
	fmt.Println("AGI extensions registered successfully")
}

func ExampleAgiRootArticulations() {
	// Register the extensions first
	RegisterExtensions()

	// Create a new glTF document
	doc := gltf.NewDocument()

	// Create root articulations extension
	rootArticulations := &AgiRootArticulations{}

	// Create an articulation
	articulation := rootArticulations.CreateArticulation("SampleArticulation")

	// Set a pointing vector (unit vector)
	vector := &[3]float64{1.0, 0.0, 0.0}
	articulation.SetPointingVector(vector)

	// Create an articulation stage
	stage, err := articulation.CreateArticulationStage("RotationStage", AgiArticulationTransformTypeXRotate)
	if err != nil {
		panic(err)
	}

	// Set stage values
	stage.SetValues(-180.0, 0.0, 180.0)

	// Add the extension to the document
	extData, err := json.Marshal(rootArticulations)
	if err != nil {
		panic(err)
	}

	if doc.Extensions == nil {
		doc.Extensions = make(gltf.Extensions)
	}
	doc.Extensions[ArticulationsExtensionName] = extData
	doc.AddExtensionUsed(ArticulationsExtensionName)

	fmt.Printf("Created articulation with %d stages\n", len(articulation.Stages))
	// Output: Created articulation with 1 stages
}

func ExampleAgiNodeArticulations() {
	// Register the extensions first
	RegisterExtensions()

	// Create a node
	node := &gltf.Node{
		Name:       "SampleNode",
		Extensions: make(gltf.Extensions),
	}

	// Create node articulations extension
	nodeArticulations := &AgiNodeArticulations{
		ArticulationName: stringPtr("SampleArticulation"),
		IsAttachPoint:    boolPtr(true),
	}

	// Add the extension to the node
	extData, err := json.Marshal(nodeArticulations)
	if err != nil {
		panic(err)
	}

	node.Extensions[ArticulationsExtensionName] = extData

	fmt.Printf("Node articulation name: %s\n", *nodeArticulations.ArticulationName)
	// Output: Node articulation name: SampleArticulation
}

func ExampleAgiRootStkMetadata() {
	// Register the extensions first
	RegisterExtensions()

	// Create a new glTF document
	doc := gltf.NewDocument()

	// Create root STK metadata extension
	rootStkMetadata := &AgiRootStkMetadata{}

	// Create a solar panel group
	group := rootStkMetadata.CreateSolarPanelGroup("SampleSolarPanelGroup")
	group.SetEfficiency(0.85)

	// Add the extension to the document
	extData, err := json.Marshal(rootStkMetadata)
	if err != nil {
		panic(err)
	}

	if doc.Extensions == nil {
		doc.Extensions = make(gltf.Extensions)
	}
	doc.Extensions[StkMetadataExtensionName] = extData
	doc.AddExtensionUsed(StkMetadataExtensionName)

	fmt.Printf("Created solar panel group with efficiency: %.2f\n", group.Efficiency)
	// Output: Created solar panel group with efficiency: 0.85
}

func ExampleAgiNodeStkMetadata() {
	// Register the extensions first
	RegisterExtensions()

	// Create a node
	node := &gltf.Node{
		Name:       "SampleNode",
		Extensions: make(gltf.Extensions),
	}

	// Create node STK metadata extension
	nodeStkMetadata := &AgiNodeStkMetadata{
		SolarPanelGroupName: stringPtr("SampleSolarPanelGroup"),
	}

	// Set no obscuration flag
	nodeStkMetadata.SetNoObscuration(true)

	// Add the extension to the node
	extData, err := json.Marshal(nodeStkMetadata)
	if err != nil {
		panic(err)
	}

	node.Extensions[StkMetadataExtensionName] = extData

	fmt.Printf("Node solar panel group: %s\n", *nodeStkMetadata.SolarPanelGroupName)
	fmt.Printf("No obscuration: %t\n", nodeStkMetadata.GetNoObscuration())
	// Output:
	// Node solar panel group: SampleSolarPanelGroup
	// No obscuration: true
}

// Helper functions for creating pointers
func stringPtr(s string) *string {
	return &s
}

func boolPtr(b bool) *bool {
	return &b
}
