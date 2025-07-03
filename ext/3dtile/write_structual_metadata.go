package tile3d

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"reflect"
	"unsafe"

	"github.com/flywave/gltf"
	ext_gltf "github.com/flywave/gltf/ext/3dtile/gltf"
	"github.com/flywave/go3d/mat3"
	"github.com/flywave/go3d/mat4"

	mat3d "github.com/flywave/go3d/float64/mat3"
	mat4d "github.com/flywave/go3d/float64/mat4"
)

func WriteStructuralMetadata(doc *gltf.Document, class string, propertiesArray []map[string]interface{}) error {
	if doc == nil {
		return fmt.Errorf("GLTF document cannot be nil")
	}
	if len(propertiesArray) == 0 {
		return nil
	}
	properties := rackProps(propertiesArray)
	if doc.Extensions == nil {
		doc.Extensions = make(map[string]interface{})
	}
	var metadata ext_gltf.ExtStructuralMetadata

	if metadata.Schema == nil {
		metadata.Schema = &ext_gltf.Schema{
			Classes: make(map[string]ext_gltf.Class),
		}
	}
	if _, exists := metadata.Schema.Classes[class]; !exists {
		metadata.Schema.Classes[class] = ext_gltf.Class{
			Properties: make(map[string]ext_gltf.ClassProperty),
		}
	}

	propTable := ext_gltf.PropertyTable{
		Class:      class,
		Properties: make(map[string]ext_gltf.PropertyTableProperty),
	}

	count := -1
	for propName, values := range properties {
		propType, componentType, err := inferPropertyType(values)
		if err != nil {
			return fmt.Errorf("failed to infer the type of property %s: %w", propName, err)
		}
		metadata.Schema.Classes[class].Properties[propName] = ext_gltf.ClassProperty{
			Type:          propType,
			ComponentType: componentType,
		}

		if count == -1 {
			count = len(values)
		} else {
			if count != len(values) {
				return fmt.Errorf("property %s has different count with other properties", propName)
			}
		}
		accessor, err := createPropertyAccessor(doc, values)
		if err != nil {
			// 修改错误信息首字母为小写
			return fmt.Errorf("failed to create the accessor for property %s: %w", propName, err)
		}
		propTable.Properties[propName] = *accessor
	}

	// 5. Set the property table count
	propTable.Count = uint32(count)
	metadata.PropertyTables = append(metadata.PropertyTables, propTable)
	doc.Extensions[ext_gltf.ExtensionName] = metadata
	addExtensionUsed(doc, ext_gltf.ExtensionName)

	return nil
}

