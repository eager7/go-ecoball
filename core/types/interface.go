package types


type Payload interface {
	Serialize() ([]byte, error)
	Deserialize(data []byte) error
	GetObject() interface{}
	Type() uint32
	JsonString() string
}
