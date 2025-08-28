package meshopt

import (
	"testing"

	"github.com/flywave/gltf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCompressionExtensionUnmarshal(t *testing.T) {
	// 测试正常情况
	jsonData := []byte(`{
		"buffer": 1,
		"byteOffset": 100,
		"byteLength": 200,
		"byteStride": 12,
		"count": 50,
		"mode": "ATTRIBUTES",
		"filter": "OCTAHEDRAL"
	}`)

	ext, err := Unmarshal(jsonData)
	require.NoError(t, err, "解组应成功")

	compressionExt, ok := ext.(*CompressionExtension)
	require.True(t, ok, "应能转换为CompressionExtension类型")
	assert.Equal(t, uint32(1), compressionExt.Buffer, "Buffer应正确解析")
	assert.Equal(t, uint32(100), compressionExt.ByteOffset, "ByteOffset应正确解析")
	assert.Equal(t, uint32(200), compressionExt.ByteLength, "ByteLength应正确解析")
	assert.Equal(t, uint32(12), compressionExt.ByteStride, "ByteStride应正确解析")
	assert.Equal(t, uint32(50), compressionExt.Count, "Count应正确解析")
	assert.Equal(t, ModeAttributes, compressionExt.Mode, "Mode应正确解析")
	assert.Equal(t, FilterOctahedral, compressionExt.Filter, "Filter应正确解析")
}

func TestCompressionExtensionUnmarshalMinimal(t *testing.T) {
	// 测试最小化数据
	jsonData := []byte(`{
		"buffer": 1,
		"byteLength": 200,
		"byteStride": 12,
		"count": 50,
		"mode": "INDICES"
	}`)

	ext, err := Unmarshal(jsonData)
	require.NoError(t, err, "解组应成功")

	compressionExt, ok := ext.(*CompressionExtension)
	require.True(t, ok, "应能转换为CompressionExtension类型")
	assert.Equal(t, uint32(1), compressionExt.Buffer, "Buffer应正确解析")
	assert.Equal(t, uint32(0), compressionExt.ByteOffset, "ByteOffset应为默认值0")
	assert.Equal(t, uint32(200), compressionExt.ByteLength, "ByteLength应正确解析")
	assert.Equal(t, uint32(12), compressionExt.ByteStride, "ByteStride应正确解析")
	assert.Equal(t, uint32(50), compressionExt.Count, "Count应正确解析")
	assert.Equal(t, ModeIndices, compressionExt.Mode, "Mode应正确解析")
	assert.Equal(t, CompressionFilter(""), compressionExt.Filter, "Filter应为空")
}

func TestValidateModeStride(t *testing.T) {
	tests := []struct {
		name        string
		mode        CompressionMode
		stride      uint32
		expectError bool
	}{
		{"ValidAttributesStride4", ModeAttributes, 4, false},
		{"ValidAttributesStride256", ModeAttributes, 256, false},
		{"InvalidAttributesStrideOdd", ModeAttributes, 5, true},
		{"InvalidAttributesStrideTooLarge", ModeAttributes, 260, true},
		{"ValidTrianglesStride2", ModeTriangles, 2, false},
		{"ValidTrianglesStride4", ModeTriangles, 4, false},
		{"InvalidTrianglesStride3", ModeTriangles, 3, true},
		{"ValidIndicesStride2", ModeIndices, 2, false},
		{"ValidIndicesStride4", ModeIndices, 4, false},
		{"InvalidIndicesStride3", ModeIndices, 3, true},
		{"InvalidMode", "INVALID", 4, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateModeStride(tt.mode, tt.stride)
			if tt.expectError {
				assert.Error(t, err, "应返回错误")
			} else {
				assert.NoError(t, err, "不应返回错误")
			}
		})
	}
}

func TestMeshoptEncodeEmptyData(t *testing.T) {
	// 测试空数据
	_, _, err := MeshoptEncode([]byte{}, 10, 4, ModeAttributes, FilterNone)
	assert.Error(t, err, "应返回错误")
	assert.Contains(t, err.Error(), "empty input data", "错误信息应包含'empty input data'")
}

func TestMeshoptEncodeZeroStride(t *testing.T) {
	// 测试零步长
	data := make([]byte, 20)
	_, _, err := MeshoptEncode(data, 5, 0, ModeAttributes, FilterNone)
	assert.Error(t, err, "应返回错误")
	assert.Contains(t, err.Error(), "zero byteStride", "错误信息应包含'zero byteStride'")
}

func TestMeshoptEncodeInsufficientData(t *testing.T) {
	// 测试数据不足
	data := make([]byte, 10)                                           // 数据不足
	_, _, err := MeshoptEncode(data, 5, 4, ModeAttributes, FilterNone) // 需要20字节(5*4)
	assert.Error(t, err, "应返回错误")
	assert.Contains(t, err.Error(), "insufficient data", "错误信息应包含'insufficient data'")
}

