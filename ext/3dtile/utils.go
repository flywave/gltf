package tile3d

import (
	"encoding/json"
	"fmt"
	"reflect"

	extgltf "github.com/flywave/gltf/ext/3dtile/gltf"

	mat3d "github.com/flywave/go3d/float64/mat3"
	mat4d "github.com/flywave/go3d/float64/mat4"

	"github.com/flywave/go3d/mat3"
	"github.com/flywave/go3d/mat4"
)

func CreateInlinePropertyTableProperty(value interface{}) (*extgltf.PropertyTableProperty, error) {
	if !isAllowedInlineType(value) {
		return nil, fmt.Errorf("type %T is not allowed to inline", value)
	}

	prop := &extgltf.PropertyTableProperty{
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

// ptr helper function: create a pointer
func ptr[T any](v T) *T {
	return &v
}

// inferPropertyType infers the property type
func inferPropertyType(values interface{}) (extgltf.ClassPropertyType, *extgltf.ClassPropertyComponentType, error) {
	switch v := reflect.ValueOf(values).Index(0).Interface().(type) {
	case string:
		return extgltf.ClassPropertyTypeString, nil, nil
	case bool:
		return extgltf.ClassPropertyTypeBoolean, nil, nil
	case float32:
		return extgltf.ClassPropertyTypeScalar, ptr(extgltf.ClassPropertyComponentTypeFloat32), nil
	case float64:
		return extgltf.ClassPropertyTypeScalar, ptr(extgltf.ClassPropertyComponentTypeFloat64), nil
	case int:
		return extgltf.ClassPropertyTypeScalar, ptr(extgltf.ClassPropertyComponentTypeInt64), nil
	case int8:
		return extgltf.ClassPropertyTypeScalar, ptr(extgltf.ClassPropertyComponentTypeInt8), nil
	case int16:
		return extgltf.ClassPropertyTypeScalar, ptr(extgltf.ClassPropertyComponentTypeInt16), nil
	case int32:
		return extgltf.ClassPropertyTypeScalar, ptr(extgltf.ClassPropertyComponentTypeInt32), nil
	case int64:
		return extgltf.ClassPropertyTypeScalar, ptr(extgltf.ClassPropertyComponentTypeInt64), nil
	case uint:
		return extgltf.ClassPropertyTypeScalar, ptr(extgltf.ClassPropertyComponentTypeUint64), nil
	case uint8:
		return extgltf.ClassPropertyTypeScalar, ptr(extgltf.ClassPropertyComponentTypeUint8), nil
	case uint16:
		return extgltf.ClassPropertyTypeScalar, ptr(extgltf.ClassPropertyComponentTypeUint16), nil
	case uint32:
		return extgltf.ClassPropertyTypeScalar, ptr(extgltf.ClassPropertyComponentTypeUint32), nil
	case uint64:
		return extgltf.ClassPropertyTypeScalar, ptr(extgltf.ClassPropertyComponentTypeUint64), nil
	case []float32:
		switch len(v) {
		case 2:
			return extgltf.ClassPropertyTypeVec2, ptr(extgltf.ClassPropertyComponentTypeFloat32), nil
		case 3:
			return extgltf.ClassPropertyTypeVec3, ptr(extgltf.ClassPropertyComponentTypeFloat32), nil
		case 4:
			return extgltf.ClassPropertyTypeVec4, ptr(extgltf.ClassPropertyComponentTypeFloat32), nil
		}
	case mat3.T:
		return extgltf.ClassPropertyTypeMat3, ptr(extgltf.ClassPropertyComponentTypeFloat32), nil
	case mat4.T:
		return extgltf.ClassPropertyTypeMat4, ptr(extgltf.ClassPropertyComponentTypeFloat32), nil
	case mat3d.T:
		return extgltf.ClassPropertyTypeMat3, ptr(extgltf.ClassPropertyComponentTypeFloat64), nil
	case mat4d.T:
		return extgltf.ClassPropertyTypeMat4, ptr(extgltf.ClassPropertyComponentTypeFloat64), nil
	case []float64:
		switch len(v) {
		case 2:
			return extgltf.ClassPropertyTypeVec2, ptr(extgltf.ClassPropertyComponentTypeFloat64), nil
		case 3:
			return extgltf.ClassPropertyTypeVec3, ptr(extgltf.ClassPropertyComponentTypeFloat64), nil
		case 4:
			return extgltf.ClassPropertyTypeVec4, ptr(extgltf.ClassPropertyComponentTypeFloat64), nil
		}
	default:
		return "", nil, fmt.Errorf("usupported type: %T", v)
	}
	return "", nil, fmt.Errorf("unable to infer the type")
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

func unmarshalExtension(ext interface{}, target interface{}) error {
	raw, err := json.Marshal(ext)
	if err != nil {
		return err
	}
	return json.Unmarshal(raw, target)
}
