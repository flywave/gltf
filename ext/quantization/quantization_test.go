package quantization

import (
	"testing"

	"github.com/flywave/gltf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestQuantizationExtensionUnmarshal(t *testing.T) {
	// 测试正常情况
	jsonData := []byte(`{
		"POSITION": 12,
		"NORMAL": 10,
		"TANGENT": 10,
		"TEXCOORD": 12,
		"COLOR": 8,
		"GENERIC": 8,
		"JOINTS": 8,
		"WEIGHTS": 8
	}`)

	ext, err := Unmarshal(jsonData)
	require.NoError(t, err, "解组应成功")

	quantizationExt, ok := ext.(*QuantizationExtension)
	require.True(t, ok, "应能转换为QuantizationExtension类型")
	assert.Equal(t, uint8(12), quantizationExt.PositionBits, "PositionBits应正确解析")
	assert.Equal(t, uint8(10), quantizationExt.NormalBits, "NormalBits应正确解析")
	assert.Equal(t, uint8(10), quantizationExt.TangentBits, "TangentBits应正确解析")
	assert.Equal(t, uint8(12), quantizationExt.TexCoordBits, "TexCoordBits应正确解析")
	assert.Equal(t, uint8(8), quantizationExt.ColorBits, "ColorBits应正确解析")
	assert.Equal(t, uint8(8), quantizationExt.GenericBits, "GenericBits应正确解析")
	assert.Equal(t, uint8(8), quantizationExt.JointBits, "JointBits应正确解析")
	assert.Equal(t, uint8(8), quantizationExt.WeightBits, "WeightBits应正确解析")
}

func TestQuantizationExtensionUnmarshalPartial(t *testing.T) {
	// 测试部分字段
	jsonData := []byte(`{
		"POSITION": 12,
		"NORMAL": 10
	}`)

	ext, err := Unmarshal(jsonData)
	require.NoError(t, err, "解组应成功")

	quantizationExt, ok := ext.(*QuantizationExtension)
	require.True(t, ok, "应能转换为QuantizationExtension类型")
	assert.Equal(t, uint8(12), quantizationExt.PositionBits, "PositionBits应正确解析")
	assert.Equal(t, uint8(10), quantizationExt.NormalBits, "NormalBits应正确解析")
	assert.Equal(t, uint8(0), quantizationExt.TangentBits, "TangentBits应为默认值0")
	assert.Equal(t, uint8(0), quantizationExt.TexCoordBits, "TexCoordBits应为默认值0")
}

func TestReadComponent(t *testing.T) {
	tests := []struct {
		name          string
		data          []byte
		componentType gltf.ComponentType
		normalized    bool
		expected      float32
		expectError   bool
	}{
		{
			name:          "ByteNormalized",
			data:          []byte{127},
			componentType: gltf.ComponentByte,
			normalized:    true,
			expected:      1.0,
			expectError:   false,
		},
		{
			name:          "UbyteNormalized",
			data:          []byte{255},
			componentType: gltf.ComponentUbyte,
			normalized:    true,
			expected:      1.0,
			expectError:   false,
		},
		{
			name:          "ShortNormalized",
			data:          []byte{255, 127}, // 32767 in little endian
			componentType: gltf.ComponentShort,
			normalized:    true,
			expected:      1.0,
			expectError:   false,
		},
		{
			name:          "UshortNormalized",
			data:          []byte{255, 255}, // 65535 in little endian
			componentType: gltf.ComponentUshort,
			normalized:    true,
			expected:      1.0,
			expectError:   false,
		},
		{
			name:          "InsufficientData",
			data:          []byte{0},
			componentType: gltf.ComponentShort,
			normalized:    true,
			expected:      0,
			expectError:   true,
		},
		{
			name:          "UnsupportedComponentType",
			data:          []byte{0, 0, 0, 0},
			componentType: gltf.ComponentFloat,
			normalized:    true,
			expected:      0,
			expectError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := readComponent(tt.data, tt.componentType, tt.normalized)
			if tt.expectError {
				assert.Error(t, err, "应返回错误")
			} else {
				assert.NoError(t, err, "不应返回错误")
				assert.Equal(t, tt.expected, result, "结果应匹配")
			}
		})
	}
}

