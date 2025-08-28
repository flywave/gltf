package tile3d

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"reflect"
	"unicode/utf8"

	extgltf "github.com/flywave/gltf/ext/3dtile/gltf"

	"github.com/flywave/gltf"
)

func init() {
	gltf.RegisterExtension(extgltf.ExtensionName, UnmarshalStructuralMetadata)
}

// UnmarshalStructuralMetadata 反序列化结构元数据扩展
func UnmarshalStructuralMetadata(data []byte) (interface{}, error) {
	var ext extgltf.ExtStructuralMetadata
	if err := json.Unmarshal(data, &ext); err != nil {
		return nil, fmt.Errorf("EXT_structural_metadata解析失败: %w", err)
	}

	// 基础校验
	if ext.Schema == nil {
		return nil, errors.New("schema为必填字段")
	}
	if len(ext.PropertyTables) == 0 && len(ext.PropertyTextures) == 0 && len(ext.PropertyAttributes) == 0 {
		return nil, errors.New("至少需要一个属性表、纹理或属性")
	}

	// 校验属性表与schema的一致性
	for _, table := range ext.PropertyTables {
		if _, exists := ext.Schema.Classes[table.Class]; !exists {
			return nil, fmt.Errorf("未定义的class: %s", table.Class)
		}
	}

	return ext, nil
}

type (
	Schema            = extgltf.Schema
	Class             = extgltf.Class
	ClassProperty     = extgltf.ClassProperty
	PropertyTable     = extgltf.PropertyTable
	TableProperty     = extgltf.PropertyTableProperty
	PropertyAttribute = extgltf.PropertyAttribute
)

// PropertyData 定义属性数据
type PropertyData struct {
	Name          string
	ElementType   extgltf.ClassPropertyType
	ComponentType extgltf.ClassPropertyComponentType
	Values        interface{} // []float64, []int, []string, 等
}

// PropertyTableManager 管理属性表的创建和操作
type PropertyTableManager struct{}

// NewPropertyTableManager 创建新的属性表管理器
func NewPropertyTableManager() *PropertyTableManager {
	return &PropertyTableManager{}
}

// createPropertyTable 创建属性表
func (m *PropertyTableManager) createPropertyTable(
	doc *gltf.Document,
	classID string,
	properties []PropertyData,
	schema *Schema,
) (*PropertyTable, error) {
	// 验证schema有效性
	if schema == nil {
		return nil, errors.New("schema is required")
	}
	schemaClass, exists := schema.Classes[classID]
	if !exists {
		return nil, fmt.Errorf("class %s not found in schema", classID)
	}

	// 验证属性
	if len(properties) == 0 {
		return nil, errors.New("no properties provided")
	}

	// 检查所有属性是否在schema中定义
	for _, prop := range properties {
		if _, exists := schemaClass.Properties[prop.Name]; !exists {
			return nil, fmt.Errorf("property %s not defined in class %s", prop.Name, classID)
		}
	}

	// 确定行数
	rowCount := m.getValueCount(properties[0].Values)
	for _, prop := range properties {
		if m.getValueCount(prop.Values) != rowCount {
			return nil, errors.New("all properties must have the same number of values")
		}
	}

	// 创建属性表
	table := &PropertyTable{
		Class:      classID,
		Count:      uint32(rowCount),
		Properties: make(map[string]extgltf.PropertyTableProperty),
	}

	// 添加属性到表
	for _, prop := range properties {
		tableProp, err := m.encodeProperty(doc, prop)
		if err != nil {
			return nil, fmt.Errorf("error encoding property %s: %w", prop.Name, err)
		}
		table.Properties[prop.Name] = *tableProp
	}

	return table, nil
}

