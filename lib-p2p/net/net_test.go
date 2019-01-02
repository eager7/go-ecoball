package net_test

import (
	"testing"
	"fmt"
	"gx/ipfs/QmdVrMn1LhB4ybb8hMVaMLXnA8XRSewMnK6YqXKXoTcRvN/go-libp2p-peer"
	"github.com/ecoball/go-ecoball/lib-p2p/net"
)

func TestPeerMap_Iterator(t *testing.T) {
	p := new(net.SenderMap)
	p.Initialize()
	p.Add(peer.ID("test1"), nil, nil)
	p.Add(peer.ID("test2"), nil, nil)
	p.Add(peer.ID("test3"), nil, nil)
	p.Add(peer.ID("test4"), nil, nil)
	p.Add(peer.ID("test5"), nil, nil)

	fmt.Println(p)

	for v := range p.Iterator() {
		fmt.Println(v)
	}
}
