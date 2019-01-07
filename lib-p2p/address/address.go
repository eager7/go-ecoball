package address

import (
	"fmt"
	"github.com/ecoball/go-ecoball/common/config"
	"github.com/ecoball/go-ecoball/common/elog"
	"github.com/ecoball/go-ecoball/common/errors"
	"gx/ipfs/QmSMZwvs3n4GBikZ7hKzT17c3bk65FmyZo2JqtJ16swqCv/multiaddr-filter"
	"gx/ipfs/QmSW4uNHbvQia8iZDXzbwjiyHQtnyo9aFqfQAMasj3TJ6Y/go-maddr-filter"
	"gx/ipfs/QmY51bqSM5XgxQZqsBrQcRkKTnCb8EKpJpR9K6Qax7Njco/go-libp2p/p2p/host/basic"
	"gx/ipfs/QmYmsdtJ3HsodkePE3eU3TsCaP2YvPZJ4LoXnNkDE5Tpt7/go-multiaddr"
	"gx/ipfs/QmcQ56iqKP8ZRhRGLe5EReJVvrJZDaGzkuatrPv4Z1B6cG/go-libp2p-circuit"
	"gx/ipfs/QmdVrMn1LhB4ybb8hMVaMLXnA8XRSewMnK6YqXKXoTcRvN/go-libp2p-peer"
	"gx/ipfs/Qme1knMqwt1hKZbc1BmQFmnm9f36nyQGwXxPGVpVJ9rMK5/go-libp2p-crypto"
	"strconv"
	"strings"
)

const nBitsForKeyPairDef = 1024

/**
** 从配置文件中获取私钥，然后解析成lib p2p加密格式私钥，如果配置文件中未填充，则生成一个新的私钥
 */
func GetNodePrivateKey(b64Pri string) (crypto.PrivKey, error) {
	var err error
	var private crypto.PrivKey
	var public crypto.PubKey
	if b64Pri == "" {
		private, public, err = crypto.GenerateKeyPair(crypto.RSA, nBitsForKeyPairDef)
		if err != nil {
			return nil, errors.New(err.Error())
		}
		b, _ := private.Bytes()
		elog.Log.Info("generate private b64 key:", crypto.ConfigEncodeKey(b))
		b, _ = public.Bytes()
		elog.Log.Info("generate public b64 key:", crypto.ConfigEncodeKey(b))
	} else {
		key, err := crypto.ConfigDecodeKey(b64Pri)
		if err != nil {
			return nil, errors.New(err.Error())
		}
		private, err = crypto.UnmarshalPrivateKey(key)
		if err != nil {
			return nil, errors.New(err.Error())
		}
	}
	return private, nil
}

/**
** 从配置文件中获取白名单和黑名单，并生成地址过滤器，然后返回过滤器函数，地址可以用此函数过滤
 */
func MakeAddressesFactory(cfg config.P2pConfig) (basichost.AddrsFactory, error) {
	var annAdds []multiaddr.Multiaddr
	for _, addr := range cfg.AnnounceAddr {
		mAddr, err := multiaddr.NewMultiaddr(addr)
		if err != nil {
			return nil, errors.New(err.Error())
		}
		annAdds = append(annAdds, mAddr)
	}

	filters := filter.NewFilters()
	noAnnAdds := map[string]bool{}
	for _, addr := range cfg.NoAnnounceAddr {
		f, err := mask.NewMask(addr)
		if err == nil {
			filters.AddDialFilter(f)
			continue
		}
		mAddr, err := multiaddr.NewMultiaddr(addr)
		if err != nil {
			return nil, errors.New(err.Error())
		}
		noAnnAdds[mAddr.String()] = true
	}

	return func(allAddresses []multiaddr.Multiaddr) []multiaddr.Multiaddr {
		var adds []multiaddr.Multiaddr
		if len(annAdds) > 0 {
			adds = annAdds
		} else {
			adds = allAddresses
		}

		var out []multiaddr.Multiaddr
		for _, mAddr := range adds { // check for exact matches
			ok, _ := noAnnAdds[mAddr.String()]
			if !ok && !filters.AddrBlocked(mAddr) { // check for /ipcidr matches
				out = append(out, mAddr)
			}
		}
		return out
	}, nil
}

/**
** 似乎是中继地址过滤
 */
func FilterRelayAddresses(addresses []multiaddr.Multiaddr) []multiaddr.Multiaddr {
	var rAdds []multiaddr.Multiaddr
	for _, addr := range addresses {
		_, err := addr.ValueForProtocol(relay.P_CIRCUIT)
		if err == nil {
			continue
		}
		rAdds = append(rAdds, addr)
	}
	return rAdds
}

/**
** 组装过滤器函数
 */
func ComposeAddressesFactory(f, g basichost.AddrsFactory) basichost.AddrsFactory {
	if !config.PConfig.DisableRelay {
		return func(addresses []multiaddr.Multiaddr) []multiaddr.Multiaddr {
			return f(g(addresses))
		}
	} else {
		return f
	}
}

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