// getValueCount 获取值数量
func (m *PropertyTableManager) getValueCount(values interface{}) int {
	switch v := values.(type) {
	case []string:
		return len(v)
	case []float64:
		return len(v)
	case []float32:
		return len(v)
	case []int:
		return len(v)
	case []int8:
		return len(v)
	case []int16:
		return len(v)
	case []int32:
		return len(v)
	case []int64:
		return len(v)
	case []uint:
		return len(v)
	case []uint8:
		return len(v)
	case []uint16:
		return len(v)
	case []uint32:
		return len(v)
	case []uint64:
		return len(v)
	case []bool:
		return len(v)
	case [][]string:
		return len(v)
	case [][]float64:
		return len(v)
	case [][]float32:
		return len(v)
	case [][]int:
		return len(v)
	case [][]int8:
		return len(v)
	case [][]int16:
		return len(v)
	case [][]int32:
		return len(v)
	case [][]int64:
		return len(v)
	case [][]uint:
		return len(v)
	case [][]uint8:
		return len(v)
	case [][]uint16:
		return len(v)
	case [][]uint32:
		return len(v)
	case [][]uint64:
		return len(v)
	case [][]bool:
		return len(v)
	default:
		return 0
	}
}

// encodeProperty 编码单个属性
func (m *PropertyTableManager) encodeProperty(
	doc *gltf.Document,
	prop PropertyData,
) (*extgltf.PropertyTableProperty, error) {
	tableProp := &extgltf.PropertyTableProperty{}

	switch values := reflect.ValueOf(prop.Values).Interface().(type) {
	case []string:
		// 处理字符串数组
		data, offsets, err := m.encodeStringArray(values)
		if err != nil {
			return nil, err
		}

		// 添加值缓冲区视图
		valuesIndex, err := m.addBufferView(doc, data)
		if err != nil {
			return nil, err
		}
		tableProp.Values = uint32(valuesIndex)

		// 添加偏移量缓冲区视图
		offsetsIndex, err := m.addBufferView(doc, offsets)
		if err != nil {
			return nil, err
		}
		tableProp.StringOffsets = new(uint32)
		*tableProp.StringOffsets = uint32(offsetsIndex)
		tableProp.StringOffsetType = extgltf.OffsetTypeUint32
	case [][]string:
		// 处理字符串数组
		data, innerOffsets, outerOffset, err := m.encodeStringMatrix(values)
		if err != nil {
			return nil, err
		}

		// 添加值缓冲区视图
		valuesIndex, err := m.addBufferView(doc, data)
		if err != nil {
			return nil, err
		}
		tableProp.Values = uint32(valuesIndex)

		// 添加偏移量缓冲区视图
		innerOffsetsIndex, err := m.addBufferView(doc, innerOffsets)
		if err != nil {
			return nil, err
		}
		tableProp.StringOffsets = new(uint32)
		*tableProp.StringOffsets = uint32(innerOffsetsIndex)
		tableProp.StringOffsetType = extgltf.OffsetTypeUint32

		outOffsetsIndex, err := m.addBufferView(doc, outerOffset)
		if err != nil {
			return nil, err
		}

		tableProp.ArrayOffsets = new(uint32)
		*tableProp.ArrayOffsets = uint32(outOffsetsIndex)
		tableProp.ArrayOffsetType = extgltf.OffsetTypeUint32
	case []bool:
		// 处理布尔数组
		data := m.encodeBoolArray(values)
		index, err := m.addBufferView(doc, data)
		if err != nil {
			return nil, err
		}
		tableProp.Values = uint32(index)

	default:
		// 处理数值数组
		data, err := m.encodeNumericArray(prop.Values, prop.ComponentType)
		if err != nil {
			return nil, err
		}
		index, err := m.addBufferView(doc, data)
		if err != nil {
			return nil, err
		}
		tableProp.Values = uint32(index)
	}

	return tableProp, nil
}

// encodeStringArray 编码字符串数组
func (m *PropertyTableManager) encodeStringArray(values []string) ([]byte, []byte, error) {
	// 计算总字节数
	totalBytes := 0
	for _, s := range values {
		if !utf8.ValidString(s) {
			return nil, nil, errors.New("invalid UTF-8 string")
		}
		totalBytes += len(s)
	}

	// 创建字符串数据缓冲区
	dataBuffer := make([]byte, 0, totalBytes)
	offsets := make([]uint32, len(values)+1)

	// 填充数据
	offset := 0
	for i, s := range values {
		dataBuffer = append(dataBuffer, []byte(s)...)
		offsets[i] = uint32(offset)
		offset += len(s)
	}
	offsets[len(values)] = uint32(offset) // 最终偏移量

	// 将偏移量转换为字节
	offsetBuffer := make([]byte, len(offsets)*4)
	for i, off := range offsets {
		binary.LittleEndian.PutUint32(offsetBuffer[i*4:], off)
	}

	return dataBuffer, offsetBuffer, nil
}

