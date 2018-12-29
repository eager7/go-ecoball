package net

import "github.com/ecoball/go-ecoball/net/message/pb"

type Network interface {
	SendMessage(b64Pub, address, port string, payload *pb.Message) error
}

type Payload interface {
	Serialize() ([]byte, error)
	Deserialize(data []byte) error
}