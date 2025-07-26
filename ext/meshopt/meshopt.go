package meshopt

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"unsafe"

	"github.com/flywave/go-meshopt"

	"github.com/flywave/gltf"
)

const ExtensionName = "EXT_meshopt_compression"

type CompressionMode string
type CompressionFilter string

const (
	ModeAttributes CompressionMode = "ATTRIBUTES"
	ModeTriangles  CompressionMode = "TRIANGLES"
	ModeIndices    CompressionMode = "INDICES"
)

const (
	FilterNone        CompressionFilter = "NONE"
	FilterOctahedral  CompressionFilter = "OCTAHEDRAL"
	FilterQuaternion  CompressionFilter = "QUATERNION"
	FilterExponential CompressionFilter = "EXPONENTIAL"
)

type CompressionExtension struct {
	Buffer     uint32            `json:"buffer"`
	ByteOffset uint32            `json:"byteOffset,omitempty"`
	ByteLength uint32            `json:"byteLength"`
	ByteStride uint32            `json:"byteStride"`
	Count      uint32            `json:"count"`
	Mode       CompressionMode   `json:"mode"`
	Filter     CompressionFilter `json:"filter,omitempty"`
}

func init() {
	gltf.RegisterExtension(ExtensionName, Unmarshal)
}

func Unmarshal(data []byte) (interface{}, error) {
	ext := &CompressionExtension{}
	if err := json.Unmarshal(data, ext); err != nil {
		return nil, fmt.Errorf("EXT_meshopt_compression unmarshal error: %w", err)
	}
	return ext, nil
}

func DecodeAll(doc *gltf.Document) error {
	for i := range doc.BufferViews {
		if err := decodeBufferView(doc, doc.BufferViews[i]); err != nil {
			return fmt.Errorf("bufferView[%d]: %w", i, err)
		}
	}
	return nil
}

func decodeBufferView(doc *gltf.Document, bufView *gltf.BufferView) error {
	extData, exists := bufView.Extensions[ExtensionName]
	if !exists {
		return nil
	}

	ext, ok := extData.(*CompressionExtension)
	if !ok {
		return errors.New("invalid extension type")
	}

	// 验证模式与步长
	if err := validateModeStride(ext.Mode, ext.ByteStride); err != nil {
		return fmt.Errorf("invalid stride: %w", err)
	}

	// 获取源缓冲区（压缩数据）
	if int(ext.Buffer) >= len(doc.Buffers) {
		return fmt.Errorf("source buffer index out of range: %d", ext.Buffer)
	}
	srcBuffer := doc.Buffers[ext.Buffer]

	// 获取目标缓冲区（解压位置）
	if int(bufView.Buffer) >= len(doc.Buffers) {
		return fmt.Errorf("target buffer index out of range: %d", bufView.Buffer)
	}
	dstBuffer := doc.Buffers[bufView.Buffer]

	// 验证字节范围
	srcStart := ext.ByteOffset
	srcEnd := srcStart + ext.ByteLength
	if srcEnd > uint32(len(srcBuffer.Data)) {
		return fmt.Errorf("source buffer overflow: %d > %d", srcEnd, len(srcBuffer.Data))
	}

	// 解压数据
	dstData, err := MeshoptDecode(
		ext.Count,
		ext.ByteStride,
		srcBuffer.Data[srcStart:srcEnd],
		ext.Mode,
		ext.Filter,
	)
	if err != nil {
		return fmt.Errorf("decompression failed: %w", err)
	}

	// 准备目标缓冲区
	dstStart := bufView.ByteOffset
	dstEnd := dstStart + uint32(len(dstData))
	if dstEnd > uint32(len(dstBuffer.Data)) {
		newData := make([]byte, dstEnd)
		copy(newData, dstBuffer.Data)
		dstBuffer.Data = newData
		dstBuffer.ByteLength = dstEnd
	}

	// 写入解压数据
	copy(dstBuffer.Data[dstStart:], dstData)
	delete(bufView.Extensions, ExtensionName)
	return nil
}

func validateModeStride(mode CompressionMode, stride uint32) error {
	switch mode {
	case ModeAttributes:
		if stride%4 != 0 || stride > 256 {
			return fmt.Errorf("stride must be divisible by 4 and <= 256 (got %d)", stride)
		}
	case ModeTriangles, ModeIndices:
		if stride != 2 && stride != 4 {
			return fmt.Errorf("stride must be 2 or 4 (got %d)", stride)
		}
	default:
		return fmt.Errorf("unknown mode: %s", mode)
	}
	return nil
}