func (m *PropertyTableManager) encodeStringMatrix(matrix [][]string) ([]byte, []byte, []byte, error) {
	// 1. 验证所有字符串并计算总大小
	totalStringsSize := 0
	totalInnerOffsets := 0
	for _, arr := range matrix {
		for _, s := range arr {
			if !utf8.ValidString(s) {
				return nil, nil, nil, fmt.Errorf("invalid UTF-8 string: %q", s)
			}
			totalStringsSize += len(s)
		}
		totalInnerOffsets += len(arr) + 1 // 每个内层数组需要 len(arr)+1 个偏移量
	}

	// 2. 分配缓冲区
	dataBuffer := make([]byte, totalStringsSize)
	innerOffsetsBuffer := make([]byte, totalInnerOffsets*4)
	outerOffsetsBuffer := make([]byte, (len(matrix)+1)*4)

	// 3. 填充数据
	stringPos := 0
	innerOffsetPos := 0
	outerOffsetPos := 0

	for _, arr := range matrix {
		// 记录外层偏移（指向内层偏移量数组）
		binary.LittleEndian.PutUint32(outerOffsetsBuffer[outerOffsetPos:], uint32(innerOffsetPos/4))
		outerOffsetPos += 4

		// 处理内层数组
		currentStringPos := stringPos

		for _, s := range arr {
			// 记录内层偏移
			binary.LittleEndian.PutUint32(innerOffsetsBuffer[innerOffsetPos:], uint32(currentStringPos))
			innerOffsetPos += 4

			// 复制字符串数据
			copy(dataBuffer[currentStringPos:], s)
			currentStringPos += len(s)
		}

		// 内层数组结束标记
		binary.LittleEndian.PutUint32(innerOffsetsBuffer[innerOffsetPos:], uint32(currentStringPos))
		innerOffsetPos += 4
		stringPos = currentStringPos
	}

	// 外层数组结束标记
	binary.LittleEndian.PutUint32(outerOffsetsBuffer[outerOffsetPos:], uint32(innerOffsetPos/4))

	return dataBuffer, innerOffsetsBuffer, outerOffsetsBuffer, nil
}

// encodeBoolArray 编码布尔数组
func (m *PropertyTableManager) encodeBoolArray(values []bool) []byte {
	// 每字节存储8个布尔值
	byteCount := (len(values) + 7) / 8
	data := make([]byte, byteCount)

	for i, val := range values {
		if val {
			byteIndex := i / 8
			bitIndex := uint(i % 8)
			data[byteIndex] |= 1 << bitIndex
		}
	}

	return data
}

// encodeNumericArray 编码数值数组
func (m *PropertyTableManager) encodeNumericArray(
	values interface{},
	componentType extgltf.ClassPropertyComponentType,
) ([]byte, error) {
	var buf bytes.Buffer

	switch v := values.(type) {
	case []float32:
		for _, val := range v {
			if err := binary.Write(&buf, binary.LittleEndian, val); err != nil {
				return nil, err
			}
		}
	case []float64:
		for _, val := range v {
			// 将float64转换为float32
			f32 := float32(val)
			if math.IsNaN(float64(f32)) || math.IsInf(float64(f32), 0) {
				return nil, errors.New("invalid float value")
			}
			if err := binary.Write(&buf, binary.LittleEndian, f32); err != nil {
				return nil, err
			}
		}
	case []int:
		for _, val := range v {
			if err := m.encodeInteger(&buf, int64(val), componentType); err != nil {
				return nil, err
			}
		}
	case []int8:
		for _, val := range v {
			if err := binary.Write(&buf, binary.LittleEndian, val); err != nil {
				return nil, err
			}
		}
	case []int16:
		for _, val := range v {
			if err := binary.Write(&buf, binary.LittleEndian, val); err != nil {
				return nil, err
			}
		}
	case []int32:
		for _, val := range v {
			if err := binary.Write(&buf, binary.LittleEndian, val); err != nil {
				return nil, err
			}
		}
	case []int64:
		for _, val := range v {
			if err := m.encodeInteger(&buf, val, componentType); err != nil {
				return nil, err
			}
		}
	case []uint:
		for _, val := range v {
			if err := m.encodeInteger(&buf, int64(val), componentType); err != nil {
				return nil, err
			}
		}
	case []uint8:
		for _, val := range v {
			if err := binary.Write(&buf, binary.LittleEndian, val); err != nil {
				return nil, err
			}
		}
	case []uint16:
		for _, val := range v {
			if err := binary.Write(&buf, binary.LittleEndian, val); err != nil {
				return nil, err
			}
		}
	case []uint32:
		for _, val := range v {
			if err := binary.Write(&buf, binary.LittleEndian, val); err != nil {
				return nil, err
			}
		}
	case []uint64:
		for _, val := range v {
			if err := m.encodeInteger(&buf, int64(val), componentType); err != nil {
				return nil, err
			}
		}
	default:
		return nil, errors.New("unsupported numeric type")
	}

	return buf.Bytes(), nil
}

