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
	ext_gltf "github.com/flywave/gltf/3dtile/gltf"
	"github.com/flywave/go3d/mat3"
	"github.com/flywave/go3d/mat4"
)

func Gen3DTileMetadata(doc *gltf.Document, class string, properties map[string]interface{}) error {
	if doc == nil {
		return fmt.Errorf("GLTF document cannot be nil")
	}
	if len(properties) == 0 {
		return nil // No properties, no operation
	}

	// 2. Initialize the extension structure
	if doc.Extensions == nil {
		doc.Extensions = make(map[string]interface{})
	}
	var metadata ext_gltf.ExtStructuralMetadata

	// 3. Build the Schema (if it doesn't exist)
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
			return fmt.Errorf("Failed to infer the type of property %s: %w", propName, err)
		}

		metadata.Schema.Classes[class].Properties[propName] = ext_gltf.ClassProperty{
			Type:          propType,
			ComponentType: componentType,
		}
		valSlice := reflect.ValueOf(values)
		if valSlice.Kind() == reflect.Slice {
			if count == -1 {
				count = valSlice.Len()
			} else if count != valSlice.Len() {
				return fmt.Errorf("slice count must be equal")
			}

			// 4.3 Create the property accessor
			accessor, err := createPropertyAccessor(doc, values)
			if err != nil {
				return fmt.Errorf("Failed to create the accessor for property %s: %w", propName, err)
			}
			propTable.Properties[propName] = *accessor
		}

	}

	// 5. Set the property table count
	propTable.Count = uint32(count)
	metadata.PropertyTables = append(metadata.PropertyTables, propTable)
	doc.Extensions[ext_gltf.EXT_structural_metadata_Name] = metadata
	return nil
}

// --- Helper functions ---

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
	case int, int8, int16, int32, int64:
		return ext_gltf.ClassPropertyTypeScalar, ptr(ext_gltf.ClassPropertyComponentTypeInt32), nil
	case uint, uint8, uint16, uint32, uint64:
		return ext_gltf.ClassPropertyTypeScalar, ptr(ext_gltf.ClassPropertyComponentTypeUint32), nil
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
		return "", nil, fmt.Errorf("Unsupported type: %T", v)
	}
	return "", nil, fmt.Errorf("Unable to infer the type")
}

// createPropertyAccessor creates a property accessor
func createPropertyAccessor(doc *gltf.Document, values interface{}) (*ext_gltf.PropertyTableProperty, error) {
	switch v := values.(type) {
	case string, bool, float32, float64, int32, uint32:
		return CreateInlinePropertyTableProperty(values)
	case []string:
		return createStringAccessor(doc, v)
	case []float32:
		return createFloatAccessor(doc, v, ext_gltf.ClassProperty{Type: ext_gltf.ClassPropertyTypeScalar})
	case [][]float32:
		return createVectorAccessor(doc, v)
	case []float64:
		return createFloatAccessor(doc, v, ext_gltf.ClassProperty{Type: ext_gltf.ClassPropertyTypeScalar})
	case [][]float64:
		return createVectorAccessor(doc, v)
	case []int:
		return createFloatAccessor(doc, v, ext_gltf.ClassProperty{Type: ext_gltf.ClassPropertyTypeScalar})
	case [][]int:
		return createVectorAccessor(doc, v)
	case []int8:
		return createFloatAccessor(doc, v, ext_gltf.ClassProperty{Type: ext_gltf.ClassPropertyTypeScalar})
	case [][]int8:
		return createVectorAccessor(doc, v)
	case []uint8:
		return createFloatAccessor(doc, v, ext_gltf.ClassProperty{Type: ext_gltf.ClassPropertyTypeScalar})
	case [][]uint8:
		return createVectorAccessor(doc, v)
	case []int16:
		return createFloatAccessor(doc, v, ext_gltf.ClassProperty{Type: ext_gltf.ClassPropertyTypeScalar})
	case [][]int16:
		return createVectorAccessor(doc, v)
	case []uint16:
		return createFloatAccessor(doc, v, ext_gltf.ClassProperty{Type: ext_gltf.ClassPropertyTypeScalar})
	case [][]uint16:
		return createVectorAccessor(doc, v)
	case []uint32:
		return createFloatAccessor(doc, v, ext_gltf.ClassProperty{Type: ext_gltf.ClassPropertyTypeScalar})
	case [][]uint32:
		return createVectorAccessor(doc, v)
	case []int32:
		return createFloatAccessor(doc, v, ext_gltf.ClassProperty{Type: ext_gltf.ClassPropertyTypeScalar})
	case [][]int32:
		return createVectorAccessor(doc, v)
	case []uint64:
		return createFloatAccessor(doc, v, ext_gltf.ClassProperty{Type: ext_gltf.ClassPropertyTypeScalar})
	case [][]uint64:
		return createVectorAccessor(doc, v)
	case []int64:
		return createFloatAccessor(doc, v, ext_gltf.ClassProperty{Type: ext_gltf.ClassPropertyTypeScalar})
	case [][]int64:
		return createVectorAccessor(doc, v)
	default:
		return nil, fmt.Errorf("Unimplemented type handling: %T", values)
	}
}

