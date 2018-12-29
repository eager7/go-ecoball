package net

import (
	"fmt"

	"strconv"
	"strings"
	"gx/ipfs/QmdVrMn1LhB4ybb8hMVaMLXnA8XRSewMnK6YqXKXoTcRvN/go-libp2p-peer"
	"gx/ipfs/Qme1knMqwt1hKZbc1BmQFmnm9f36nyQGwXxPGVpVJ9rMK5/go-libp2p-crypto"
	"github.com/ecoball/go-ecoball/common/errors"
)

func IdFromPublicKey(pubKey string) (peer.ID, error) {
	key, err := crypto.ConfigDecodeKey(pubKey)
	if err != nil {
		return "", errors.New(err.Error())
	}
	pk, err := crypto.UnmarshalPublicKey(key)
	if err != nil {
		return "", errors.New(err.Error())
	}
	id, err := peer.IDFromPublicKey(pk)
	if err != nil {
		return "", errors.New(err.Error())
	}
	return id, nil
}

func NewAddrInfo(ip, port string) (addr string) {
	tcpPort, _ := strconv.Atoi(port)
	if strings.Contains(ip, ":") {
		addr = fmt.Sprintf("/ip6/%s/tcp/%d", ip, tcpPort)
	} else {
		addr = fmt.Sprintf("/ip4/%s/tcp/%d", ip, tcpPort)
	}
	return
}