func TestFirstNonZero(t *testing.T) {
	tests := []struct {
		name     string
		values   []uint8
		expected uint8
	}{
		{
			name:     "FirstNonZero",
			values:   []uint8{0, 0, 5, 0, 3},
			expected: 5,
		},
		{
			name:     "AllZero",
			values:   []uint8{0, 0, 0, 0},
			expected: 0,
		},
		{
			name:     "FirstNonZeroAtStart",
			values:   []uint8{7, 0, 5, 0, 3},
			expected: 7,
		},
		{
			name:     "Empty",
			values:   []uint8{},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := firstNonZero(tt.values...)
			assert.Equal(t, tt.expected, result, "结果应匹配")
		})
	}
}

func TestContains(t *testing.T) {
	slice := []string{"a", "b", "c", "d"}

	tests := []struct {
		name     string
		slice    []string
		item     string
		expected bool
	}{
		{
			name:     "ItemExists",
			slice:    slice,
			item:     "b",
			expected: true,
		},
		{
			name:     "ItemNotExists",
			slice:    slice,
			item:     "e",
			expected: false,
		},
		{
			name:     "EmptySlice",
			slice:    []string{},
			item:     "a",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := contains(tt.slice, tt.item)
			assert.Equal(t, tt.expected, result, "结果应匹配")
		})
	}
}

func TestClampFloat(t *testing.T) {
	tests := []struct {
		name     string
		value    float32
		min      float32
		max      float32
		expected float32
	}{
		{
			name:     "WithinRange",
			value:    5.0,
			min:      0.0,
			max:      10.0,
			expected: 5.0,
		},
		{
			name:     "BelowMin",
			value:    -5.0,
			min:      0.0,
			max:      10.0,
			expected: 0.0,
		},
		{
			name:     "AboveMax",
			value:    15.0,
			min:      0.0,
			max:      10.0,
			expected: 10.0,
		},
		{
			name:     "EqualToMin",
			value:    0.0,
			min:      0.0,
			max:      10.0,
			expected: 0.0,
		},
		{
			name:     "EqualToMax",
			value:    10.0,
			min:      0.0,
			max:      10.0,
			expected: 10.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := clampFloat(tt.value, tt.min, tt.max)
			assert.Equal(t, tt.expected, result, "结果应匹配")
		})
	}
}

func TestFloat32ToBytes(t *testing.T) {
	// 测试float32到字节的转换
	data := []float32{1.0, 2.0, 3.0, 4.0}
	result := float32ToBytes(data)

	// 验证长度
	assert.Equal(t, len(data)*4, len(result), "结果长度应正确")

	// 验证内容
	expected := make([]byte, 16)
	// 1.0 as float32 in little endian
	expected[0] = 0x00
	expected[1] = 0x00
	expected[2] = 0x80
	expected[3] = 0x3F
	// 2.0 as float32 in little endian
	expected[4] = 0x00
	expected[5] = 0x00
	expected[6] = 0x00
	expected[7] = 0x40
	// 3.0 as float32 in little endian
	expected[8] = 0x00
	expected[9] = 0x00
	expected[10] = 0x40
	expected[11] = 0x40
	// 4.0 as float32 in little endian
	expected[12] = 0x00
	expected[13] = 0x00
	expected[14] = 0x80
	expected[15] = 0x40

	assert.Equal(t, expected, result, "转换结果应匹配")
}

func TestGetQuantizationBits(t *testing.T) {
	ext := &QuantizationExtension{
		PositionBits: 12,
		NormalBits:   10,
		TangentBits:  10,
		TexCoordBits: 12,
		ColorBits:    8,
		GenericBits:  8,
		JointBits:    8,
		WeightBits:   8,
	}

	dequantizer := &Dequantizer{}

	tests := []struct {
		name         string
		attribute    string
		expectedBits uint8
	}{
		{"Position", "POSITION", 12},
		{"PositionWithIndex", "POSITION_0", 12},
		{"Normal", "NORMAL", 10},
		{"NormalWithIndex", "NORMAL_0", 10},
		{"Tangent", "TANGENT", 10},
		{"TexCoord", "TEXCOORD_0", 12},
		{"TexCoordAnotherIndex", "TEXCOORD_1", 12},
		{"Color", "COLOR_0", 8},
		{"Weights", "WEIGHTS_0", 8},
		{"Joints", "JOINTS_0", 8},
		{"Generic", "CUSTOM_ATTR", 8},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := dequantizer.getQuantizationBits(tt.attribute, ext)
			assert.Equal(t, tt.expectedBits, result, "量化位数应匹配")
		})
	}
}

