package tile3d

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"

	extgltf "github.com/flywave/gltf/ext/3dtile/gltf"
	extmesh "github.com/flywave/gltf/ext/3dtile/mesh"

	"github.com/flywave/gltf"
)

func init() {
	gltf.RegisterExtension(extmesh.ExtensionName, UnmarshalMeshFeatures)
}

// UnmarshalMeshFeatures 反序列化EXT_mesh_features扩展数据
func UnmarshalMeshFeatures(data []byte) (interface{}, error) {
	var ext struct {
		FeatureIDs []extmesh.FeatureID `json:"featureIds"`
	}

	if err := json.Unmarshal(data, &ext); err != nil {
		return nil, fmt.Errorf("EXT_mesh_features解析失败: %w", err)
	}

	// 验证特征ID有效性
	for _, feature := range ext.FeatureIDs {
		if err := validateFeatureID(feature); err != nil {
			return nil, fmt.Errorf("无效的特征ID配置: %w", err)
		}
	}

	return extmesh.ExtMeshFeatures{FeatureIDs: ext.FeatureIDs}, nil
}

// 私有验证函数
func validateFeatureID(feature extmesh.FeatureID) error {
	if feature.FeatureCount == 0 {
		return errors.New("featureCount必须大于0")
	}

	referenceMethods := 0
	if feature.Attribute != nil {
		referenceMethods++
	}
	if feature.Texture != nil {
		referenceMethods++
	}
	if feature.PropertyTable != nil {
		referenceMethods++
	}

	if referenceMethods != 1 {
		return errors.New("必须且只能指定一种特征ID引用方式")
	}
	return nil
}

// MeshFeaturesEncoder 提供网格特征和结构元数据的编码功能
type MeshFeaturesEncoder struct {
	metaEncoder *StructuralMetadataEncoder // 复用元数据编码器
}

// NewMeshFeaturesEncoder 创建新的网格特征编码器
func NewMeshFeaturesEncoder() *MeshFeaturesEncoder {
	return &MeshFeaturesEncoder{
		metaEncoder: NewStructuralMetadataEncoder(),
	}
}

// AddMeshFeatures 添加网格特征扩展到原始图元
func (e *MeshFeaturesEncoder) AddMeshFeatures(
	primitive *gltf.Primitive,
	featureIDs []extmesh.FeatureID,
) error {
	if primitive.Extensions == nil {
		primitive.Extensions = make(gltf.Extensions)
	}

	// 创建扩展对象
	ext := extmesh.ExtMeshFeatures{
		FeatureIDs: featureIDs,
	}

	// 序列化扩展
	extData, err := json.Marshal(ext)
	if err != nil {
		return fmt.Errorf("error marshaling mesh features extension: %w", err)
	}

	// 添加到原始图元
	primitive.Extensions[extmesh.ExtensionName] = extData
	return nil
}

// AddStructuralMetadata 添加结构元数据扩展到原始图元
func (e *MeshFeaturesEncoder) AddStructuralMetadata(
	primitive *gltf.Primitive,
	propertyTextures, propertyAttributes []uint32,
) error {
	if primitive.Extensions == nil {
		primitive.Extensions = make(gltf.Extensions)
	}

	// 创建扩展对象
	ext := extmesh.ExtStructuralMetadata{
		PropertyTextures:   propertyTextures,
		PropertyAttributes: propertyAttributes,
	}

	// 序列化扩展
	extData, err := json.Marshal(ext)
	if err != nil {
		return fmt.Errorf("error marshaling structural metadata extension: %w", err)
	}

	// 添加到原始图元
	primitive.Extensions["EXT_structural_metadata"] = extData
	return nil
}

// CreateFeatureID 创建特征ID对象
func (e *MeshFeaturesEncoder) CreateFeatureID(
	featureCount uint32,
	options ...FeatureIDOption,
) extmesh.FeatureID {
	featureID := extmesh.FeatureID{
		FeatureCount: featureCount,
	}

	// 应用选项
	for _, option := range options {
		option(&featureID)
	}

	return featureID
}