// inferPropertyType infers the property type
func inferPropertyType(values interface{}) (ext_gltf.ClassPropertyType, *ext_gltf.ClassPropertyComponentType, error) {
	switch v := reflect.ValueOf(values).Index(0).Interface().(type) {
	case string:
		return ext_gltf.ClassPropertyTypeString, nil, nil
	case bool:
		return ext_gltf.ClassPropertyTypeBoolean, nil, nil
	case float32:
		return ext_gltf.ClassPropertyTypeScalar, ptr(ext_gltf.ClassPropertyComponentTypeFloat32), nil
	case float64:
		return ext_gltf.ClassPropertyTypeScalar, ptr(ext_gltf.ClassPropertyComponentTypeFloat64), nil
	case int:
		return ext_gltf.ClassPropertyTypeScalar, ptr(ext_gltf.ClassPropertyComponentTypeInt64), nil
	case int8:
		return ext_gltf.ClassPropertyTypeScalar, ptr(ext_gltf.ClassPropertyComponentTypeInt8), nil
	case int16:
		return ext_gltf.ClassPropertyTypeScalar, ptr(ext_gltf.ClassPropertyComponentTypeInt16), nil
	case int32:
		return ext_gltf.ClassPropertyTypeScalar, ptr(ext_gltf.ClassPropertyComponentTypeInt32), nil
	case int64:
		return ext_gltf.ClassPropertyTypeScalar, ptr(ext_gltf.ClassPropertyComponentTypeInt64), nil
	case uint:
		return ext_gltf.ClassPropertyTypeScalar, ptr(ext_gltf.ClassPropertyComponentTypeUint64), nil
	case uint8:
		return ext_gltf.ClassPropertyTypeScalar, ptr(ext_gltf.ClassPropertyComponentTypeUint8), nil
	case uint16:
		return ext_gltf.ClassPropertyTypeScalar, ptr(ext_gltf.ClassPropertyComponentTypeUint16), nil
	case uint32:
		return ext_gltf.ClassPropertyTypeScalar, ptr(ext_gltf.ClassPropertyComponentTypeUint32), nil
	case uint64:
		return ext_gltf.ClassPropertyTypeScalar, ptr(ext_gltf.ClassPropertyComponentTypeUint64), nil
	case []float32:
		switch len(v) {
		case 2:
			return ext_gltf.ClassPropertyTypeVec2, ptr(ext_gltf.ClassPropertyComponentTypeFloat32), nil
		case 3:
			return ext_gltf.ClassPropertyTypeVec3, ptr(ext_gltf.ClassPropertyComponentTypeFloat32), nil
		case 4:
			return ext_gltf.ClassPropertyTypeVec4, ptr(ext_gltf.ClassPropertyComponentTypeFloat32), nil
		}
	case mat3.T:
		return ext_gltf.ClassPropertyTypeMat3, ptr(ext_gltf.ClassPropertyComponentTypeFloat32), nil
	case mat4.T:
		return ext_gltf.ClassPropertyTypeMat4, ptr(ext_gltf.ClassPropertyComponentTypeFloat32), nil
	case mat3d.T:
		return ext_gltf.ClassPropertyTypeMat3, ptr(ext_gltf.ClassPropertyComponentTypeFloat64), nil
	case mat4d.T:
		return ext_gltf.ClassPropertyTypeMat4, ptr(ext_gltf.ClassPropertyComponentTypeFloat64), nil
	case []float64:
		switch len(v) {
		case 2:
			return ext_gltf.ClassPropertyTypeVec2, ptr(ext_gltf.ClassPropertyComponentTypeFloat64), nil
		case 3:
			return ext_gltf.ClassPropertyTypeVec3, ptr(ext_gltf.ClassPropertyComponentTypeFloat64), nil
		case 4:
			return ext_gltf.ClassPropertyTypeVec4, ptr(ext_gltf.ClassPropertyComponentTypeFloat64), nil
		}
	default:
		return "", nil, fmt.Errorf("usupported type: %T", v)
	}
	return "", nil, fmt.Errorf("unable to infer the type")
}

// createPropertyAccessor creates a property accessor
func createPropertyAccessor(doc *gltf.Document, values interface{}) (*ext_gltf.PropertyTableProperty, error) {
	switch v := values.(type) {
	case string, bool, float32, float64, int32, uint32, int, int8, int16, int64, uint, uint8, uint16, uint64:
		return CreateInlinePropertyTableProperty(values)
	case []string:
		return createStringAccessor(doc, v)
	case []float32:
		return createFloatAccessor(doc, v)
	case [][]float32:
		return createVectorAccessor(doc, v)
	case []float64:
		return createFloatAccessor(doc, v)
	case [][]float64:
		return createVectorAccessor(doc, v)
	case []int:
		return createFloatAccessor(doc, v)
	case [][]int:
		return createVectorAccessor(doc, v)
	case []int8:
		return createFloatAccessor(doc, v)
	case [][]int8:
		return createVectorAccessor(doc, v)
	case []uint8:
		return createFloatAccessor(doc, v)
	case [][]uint8:
		return createVectorAccessor(doc, v)
	case []int16:
		return createFloatAccessor(doc, v)
	case [][]int16:
		return createVectorAccessor(doc, v)
	case []uint16:
		return createFloatAccessor(doc, v)
	case [][]uint16:
		return createVectorAccessor(doc, v)
	case []uint32:
		return createFloatAccessor(doc, v)
	case [][]uint32:
		return createVectorAccessor(doc, v)
	case []int32:
		return createFloatAccessor(doc, v)
	case [][]int32:
		return createVectorAccessor(doc, v)
	case []uint64:
		return createFloatAccessor(doc, v)
	case [][]uint64:
		return createVectorAccessor(doc, v)
	case []int64:
		return createFloatAccessor(doc, v)
	case [][]int64:
		return createVectorAccessor(doc, v)
	default:
		return nil, fmt.Errorf("unimplemented type handling: %T", values)
	}
}

