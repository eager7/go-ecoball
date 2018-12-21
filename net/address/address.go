package address

import (
	"github.com/ecoball/go-ecoball/common/config"
	"github.com/ecoball/go-ecoball/common/errors"
	"gx/ipfs/Qme1knMqwt1hKZbc1BmQFmnm9f36nyQGwXxPGVpVJ9rMK5/go-libp2p-crypto"
)

const nBitsForKeyPairDef = 1024

/**
** 从配置文件中获取私钥，然后解析成lib p2p加密格式私钥，如果配置文件中未填充，则生成一个新的私钥
 */
func GetNodePrivateKey() (crypto.PrivKey, error) {
	var err error
	var private crypto.PrivKey
	if config.SwarmConfig.PrivateKey == "" {
		private, _, err = crypto.GenerateKeyPair(crypto.RSA, nBitsForKeyPairDef)
		if err != nil {
			return nil, errors.New(err.Error())
		}
	} else {
		key, err := crypto.ConfigDecodeKey(config.SwarmConfig.PrivateKey)
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
