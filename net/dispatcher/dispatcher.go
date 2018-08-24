package dispatcher

import (
	"fmt"
	"github.com/ecoball/go-ecoball/net/message"
	"gx/ipfs/QmdbxjQWogRCHRaxhhGnYdT1oQJzL9GdqSKzCdqWr85AP2/pubsub"
	"gx/ipfs/QmdVrMn1LhB4ybb8hMVaMLXnA8XRSewMnK6YqXKXoTcRvN/go-libp2p-peer"
)

const (
	bufferSize = 16
	errorStr = "dispatcher is not ready"
)

var (
	dispatcher *Dispatcher
)

func InitMsgDispatcher(sender MsgNode){
	if dispatcher == nil {
		dispatcher = &Dispatcher{
			pubsub.New(bufferSize),
			sender,
		}
	}
}

type Dispatcher struct {
	ps     *pubsub.PubSub
	sender MsgNode
}

func (ds *Dispatcher) publish(msg message.EcoBallNetMsg) {
	ds.ps.Pub(msg, message.MessageToStr[msg.Type()])
}

func (ds *Dispatcher) subscribe(msgType ...uint32) <-chan interface{} {
	var msgstr []string
	for _, msg := range msgType {
		msgstr = append(msgstr, message.MessageToStr[msg])
	}
	if len(msgstr) > 0 {
		return ds.ps.Sub(msgstr...)
	}

	return nil
}

func (ds *Dispatcher) unsubscribe(chn chan interface{}, msgType ...uint32) {
	var msgstr []string
	for _, msg := range msgType {
		msgstr = append(msgstr, message.MessageToStr[msg])
	}

	ds.ps.Unsub(chn, msgstr...)
}

// Not safe to call more than once.
func (ds *Dispatcher) shutdown() {
	// shutdown the pubsub.
	ds.ps.Shutdown()
}

func Subscribe (msgType ...uint32) (<-chan interface{}, error) {
	if dispatcher == nil {
		return nil, fmt.Errorf(errorStr)
	}
	return dispatcher.subscribe(msgType...), nil
}

func UnSubscribe (chn chan interface{}, msgs ...uint32) error {
	if dispatcher == nil {
		return fmt.Errorf(errorStr)
	}
	dispatcher.unsubscribe(chn, msgs...)

	return nil
}

func Publish (msg message.EcoBallNetMsg) error {
	if dispatcher == nil {
		return fmt.Errorf(errorStr)
	}
	dispatcher.publish(msg)

	return nil
}

func SendMessage (pid peer.ID, msg message.EcoBallNetMsg) error {
	if dispatcher == nil {
		return fmt.Errorf(errorStr)
	}
	dispatcher.sender.SendMsg2Peer(pid, msg)

	return nil
}

func SendMsgToRandomPeers (peerCounts int, msg message.EcoBallNetMsg) error {
	if dispatcher == nil {
		return fmt.Errorf(errorStr)
	}
	dispatcher.sender.SendMsg2RandomPeers(peerCounts, msg)

	return nil
}

func BroadcastMessage (msg message.EcoBallNetMsg) error {
	if dispatcher == nil {
		return fmt.Errorf(errorStr)
	}
	dispatcher.sender.SendBroadcastMsg(msg)

	return nil
}

func GetPeerID () (peer.ID, error) {
	if dispatcher == nil {
		return "", fmt.Errorf(errorStr)
	}
	return dispatcher.sender.SelfRawId(), nil
}

func GetRandomPeers (k int) ([]peer.ID, error) {
	if dispatcher == nil {
		return []peer.ID{}, fmt.Errorf(errorStr)
	}
	return dispatcher.sender.SelectRandomPeers(k), nil
}