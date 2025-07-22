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

// 新增压缩模式和过滤器类型
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
		return nil, fmt.Errorf("EXT_meshopt_compression解析失败: %w", err)
	}
	return ext, nil
}

func DecodeAll(doc *gltf.Document) error {
	for i, bufView := range doc.BufferViews {
		if err := decodeBufferView(doc, bufView); err != nil {
			return fmt.Errorf("缓冲区视图%d解码失败: %w", i, err)
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
		return errors.New("无效的扩展格式")
	}

	// 关键修复1：使用压缩数据源buffer
	if int(ext.Buffer) >= len(doc.Buffers) {
		return fmt.Errorf("无效的buffer索引: %d", ext.Buffer)
	}
	srcBuffer := doc.Buffers[ext.Buffer]

	// 关键修复2：处理没有URI的buffer
	if int(bufView.Buffer) >= len(doc.Buffers) {
		return fmt.Errorf("无效的目标buffer索引: %d", bufView.Buffer)
	}
	dstBuffer := doc.Buffers[bufView.Buffer]

	// 确保目标buffer有数据存储空间
	if dstBuffer.Data == nil {
		dstBuffer.Data = make([]byte, dstBuffer.ByteLength)
	}

	// 验证字节范围有效性
	if int(ext.ByteOffset+ext.ByteLength) > len(srcBuffer.Data) {
		return fmt.Errorf("字节范围越界 [%d-%d] (buffer长度:%d)",
			ext.ByteOffset, ext.ByteOffset+ext.ByteLength, len(srcBuffer.Data))
	}

	srcData := srcBuffer.Data[ext.ByteOffset : ext.ByteOffset+ext.ByteLength]

	dstData, err := MeshoptDecode(
		ext.Count,
		ext.ByteStride,
		srcData,
		ext.Mode,
		ext.Filter,
	)
	if err != nil {
		return fmt.Errorf("解压失败: %w", err)
	}

	// 关键修复3：确保目标buffer有足够容量
	requiredLen := int(bufView.ByteOffset) + len(dstData)
	if len(dstBuffer.Data) < requiredLen {
		newData := make([]byte, requiredLen)
		copy(newData, dstBuffer.Data)
		dstBuffer.Data = newData
		dstBuffer.ByteLength = uint32(requiredLen)
	}

	copy(dstBuffer.Data[bufView.ByteOffset:], dstData)
	delete(bufView.Extensions, ExtensionName)
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
		return nil, nil, errors.New("空输入数据")
	}
	if byteStride == 0 {
		return nil, nil, errors.New("无效的byteStride值")
	}
	if expectedLen := int(count) * int(byteStride); len(data) < expectedLen {
		return nil, nil, fmt.Errorf("数据长度不足: 需要%d字节, 实际%d字节", expectedLen, len(data))
	}

	var buf bytes.Buffer
	// 根据压缩模式调用不同的压缩方法
	switch mode {
	case ModeAttributes:
		// 处理顶点属性过滤器
		if filter != FilterNone {
			decoded, err := applyFilter(data, byteStride, filter)
			if err != nil {
				return nil, nil, err
			}
			data = decoded
		}
		if err := meshopt.CompressVertexStream(&buf, data, int(count), int(byteStride)); err != nil {
			return nil, nil, fmt.Errorf("顶点压缩失败: %w", err)
		}

	case ModeTriangles:
		if byteStride != 2 && byteStride != 4 {
			return nil, nil, errors.New("三角形模式只支持2或4字节步长")
		}
		// 三角形模式暂不需要filter处理
		if filter != FilterNone {
			return nil, nil, errors.New("TRIANGLES模式不支持过滤器")
		}
		if err := meshopt.CompressIndexStream(&buf, data, int(count), int(byteStride)); err != nil {
			return nil, nil, fmt.Errorf("索引流压缩失败: %w", err)
		}

	case ModeIndices:
		if byteStride != 2 && byteStride != 4 {
			return nil, nil, errors.New("索引模式只支持2或4字节步长")
		}
		// 处理索引过滤器
		if filter != FilterNone {
			decoded, err := applyFilter(data, byteStride, filter)
			if err != nil {
				return nil, nil, err
			}
			data = decoded
		}
		if err := meshopt.CompressIndexSequence(&buf, data, int(count), int(byteStride)); err != nil {
			return nil, nil, fmt.Errorf("索引序列压缩失败: %w", err)
		}

	default:
		return nil, nil, fmt.Errorf("不支持的压缩模式: %s", mode)
	}

	// 构建扩展配置对象 (Buffer和ByteOffset由调用方后续填充)
	ext = &CompressionExtension{
		ByteLength: uint32(buf.Len()),
		ByteStride: byteStride,
		Count:      count,
		Mode:       mode,
		Filter:     filter,
	}

	return buf.Bytes(), ext, nil
}