// FeatureIDOption 用于配置特征ID的选项函数
type FeatureIDOption func(*extmesh.FeatureID)

// WithNullFeatureID 设置空特征ID
func WithNullFeatureID(nullID uint32) FeatureIDOption {
	return func(f *extmesh.FeatureID) {
		f.NullFeatureID = &nullID
	}
}

// WithLabel 设置特征标签
func WithLabel(label string) FeatureIDOption {
	return func(f *extmesh.FeatureID) {
		f.Label = &label
	}
}

// WithAttribute 设置属性索引
func WithAttribute(attribute uint32) FeatureIDOption {
	return func(f *extmesh.FeatureID) {
		f.Attribute = &attribute
	}
}

// WithTexture 设置特征纹理
func WithTexture(texture extmesh.FeatureIDTexture) FeatureIDOption {
	return func(f *extmesh.FeatureID) {
		f.Texture = &texture
	}
}

// WithPropertyTable 设置属性表索引
func WithPropertyTable(table uint32) FeatureIDOption {
	return func(f *extmesh.FeatureID) {
		f.PropertyTable = &table
	}
}

// CreateFeatureIDTexture 创建特征ID纹理对象
func (e *MeshFeaturesEncoder) CreateFeatureIDTexture(
	textureIndex uint32,
	options ...FeatureIDTextureOption,
) extmesh.FeatureIDTexture {
	texture := extmesh.FeatureIDTexture{
		Index:    textureIndex,
		Channels: extmesh.DefaultChannels(), // 默认通道
	}

	// 应用选项
	for _, option := range options {
		option(&texture)
	}

	return texture
}

// FeatureIDTextureOption 用于配置特征ID纹理的选项函数
type FeatureIDTextureOption func(*extmesh.FeatureIDTexture)

// WithTextureChannels 设置纹理通道
func WithTextureChannels(channels []uint32) FeatureIDTextureOption {
	return func(t *extmesh.FeatureIDTexture) {
		t.Channels = channels
	}
}

// WithTextureCoord 设置纹理坐标
func WithTextureCoord(texCoord uint32) FeatureIDTextureOption {
	return func(t *extmesh.FeatureIDTexture) {
		t.TexCoord = texCoord
	}
}

// GetMeshFeatures 从原始图元获取网格特征扩展
func (e *MeshFeaturesEncoder) GetMeshFeatures(
	primitive *gltf.Primitive,
) (*extmesh.ExtMeshFeatures, error) {
	if primitive.Extensions == nil {
		return nil, errors.New("no extensions found")
	}

	extData, exists := primitive.Extensions[extmesh.ExtensionName]
	if !exists {
		return nil, fmt.Errorf("%s extension not found", extmesh.ExtensionName)
	}
	// 添加类型断言
	extDataBytes, ok := extData.([]byte)
	if !ok {
		return nil, fmt.Errorf("extension data is not in expected format ([]byte)")
	}
	var ext extmesh.ExtMeshFeatures
	if err := json.Unmarshal(extDataBytes, &ext); err != nil {
		return nil, fmt.Errorf("error unmarshaling mesh features extension: %w", err)
	}

	return &ext, nil
}

// GetStructuralMetadata 从原始图元获取结构元数据扩展
func (e *MeshFeaturesEncoder) GetStructuralMetadata(
	primitive *gltf.Primitive,
) (*extmesh.ExtStructuralMetadata, error) {
	if primitive.Extensions == nil {
		return nil, errors.New("no extensions found")
	}

	extData, exists := primitive.Extensions["EXT_structural_metadata"]
	if !exists {
		return nil, errors.New("EXT_structural_metadata extension not found")
	}
	// 添加类型断言
	extDataBytes, ok := extData.([]byte)
	if !ok {
		return nil, fmt.Errorf("extension data is not in expected format ([]byte)")
	}
	var ext extmesh.ExtStructuralMetadata
	if err := json.Unmarshal(extDataBytes, &ext); err != nil {
		return nil, fmt.Errorf("error unmarshaling structural metadata extension: %w", err)
	}

	return &ext, nil
}