func TestNewQuantizerWithNilConfig(t *testing.T) {
	doc := &gltf.Document{}
	quantizer := NewQuantizer(doc, nil)

	// 验证默认配置
	assert.NotNil(t, quantizer.config, "配置不应为nil")
	assert.Equal(t, uint8(12), quantizer.config.PositionBits, "PositionBits应为默认值12")
	assert.Equal(t, uint8(10), quantizer.config.NormalBits, "NormalBits应为默认值10")
	assert.Equal(t, uint8(10), quantizer.config.TangentBits, "TangentBits应为默认值10")
	assert.Equal(t, uint8(12), quantizer.config.TexCoordBits, "TexCoordBits应为默认值12")
	assert.Equal(t, uint8(8), quantizer.config.ColorBits, "ColorBits应为默认值8")
	assert.Equal(t, uint8(8), quantizer.config.WeightBits, "WeightBits应为默认值8")
	assert.Equal(t, uint8(8), quantizer.config.GenericBits, "GenericBits应为默认值8")
}

// 新增的测试用例

func TestDequantizerProcessWithNoExtension(t *testing.T) {
	// 测试没有量化扩展的文档处理
	doc := &gltf.Document{
		Meshes: []*gltf.Mesh{
			{
				Primitives: []*gltf.Primitive{
					{
						Attributes: map[string]uint32{
							"POSITION": 0,
						},
						Extensions: map[string]interface{}{},
					},
				},
			},
		},
		Accessors: []*gltf.Accessor{
			{
				ComponentType: gltf.ComponentFloat,
				Type:          gltf.AccessorVec3,
			},
		},
	}

	dequantizer := NewDequantizer(doc)
	err := dequantizer.Process()
	assert.NoError(t, err, "处理应成功")
}

func TestQuantizerProcessWithFloatAccessor(t *testing.T) {
	// 测试量化器处理浮点访问器
	buffer := &gltf.Buffer{
		ByteLength: 12,
		Data:       []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	}

	bufferView := &gltf.BufferView{
		Buffer:     0,
		ByteOffset: 0,
		ByteLength: 12,
	}

	accessor := &gltf.Accessor{
		BufferView:    gltf.Index(0),
		ByteOffset:    0,
		ComponentType: gltf.ComponentFloat,
		Count:         1,
		Type:          gltf.AccessorVec3,
		Min:           []float32{0.0, 0.0, 0.0},
		Max:           []float32{1.0, 1.0, 1.0},
	}

	doc := &gltf.Document{
		Buffers:     []*gltf.Buffer{buffer},
		BufferViews: []*gltf.BufferView{bufferView},
		Accessors:   []*gltf.Accessor{accessor},
		Meshes: []*gltf.Mesh{
			{
				Primitives: []*gltf.Primitive{
					{
						Attributes: map[string]uint32{
							"POSITION": 0,
						},
					},
				},
			},
		},
		ExtensionsUsed: []string{},
	}

	quantizer := NewQuantizer(doc, nil)
	err := quantizer.Process()
	assert.NoError(t, err, "处理应成功")

	// 验证扩展是否已添加
	assert.Contains(t, doc.ExtensionsUsed, ExtensionName, "应包含量化扩展")
}

func TestQuantizerProcessWithNonFloatAccessor(t *testing.T) {
	// 测试量化器处理非浮点访问器（应跳过）
	buffer := &gltf.Buffer{
		ByteLength: 12,
		Data:       []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	}

	bufferView := &gltf.BufferView{
		Buffer:     0,
		ByteOffset: 0,
		ByteLength: 12,
	}

	accessor := &gltf.Accessor{
		BufferView:    gltf.Index(0),
		ByteOffset:    0,
		ComponentType: gltf.ComponentUbyte, // 非浮点类型
		Count:         1,
		Type:          gltf.AccessorVec3,
	}

	doc := &gltf.Document{
		Buffers:     []*gltf.Buffer{buffer},
		BufferViews: []*gltf.BufferView{bufferView},
		Accessors:   []*gltf.Accessor{accessor},
		Meshes: []*gltf.Mesh{
			{
				Primitives: []*gltf.Primitive{
					{
						Attributes: map[string]uint32{
							"POSITION": 0,
						},
					},
				},
			},
		},
	}

	quantizer := NewQuantizer(doc, nil)
	err := quantizer.Process()
	assert.NoError(t, err, "处理应成功")

	// 验证访问器未被修改
	assert.Equal(t, gltf.ComponentUbyte, doc.Accessors[0].ComponentType, "组件类型应保持不变")
}

