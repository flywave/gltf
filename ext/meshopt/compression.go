package meshopt

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"

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

	srcBuffer := doc.Buffers[ext.Buffer]
	dstBuffer := doc.Buffers[bufView.Buffer]

	// 获取压缩数据
	srcData := srcBuffer.Data[ext.ByteOffset : ext.ByteOffset+ext.ByteLength]

	// 调用meshopt解码库 (需实现)
	if dstData, err := meshoptDecode(
		ext.Count,
		ext.ByteStride,
		srcData,
		ext.Mode,
		ext.Filter,
	); err != nil {
		return err
	} else {
		copy(dstBuffer.Data[bufView.ByteOffset:], dstData)
	}

	// 移除扩展标记
	delete(bufView.Extensions, ExtensionName)
	return nil
}

func meshoptDecode(count uint32, stride uint32, data []byte, mode CompressionMode, filter CompressionFilter) ([]byte, error) {
	var buf bytes.Buffer

	// TODO decode

	if filter != FilterNone {
		return decodeBufferViewWithFilter(buf.Bytes(), filter, stride)
	}
	return buf.Bytes(), nil
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

// 新增过滤器处理函数
func applyFilter(data []byte, stride uint32, filter CompressionFilter) ([]byte, error) {
	switch filter {
	case FilterOctahedral:
		return encodeOctahedral(data, stride)
	case FilterQuaternion:
		return encodeQuaternion(data)
	case FilterExponential:
		return encodeExponential(data)
	case FilterNone:
		return data, nil
	default:
		return nil, fmt.Errorf("不支持的过滤器类型: %s", filter)
	}
}

func decodeBufferViewWithFilter(data []byte, filter CompressionFilter, stride uint32) ([]byte, error) {
	switch filter {
	case FilterOctahedral:
		return decodeOctahedral(data, stride)
	case FilterQuaternion:
		return decodeQuaternion(data)
	case FilterExponential:
		return decodeExponential(data)
	case FilterNone:
		return data, nil
	default:
		return nil, fmt.Errorf("不支持的过滤器类型: %s", filter)
	}
}
