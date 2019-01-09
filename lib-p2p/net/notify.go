package net

import (
	"gx/ipfs/QmPjvxTpVH8qJyQDnxnsxF9kv9jezKD1kozz1hs3fCGsNh/go-libp2p-net"
	"gx/ipfs/QmYmsdtJ3HsodkePE3eU3TsCaP2YvPZJ4LoXnNkDE5Tpt7/go-multiaddr"
)

func (i *Instance) Listen(n net.Network, a multiaddr.Multiaddr)      { log.Debug("Listen...") }
func (i *Instance) ListenClose(n net.Network, a multiaddr.Multiaddr) { log.Debug("ListenClose...") }
func (i *Instance) Connected(n net.Network, v net.Conn) {
	log.Debug("Connected ID:", v.RemotePeer().Pretty(), "Addr:", n.Peerstore().Addrs(v.RemotePeer()))
}
func (i *Instance) Disconnected(n net.Network, v net.Conn) {
	log.Debug("Disconnected:", n.Peerstore().Addrs(v.RemotePeer()))

}
func (i *Instance) OpenedStream(n net.Network, v net.Stream) {
	log.Debug("OpenedStream", v.Conn().RemotePeer().Pretty(), v.Conn().RemoteMultiaddr())
}
func (i *Instance) ClosedStream(n net.Network, v net.Stream) {
	log.Debug("ClosedStream", v.Conn().RemotePeer().Pretty(), n.Peerstore().Addrs(v.Conn().RemotePeer()), v)
}