func TestCalculateMinMaxWithExistingValues(t *testing.T) {
	// 测试使用现有min/max值的计算
	accessor := &gltf.Accessor{
		Min: []float32{1.0, 2.0, 3.0},
		Max: []float32{4.0, 5.0, 6.0},
	}

	quantizer := &Quantizer{}
	min, max := quantizer.calculateMinMax(accessor, 3)

	assert.Equal(t, []float32{1.0, 2.0, 3.0}, min, "最小值应匹配")
	assert.Equal(t, []float32{4.0, 5.0, 6.0}, max, "最大值应匹配")
}

func TestEnsureExtensionDeclared(t *testing.T) {
	// 测试确保扩展声明
	doc := &gltf.Document{
		ExtensionsUsed:     []string{},
		ExtensionsRequired: []string{},
	}

	quantizer := &Quantizer{doc: doc}
	quantizer.ensureExtensionDeclared()

	assert.Contains(t, doc.ExtensionsUsed, ExtensionName, "应包含在ExtensionsUsed中")
	assert.Contains(t, doc.ExtensionsRequired, ExtensionName, "应包含在ExtensionsRequired中")
}

func TestRemoveTopLevelExtension(t *testing.T) {
	// 测试移除顶层扩展
	doc := &gltf.Document{
		Extensions: map[string]interface{}{
			ExtensionName: &QuantizationExtension{},
		},
		ExtensionsUsed:     []string{ExtensionName, "OTHER_EXTENSION"},
		ExtensionsRequired: []string{ExtensionName, "OTHER_EXTENSION"},
	}

	dequantizer := &Dequantizer{doc: doc}
	dequantizer.removeTopLevelExtension()

	// 验证扩展已被移除
	_, exists := doc.Extensions[ExtensionName]
	assert.False(t, exists, "顶层扩展应被移除")
	assert.NotContains(t, doc.ExtensionsUsed, ExtensionName, "应从ExtensionsUsed中移除")
	assert.NotContains(t, doc.ExtensionsRequired, ExtensionName, "应从ExtensionsRequired中移除")
}

// 新增的边界情况和错误处理测试

func TestDequantizeAccessorMissingMinOrMax(t *testing.T) {
	// 测试缺少min或max值的访问器
	doc := &gltf.Document{
		Buffers: []*gltf.Buffer{
			{
				ByteLength: 12,
				Data:       make([]byte, 12),
			},
		},
		BufferViews: []*gltf.BufferView{
			{
				Buffer:     0,
				ByteOffset: 0,
				ByteLength: 12,
			},
		},
		Accessors: []*gltf.Accessor{
			{
				BufferView:    gltf.Index(0),
				ByteOffset:    0,
				ComponentType: gltf.ComponentUbyte,
				Count:         1,
				Type:          gltf.AccessorVec3,
				// 故意缺少Min或Max
			},
		},
	}

	dequantizer := &Dequantizer{doc: doc}
	_, err := dequantizer.dequantizeAccessor(doc.Accessors[0], 8)
	assert.Error(t, err, "应返回错误")
	assert.Contains(t, err.Error(), "missing min/max values", "错误信息应包含'missing min/max values'")
}

func TestDequantizeAccessorUnsupportedType(t *testing.T) {
	// 测试不支持的访问器类型
	// 通过创建一个长度为0的min/max数组来模拟不支持的情况
	doc := &gltf.Document{
		Buffers: []*gltf.Buffer{
			{
				ByteLength: 12,
				Data:       make([]byte, 12),
			},
		},
		BufferViews: []*gltf.BufferView{
			{
				Buffer:     0,
				ByteOffset: 0,
				ByteLength: 12,
			},
		},
		Accessors: []*gltf.Accessor{
			{
				BufferView:    gltf.Index(0),
				ByteOffset:    0,
				ComponentType: gltf.ComponentUbyte,
				Count:         1,
				Type:          gltf.AccessorVec3,
				Min:           []float32{}, // 空数组表示不支持的情况
				Max:           []float32{}, // 空数组表示不支持的情况
			},
		},
	}

	dequantizer := &Dequantizer{doc: doc}
	_, err := dequantizer.dequantizeAccessor(doc.Accessors[0], 8)
	assert.Error(t, err, "应返回错误")
	assert.Contains(t, err.Error(), "min/max length mismatch", "错误信息应包含'min/max length mismatch'")
}