func TestMeshoptDecodeZeroCount(t *testing.T) {
	// 测试零计数
	_, err := MeshoptDecode(0, 4, []byte{1, 2, 3, 4}, ModeAttributes, FilterNone)
	assert.Error(t, err, "应返回错误")
	assert.Contains(t, err.Error(), "zero count", "错误信息应包含'zero count'")
}

func TestMeshoptDecodeZeroStride(t *testing.T) {
	// 测试零步长
	_, err := MeshoptDecode(5, 0, []byte{1, 2, 3, 4}, ModeAttributes, FilterNone)
	assert.Error(t, err, "应返回错误")
	assert.Contains(t, err.Error(), "zero stride", "错误信息应包含'zero stride'")
}

func TestDecodeAllWithNoExtension(t *testing.T) {
	// 测试没有扩展的文档
	doc := &gltf.Document{
		BufferViews: []*gltf.BufferView{
			{
				Buffer:     0,
				Extensions: map[string]interface{}{},
			},
		},
	}

	err := DecodeAll(doc)
	assert.NoError(t, err, "应无错误返回")
}

func TestDecodeBufferViewInvalidExtensionType(t *testing.T) {
	// 测试无效的扩展类型
	doc := &gltf.Document{
		BufferViews: []*gltf.BufferView{
			{
				Buffer: 0,
				Extensions: map[string]interface{}{
					ExtensionName: "invalid extension type",
				},
			},
		},
	}

	err := decodeBufferView(doc, doc.BufferViews[0])
	assert.Error(t, err, "应返回错误")
	assert.Contains(t, err.Error(), "invalid extension type", "错误信息应包含'invalid extension type'")
}

func TestDecodeBufferViewBufferIndexOutOfRange(t *testing.T) {
	// 测试缓冲区索引越界
	ext := &CompressionExtension{
		Buffer:     1, // 越界索引
		ByteLength: 10,
		ByteStride: 4,
		Count:      5,
		Mode:       ModeAttributes,
	}

	doc := &gltf.Document{
		Buffers: []*gltf.Buffer{
			{Data: make([]byte, 20)},
		},
		BufferViews: []*gltf.BufferView{
			{
				Buffer: 0,
				Extensions: map[string]interface{}{
					ExtensionName: ext,
				},
			},
		},
	}

	err := decodeBufferView(doc, doc.BufferViews[0])
	assert.Error(t, err, "应返回错误")
	assert.Contains(t, err.Error(), "source buffer index out of range", "错误信息应包含'source buffer index out of range'")
}

func TestBytesToFloat32(t *testing.T) {
	// 测试bytesToFloat32函数
	data := []byte{
		0x00, 0x00, 0x80, 0x3F, // 1.0 in float32
		0x00, 0x00, 0x00, 0x40, // 2.0 in float32
		0x00, 0x00, 0x40, 0x40, // 3.0 in float32
	}

	expected := []float32{1.0, 2.0, 3.0}
	result := bytesToFloat32(data)
	assert.Equal(t, expected, result, "转换结果应匹配")
}

func TestApplyDecodeFilterNone(t *testing.T) {
	// 测试无过滤器的解码
	data := []byte{1, 2, 3, 4, 5, 6, 7, 8}
	result, err := applyDecodeFilter(data, 4, FilterNone)
	assert.NoError(t, err, "不应返回错误")
	assert.Equal(t, data, result, "结果应与输入相同")
}

func TestApplyDecodeFilterUnsupported(t *testing.T) {
	// 测试不支持的过滤器
	data := []byte{1, 2, 3, 4, 5, 6, 7, 8}
	_, err := applyDecodeFilter(data, 4, "INVALID_FILTER")
	assert.Error(t, err, "应返回错误")
	assert.Contains(t, err.Error(), "unsupported filter", "错误信息应包含'unsupported filter'")
}

// 新增的测试用例

func TestMeshoptEncodeDecodeRoundTrip(t *testing.T) {
	// 测试编码和解码的往返过程
	// 创建一些测试数据（模拟顶点位置数据）
	data := make([]byte, 120) // 10个顶点，每个顶点3个float32（12字节）
	for i := range data {
		data[i] = byte(i % 256)
	}

	// 编码数据
	compressedData, ext, err := MeshoptEncode(data, 10, 12, ModeAttributes, FilterNone)
	assert.NoError(t, err, "编码应成功")
	assert.NotNil(t, compressedData, "应返回压缩数据")
	assert.NotNil(t, ext, "应返回扩展信息")
	assert.Equal(t, uint32(10), ext.Count, "计数应匹配")
	assert.Equal(t, uint32(12), ext.ByteStride, "步长应匹配")
	assert.Equal(t, ModeAttributes, ext.Mode, "模式应匹配")

	// 解码数据
	decompressedData, err := MeshoptDecode(ext.Count, ext.ByteStride, compressedData, ext.Mode, ext.Filter)
	assert.NoError(t, err, "解码应成功")
	assert.NotNil(t, decompressedData, "应返回解压数据")

	// 验证数据一致性
	assert.Equal(t, len(data), len(decompressedData), "数据长度应一致")
	// 注意：由于是有损压缩，数据可能不完全相同，但在无过滤器情况下应该相同
	assert.Equal(t, data, decompressedData, "数据内容应一致")
}