// ValidateFeatureID 验证特征ID对象是否有效
func (e *MeshFeaturesEncoder) ValidateFeatureID(featureID extmesh.FeatureID) error {
	if featureID.FeatureCount == 0 {
		return errors.New("featureCount must be greater than zero")
	}

	// 检查至少有一个引用方法被设置
	referenceMethods := 0
	if featureID.Attribute != nil {
		referenceMethods++
	}
	if featureID.Texture != nil {
		referenceMethods++
	}
	if featureID.PropertyTable != nil {
		referenceMethods++
	}

	if referenceMethods == 0 {
		return errors.New("at least one reference method (attribute, texture, or propertyTable) must be set")
	}
	if referenceMethods > 1 {
		return errors.New("only one reference method (attribute, texture, or propertyTable) can be set")
	}

	// 验证纹理配置
	if featureID.Texture != nil {
		if len(featureID.Texture.Channels) == 0 {
			return errors.New("texture channels must not be empty")
		}
		for _, channel := range featureID.Texture.Channels {
			if channel > 3 {
				return errors.New("texture channel must be between 0 and 3")
			}
		}
	}

	return nil
}

// UpdateFeatureID 更新原始图元中的特征ID
func (e *MeshFeaturesEncoder) UpdateFeatureID(
	primitive *gltf.Primitive,
	index int,
	featureID extmesh.FeatureID,
) error {
	ext, err := e.GetMeshFeatures(primitive)
	if err != nil {
		return err
	}

	// 验证索引范围
	if index < 0 || index >= len(ext.FeatureIDs) {
		return errors.New("invalid feature ID index")
	}

	// 验证新的特征ID
	if err := e.ValidateFeatureID(featureID); err != nil {
		return err
	}

	// 更新特征ID
	ext.FeatureIDs[index] = featureID

	// 重新序列化并更新扩展
	extData, err := json.Marshal(ext)
	if err != nil {
		return fmt.Errorf("error marshaling updated mesh features extension: %w", err)
	}

	primitive.Extensions[extmesh.ExtensionName] = extData
	return nil
}

// WriteFeatureData 批量写入特征数据和属性表（高级API）
func (e *MeshFeaturesEncoder) WriteFeatureData(
	doc *gltf.Document,
	class string,
	propertiesArray []map[string]interface{},
	featureIDs [][][]uint16,
) error {
	// 1. 参数校验
	if doc == nil {
		return fmt.Errorf("glTF document cannot be nil")
	}
	if len(propertiesArray) == 0 || len(featureIDs) == 0 {
		return fmt.Errorf("properties and featureIDs cannot be empty")
	}

	// 2. 创建元数据扩展
	metadata, err := e.getOrCreateMetadata(doc)
	if err != nil {
		return fmt.Errorf("metadata initialization failed: %w", err)
	}

	props := rackProps(propertiesArray)
	if err := e.updateSchema(metadata, class, props); err != nil {
		return fmt.Errorf("schema update failed: %w", err)
	}

	// 4. 创建属性表
	tableIndex, err := e.createPropertyTable(doc, metadata, class, props)
	if err != nil {
		return fmt.Errorf("property table creation failed: %w", err)
	}

	// 5. 创建FeatureID属性
	if err := e.createFeatureIDAttributes(doc, featureIDs, uint32(tableIndex)); err != nil {
		return fmt.Errorf("featureID attributes creation failed: %w", err)
	}

	// 6. 更新文档扩展
	doc.Extensions[extmesh.ExtensionName] = metadata
	return nil
}