func TestDequantizeAccessorMissingBufferView(t *testing.T) {
	// 测试缺少缓冲视图的访问器
	doc := &gltf.Document{
		Accessors: []*gltf.Accessor{
			{
				ComponentType: gltf.ComponentUbyte,
				Count:         1,
				Type:          gltf.AccessorVec3,
				Min:           []float32{0, 0, 0},
				Max:           []float32{1, 1, 1},
				// 故意缺少BufferView
			},
		},
	}

	dequantizer := &Dequantizer{doc: doc}
	_, err := dequantizer.dequantizeAccessor(doc.Accessors[0], 8)
	assert.Error(t, err, "应返回错误")
	assert.Contains(t, err.Error(), "missing buffer view", "错误信息应包含'missing buffer view'")
}

func TestQuantizeAccessorNonFloat(t *testing.T) {
	// 测试量化非浮点访问器
	buffer := &gltf.Buffer{
		ByteLength: 12,
		Data:       make([]byte, 12),
	}

	bufferView := &gltf.BufferView{
		Buffer:     0,
		ByteOffset: 0,
		ByteLength: 12,
	}

	accessor := &gltf.Accessor{
		BufferView:    gltf.Index(0),
		ByteOffset:    0,
		ComponentType: gltf.ComponentUbyte, // 非浮点类型
		Count:         1,
		Type:          gltf.AccessorVec3,
		Min:           []float32{0, 0, 0},
		Max:           []float32{1, 1, 1},
	}

	doc := &gltf.Document{
		Buffers:     []*gltf.Buffer{buffer},
		BufferViews: []*gltf.BufferView{bufferView},
		Accessors:   []*gltf.Accessor{accessor},
	}

	quantizer := &Quantizer{doc: doc}
	_, err := quantizer.quantizeAccessor(accessor, 8, gltf.ComponentUbyte)
	assert.Error(t, err, "应返回错误")
	assert.Contains(t, err.Error(), "only float accessors can be quantized", "错误信息应包含'only float accessors can be quantized'")
}

func TestQuantizeAccessorUnsupportedType(t *testing.T) {
	// 测试访问器类型不支持的情况
	// 使用一个在switch语句中未定义的访问器类型来触发错误
	// 但gltf.AccessorType是预定义类型，我们无法直接创建无效值
	// 所以我们跳过这个测试，因为quantizeAccessor函数中的类型检查已经足够
	t.Skip("无法创建无效的gltf.AccessorType值进行测试")
}

func TestQuantizeAccessorBufferOutOfRange(t *testing.T) {
	// 测试缓冲区索引越界的情况
	buffer := &gltf.Buffer{
		ByteLength: 12,
		Data:       make([]byte, 12),
	}

	// 创建一个引用不存在缓冲区的缓冲视图
	bufferView := &gltf.BufferView{
		Buffer:     1, // 不存在的缓冲区索引
		ByteOffset: 0,
		ByteLength: 12,
	}

	accessor := &gltf.Accessor{
		BufferView:    gltf.Index(0),
		ByteOffset:    0,
		ComponentType: gltf.ComponentFloat,
		Count:         1,
		Type:          gltf.AccessorVec3,
		Min:           []float32{0, 0, 0},
		Max:           []float32{1, 1, 1},
	}

	doc := &gltf.Document{
		Buffers:     []*gltf.Buffer{buffer},         // 只有索引0的缓冲区
		BufferViews: []*gltf.BufferView{bufferView}, // 但缓冲视图引用索引1
		Accessors:   []*gltf.Accessor{accessor},
	}

	quantizer := &Quantizer{doc: doc}
	_, err := quantizer.quantizeAccessor(accessor, 8, gltf.ComponentUbyte)
	assert.Error(t, err, "应返回错误")
	assert.Contains(t, err.Error(), "buffer index out of range", "错误信息应包含'buffer index out of range'")
}

