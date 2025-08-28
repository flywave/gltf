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

// MeshFeaturesManager 管理网格特征扩展
type MeshFeaturesManager struct {
	metaManager *StructuralMetadataManager // 使用结构元数据管理器
}

// NewMeshFeaturesManager 创建新的网格特征管理器
func NewMeshFeaturesManager() *MeshFeaturesManager {
	return &MeshFeaturesManager{
		metaManager: NewStructuralMetadataManager(),
	}
}

// AddMeshFeatures 添加网格特征扩展到原始图元
func (m *MeshFeaturesManager) AddMeshFeatures(
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
func (m *MeshFeaturesManager) AddStructuralMetadata(
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
	primitive.Extensions[extmesh.StructuralMetadataExtensionName] = extData
	return nil
}

// CreateFeatureID 创建特征ID对象
func (m *MeshFeaturesManager) CreateFeatureID(
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
func (m *MeshFeaturesManager) CreateFeatureIDTexture(
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
func (m *MeshFeaturesManager) GetMeshFeatures(
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
func (m *MeshFeaturesManager) GetStructuralMetadata(
	primitive *gltf.Primitive,
) (*extmesh.ExtStructuralMetadata, error) {
	if primitive.Extensions == nil {
		return nil, errors.New("no extensions found")
	}

	extData, exists := primitive.Extensions[extmesh.StructuralMetadataExtensionName]
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
func (m *MeshFeaturesManager) ValidateFeatureID(featureID extmesh.FeatureID) error {
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
func (m *MeshFeaturesManager) UpdateFeatureID(
	primitive *gltf.Primitive,
	index int,
	featureID extmesh.FeatureID,
) error {
	ext, err := m.GetMeshFeatures(primitive)
	if err != nil {
		return err
	}

	// 验证索引范围
	if index < 0 || index >= len(ext.FeatureIDs) {
		return errors.New("invalid feature ID index")
	}

	// 验证新的特征ID
	if err = m.ValidateFeatureID(featureID); err != nil {
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
func (m *MeshFeaturesManager) WriteFeatureData(
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

	// 2. 创建属性数据
	props := rackProps(propertiesArray)
	propData := make([]PropertyData, 0, len(props))

	for name, values := range props {
		propType, componentType, _, err := inferPropertyType(values)
		if err != nil {
			return fmt.Errorf("属性类型推断失败: %w", err)
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

	// 3. 添加属性表到文档
	tableIndex, err := m.metaManager.AddPropertyTable(doc, class, propData)
	if err != nil {
		return fmt.Errorf("属性表创建失败: %w", err)
	}

	// 4. 创建FeatureID属性
	if err := m.createFeatureIDAttributes(doc, featureIDs, uint32(tableIndex)); err != nil {
		return fmt.Errorf("featureID attributes creation failed: %w", err)
	}

	return nil
}

func (m *MeshFeaturesManager) WriteInstanceFeatureData(doc *gltf.Document, class string, propertiesArray []map[string]interface{}) error {
	if doc == nil {
		return fmt.Errorf("glTF document cannot be nil")
	}
	if len(propertiesArray) == 0 {
		return nil
	}

	if len(doc.Meshes) == 0 {
		return fmt.Errorf("document must have mesh")
	}

	// 创建属性数据
	properties := rackProps(propertiesArray)
	propData := make([]PropertyData, 0, len(properties))

	for name, values := range properties {
		propType, componentType, _, err := inferPropertyType(values)
		if err != nil {
			return fmt.Errorf("属性类型推断失败: %w", err)
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

	// 添加属性表到文档
	tableIndex, err := m.metaManager.AddPropertyTable(doc, class, propData)
	if err != nil {
		return fmt.Errorf("属性表创建失败: %w", err)
	}

	if err := m.createInstanceFeatureIDAttributes(doc, len(propertiesArray), uint32(tableIndex)); err != nil {
		return err
	}

	return nil
}

func (m *MeshFeaturesManager) createInstanceFeatureIDAttributes(doc *gltf.Document, featureCount int, tableIndex uint32) error {
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

	featureID := m.CreateFeatureID(uint32(len(ids)),
		WithPropertyTable(tableIndex),
		WithAttribute(m.createAccessor(doc, ids)),
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

func (m *MeshFeaturesManager) createFeatureIDAttributes(doc *gltf.Document, featureIDs [][][]uint16, tableIndex uint32) error {
	for meshIdx, idss := range featureIDs {
		mesh := doc.Meshes[meshIdx]
		for primIdx, ids := range idss {
			featureID := m.CreateFeatureID(uint32(len(ids)),
				WithPropertyTable(tableIndex),
				WithAttribute(m.createAccessor(doc, ids)),
			)

			if err := m.AddMeshFeatures(mesh.Primitives[primIdx], []extmesh.FeatureID{featureID}); err != nil {
				return err
			}
		}
	}
	return nil
}

// 创建数值访问器
func (m *MeshFeaturesManager) createAccessor(doc *gltf.Document, data interface{}) uint32 {
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

// WriteFeatureData 批量写入特征数据和属性表（独立函数）
func WriteFeatureData(
	doc *gltf.Document,
	class string,
	propertiesArray []map[string]interface{},
	featureIDs [][][]uint16,
) error {
	// 创建管理器实例
	manager := NewMeshFeaturesManager()

	// 复用管理器实现
	if err := manager.WriteFeatureData(doc, class, propertiesArray, featureIDs); err != nil {
		return fmt.Errorf("写入特征数据失败: %w", err)
	}

	// 确保扩展声明
	doc.AddExtensionUsed(extmesh.ExtensionName)
	doc.AddExtensionUsed(extgltf.ExtensionName)
	return nil
}

func WriteInstanceFeatureData(doc *gltf.Document, class string, propertiesArray []map[string]interface{}) error {
	manager := NewMeshFeaturesManager()
	if err := manager.WriteInstanceFeatureData(doc, class, propertiesArray); err != nil {
		return fmt.Errorf("写入特征数据失败: %w", err)
	}

	doc.AddExtensionUsed(extmesh.ExtensionName)
	doc.AddExtensionUsed(extgltf.ExtensionName)
	return nil
}