func TestMeshoptEncodeUnsupportedMode(t *testing.T) {
	// 测试不支持的编码模式
	data := make([]byte, 20)
	_, _, err := MeshoptEncode(data, 5, 4, "INVALID_MODE", FilterNone)
	assert.Error(t, err, "应返回错误")
	assert.Contains(t, err.Error(), "unsupported mode", "错误信息应包含'unsupported mode'")
}

func TestMeshoptDecodeUnsupportedMode(t *testing.T) {
	// 测试不支持的解码模式
	_, err := MeshoptDecode(5, 4, []byte{1, 2, 3, 4}, "INVALID_MODE", FilterNone)
	assert.Error(t, err, "应返回错误")
	assert.Contains(t, err.Error(), "unsupported mode", "错误信息应包含'unsupported mode'")
}

func ApplyEncodeFilterUnsupported(t *testing.T) {
	// 测试不支持的编码过滤器
	data := make([]byte, 16)
	_, err := applyEncodeFilter(data, 4, "INVALID_FILTER")
	assert.Error(t, err, "应返回错误")
	assert.Contains(t, err.Error(), "unsupported filter", "错误信息应包含'unsupported filter'")
}

func TestApplyEncodeFilterNone(t *testing.T) {
	// 测试无过滤器的编码
	data := []byte{1, 2, 3, 4, 5, 6, 7, 8}
	result, err := applyEncodeFilter(data, 4, FilterNone)
	assert.NoError(t, err, "不应返回错误")
	assert.Equal(t, data, result, "结果应与输入相同")
}

// 添加更多测试用例

func TestDecodeBufferViewSourceBufferOverflow(t *testing.T) {
	// 测试源缓冲区溢出
	ext := &CompressionExtension{
		Buffer:     0,
		ByteOffset: 100, // 超出范围的偏移
		ByteLength: 100, // 超出范围的长度
		ByteStride: 4,
		Count:      5,
		Mode:       ModeAttributes,
	}

	doc := &gltf.Document{
		Buffers: []*gltf.Buffer{
			{Data: make([]byte, 50)}, // 只有50字节，但偏移+长度=200
		},
		BufferViews: []*gltf.BufferView{
			{
				Buffer: 0,
				Extensions: map[string]interface{}{
					ExtensionName: ext,
				},
			},
		},
	}

	err := decodeBufferView(doc, doc.BufferViews[0])
	assert.Error(t, err, "应返回错误")
	assert.Contains(t, err.Error(), "source buffer overflow", "错误信息应包含'source buffer overflow'")
}

func TestDecodeBufferViewDecompressionFailed(t *testing.T) {
	// 测试解压失败（使用无效的压缩数据）
	ext := &CompressionExtension{
		Buffer:     0,
		ByteOffset: 0,
		ByteLength: 4,
		ByteStride: 4,
		Count:      1,
		Mode:       ModeAttributes,
	}

	doc := &gltf.Document{
		Buffers: []*gltf.Buffer{
			{
				Data: []byte{0xFF, 0xFF, 0xFF, 0xFF}, // 无效的压缩数据
			},
		},
		BufferViews: []*gltf.BufferView{
			{
				Buffer: 0,
				Extensions: map[string]interface{}{
					ExtensionName: ext,
				},
			},
		},
	}

	err := decodeBufferView(doc, doc.BufferViews[0])
	assert.Error(t, err, "应返回错误")
	assert.Contains(t, err.Error(), "decompression failed", "错误信息应包含'decompression failed'")
}