// encodeInteger 编码整数值
func (m *PropertyTableManager) encodeInteger(
	buf *bytes.Buffer,
	val int64,
	componentType extgltf.ClassPropertyComponentType,
) error {
	switch componentType {
	case extgltf.ClassPropertyComponentTypeInt8:
		return binary.Write(buf, binary.LittleEndian, int8(val))
	case extgltf.ClassPropertyComponentTypeUint8:
		return binary.Write(buf, binary.LittleEndian, uint8(val))
	case extgltf.ClassPropertyComponentTypeInt16:
		return binary.Write(buf, binary.LittleEndian, int16(val))
	case extgltf.ClassPropertyComponentTypeUint16:
		return binary.Write(buf, binary.LittleEndian, uint16(val))
	case extgltf.ClassPropertyComponentTypeInt32:
		return binary.Write(buf, binary.LittleEndian, int32(val))
	case extgltf.ClassPropertyComponentTypeUint32:
		return binary.Write(buf, binary.LittleEndian, uint32(val))
	case extgltf.ClassPropertyComponentTypeInt64:
		return binary.Write(buf, binary.LittleEndian, val)
	case extgltf.ClassPropertyComponentTypeUint64:
		return binary.Write(buf, binary.LittleEndian, uint64(val))
	default:
		return binary.Write(buf, binary.LittleEndian, int32(val))
	}
}

// addBufferView 添加缓冲区视图到GLTF文档
func (m *PropertyTableManager) addBufferView(doc *gltf.Document, data []byte) (int, error) {
	// 确保缓冲区存在
	if len(doc.Buffers) == 0 {
		doc.Buffers = append(doc.Buffers, &gltf.Buffer{})
	}

	// 使用第一个缓冲区
	buffer := doc.Buffers[0]
	buffer.Data = append(buffer.Data, data...)
	buffer.ByteLength += uint32(len(data))

	pad := PaddingByte(int(buffer.ByteLength))
	buffer.Data = append(buffer.Data, pad...)
	buffer.ByteLength += uint32(len(pad))

	// 创建缓冲区视图
	view := &gltf.BufferView{
		Buffer:     0,
		ByteOffset: buffer.ByteLength - uint32(len(data)),
		ByteLength: uint32(len(data)),
	}

	// 添加到文档
	doc.BufferViews = append(doc.BufferViews, view)
	return len(doc.BufferViews) - 1, nil
}

// DecodeProperty 解码属性数据
func (m *PropertyTableManager) DecodeProperty(
	doc *gltf.Document,
	tableIndex int,
	propertyName string,
) (interface{}, error) {
	// 获取扩展
	extData, exists := doc.Extensions[extgltf.ExtensionName]
	if !exists {
		return nil, errors.New("extension not found")
	}
	// 添加类型断言
	extDataBytes, ok := extData.([]byte)
	if !ok {
		return nil, fmt.Errorf("extension data is not in expected format ([]byte)")
	}
	var ext extgltf.ExtStructuralMetadata
	if err := json.Unmarshal(extDataBytes, &ext); err != nil {
		return nil, err
	}

	// 检查表索引
	if tableIndex < 0 || tableIndex >= len(ext.PropertyTables) {
		return nil, errors.New("invalid table index")
	}

	table := ext.PropertyTables[tableIndex]
	prop, exists := table.Properties[propertyName]
	if !exists {
		return nil, errors.New("property not found")
	}

	// 获取值缓冲区视图
	valuesView := doc.BufferViews[prop.Values]
	valuesBuffer := doc.Buffers[valuesView.Buffer].Data
	valuesData := valuesBuffer[valuesView.ByteOffset : valuesView.ByteOffset+uint32(valuesView.ByteLength)]

	// 新增：从schema获取class属性
	schemaClass := ext.Schema.Classes[table.Class]
	classProperty, exists := schemaClass.Properties[propertyName]
	if !exists {
		return nil, fmt.Errorf("property %s not defined in class %s", propertyName, table.Class)
	}

	// 根据类型解码
	switch {
	case prop.StringOffsets != nil:
		// 字符串类型
		return m.decodeStringProperty(doc, prop, valuesData)
	default:
		// 数值或布尔类型
		return m.decodeNumericProperty(doc, classProperty, valuesData)
	}
}