func applyFilter(data []byte, stride uint32, filter CompressionFilter) ([]byte, error) {
	// 转换为float32切片用于编码过滤器
	floatData := bytesToFloat32(data)

	dst := make([]byte, len(data))
	count := len(floatData) / int(stride/4) // stride以字节为单位，float32占4字节

	switch filter {
	case FilterOctahedral:
		meshopt.CompressFilterOct(
			dst,
			count,
			int(stride),
			12,
			floatData,
		)
	case FilterQuaternion:
		meshopt.CompressFilterQuat(
			dst,
			count,
			int(stride),
			12,
			floatData,
		)
	case FilterExponential:
		meshopt.CompressFilterExp(
			dst,
			count,
			int(stride),
			15,
			floatData,
		)
	case FilterNone:
		return data, nil
	default:
		return nil, fmt.Errorf("不支持的过滤器类型: %s", filter)
	}
	return dst, nil
}

func decodeBufferViewWithFilter(data []byte, filter CompressionFilter, stride uint32) ([]byte, error) {
	count := len(data) / int(stride)

	switch filter {
	case FilterOctahedral:
		meshopt.DecompressFilterOct(data, count, int(stride))
	case FilterQuaternion:
		meshopt.DecompressFilterQuat(data, count, int(stride))
	case FilterExponential:
		meshopt.DecompressFilterExp(data, count, int(stride))
	case FilterNone:
		return data, nil
	default:
		return nil, fmt.Errorf("不支持的过滤器类型: %s", filter)
	}
	return data, nil
}

func bytesToFloat32(b []byte) []float32 {
	floatSlice := make([]float32, len(b)/4)
	for i := 0; i < len(floatSlice); i++ {
		floatSlice[i] = *(*float32)(unsafe.Pointer(&b[i*4]))
	}
	return floatSlice
}

func MeshoptDecode(count uint32, stride uint32, data []byte, mode CompressionMode, filter CompressionFilter) ([]byte, error) {
	if count == 0 {
		return nil, errors.New("无效的count值")
	}
	if stride == 0 {
		return nil, errors.New("无效的stride值")
	}

	expectedSize := int(count) * int(stride)
	if expectedSize <= 0 {
		return nil, fmt.Errorf("无效的缓冲区大小计算 count:%d stride:%d", count, stride)
	}

	dst := make([]byte, expectedSize)

	var err error
	switch mode {
	case ModeAttributes:
		err = meshopt.DecompressVertexStream(
			dst,
			int(count),
			int(stride),
			data,
		)
	case ModeTriangles:
		strideInt := int(stride)
		if strideInt != 2 && strideInt != 4 {
			return nil, errors.New("invalid stride for index decompression")
		}
		err = meshopt.DecompressIndexStream(
			dst,
			int(count),
			strideInt,
			data,
		)
	case ModeIndices:
		strideInt := int(stride)
		if strideInt != 2 && strideInt != 4 {
			return nil, errors.New("invalid stride for index sequence decompression")
		}
		err = meshopt.DecompressIndexSequence(
			dst,
			int(count),
			strideInt,
			data,
		)
	default:
		return nil, fmt.Errorf("unsupported compression mode: %s", mode)
	}

	if err != nil {
		return nil, fmt.Errorf("解压失败: %w", err)
	}

	if len(dst) != expectedSize {
		return nil, fmt.Errorf("解压数据大小不匹配 预期:%d 实际:%d",
			expectedSize, len(dst))
	}

	if filter != FilterNone && filter != "" {
		return decodeBufferViewWithFilter(dst, filter, stride)
	}
	return dst, nil
}