func (e *MeshFeaturesEncoder) WriteInstanceFeatureData(doc *gltf.Document, class string, propertiesArray []map[string]interface{}) error {
	if doc == nil {
		return fmt.Errorf("glTF document cannot be nil")
	}
	if len(propertiesArray) == 0 {
		return nil
	}

	if len(doc.Meshes) == 0 {
		return fmt.Errorf("document must have mesh")
	}

	metadata, err := e.getOrCreateMetadata(doc)
	if err != nil {
		return err
	}

	properties := rackProps(propertiesArray)

	if err := e.updateSchema(metadata, class, properties); err != nil {
		return err
	}

	tableIndex, err := e.createPropertyTable(doc, metadata, class, properties)
	if err != nil {
		return err
	}

	if err := e.createInstanceFeatureIDAttributes(doc, len(propertiesArray), uint32(tableIndex)); err != nil {
		return err
	}

	doc.Extensions[extmesh.ExtensionName] = metadata
	return nil
}

func (e *MeshFeaturesEncoder) createInstanceFeatureIDAttributes(doc *gltf.Document, featureCount int, tableIndex uint32) error {
	meshFeatures := extmesh.ExtMeshFeatures{}
	if ext, exists := doc.Extensions[extmesh.ExtensionName]; exists {
		if err := unmarshalExtension(ext, &meshFeatures); err != nil {
			return err
		}
	}
	ids := make([]uint32, featureCount)
	for i := 0; i < featureCount; i++ {
		ids[i] = uint32(i)
	}

	featureID := e.CreateFeatureID(uint32(len(ids)),
		WithPropertyTable(tableIndex),
		WithAttribute(e.createAccessor(doc, ids)),
	)

	for _, mesh := range doc.Meshes {
		for _, primitive := range mesh.Primitives {
			if existing, ok := primitive.Extensions[extmesh.ExtensionName].(*extmesh.ExtMeshFeatures); ok {
				existing.FeatureIDs = append(existing.FeatureIDs, featureID)
			} else {
				primitive.Extensions[extmesh.ExtensionName] = &extmesh.ExtMeshFeatures{
					FeatureIDs: []extmesh.FeatureID{featureID},
				}
			}
		}
	}
	doc.Extensions[extmesh.ExtensionName] = meshFeatures
	return nil
}

// 私有辅助方法
func (e *MeshFeaturesEncoder) getOrCreateMetadata(doc *gltf.Document) (*extgltf.ExtStructuralMetadata, error) {
	if doc.Extensions == nil {
		doc.Extensions = make(gltf.Extensions)
	}

	extData, exists := doc.Extensions[extmesh.ExtensionName]
	if !exists {
		return &extgltf.ExtStructuralMetadata{
			Schema: &extgltf.Schema{
				Classes: make(map[string]extgltf.Class),
			},
		}, nil
	}

	// 类型断言处理
	extDataBytes, ok := extData.([]byte)
	if !ok {
		return nil, fmt.Errorf("invalid metadata format")
	}

	var metadata extgltf.ExtStructuralMetadata
	if err := json.Unmarshal(extDataBytes, &metadata); err != nil {
		return nil, err
	}
	return &metadata, nil
}

func (e *MeshFeaturesEncoder) updateSchema(metadata *extgltf.ExtStructuralMetadata, class string, sampleProps map[string]interface{}) error {
	if _, exists := metadata.Schema.Classes[class]; exists {
		return nil
	}

	props := make(map[string]extgltf.ClassProperty)
	for name, val := range sampleProps {
		propType, componentType, err := inferPropertyType(val)
		if err != nil {
			return err
		}
		props[name] = extgltf.ClassProperty{
			Type:          propType,
			ComponentType: componentType,
		}
	}
	metadata.Schema.Classes[class] = extgltf.Class{Properties: props}
	return nil
}

func (e *MeshFeaturesEncoder) createFeatureIDAttributes(doc *gltf.Document, featureIDs [][][]uint16, tableIndex uint32) error {
	for meshIdx, idss := range featureIDs {
		mesh := doc.Meshes[meshIdx]
		for primIdx, ids := range idss {
			featureID := e.CreateFeatureID(uint32(len(ids)),
				WithPropertyTable(tableIndex),
				WithAttribute(e.createAccessor(doc, ids)),
			)

			if err := e.AddMeshFeatures(mesh.Primitives[primIdx], []extmesh.FeatureID{featureID}); err != nil {
				return err
			}
		}
	}
	return nil
}