func CreateInlinePropertyTableProperty(value interface{}) (*ext_gltf.PropertyTableProperty, error) {
	if !isAllowedInlineType(value) {
		return nil, fmt.Errorf("类型%T不支持直接内联", value)
	}

	prop := &ext_gltf.PropertyTableProperty{
		Extras: buildInlineValueExtra(value),
	}
	switch value.(type) {
	case float32, float64, int, int32, uint32:
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
	indices := make([]uint32, len(values))
	indexMap := make(map[string]uint32)
	stringData := bytes.Buffer{}

	// 构建字符串表和二进制数据
	for i, s := range values {
		if idx, exists := indexMap[s]; exists {
			indices[i] = idx
			continue
		}

		// 新字符串处理
		idx := uint32(len(stringTable))
		stringTable = append(stringTable, s)
		indexMap[s] = idx
		indices[i] = idx

		// 写入字符串数据（带null终止符）
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

	// 3. 创建字符串索引的BufferView
	indicesByteLength := len(indices) * 4 // uint32占4字节
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

	// 7. 内存对齐（GLTF要求4字节对齐）
	pad := (4 - (doc.Buffers[0].ByteLength % 4)) % 4
	if pad > 0 {
		doc.Buffers[0].Data = append(doc.Buffers[0].Data, make([]byte, pad)...)
		doc.Buffers[0].ByteLength += uint32(pad)
	}

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

func createFloatAccessor[T float32 | float64 | int | int8 | uint8 | int16 | uint16 | int32 | uint32 | uint64 | int64](doc *gltf.Document, values []T, prop ext_gltf.ClassProperty) (*ext_gltf.PropertyTableProperty, error) {
	if len(values) == 0 {
		return nil, errors.New("empty values array")
	}

	// 1. Calculate min/max in a single pass
	minVal, maxVal := values[0], values[0]
	for _, v := range values[1:] {
		if v < minVal {
			minVal = v
		}
		if v > maxVal {
			maxVal = v
		}
	}

	// 2. Prepare buffer data with direct byte conversion (faster than binary.Write)
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

	// 5. Add padding if needed (4-byte alignment)
	if padding := (4 - (byteLength % 4)) % 4; padding > 0 {
		doc.Buffers[0].Data = append(doc.Buffers[0].Data, make([]byte, padding)...)
		doc.Buffers[0].ByteLength += uint32(padding)
	}

	// 6. Return property table property
	return &ext_gltf.PropertyTableProperty{
		Values: bufViewIndex,
		Min:    mustMarshal(minVal),
		Max:    mustMarshal(maxVal),
	}, nil
}

// calculateFloatRange calculates the minimum and maximum values of a float array
func calculateFloatRange(values []float32) (min, max float32) {
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
		return nil, fmt.Errorf("Vector array cannot be empty")
	}

	// 1. Determine the vector dimension and component type
	dim := len(vectors[0])

	// 2. Check the consistency of all vector dimensions
	for i, vec := range vectors {
		if len(vec) != dim {
			return nil, fmt.Errorf("The dimension of vector %d does not match. Expected %d, actual %d", i, dim, len(vec))
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

	// 6. Memory alignment (GLTF requires BufferView starting position to be 4-byte aligned)
	pad := (4 - (doc.Buffers[0].ByteLength % 4)) % 4
	if pad > 0 {
		doc.Buffers[0].Data = append(doc.Buffers[0].Data, make([]byte, pad)...)
		doc.Buffers[0].ByteLength += uint32(pad)
	}

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
