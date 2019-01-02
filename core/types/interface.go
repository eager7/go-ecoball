package types

import "github.com/ecoball/go-ecoball/common/message/mpb"

/*type Payload interface {
	Type() uint32
	GetObject() interface{}
	Serialize() ([]byte, error)
	Deserialize(data []byte) error
	JsonString() string
}*/

type EcoMessage interface {
	Identify() mpb.Identify
	String() string
	GetInstance() interface{}
	Serialize() ([]byte, error)
	Deserialize(data []byte) error
}