// 创建数值访问器
func (e *MeshFeaturesEncoder) createAccessor(doc *gltf.Document, data interface{}) uint32 {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, data)

	// 确保主缓冲区存在
	if len(doc.Buffers) == 0 {
		doc.Buffers = append(doc.Buffers, &gltf.Buffer{})
	}

	bufferView := &gltf.BufferView{
		Buffer:     0,
		ByteOffset: uint32(len(doc.Buffers[0].Data)),
		ByteLength: uint32(buf.Len()),
	}
	doc.BufferViews = append(doc.BufferViews, bufferView)

	doc.Buffers[0].Data = append(doc.Buffers[0].Data, buf.Bytes()...)
	doc.Buffers[0].ByteLength += uint32(buf.Len())

	pad := PaddingByte(int(doc.Buffers[0].ByteLength))
	doc.Buffers[0].Data = append(doc.Buffers[0].Data, pad...)
	doc.Buffers[0].ByteLength += uint32(len(pad))

	accessor := &gltf.Accessor{
		BufferView:    gltf.Index(uint32(len(doc.BufferViews) - 1)),
		ComponentType: gltf.ComponentShort,
		Count:         uint32(reflect.ValueOf(data).Len()),
		Type:          gltf.AccessorScalar,
	}
	doc.Accessors = append(doc.Accessors, accessor)
	return uint32(len(doc.Accessors) - 1)
}

// 修改后的createPropertyTable方法
func (e *MeshFeaturesEncoder) createPropertyTable(
	doc *gltf.Document,
	metadata *extgltf.ExtStructuralMetadata,
	class string,
	props map[string]interface{},
) (int, error) {
	propData := make([]PropertyData, 0, len(props))
	for name, values := range props {
		propType, componentType, err := inferPropertyType(values) // 添加错误处理
		if err != nil {
			return 0, fmt.Errorf("infer property type for %s: %w", name, err)
		}
		p := PropertyData{
			Name:        name,
			ElementType: propType,
			Values:      values,
		}
		if componentType != nil {
			p.ComponentType = *componentType
		}
		propData = append(propData, p)
	}

	// 复用已有实现
	tableIndex, err := e.metaEncoder.AddPropertyTable(doc, class, propData)
	if err != nil {
		return 0, fmt.Errorf("create property table via meta encoder: %w", err)
	}

	// 获取更新后的元数据扩展
	ext, err := e.metaEncoder.getOrCreateExtension(doc)
	if err != nil {
		return 0, fmt.Errorf("get updated metadata: %w", err)
	}

	// 同步到当前metadata对象
	*metadata = *ext
	return tableIndex, nil
}

// WriteFeatureData 批量写入特征数据和属性表（独立函数）
func WriteFeatureData(
	doc *gltf.Document,
	class string,
	propertiesArray []map[string]interface{},
	featureIDs [][][]uint16,
) error {
	// 创建编码器实例
	encoder := NewMeshFeaturesEncoder()

	// 复用编码器实现
	if err := encoder.WriteFeatureData(doc, class, propertiesArray, featureIDs); err != nil {
		return fmt.Errorf("写入特征数据失败: %w", err)
	}

	// 确保扩展声明
	doc.AddExtensionUsed(extmesh.ExtensionName)
	doc.AddExtensionUsed(extgltf.ExtensionName)
	return nil
}

func WriteInstanceFeatureData(doc *gltf.Document, class string, propertiesArray []map[string]interface{}) error {
	encoder := NewMeshFeaturesEncoder()
	if err := encoder.WriteInstanceFeatureData(doc, class, propertiesArray); err != nil {
		return fmt.Errorf("写入特征数据失败: %w", err)
	}

	doc.AddExtensionUsed(extmesh.ExtensionName)
	doc.AddExtensionUsed(extgltf.ExtensionName)
	return nil
}
