package main

import (
	"context"
	"github.com/ecoball/go-ecoball/net"
	ma "gx/ipfs/QmYmsdtJ3HsodkePE3eU3TsCaP2YvPZJ4LoXnNkDE5Tpt7/go-multiaddr"
	"github.com/ecoball/go-ecoball/net/network"
	"time"
	"fmt"
	"gx/ipfs/QmdVrMn1LhB4ybb8hMVaMLXnA8XRSewMnK6YqXKXoTcRvN/go-libp2p-peer"
	"github.com/ecoball/go-ecoball/net/message"
)

var netInst network.EcoballNetwork

func gossipMsgTest() {
	msg := message.New(message.APP_MSG_MAX, []byte{1,2,3,4,5,6,7,8,9})
	netInst.GossipMsg(msg)
}

func invalidMsgTest(p peer.ID) {
	msg := message.New(message.APP_MSG_MAX, []byte{1,2,3,4,5,6,7,8,9})
	netInst.SendMsgToPeerWithId(p, msg)
}

func main() {
	ctx := context.Background()
	net.InitNetWork(ctx)

	net.StartNetWork()

	remotePeer := "/ip4/192.168.8.221/tcp/4013"

	id, err :=  peer.IDB58Decode("QmW33JWeTeBhjviWaHLUTt7jNiNj7Z86jngC7ogBZSXmSt")
	if err != nil {
		fmt.Println("failed to decode peer id")
		return
	}
	addr , err := ma.NewMultiaddr(remotePeer)
	if err != nil {
		fmt.Println("failed to create peer address")
		return
	}

	netInst = network.GetNetInstance()
	if netInst == nil {
		fmt.Println("inst of network is nil")
		return
	}
	netInst.Host().Peerstore().AddAddr(id, addr, time.Second * 10)

	fmt.Println("begin to send message ......")
	go func() {
		for {
			invalidMsgTest(id)
			time.Sleep(time.Second * 2)
			gossipMsgTest()
			return
		}
	}()

	for {
		select {
		case <- ctx.Done():
			return
		}
	}
}


