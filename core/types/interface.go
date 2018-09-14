package types

type Payload interface {
	Serialize() ([]byte, error)
	Deserialize(data []byte) error
	GetObject() interface{}
	Type() uint32
	JsonString() string
}

type HeaderType uint32

const (
	HeMinorBlock HeaderType = 1
	HeCmBlock    HeaderType = 2
	HeFinalBlock HeaderType = 3
)

type HeInterface interface {

}

type BInterface interface {
	
}