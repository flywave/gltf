package dynamicattr

const ExtensionName = "EXT_dynamic_attr_metadata"

type DynamicAttrEncoding string

const (
	DynamicAttrEncodingJSON    DynamicAttrEncoding = "json"
	DynamicAttrEncodingMsgpack DynamicAttrEncoding = "msgpack"
)

type DynamicAttrExtension struct {
	Encoding   DynamicAttrEncoding `json:"encoding"`
	BufferView uint32              `json:"bufferView"`
	RecordSize uint32              `json:"recordSize"`
	FeatureID  uint32              `json:"featureIDs"`
}
