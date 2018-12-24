package net_test

import (
	"context"
	"encoding/json"
	"github.com/ecoball/go-ecoball/common/elog"
	"github.com/ecoball/go-ecoball/common/errors"
	"github.com/ecoball/go-ecoball/net"
	"github.com/ecoball/go-ecoball/sharding/common"
	"github.com/ecoball/go-ecoball/test/example"
	"testing"
	"github.com/ecoball/go-ecoball/net/network"
)

func TestNet(t *testing.T) {
	elog.Log.Debug("net test program...")

	c := make(chan interface{})
	ctx, cancel := context.WithCancel(context.Background())
	net.InitNetWork(ctx)
	net.StartNetWork(c)

	toPoInfo := `{
	"ShardId": 1,
	"Pubkey": "CAASogEwgZ8wDQYJKoZIhvcNAQEBBQADgY0AMIGJAoGBAL98FFiv93FzIwFS6hTIXRlJoY2rTrU9jQR/uzYk1UlnfCBJ4Q7fsWSBwILmGttmyr1599Si+XlH4AFYZ9bWXGwZXuLx9VYVG/faB1fu4/kzPF9IZJ80M0XPMNhR4/IX69Fde0Mpr1muFTAFgmNcmrZoXymTHsEOM+bP4VfitOtpAgMBAAE=",
	"ShardingInfo": [
		[{
			"Pubkey": "CAASogEwgZ8wDQYJKoZIhvcNAQEBBQADgY0AMIGJAoGBANWB0oHE4Qebj9vRSPTWrKRrzwy73xm9JoBr5j57J6qb5f93gZqkaWOl3oMr6pZIyBOH6fPqvsKAagqIJQlkgch4NjV4LZmWjdCEcK9UdyTT0pD+MdkuqlGcOXKG913wWFPlRNbEKkT+/jO+SC+k+iStRr50yFah074QbIIxIeNbAgMBAAE=",
			"Address": "192.168.8.35",
			"Port": "9901"
		}],
		[{
			"Pubkey": "CAASogEwgZ8wDQYJKoZIhvcNAQEBBQADgY0AMIGJAoGBAL98FFiv93FzIwFS6hTIXRlJoY2rTrU9jQR/uzYk1UlnfCBJ4Q7fsWSBwILmGttmyr1599Si+XlH4AFYZ9bWXGwZXuLx9VYVG/faB1fu4/kzPF9IZJ80M0XPMNhR4/IX69Fde0Mpr1muFTAFgmNcmrZoXymTHsEOM+bP4VfitOtpAgMBAAE=",
			"Address": "192.168.8.35",
			"Port": "9902"
		}],
		[{
			"Pubkey": "CAASogEwgZ8wDQYJKoZIhvcNAQEBBQADgY0AMIGJAoGBALdD7uB7N/KrVUNxoVqTxFduv4EIJ+re04QsljS5Gv4xtPu/1J45EDffdPS1XRrf76HhJpEgL/SKR5lq3DLj6PAgWdtBn7CE6XJKUkOn/rR+8avJ8gJ+KCwf2MvS7UbbOtp5vl4oFraNpGCEwLFcFG2u32pgIb6CwbVG42YCHDeRAgMBAAE=",
			"Address": "192.168.8.35",
			"Port": "9903"
		}],
		[{
			"Pubkey": "CAASogEwgZ8wDQYJKoZIhvcNAQEBBQADgY0AMIGJAoGBAJ5FjqG9/JDt1WHFDFgHT6L2lxOCb8q9/HlrpI1LVRZ/duPWJevdzjHKgU6qhLdHAX12juewgNm+Jh9feFfr+yiP6L/0dU4Rum9G3NMhVvg7wI3Og0erYHFE+M8XPjU2DWhDA3w7zY0Pfn4Tr73MMMb2YxxvySonpw1G81rjwvrjAgMBAAE=",
			"Address": "192.168.8.35",
			"Port": "9904"
		}], null
	]
}`
	toPo := &common.ShardingTopo{}
	errors.CheckErrorPanic(json.Unmarshal([]byte(toPoInfo), toPo))
	c <- toPo

	_, err := network.GetNetInstance()
	errors.CheckErrorPanic(err)

	example.Wait()
	cancel()
}
