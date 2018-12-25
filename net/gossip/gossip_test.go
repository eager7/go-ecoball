package gossip_test

import (
	"context"
	"github.com/ecoball/go-ecoball/common/errors"
	"github.com/ecoball/go-ecoball/net"
	"github.com/ecoball/go-ecoball/net/message"
	"github.com/ecoball/go-ecoball/net/message/pb"
	"github.com/ecoball/go-ecoball/net/network"
	"github.com/ecoball/go-ecoball/test/net/gossippull"
	"gx/ipfs/QmcJukH2sAFjY3HdBKq35WDzWoL3UUu2gt9wdfqZTUyM74/go-libp2p-peer"
	"testing"
)

func TestGossip(t *testing.T) {
	ctx := context.Background()
	net.InitNetWork(ctx)
	net.StartNetWork(nil)
	if m := gossippull.StartBlockPuller(ctx); m == nil {
		t.Fatal("net not initialize")
	}

}

func SendMessage(p peer.ID) {
	msg := message.New(pb.MsgType_APP_MSG_UNDEFINED, []byte{1, 2, 3, 4, 5, 6, 7, 8, 9})
	netInst, err := network.GetNetInstance()
	errors.CheckErrorPanic(err)
	//netInst.SendMsgToPeerWithId(, msg)
}