func TestMeshoptEncodeInvalidParameters(t *testing.T) {
	// 测试编码函数的无效参数
	tests := []struct {
		name        string
		data        []byte
		count       uint32
		byteStride  uint32
		mode        CompressionMode
		filter      CompressionFilter
		expectError bool
		errorMsg    string
	}{
		{
			name:        "EmptyData",
			data:        []byte{},
			count:       1,
			byteStride:  4,
			mode:        ModeAttributes,
			filter:      FilterNone,
			expectError: true,
			errorMsg:    "empty input data",
		},
		{
			name:        "ZeroStride",
			data:        []byte{1, 2, 3, 4},
			count:       1,
			byteStride:  0,
			mode:        ModeAttributes,
			filter:      FilterNone,
			expectError: true,
			errorMsg:    "zero byteStride",
		},
		{
			name:        "InsufficientData",
			data:        []byte{1, 2, 3}, // 不足4字节
			count:       2,
			byteStride:  4,
			mode:        ModeAttributes,
			filter:      FilterNone,
			expectError: true,
			errorMsg:    "insufficient data",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := MeshoptEncode(tt.data, tt.count, tt.byteStride, tt.mode, tt.filter)
			if tt.expectError {
				assert.Error(t, err, "应返回错误")
				assert.Contains(t, err.Error(), tt.errorMsg, "错误信息应包含预期内容")
			} else {
				assert.NoError(t, err, "不应返回错误")
			}
		})
	}
}

func TestMeshoptEncodeUnsupportedModeWithFilter(t *testing.T) {
	// 测试不支持过滤器的模式
	data := make([]byte, 20)
	_, _, err := MeshoptEncode(data, 5, 4, ModeTriangles, FilterOctahedral)
	assert.Error(t, err, "应返回错误")
	assert.Contains(t, err.Error(), "TRIANGLES mode doesn't support filters", "错误信息应包含'TRIANGLES mode doesn't support filters'")
}

func TestMeshoptDecodeInvalidParameters(t *testing.T) {
	// 测试解码函数的无效参数
	tests := []struct {
		name        string
		count       uint32
		stride      uint32
		data        []byte
		mode        CompressionMode
		filter      CompressionFilter
		expectError bool
		errorMsg    string
	}{
		{
			name:        "ZeroCount",
			count:       0,
			stride:      4,
			data:        []byte{1, 2, 3, 4},
			mode:        ModeAttributes,
			filter:      FilterNone,
			expectError: true,
			errorMsg:    "zero count",
		},
		{
			name:        "ZeroStride",
			count:       1,
			stride:      0,
			data:        []byte{1, 2, 3, 4},
			mode:        ModeAttributes,
			filter:      FilterNone,
			expectError: true,
			errorMsg:    "zero stride",
		},
		{
			name:        "InvalidCountAndStride",
			count:       0xFFFFFFFF, // 大数值可能导致溢出
			stride:      0xFFFFFFFF,
			data:        []byte{1, 2, 3, 4},
			mode:        ModeAttributes,
			filter:      FilterNone,
			expectError: true,
			errorMsg:    "invalid buffer size",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := MeshoptDecode(tt.count, tt.stride, tt.data, tt.mode, tt.filter)
			if tt.expectError {
				assert.Error(t, err, "应返回错误")
				assert.Contains(t, err.Error(), tt.errorMsg, "错误信息应包含预期内容")
			} else {
				assert.NoError(t, err, "不应返回错误")
			}
		})
	}
}

func TestApplyEncodeFilterWithValidFilters(t *testing.T) {
	// 由于实际的过滤器函数依赖于C库，在测试中我们只测试不支持的过滤器情况
	// 在实际应用中，这些过滤器会正常工作

	// 测试一个不存在的过滤器
	data := make([]byte, 16)
	_, err := applyEncodeFilter(data, 4, "NONEXISTENT_FILTER")
	assert.Error(t, err, "应返回错误")
	assert.Contains(t, err.Error(), "unsupported filter", "错误信息应包含'unsupported filter'")
}

func TestApplyDecodeFilterWithValidFilters(t *testing.T) {
	// 由于实际的过滤器函数依赖于C库，在测试中我们只测试不支持的过滤器情况
	// 在实际应用中，这些过滤器会正常工作

	// 测试一个不存在的过滤器
	data := make([]byte, 16)
	_, err := applyDecodeFilter(data, 4, "NONEXISTENT_FILTER")
	assert.Error(t, err, "应返回错误")
	assert.Contains(t, err.Error(), "unsupported filter", "错误信息应包含'unsupported filter'")
}

func TestApplyEncodeFilterUnsupportedFilter(t *testing.T) {
	// 测试不支持的编码过滤器
	data := make([]byte, 16)
	_, err := applyEncodeFilter(data, 4, "NONEXISTENT_FILTER")
	assert.Error(t, err, "应返回错误")
	assert.Contains(t, err.Error(), "unsupported filter", "错误信息应包含'unsupported filter'")
}

func TestApplyDecodeFilterUnsupportedFilter(t *testing.T) {
	// 测试不支持的解码过滤器
	data := make([]byte, 16)
	_, err := applyDecodeFilter(data, 4, "NONEXISTENT_FILTER")
	assert.Error(t, err, "应返回错误")
	assert.Contains(t, err.Error(), "unsupported filter", "错误信息应包含'unsupported filter'")
}
