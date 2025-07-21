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
	padding := size % 4
	if padding != 0 {
		padding = 4 - padding
	}
	if padding == 0 {
		return []byte{}
	}
	pad := make([]byte, padding)
	for i := range pad {
		pad[i] = 0x00
	}
	return pad
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
	case reflect.String:
		switch v := value.(type) {
		case string:
			return v
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
func inferPropertyType(values interface{}) (extgltf.ClassPropertyType, *extgltf.ClassPropertyComponentType, bool, error) {
	switch v := reflect.ValueOf(values).Index(0).Interface().(type) {
	case string:
		return extgltf.ClassPropertyTypeString, nil, false, nil
	case bool:
		return extgltf.ClassPropertyTypeBoolean, nil, false, nil
	case float32:
		return extgltf.ClassPropertyTypeScalar, ptr(extgltf.ClassPropertyComponentTypeFloat32), false, nil
	case float64:
		return extgltf.ClassPropertyTypeScalar, ptr(extgltf.ClassPropertyComponentTypeFloat64), false, nil
	case int:
		return extgltf.ClassPropertyTypeScalar, ptr(extgltf.ClassPropertyComponentTypeInt64), false, nil
	case int8:
		return extgltf.ClassPropertyTypeScalar, ptr(extgltf.ClassPropertyComponentTypeInt8), false, nil
	case int16:
		return extgltf.ClassPropertyTypeScalar, ptr(extgltf.ClassPropertyComponentTypeInt16), false, nil
	case int32:
		return extgltf.ClassPropertyTypeScalar, ptr(extgltf.ClassPropertyComponentTypeInt32), false, nil
	case int64:
		return extgltf.ClassPropertyTypeScalar, ptr(extgltf.ClassPropertyComponentTypeInt64), false, nil
	case uint:
		return extgltf.ClassPropertyTypeScalar, ptr(extgltf.ClassPropertyComponentTypeUint64), false, nil
	case uint8:
		return extgltf.ClassPropertyTypeScalar, ptr(extgltf.ClassPropertyComponentTypeUint8), false, nil
	case uint16:
		return extgltf.ClassPropertyTypeScalar, ptr(extgltf.ClassPropertyComponentTypeUint16), false, nil
	case uint32:
		return extgltf.ClassPropertyTypeScalar, ptr(extgltf.ClassPropertyComponentTypeUint32), false, nil
	case uint64:
		return extgltf.ClassPropertyTypeScalar, ptr(extgltf.ClassPropertyComponentTypeUint64), false, nil
	case []string:
		return extgltf.ClassPropertyTypeString, nil, true, nil
	case []bool:
		return extgltf.ClassPropertyTypeBoolean, nil, true, nil

	case []int:
		return extgltf.ClassPropertyTypeScalar, ptr(extgltf.ClassPropertyComponentTypeInt64), true, nil
	case []int8:
		return extgltf.ClassPropertyTypeScalar, ptr(extgltf.ClassPropertyComponentTypeInt8), true, nil
	case []int16:
		return extgltf.ClassPropertyTypeScalar, ptr(extgltf.ClassPropertyComponentTypeInt16), true, nil
	case []int32:
		return extgltf.ClassPropertyTypeScalar, ptr(extgltf.ClassPropertyComponentTypeInt32), true, nil
	case []int64:
		return extgltf.ClassPropertyTypeScalar, ptr(extgltf.ClassPropertyComponentTypeInt64), true, nil
	case []uint:
		return extgltf.ClassPropertyTypeScalar, ptr(extgltf.ClassPropertyComponentTypeUint64), true, nil
	case []uint8:
		return extgltf.ClassPropertyTypeScalar, ptr(extgltf.ClassPropertyComponentTypeUint8), true, nil
	case []uint16:
		return extgltf.ClassPropertyTypeScalar, ptr(extgltf.ClassPropertyComponentTypeUint16), true, nil
	case []uint32:
		return extgltf.ClassPropertyTypeScalar, ptr(extgltf.ClassPropertyComponentTypeUint32), true, nil
	case []uint64:
		return extgltf.ClassPropertyTypeScalar, ptr(extgltf.ClassPropertyComponentTypeUint64), true, nil
	case []float32:
		switch len(v) {
		case 2:
			return extgltf.ClassPropertyTypeVec2, ptr(extgltf.ClassPropertyComponentTypeFloat32), false, nil
		case 3:
			return extgltf.ClassPropertyTypeVec3, ptr(extgltf.ClassPropertyComponentTypeFloat32), false, nil
		case 4:
			return extgltf.ClassPropertyTypeVec4, ptr(extgltf.ClassPropertyComponentTypeFloat32), false, nil
		default:
			return extgltf.ClassPropertyTypeVec4, ptr(extgltf.ClassPropertyComponentTypeFloat32), true, nil
		}
	case mat3.T:
		return extgltf.ClassPropertyTypeMat3, ptr(extgltf.ClassPropertyComponentTypeFloat32), false, nil
	case mat4.T:
		return extgltf.ClassPropertyTypeMat4, ptr(extgltf.ClassPropertyComponentTypeFloat32), false, nil
	case mat3d.T:
		return extgltf.ClassPropertyTypeMat3, ptr(extgltf.ClassPropertyComponentTypeFloat64), false, nil
	case mat4d.T:
		return extgltf.ClassPropertyTypeMat4, ptr(extgltf.ClassPropertyComponentTypeFloat64), false, nil
	case []float64:
		switch len(v) {
		case 2:
			return extgltf.ClassPropertyTypeVec2, ptr(extgltf.ClassPropertyComponentTypeFloat64), false, nil
		case 3:
			return extgltf.ClassPropertyTypeVec3, ptr(extgltf.ClassPropertyComponentTypeFloat64), false, nil
		case 4:
			return extgltf.ClassPropertyTypeVec4, ptr(extgltf.ClassPropertyComponentTypeFloat64), false, nil
		default:
			return extgltf.ClassPropertyTypeVec4, ptr(extgltf.ClassPropertyComponentTypeFloat32), true, nil
		}
	default:
		return "", nil, false, fmt.Errorf("usupported type: %T", v)
	}
	return "", nil, false, fmt.Errorf("unable to infer the type")
}

func rackProps(props []map[string]interface{}) map[string]interface{} {
	// 第一阶段：收集字段元信息
	fieldMeta := make(map[string]struct {
		typ          reflect.Type
		defaultValue interface{}
	})
	var fieldOrder []string // 保持字段发现顺序

	for _, prop := range props {
		for name, val := range prop {
			meta, exists := fieldMeta[name]
			if !exists && val != nil {
				fieldOrder = append(fieldOrder, name)
				typ := GetUnderlyingType(val)
				meta = struct {
					typ          reflect.Type
					defaultValue interface{}
				}{
					typ:          typ,
					defaultValue: CreateDefaultValue(typ),
				}
				fieldMeta[name] = meta
			}
			res := convertValue(meta.typ, val)
			prop[name] = res
		}
	}

	// 第二阶段：构建结果
	result := make(map[string]interface{}, len(fieldMeta))
	for _, name := range fieldOrder {
		meta := fieldMeta[name]
		sliceType := reflect.SliceOf(meta.typ)
		values := reflect.MakeSlice(sliceType, len(props), len(props))
		for i, prop := range props {
			if val, exists := prop[name]; exists {
				if reflect.TypeOf(val) == meta.typ {
					values.Index(i).Set(reflect.ValueOf(val))
				} else {
					values.Index(i).Set(reflect.ValueOf(convertType(val, meta.typ, meta.defaultValue)))
				}
			} else {
				values.Index(i).Set(reflect.ValueOf(meta.defaultValue))
			}
		}
		result[name] = values.Interface()
	}
	return result
}

func GetUnderlyingType(val interface{}) reflect.Type {
	if val == nil {
		return nil
	}

	typ := reflect.TypeOf(val)
	if typ == nil {
		return nil
	}
	return getBaseType(typ, val)
}

func getBaseType(typ reflect.Type, val interface{}) reflect.Type {
	var val1 interface{}
	var ty1 reflect.Type
	switch typ.Kind() {
	case reflect.Slice:
		vals := val.([]interface{})
		val1 = vals[0]
		ty1 = typ.Elem()
		t := getBaseType(ty1, val1)
		return reflect.SliceOf(t)
	case reflect.Array:
		vals := val.([]interface{})
		val1 = vals[0]
		ty1 = typ.Elem()
		t := getBaseType(ty1, val1)
		return reflect.ArrayOf(typ.Len(), t)
	case reflect.Map:
		ty1 = typ.Elem()
		vals := val.(map[string]interface{})
		for _, v := range vals {
			val1 = v
			break
		}
		t := getBaseType(ty1, val1)
		return reflect.MapOf(typ.Key(), t)
	default:
		switch val.(type) {
		case string:
			return reflect.TypeOf("")
		case int:
			return reflect.TypeOf(0)
		case float64:
			return reflect.TypeOf(float64(0.0))
		case bool:
			return reflect.TypeOf(false)
		default:
			return typ
		}
	}
}

func convertValue(typ reflect.Type, val interface{}) interface{} {
	var ty1 reflect.Type
	switch typ.Kind() {
	case reflect.Slice:
		vals := val.([]interface{})
		ty1 = typ.Elem()
		res := reflect.MakeSlice(typ, len(vals), len(vals))
		for i, v := range vals {
			res.Index(i).Set(reflect.ValueOf(convertValue(ty1, v)))
		}
		return res.Interface()
	case reflect.Array:
		vals := val.([]interface{})
		ty1 = typ.Elem()
		res := reflect.New(typ)
		for i, v := range vals {
			res.Index(i).Set(reflect.ValueOf(convertValue(ty1, v)))
		}
		return res.Interface()
	case reflect.Map:
		keyType := typ.Key()
		ty1 = typ.Elem()
		vals := val.(map[interface{}]interface{})
		mapType := reflect.MapOf(keyType, ty1)
		mapValue := reflect.MakeMap(mapType)

		for k, v := range vals {
			key := convertValue(keyType, k)
			val := convertValue(ty1, v)
			mapValue.SetMapIndex(reflect.ValueOf(key), reflect.ValueOf(val))
		}
		return mapValue.Interface()
	default:
		return convertType(val, typ, nil)
	}
}

// CreateDefaultValue 创建该类型的默认值
func CreateDefaultValue(typ reflect.Type) interface{} {
	if typ == nil {
		return nil
	}
	return reflect.Zero(typ).Interface()
}

func unmarshalExtension(ext interface{}, target interface{}) error {
	raw, err := json.Marshal(ext)
	if err != nil {
		return err
	}
	return json.Unmarshal(raw, target)
}
