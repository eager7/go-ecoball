package net

import "github.com/ecoball/go-ecoball/net/message/pb"

type Network interface {
	SendMessage(b64Pub, address, port string, payload *pb.Message) error
}