func CreateInlinePropertyTableProperty(value interface{}) (*ext_gltf.PropertyTableProperty, error) {
	if !isAllowedInlineType(value) {
		return nil, fmt.Errorf("type %T is not allowed to inline", value)
	}

	prop := &ext_gltf.PropertyTableProperty{
		Extras: buildInlineValueExtra(value),
	}
	switch value.(type) {
	case string, bool:
		break
	default:
		prop.Max = mustMarshal(value)
		prop.Min = mustMarshal(value)
	}

	return prop, nil
}

// 允许内联的类型
func isAllowedInlineType(v interface{}) bool {
	switch v.(type) {
	case string, bool, float32, float64, int, int32, uint32:
		return true
	}
	return false
}

// 构建存储实际值的Extras字段
func buildInlineValueExtra(value interface{}) json.RawMessage {
	return mustMarshal(map[string]interface{}{
		"_inlineValue": value,
		"offset":       0,
	})
}

// createStringAccessor creates a string property accessor
func createStringAccessor(doc *gltf.Document, values []string) (*ext_gltf.PropertyTableProperty, error) {
	if len(values) == 0 {
		return nil, fmt.Errorf("string array cannot be empty")
	}

	stringTable := make([]string, 0)
	indices := make([]uint16, len(values))
	indexMap := make(map[string]uint16)
	stringData := bytes.Buffer{}

	for i, s := range values {
		if idx, exists := indexMap[s]; exists {
			indices[i] = idx
			continue
		}

		idx := uint16(len(stringTable))
		stringTable = append(stringTable, s)
		indexMap[s] = idx
		indices[i] = idx

		stringData.WriteString(s)
		stringData.WriteByte(0)
	}

	// 2. 确保主Buffer存在
	if len(doc.Buffers) == 0 {
		doc.Buffers = append(doc.Buffers, &gltf.Buffer{
			ByteLength: 0,
			Data:       make([]byte, 0),
		})
	}

	indicesByteLength := len(indices) * 2
	indicesBufView := gltf.BufferView{
		Buffer:     0, // 主Buffer
		ByteOffset: doc.Buffers[0].ByteLength,
		ByteLength: uint32(indicesByteLength),
		Target:     gltf.TargetArrayBuffer,
	}
	indicesBufViewIndex := uint32(len(doc.BufferViews))
	doc.BufferViews = append(doc.BufferViews, &indicesBufView)

	// 4. 写入索引数据
	indicesBuf := bytes.NewBuffer(make([]byte, 0, indicesByteLength))
	binary.Write(indicesBuf, binary.LittleEndian, indices)
	doc.Buffers[0].Data = append(doc.Buffers[0].Data, indicesBuf.Bytes()...)
	doc.Buffers[0].ByteLength += uint32(indicesByteLength)
	padBuffer(doc)

	// 8. 创建属性表属性
	return &ext_gltf.PropertyTableProperty{
		Values: indicesBufViewIndex,
		Extensions: map[string]json.RawMessage{
			"3DTILES_property_string": mustMarshal(map[string]interface{}{
				"strings": stringTable,
			}),
		},
	}, nil
}

