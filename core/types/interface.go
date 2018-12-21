package types

type Payload interface {
	Type() uint32
	GetObject() interface{}
	Serialize() ([]byte, error)
	Deserialize(data []byte) error
	JsonString() string
}