// decodeStringProperty 解码字符串属性
func (m *PropertyTableManager) decodeStringProperty(
	doc *gltf.Document,
	prop extgltf.PropertyTableProperty,
	valuesData []byte,
) ([]string, error) {
	// 获取偏移量缓冲区视图
	offsetsView := doc.BufferViews[*prop.StringOffsets]
	offsetsBuffer := doc.Buffers[offsetsView.Buffer].Data
	offsetsData := offsetsBuffer[offsetsView.ByteOffset : offsetsView.ByteOffset+uint32(offsetsView.ByteLength)]

	// 读取偏移量
	offsetCount := len(offsetsData) / 4
	offsets := make([]uint32, offsetCount)
	if err := binary.Read(bytes.NewReader(offsetsData), binary.LittleEndian, &offsets); err != nil {
		return nil, err
	}

	// 提取字符串
	strings := make([]string, offsetCount-1)
	for i := 0; i < len(strings); i++ {
		start := offsets[i]
		end := offsets[i+1]
		strBytes := valuesData[start:end]
		strings[i] = string(strBytes)
	}

	return strings, nil
}

// decodeNumericProperty 解码数值属性
func (m *PropertyTableManager) decodeNumericProperty(
	_ *gltf.Document,
	prop extgltf.ClassProperty,
	valuesData []byte,
) (interface{}, error) {
	if prop.ComponentType == nil {
		return nil, errors.New("component type is nil")
	}
	// 根据组件类型确定解码方式
	switch *prop.ComponentType {
	case extgltf.ClassPropertyComponentTypeFloat32:
		return m.decodeFloat32Array(valuesData)
	case extgltf.ClassPropertyComponentTypeInt32:
		return m.decodeInt32Array(valuesData)
	case extgltf.ClassPropertyComponentTypeUint32:
		return m.decodeUint32Array(valuesData)
	case extgltf.ClassPropertyComponentTypeInt8:
		return m.decodeInt8Array(valuesData)
	default:
		return nil, fmt.Errorf("unsupported component type: %s", *prop.ComponentType)
	}
}