func createFloatAccessor[T float32 | float64 | int | int8 | uint8 | int16 | uint16 | int32 | uint32 | uint64 | int64](doc *gltf.Document, values []T) (*ext_gltf.PropertyTableProperty, error) {
	if len(values) == 0 {
		return nil, errors.New("empty values array")
	}

	minVal, maxVal := calculateFloatRange(values)
	byteLength := len(values) * int(unsafe.Sizeof(values[0]))
	data := make([]byte, 0, byteLength)
	buf := bytes.NewBuffer(data)
	buf.Grow(byteLength)
	for _, v := range values {
		binary.Write(buf, binary.LittleEndian, v)
	}
	data = buf.Bytes()

	// 3. Update buffer view and main buffer
	bufViewIndex := uint32(len(doc.BufferViews))
	doc.BufferViews = append(doc.BufferViews, &gltf.BufferView{
		Buffer:     0, // Main Buffer
		ByteOffset: doc.Buffers[0].ByteLength,
		ByteLength: uint32(byteLength),
		Target:     gltf.TargetArrayBuffer,
	})

	// 4. Append data with padding calculation
	doc.Buffers[0].Data = append(doc.Buffers[0].Data, data...)
	doc.Buffers[0].ByteLength += uint32(byteLength)
	padBuffer(doc)

	// 6. Return property table property
	return &ext_gltf.PropertyTableProperty{
		Values: bufViewIndex,
		Min:    mustMarshal(minVal),
		Max:    mustMarshal(maxVal),
	}, nil
}

// calculateFloatRange calculates the minimum and maximum values of a float array
func calculateFloatRange[T float32 | float64 | int | int8 | uint8 | int16 | uint16 | int32 | uint32 | uint64 | int64](values []T) (min, max T) {
	if len(values) == 0 {
		return 0, 0
	}
	min, max = values[0], values[0]
	for _, v := range values[1:] {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}
	return
}

func createVectorAccessor[T float32 | float64 | int | int8 | uint8 | int16 | uint16 | int32 | uint32 | uint64 | int64](doc *gltf.Document, vectors [][]T) (*ext_gltf.PropertyTableProperty, error) {
	if len(vectors) == 0 {
		return nil, fmt.Errorf("vector array cannot be empty")
	}

	// 1. Determine the vector dimension and component type
	dim := len(vectors[0])

	// 2. Check the consistency of all vector dimensions
	for i, vec := range vectors {
		if len(vec) != dim {
			return nil, fmt.Errorf("the dimension of vector %d does not match. Expected %d, actual %d", i, dim, len(vec))
		}
	}

	// 3. Calculate the minimum and maximum values
	minVals, maxVals := calculateVectorBounds(vectors)

	// 4. Create the BufferView
	componentSize := int(unsafe.Sizeof(vectors[0][0])) // float32 occupies 4 bytes
	byteLength := len(vectors) * dim * componentSize
	bufView := gltf.BufferView{
		Buffer:     0, // Main Buffer
		ByteOffset: doc.Buffers[0].ByteLength,
		ByteLength: uint32(byteLength),
		Target:     gltf.TargetArrayBuffer,
	}
	bufViewIndex := uint32(len(doc.BufferViews))
	doc.BufferViews = append(doc.BufferViews, &bufView)

	// 5. Write binary data
	buf := bytes.NewBuffer([]byte{})
	for _, v := range vectors {
		binary.Write(buf, binary.LittleEndian, v)
	}
	data := buf.Bytes()
	doc.Buffers[0].Data = append(doc.Buffers[0].Data, data...)
	doc.Buffers[0].ByteLength += uint32(byteLength)
	padBuffer(doc)
	// 7. Return the property table property
	return &ext_gltf.PropertyTableProperty{
		Values: bufViewIndex,
		Min:    mustMarshal(minVals),
		Max:    mustMarshal(maxVals),
	}, nil
}

// calculateVectorBounds calculates the bounds of each dimension of a vector array
func calculateVectorBounds[T float32 | float64 | int32 | uint32 | int8 | uint8 | int16 | uint16 | uint64 | int | int64](vectors [][]T) (min, max []T) {
	if len(vectors) == 0 {
		return nil, nil
	}
	dim := len(vectors[0])
	min = make([]T, dim)
	max = make([]T, dim)
	var mx T
	var mi T
	a := math.MaxFloat64
	mi = T(a)
	mx = -mi
	for i := range min {
		min[i] = mi
		max[i] = mx
	}

	for _, vec := range vectors {
		for i, v := range vec {
			if v < min[i] {
				min[i] = v
			}
			if v > max[i] {
				max[i] = v
			}
		}
	}
	return
}