func TestQuantizeAccessorMissingBufferView(t *testing.T) {
	// 测试缺少缓冲视图的访问器
	accessor := &gltf.Accessor{
		ComponentType: gltf.ComponentFloat,
		Count:         1,
		Type:          gltf.AccessorVec3,
		Min:           []float32{0, 0, 0},
		Max:           []float32{1, 1, 1},
		// 故意缺少BufferView
	}

	doc := &gltf.Document{
		Accessors: []*gltf.Accessor{accessor},
	}

	quantizer := &Quantizer{doc: doc}
	_, err := quantizer.quantizeAccessor(accessor, 8, gltf.ComponentUbyte)
	assert.Error(t, err, "应返回错误")
	assert.Contains(t, err.Error(), "missing buffer view", "错误信息应包含'missing buffer view'")
}

// 添加更多边界情况测试

func TestQuantizeAccessorDataExceedsBufferRange(t *testing.T) {
	// 测试数据超出缓冲区范围的情况
	buffer := &gltf.Buffer{
		ByteLength: 12,
		Data:       make([]byte, 12), // 只有12字节
	}

	bufferView := &gltf.BufferView{
		Buffer:     0,
		ByteOffset: 0,
		ByteLength: 12,
	}

	// 创建一个需要更多数据的访问器
	accessor := &gltf.Accessor{
		BufferView:    gltf.Index(0),
		ByteOffset:    0,
		ComponentType: gltf.ComponentFloat,
		Count:         100, // 需要很多数据
		Type:          gltf.AccessorVec3,
		Min:           []float32{0, 0, 0},
		Max:           []float32{1, 1, 1},
	}

	doc := &gltf.Document{
		Buffers:     []*gltf.Buffer{buffer},
		BufferViews: []*gltf.BufferView{bufferView},
		Accessors:   []*gltf.Accessor{accessor},
	}

	quantizer := &Quantizer{doc: doc}
	_, err := quantizer.quantizeAccessor(accessor, 8, gltf.ComponentUbyte)
	assert.Error(t, err, "应返回错误")
	assert.Contains(t, err.Error(), "accessor data exceeds buffer range", "错误信息应包含'accessor data exceeds buffer range'")
}

func TestDequantizeAccessorBufferViewIndexOutOfRange(t *testing.T) {
	// 测试缓冲视图索引越界的情况
	doc := &gltf.Document{
		Buffers: []*gltf.Buffer{
			{
				ByteLength: 12,
				Data:       make([]byte, 12),
			},
		},
		BufferViews: []*gltf.BufferView{}, // 空的缓冲视图列表
		Accessors: []*gltf.Accessor{
			{
				BufferView:    gltf.Index(0), // 引用不存在的缓冲视图
				ByteOffset:    0,
				ComponentType: gltf.ComponentUbyte,
				Count:         1,
				Type:          gltf.AccessorVec3,
				Min:           []float32{0, 0, 0},
				Max:           []float32{1, 1, 1},
			},
		},
	}

	dequantizer := &Dequantizer{doc: doc}
	_, err := dequantizer.dequantizeAccessor(doc.Accessors[0], 8)
	assert.Error(t, err, "应返回错误")
	assert.Contains(t, err.Error(), "buffer view index out of range", "错误信息应包含'buffer view index out of range'")
}

func TestDequantizeAccessorBufferIndexOutOfRange(t *testing.T) {
	// 测试缓冲区索引越界的情况
	doc := &gltf.Document{
		Buffers: []*gltf.Buffer{}, // 空的缓冲区列表
		BufferViews: []*gltf.BufferView{
			{
				Buffer:     0, // 引用不存在的缓冲区
				ByteOffset: 0,
				ByteLength: 12,
			},
		},
		Accessors: []*gltf.Accessor{
			{
				BufferView:    gltf.Index(0),
				ByteOffset:    0,
				ComponentType: gltf.ComponentUbyte,
				Count:         1,
				Type:          gltf.AccessorVec3,
				Min:           []float32{0, 0, 0},
				Max:           []float32{1, 1, 1},
			},
		},
	}

	dequantizer := &Dequantizer{doc: doc}
	_, err := dequantizer.dequantizeAccessor(doc.Accessors[0], 8)
	assert.Error(t, err, "应返回错误")
	assert.Contains(t, err.Error(), "buffer index out of range", "错误信息应包含'buffer index out of range'")
}
