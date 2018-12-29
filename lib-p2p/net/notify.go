package net

import (
	"gx/ipfs/QmPjvxTpVH8qJyQDnxnsxF9kv9jezKD1kozz1hs3fCGsNh/go-libp2p-net"
	"gx/ipfs/QmYmsdtJ3HsodkePE3eU3TsCaP2YvPZJ4LoXnNkDE5Tpt7/go-multiaddr"
)

func (i *Instance) Listen(n net.Network, a multiaddr.Multiaddr)      { log.Debug("Listen") }
func (i *Instance) ListenClose(n net.Network, a multiaddr.Multiaddr) { log.Debug("ListenClose") }
func (i *Instance) Connected(n net.Network, v net.Conn) {
	log.Debug("Connected ID:", v.RemotePeer().Pretty(), "Addr:", n.Peerstore().Addrs(v.RemotePeer()))
}
func (i *Instance) Disconnected(n net.Network, v net.Conn) {
	log.Debug("Disconnected:", n.Peerstore().Addrs(v.RemotePeer()))

}
func (i *Instance) OpenedStream(n net.Network, v net.Stream) {
	//id := v.Conn().RemotePeer()
	//addresses :=  n.Peerstore().Addrs(v.Conn().RemotePeer())
	//log.Debug("OpenedStream", id.Pretty(), addresses, v.Conn().RemoteMultiaddr(), v)
}
func (i *Instance) ClosedStream(n net.Network, v net.Stream) {
	//log.Debug("ClosedStream", n.Peerstore().Addrs(v.Conn().RemotePeer()), v)
}