// decodeFloat32Array 解码float32数组
func (m *PropertyTableManager) decodeFloat32Array(data []byte) ([]float32, error) {
	count := len(data) / 4
	result := make([]float32, count)
	if err := binary.Read(bytes.NewReader(data), binary.LittleEndian, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// decodeInt32Array 解码int32数组
func (m *PropertyTableManager) decodeInt32Array(data []byte) ([]int32, error) {
	count := len(data) / 4
	result := make([]int32, count)
	if err := binary.Read(bytes.NewReader(data), binary.LittleEndian, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// decodeUint32Array 解码uint32数组
func (m *PropertyTableManager) decodeUint32Array(data []byte) ([]uint32, error) {
	count := len(data) / 4
	result := make([]uint32, count)
	if err := binary.Read(bytes.NewReader(data), binary.LittleEndian, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// decodeInt8Array 解码int8数组
func (m *PropertyTableManager) decodeInt8Array(data []byte) ([]int8, error) {
	count := len(data)
	result := make([]int8, count)
	if err := binary.Read(bytes.NewReader(data), binary.LittleEndian, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// StructuralMetadataManager 管理结构元数据扩展
type StructuralMetadataManager struct {
	propertyTableManager *PropertyTableManager
}

// NewStructuralMetadataManager 创建新的元数据管理器
func NewStructuralMetadataManager() *StructuralMetadataManager {
	return &StructuralMetadataManager{
		propertyTableManager: NewPropertyTableManager(),
	}
}

// getOrCreateExtension 获取或创建结构元数据扩展
func (m *StructuralMetadataManager) getOrCreateExtension(doc *gltf.Document) (*extgltf.ExtStructuralMetadata, error) {
	if doc.Extensions == nil {
		doc.Extensions = make(gltf.Extensions)
	}

	if extData, exists := doc.Extensions[extgltf.ExtensionName]; exists {
		extDataBytes, ok := extData.([]byte)
		if !ok {
			return nil, fmt.Errorf("extension data is not in expected format ([]byte)")
		}

		var ext extgltf.ExtStructuralMetadata
		if err := json.Unmarshal(extDataBytes, &ext); err != nil {
			return nil, fmt.Errorf("error unmarshaling existing extension: %w", err)
		}
		return &ext, nil
	}

	// 创建新扩展
	ext := &extgltf.ExtStructuralMetadata{
		Schema: &Schema{
			ID:      "default_schema",
			Classes: make(map[string]Class),
		},
	}
	return ext, nil
}

// createSchema 创建或更新元数据结构
func (m *StructuralMetadataManager) createSchema(
	classID string,
	properties []PropertyData,
	existingSchema *Schema,
) *Schema {
	schema := existingSchema
	if schema == nil {
		schema = &Schema{
			ID:      "default_schema",
			Classes: make(map[string]Class),
		}
	}

	// 获取或创建类
	class, exists := schema.Classes[classID]
	if !exists {
		class = Class{
			Properties: make(map[string]ClassProperty),
		}
	}

	// 添加属性到类
	for _, prop := range properties {
		classProperty := ClassProperty{
			Type:          prop.ElementType,
			ComponentType: &prop.ComponentType,
		}
		class.Properties[prop.Name] = classProperty
	}

	// 更新schema
	schema.Classes[classID] = class
	return schema
}

// AddPropertyTable 添加属性表到GLTF文档
func (m *StructuralMetadataManager) AddPropertyTable(
	doc *gltf.Document,
	classID string,
	properties []PropertyData,
) (int, error) {
	// 获取或创建扩展
	ext, err := m.getOrCreateExtension(doc)
	if err != nil {
		return -1, err
	}

	// 创建或更新schema
	ext.Schema = m.createSchema(classID, properties, ext.Schema)

	// 创建属性表
	table, err := m.propertyTableManager.createPropertyTable(doc, classID, properties, ext.Schema)
	if err != nil {
		return -1, err
	}

	// 添加属性表
	if ext.PropertyTables == nil {
		ext.PropertyTables = make([]PropertyTable, 0)
	}
	ext.PropertyTables = append(ext.PropertyTables, *table)

	// 保存扩展
	if err := m.SaveExtension(doc, ext); err != nil {
		return -1, err
	}

	return len(ext.PropertyTables) - 1, nil
}

// DecodeProperty 解码属性数据
func (m *StructuralMetadataManager) DecodeProperty(
	doc *gltf.Document,
	tableIndex int,
	propertyName string,
) (interface{}, error) {
	return m.propertyTableManager.DecodeProperty(doc, tableIndex, propertyName)
}

// SaveExtension 保存扩展回文档
func (m *StructuralMetadataManager) SaveExtension(doc *gltf.Document, ext *extgltf.ExtStructuralMetadata) error {
	extData, err := json.Marshal(ext)
	if err != nil {
		return err
	}

	if doc.Extensions == nil {
		doc.Extensions = make(gltf.Extensions)
	}
	doc.Extensions[extgltf.ExtensionName] = extData
	doc.AddExtensionUsed(extgltf.ExtensionName)
	return nil
}

// WriteStructuralMetadata 保存结构元数据到文档
func WriteStructuralMetadata(doc *gltf.Document, class string, propertiesArray []map[string]interface{}) error {
	manager := NewStructuralMetadataManager()

	// 转换旧格式参数为新格式
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

	// 添加属性表
	_, err := manager.AddPropertyTable(doc, class, propData)
	if err != nil {
		return fmt.Errorf("创建属性表失败: %w", err)
	}

	return nil
}
