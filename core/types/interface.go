package types

import "github.com/ecoball/go-ecoball/common/message/mpb"

type EcoMessage interface {
	Identify() mpb.Identify
	String() string
	GetInstance() interface{}
	Serialize() ([]byte, error)
	Deserialize(data []byte) error
}