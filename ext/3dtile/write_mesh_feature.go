package tile3d

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"

	"github.com/flywave/gltf"
	ext_gltf "github.com/flywave/gltf/ext/3dtile/gltf"
	ext_mesh "github.com/flywave/gltf/ext/3dtile/mesh"
)

// WriteFeatureData 写入特征数据和FeatureID到glTF文档
func WriteFeatureData(doc *gltf.Document, propertiesArray []map[string]interface{}, featureIDs [][][]int16) error {
	if doc == nil {
		return fmt.Errorf("glTF document cannot be nil")
	}
	if len(propertiesArray) == 0 {
		return nil
	}
	if len(featureIDs) == 0 {
		return fmt.Errorf("featureIDs cannot be empty")
	}

	if len(doc.Meshes) == 0 {
		return fmt.Errorf("document must have mesh")
	}

	metadata, err := getOrCreateMetadata(doc)
	if err != nil {
		return err
	}

	// 4. 更新Schema
	class := "featureClass"
	properties := rackProps(propertiesArray)

	if err := updateSchema(metadata, class, properties); err != nil {
		return err
	}

	tableIndex, err := createPropertyTable(doc, metadata, class, properties)
	if err != nil {
		return err
	}

	if err := createFeatureIDAttributes(doc, featureIDs, uint32(tableIndex)); err != nil {
		return err
	}

	doc.Extensions[ext_gltf.ExtensionName] = metadata
	addExtensionUsed(doc, ext_gltf.ExtensionName)
	addExtensionUsed(doc, ext_mesh.ExtensionName)

	return nil
}

// 获取或创建EXT_structural_metadata扩展
func getOrCreateMetadata(doc *gltf.Document) (*ext_gltf.ExtStructuralMetadata, error) {
	var metadata ext_gltf.ExtStructuralMetadata
	if ext, exists := doc.Extensions[ext_gltf.ExtensionName]; exists {
		if err := unmarshalExtension(ext, &metadata); err != nil {
			return nil, err
		}
	}

	if metadata.Schema == nil {
		metadata.Schema = &ext_gltf.Schema{
			Classes: make(map[string]ext_gltf.Class),
		}
	}
	if metadata.PropertyTables == nil {
		metadata.PropertyTables = make([]ext_gltf.PropertyTable, 0)
	}

	return &metadata, nil
}

func updateSchema(metadata *ext_gltf.ExtStructuralMetadata, class string, sampleProps map[string][]interface{}) error {
	if _, exists := metadata.Schema.Classes[class]; exists {
		return nil
	}

	props := make(map[string]ext_gltf.ClassProperty)
	for name, val := range sampleProps {
		propType, componentType, err := inferPropertyType(val[0])
		if err != nil {
			return err
		}
		props[name] = ext_gltf.ClassProperty{
			Type:          propType,
			ComponentType: componentType,
		}
	}
	metadata.Schema.Classes[class] = ext_gltf.Class{Properties: props}
	return nil
}

// 创建属性表并返回索引
func createPropertyTable(doc *gltf.Document, metadata *ext_gltf.ExtStructuralMetadata, class string, props map[string][]interface{}) (int, error) {
	// 1. 创建属性访问器
	propertyDefs := make(map[string]ext_gltf.PropertyTableProperty)
	for name, values := range props {
		accessor, err := createPropertyAccessor(doc, values)
		if err != nil {
			return 0, fmt.Errorf("create property accessor %s: %w", name, err)
		}
		propertyDefs[name] = *accessor
	}

	// 2. 添加PropertyTable
	table := ext_gltf.PropertyTable{
		Class:      class,
		Count:      uint32(len(props)),
		Properties: propertyDefs,
	}
	metadata.PropertyTables = append(metadata.PropertyTables, table)

	return len(metadata.PropertyTables) - 1, nil
}

// 创建FeatureID属性并关联到Mesh
func createFeatureIDAttributes(doc *gltf.Document, featureIDs [][][]int16, tableIndex uint32) error {
	// 1. 为每个Primitive创建FeatureID属性
	meshFeatures := ext_mesh.ExtMeshFeatures{}
	if ext, exists := doc.Extensions[ext_mesh.ExtensionName]; exists {
		if err := unmarshalExtension(ext, &meshFeatures); err != nil {
			return err
		}
	}
	count := 0
	for meshIdx, idss := range featureIDs {
		mesh := doc.Meshes[meshIdx]
		for primIdx, ids := range idss {
			// 2. 创建BufferView
			buf := new(bytes.Buffer)
			if err := binary.Write(buf, binary.LittleEndian, ids); err != nil {
				return err
			}
			data := buf.Bytes()

			bufferView := gltf.BufferView{
				Buffer:     0,
				ByteOffset: doc.Buffers[0].ByteLength,
				ByteLength: uint32(len(data)),
				Target:     gltf.TargetArrayBuffer,
			}
			bufferViewIdx := uint32(len(doc.BufferViews))
			doc.BufferViews = append(doc.BufferViews, &bufferView)

			doc.Buffers[0].Data = append(doc.Buffers[0].Data, data...)
			doc.Buffers[0].ByteLength += uint32(len(data))
			padBuffer(doc)

			// 4. 创建Accessor
			accessor := gltf.Accessor{
				BufferView:    &bufferViewIdx,
				ComponentType: gltf.ComponentShort,
				Count:         uint32(len(ids)),
				Type:          gltf.AccessorScalar,
			}
			accessorIdx := uint32(len(doc.Accessors))
			doc.Accessors = append(doc.Accessors, &accessor)

			attrName := fmt.Sprintf("_FEATURE_ID_%d", count)
			count++
			primitive := mesh.Primitives[primIdx]
			if primitive.Attributes == nil {
				primitive.Attributes = make(map[string]uint32)
			}
			primitive.Attributes[attrName] = accessorIdx

			// 6. 记录FeatureID关联
			meshFeatures.FeatureIDs = append(meshFeatures.FeatureIDs, ext_mesh.FeatureID{
				Attribute:     &accessorIdx,
				PropertyTable: &tableIndex,
			})
		}
	}

	// 7. 更新EXT_mesh_features
	doc.Extensions[ext_mesh.ExtensionName] = meshFeatures
	return nil
}

// 辅助函数：添加扩展声明
func addExtensionUsed(doc *gltf.Document, ext string) {
	for _, e := range doc.ExtensionsUsed {
		if e == ext {
			return
		}
	}
	doc.ExtensionsUsed = append(doc.ExtensionsUsed, ext)
}

// 辅助函数：解析扩展数据
func unmarshalExtension(ext interface{}, target interface{}) error {
	raw, err := json.Marshal(ext)
	if err != nil {
		return err
	}
	return json.Unmarshal(raw, target)
}