func MeshoptEncode(
	data []byte,
	count uint32,
	byteStride uint32,
	mode CompressionMode,
	filter CompressionFilter,
) (compressedData []byte, ext *CompressionExtension, err error) {
	// 验证基本参数
	if len(data) == 0 {
		return nil, nil, errors.New("empty input data")
	}
	if byteStride == 0 {
		return nil, nil, errors.New("zero byteStride")
	}
	if expected := int(count) * int(byteStride); len(data) < expected {
		return nil, nil, fmt.Errorf("insufficient data: need %d bytes, got %d", expected, len(data))
	}

	var buf bytes.Buffer
	switch mode {
	case ModeAttributes:
		// 应用预处理过滤器
		if filter != FilterNone {
			data, err = applyEncodeFilter(data, byteStride, filter)
			if err != nil {
				return nil, nil, err
			}
		}
		err = meshopt.CompressVertexStream(&buf, data, int(count), int(byteStride))
	case ModeTriangles:
		if filter != FilterNone {
			return nil, nil, errors.New("TRIANGLES mode doesn't support filters")
		}
		err = meshopt.CompressIndexStream(&buf, data, int(count), int(byteStride))
	case ModeIndices:
		if filter != FilterNone {
			data, err = applyEncodeFilter(data, byteStride, filter)
			if err != nil {
				return nil, nil, err
			}
		}
		err = meshopt.CompressIndexSequence(&buf, data, int(count), int(byteStride))
	default:
		return nil, nil, fmt.Errorf("unsupported mode: %s", mode)
	}

	if err != nil {
		return nil, nil, fmt.Errorf("compression error: %w", err)
	}

	ext = &CompressionExtension{
		ByteLength: uint32(buf.Len()),
		ByteStride: byteStride,
		Count:      count,
		Mode:       mode,
		Filter:     filter,
	}
	return buf.Bytes(), ext, nil
}

func applyEncodeFilter(data []byte, stride uint32, filter CompressionFilter) ([]byte, error) {
	// 转换为float32切片用于编码过滤器
	floatData := bytesToFloat32(data)
	dst := make([]byte, len(data))
	count := len(floatData) / int(stride/4) // stride以字节为单位，float32占4字节

	switch filter {
	case FilterOctahedral:
		meshopt.CompressFilterOct(dst, count, int(stride), 12, floatData)
	case FilterQuaternion:
		meshopt.CompressFilterQuat(dst, count, int(stride), 12, floatData)
	case FilterExponential:
		meshopt.CompressFilterExp(dst, count, int(stride), 15, floatData)
	case FilterNone:
		return data, nil
	default:
		return nil, fmt.Errorf("unsupported filter: %s", filter)
	}
	return dst, nil
}

func MeshoptDecode(count uint32, stride uint32, data []byte, mode CompressionMode, filter CompressionFilter) ([]byte, error) {
	// 验证参数
	if count == 0 {
		return nil, errors.New("zero count")
	}
	if stride == 0 {
		return nil, errors.New("zero stride")
	}

	expectedSize := int(count) * int(stride)
	if expectedSize <= 0 {
		return nil, fmt.Errorf("invalid buffer size: count=%d stride=%d", count, stride)
	}

	dst := make([]byte, expectedSize)
	var err error

	// 根据模式解码
	switch mode {
	case ModeAttributes:
		err = meshopt.DecompressVertexStream(dst, int(count), int(stride), data)
	case ModeTriangles:
		err = meshopt.DecompressIndexStream(dst, int(count), int(stride), data)
	case ModeIndices:
		err = meshopt.DecompressIndexSequence(dst, int(count), int(stride), data)
	default:
		return nil, fmt.Errorf("unsupported mode: %s", mode)
	}

	if err != nil {
		return nil, fmt.Errorf("decompression failed: %w", err)
	}

	// 应用后处理过滤器
	if filter != FilterNone {
		return applyDecodeFilter(dst, stride, filter)
	}
	return dst, nil
}

func applyDecodeFilter(data []byte, stride uint32, filter CompressionFilter) ([]byte, error) {
	count := len(data) / int(stride)

	switch filter {
	case FilterOctahedral:
		meshopt.DecompressFilterOct(data, count, int(stride))
	case FilterQuaternion:
		meshopt.DecompressFilterQuat(data, count, int(stride))
	case FilterExponential:
		meshopt.DecompressFilterExp(data, count, int(stride))
	case FilterNone, "":
		return data, nil
	default:
		return nil, fmt.Errorf("unsupported filter: %s", filter)
	}
	return data, nil
}

func bytesToFloat32(b []byte) []float32 {
	count := len(b) / 4
	result := make([]float32, count)

	// 高性能转换
	src := unsafe.Slice((*float32)(unsafe.Pointer(&b[0])), count)
	copy(result, src)
	return result
}