// componentSize returns the byte size of the component type

// ptr helper function: create a pointer
func ptr[T any](v T) *T {
	return &v
}

func mustMarshal(v interface{}) json.RawMessage {
	data, _ := json.Marshal(v)
	return data
}

func PaddingByte(size int) []byte {
	padding := size % 8
	if padding != 0 {
		padding = 8 - padding
	}
	if padding == 0 {
		return []byte{}
	}
	pad := make([]byte, padding)
	for i := range pad {
		pad[i] = 0x20
	}
	return pad
}

func rackProps(props []map[string]interface{}) map[string][]interface{} {
	// 第一阶段：收集字段元信息
	fieldMeta := make(map[string]struct {
		typ          reflect.Type
		defaultValue interface{}
	})
	var fieldOrder []string // 保持字段发现顺序

	for _, prop := range props {
		for name, val := range prop {
			if _, exists := fieldMeta[name]; !exists && val != nil {
				fieldOrder = append(fieldOrder, name)
				typ := reflect.TypeOf(val)
				fieldMeta[name] = struct {
					typ          reflect.Type
					defaultValue interface{}
				}{
					typ:          typ,
					defaultValue: createZeroValue(val),
				}
			}
		}
	}

	// 第二阶段：构建结果
	result := make(map[string][]interface{}, len(fieldMeta))
	for _, name := range fieldOrder {
		meta := fieldMeta[name]
		values := make([]interface{}, len(props))

		for i, prop := range props {
			if val, exists := prop[name]; exists {
				if reflect.TypeOf(val) == meta.typ {
					values[i] = val
				} else {
					values[i] = convertType(val, meta.typ, meta.defaultValue)
				}
			} else {
				values[i] = meta.defaultValue
			}
		}
		result[name] = values
	}
	return result
}

func createZeroValue(sample interface{}) interface{} {
	if sample == nil {
		return nil
	}

	// 通用反射处理（覆盖所有类型）
	typ := reflect.TypeOf(sample)
	switch typ.Kind() {
	case reflect.Slice, reflect.Map, reflect.Ptr, reflect.Chan, reflect.Func, reflect.Interface:
		return reflect.Zero(typ).Interface() // 返回nil
	case reflect.Array, reflect.Struct:
		return reflect.New(typ).Elem().Interface() // 返回零值实例
	case reflect.String:
		return ""
	case reflect.Bool:
		return false
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return reflect.Zero(typ).Interface() // 返回0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return reflect.Zero(typ).Interface()
	case reflect.Float32, reflect.Float64:
		return reflect.Zero(typ).Interface()
	default:
		return nil
	}
}

func convertType(value interface{}, targetType reflect.Type, defaultValue interface{}) interface{} {
	if value == nil {
		return defaultValue
	}

	// 快速路径：常见类型转换
	switch targetType.Kind() {
	case reflect.Float32, reflect.Float64:
		switch v := value.(type) {
		case float32:
			return float64(v)
		case float64:
			return v
		case int:
			return float64(v)
		}
	case reflect.Int:
		switch v := value.(type) {
		case int:
			return v
		case float64:
			return int(v)
		}
	}

	// 慢速路径：反射转换
	srcValue := reflect.ValueOf(value)
	if srcValue.Type().ConvertibleTo(targetType) {
		return srcValue.Convert(targetType).Interface()
	}

	return defaultValue
}

func padBuffer(doc *gltf.Document) {
	s := (4 - (doc.Buffers[0].ByteLength % 4)) % 4
	pad := make([]byte, s)
	doc.Buffers[0].Data = append(doc.Buffers[0].Data, pad...)
	doc.Buffers[0].ByteLength += uint32(len(pad))
}
