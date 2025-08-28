package bimdata

import (
	"encoding/json"
	"fmt"

	"github.com/flywave/gltf"
)

func ExampleBimData() {
	// Create a new glTF document
	doc := gltf.NewDocument()

	// Create nodes
	node1 := &gltf.Node{
		Name:       "Door 1",
		Extensions: make(gltf.Extensions),
	}
	doc.Nodes = append(doc.Nodes, node1)

	node2 := &gltf.Node{
		Name:       "Door 2",
		Extensions: make(gltf.Extensions),
	}
	doc.Nodes = append(doc.Nodes, node2)

	// Create BIM data extension for nodes
	nodeExt1 := &BimData{}
	nodeExt1.SetPropertyIndices([]uint32{0})
	typeIndex := uint32(0)
	nodeExt1.SetTypeIndex(typeIndex)

	nodeExt2 := &BimData{}
	nodeExt2.SetPropertyIndices([]uint32{3})
	nodeExt2.SetTypeIndex(typeIndex)

	// Add the extension to the nodes
	extData1, err := json.Marshal(nodeExt1)
	if err != nil {
		panic(err)
	}
	node1.Extensions[BimDataExtensionName] = extData1

	extData2, err := json.Marshal(nodeExt2)
	if err != nil {
		panic(err)
	}
	node2.Extensions[BimDataExtensionName] = extData2

	// Create root BIM data extension
	rootExt := &BimDataRoot{
		PropertyNames:  []string{"Height", "Width", "Material"},
		PropertyValues: []string{"2,1 m", "900 mm", "Timber", "2,4 m"},
		Properties: []BimProperty{
			{Name: 0, Value: 0},
			{Name: 1, Value: 1},
			{Name: 2, Value: 2},
			{Name: 0, Value: 3},
		},
		Types: []BimType{
			{Properties: []uint32{1, 2}},
		},
	}

	// Add the root extension to the document
	rootExtData, err := json.Marshal(rootExt)
	if err != nil {
		panic(err)
	}

	if doc.Extensions == nil {
		doc.Extensions = make(gltf.Extensions)
	}
	doc.Extensions[BimDataExtensionName] = rootExtData
	doc.AddExtensionUsed(BimDataExtensionName)

	fmt.Printf("Node 1 properties: %v\n", nodeExt1.GetPropertyIndices())
	fmt.Printf("Node 1 type: %d\n", *nodeExt1.GetTypeIndex())
	fmt.Printf("Node 2 properties: %v\n", nodeExt2.GetPropertyIndices())
	fmt.Printf("Node 2 type: %d\n", *nodeExt2.GetTypeIndex())
	fmt.Printf("Root property names: %v\n", rootExt.PropertyNames)
	fmt.Printf("Root property values: %v\n", rootExt.PropertyValues)

	// Output:
	// Node 1 properties: [0]
	// Node 1 type: 0
	// Node 2 properties: [3]
	// Node 2 type: 0
	// Root property names: [Height Width Material]
	// Root property values: [2,1 m 900 mm Timber 2,4 m]
}
