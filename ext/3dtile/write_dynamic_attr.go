package tile3d

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/flywave/gltf"
	dynamicattr "github.com/flywave/gltf/ext/3dtile/dynamic_attr"
	"github.com/vmihailenco/msgpack"
)

func WriteDynamicAttributes(doc *gltf.Document, features []map[string]interface{}, encoding dynamicattr.DynamicAttrEncoding) error {
	if len(features) == 0 {
		return fmt.Errorf("features cannot be empty")
	}
	if len(doc.Buffers) == 0 {
		return fmt.Errorf("glTF document has no buffers")
	}

	dynamics := []*dynamicattr.DynamicAttrExtension{}
	for i, mp := range features {
		var (
			dataBytes []byte
			err       error
		)
		switch encoding {
		case dynamicattr.DynamicAttrEncodingJSON:
			dataBytes, err = serializeToJSON(mp)
		case dynamicattr.DynamicAttrEncodingMsgpack:
			dataBytes, err = serializeToMsgPack(mp)
		default:
			return fmt.Errorf("unsupported encoding: %s", encoding)
		}
		if err != nil {
			return fmt.Errorf("serialization failed: %v", err)
		}

		buffer := doc.Buffers[0]
		byteOffset := buffer.ByteLength
		buffer.Data = append(buffer.Data, dataBytes...)
		buffer.ByteLength += uint32(len(dataBytes))
		padBuffer(doc)

		bufferView := &gltf.BufferView{
			Buffer:     0,
			ByteOffset: byteOffset,
			ByteLength: uint32(len(dataBytes)),
		}
		bufferViewIndex := uint32(len(doc.BufferViews))
		doc.BufferViews = append(doc.BufferViews, bufferView)

		if doc.Extensions == nil {
			doc.Extensions = make(map[string]interface{})
		}
		dynamics = append(dynamics, &dynamicattr.DynamicAttrExtension{
			Encoding:   encoding,
			BufferView: bufferViewIndex,
			FeatureID:  uint32(i),
		})
	}

	if doc.Extensions == nil {
		doc.Extensions = make(map[string]interface{})
	}
	doc.Extensions[dynamicattr.ExtensionName] = dynamics
	addExtensionUsed(doc, dynamicattr.ExtensionName)
	return nil
}

func serializeToJSON(features map[string]interface{}) ([]byte, error) {
	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	for _, feature := range features {
		if err := encoder.Encode(feature); err != nil {
			return nil, err
		}
	}
	return buf.Bytes(), nil
}

func serializeToMsgPack(features map[string]interface{}) ([]byte, error) {
	var buf bytes.Buffer
	encoder := msgpack.NewEncoder(&buf)
	for _, feature := range features {
		if err := encoder.Encode(feature); err != nil {
			return nil, err
		}
	}
	return buf.Bytes(), nil
}
