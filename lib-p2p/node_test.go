package p2p_test

import (
	"context"
	"encoding/json"
	"github.com/ecoball/go-ecoball/common/elog"
	"github.com/ecoball/go-ecoball/common/errors"
	"github.com/ecoball/go-ecoball/common/event"
	"github.com/ecoball/go-ecoball/common/message"
	"github.com/ecoball/go-ecoball/common/utils"
	"github.com/ecoball/go-ecoball/lib-p2p"
	"github.com/ecoball/go-ecoball/sharding/common"
	"github.com/ecoball/go-ecoball/test/example"
	"testing"
	"time"
)

const toPoInfo = `{
	"ShardId": 1,
	"Pubkey": "CAASogEwgZ8wDQYJKoZIhvcNAQEBBQADgY0AMIGJAoGBALna9LG/OdOImFPZ19WXzpCnCegonngYny888RvEUl/YcMpNQ1Rclpo/rtNiBlcxuXW7TepW/afQ0Y1yq8aRuRe7526RUQ8sLWc2mfCvV/HL6b1614qH8Q9HODnHTNIKzya+0PZuLNsS4Rug5dwMJHMKW8sAQK7TVvz5sdU+qa4vAgMBAAE=",
	"ShardingInfo": [
		[{
			"Pubkey": "CAASogEwgZ8wDQYJKoZIhvcNAQEBBQADgY0AMIGJAoGBANWB0oHE4Qebj9vRSPTWrKRrzwy73xm9JoBr5j57J6qb5f93gZqkaWOl3oMr6pZIyBOH6fPqvsKAagqIJQlkgch4NjV4LZmWjdCEcK9UdyTT0pD+MdkuqlGcOXKG913wWFPlRNbEKkT+/jO+SC+k+iStRr50yFah074QbIIxIeNbAgMBAAE=",
			"Address": "192.168.8.35",
			"Port": "9901"
		}],
		[{
			"Pubkey": "CAASogEwgZ8wDQYJKoZIhvcNAQEBBQADgY0AMIGJAoGBALna9LG/OdOImFPZ19WXzpCnCegonngYny888RvEUl/YcMpNQ1Rclpo/rtNiBlcxuXW7TepW/afQ0Y1yq8aRuRe7526RUQ8sLWc2mfCvV/HL6b1614qH8Q9HODnHTNIKzya+0PZuLNsS4Rug5dwMJHMKW8sAQK7TVvz5sdU+qa4vAgMBAAE=",
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
const (
	priKey1 = "CAAS4QQwggJdAgEAAoGBANqQleuG0BmzpttZ1lfkGmxyKILudJEFLgFcnguSllgdN+6GoeZmByZLoiioTTVgexmXcLGDUdHz5wREhaEo/cx2RwdaUZES6Lewzc82vkmPmp1HMQB3d5s45SMuwqDVSgfvlzdUOXu9629hTgDE//wlq47Kgk6aDCyuLA7jlLGzAgMBAAECgYB96Yukuu6Jz/hRJ6kWyx752K5D95GJth0xxaR68EDSlEqTjFYawC5gPnQ1zfdkx6dDL/5JFWj+de9hgwQkutOydDB8c6HVweTVBrPMB2qIwkWxqofSsHzELP6tF9SuS7tz0ZTmgzkXIcK69nQt/Jlwg+3ronTfkkXCs38sjqA1EQJBAP5xndgg/CPjwwbkF3uaLkz2OytGd445BhqUByK/Ptnz4w+IJ8xMg16uCgglTDIz9454Grc7DpPD3Q1c8XI9UTkCQQDb5ssLzJ0El1JHfo2DiWE1upcJXHlM10vpDL2XHi94eTIfzEj7VxqYMoyC9BJZnRUGMh7gAc9petOORZdiuxZLAkAl825WoTzaYYtiSL0T64BCbGuQ3dbROMInTrLtxNasDYttcqJ0/2iMw6qtYlrGFigzcMiTUdSvx4P+DUHaBzlJAkEAjp0cXBekUaDt3K4niwIiyFytrYWKqZoLgiYgIwyRjtlS96pePpscBU7rL9aou/OS+gSxX2ftIyRkZaWea4qYBwJBAMmHnCCfH87KQY+OwERJHb/z5g4skfLZLKBK1x2bMs2uI14Q5keDRTrb/B6cZzeKsViWK3hvFdXMq5Uc8i5uDyQ="
	pubKey1 = "CAASogEwgZ8wDQYJKoZIhvcNAQEBBQADgY0AMIGJAoGBANqQleuG0BmzpttZ1lfkGmxyKILudJEFLgFcnguSllgdN+6GoeZmByZLoiioTTVgexmXcLGDUdHz5wREhaEo/cx2RwdaUZES6Lewzc82vkmPmp1HMQB3d5s45SMuwqDVSgfvlzdUOXu9629hTgDE//wlq47Kgk6aDCyuLA7jlLGzAgMBAAE="
	priKey2 = "CAAS4AQwggJcAgEAAoGBAJXs/ovug1g4gu43I08QiyUSN9E4SSuWqFNe4qYNn6x6PhTTVDW1yatb8uE3aaFB+Jm9Pyh3eADQ9y8EFK9XN5fwJp7y3szeD/xl0HtiNk1xJKmRX+njEPZ3F6XMAL6wA6FFlif6FI9wj4bci0pk4g5xi28vQ6XBO50G71YUIhbfAgMBAAECgYA6mk2RQuTiSgybsr/BevT4w5s/06F+QUCAfhlX0QF1+L5lg4lqCSnQKnvQnslSOChFZ9zVI4WrxAKqxQyU0SGwUA0yDGIQ+MKcr85+vhrPB9qlA6+/Ruy7cqQ8ZF38Y57KSAC7jXLiuOfm580bHHWd1k0ijgR/7j7FLvjF6JChcQJBAMTDloPI99mGkUzqRZ2Gwl9ArVdTWDZZxmuuOGYpSpif5zszDYoME6w4J+ldrmSQZEr9G01sZF5djwMC/air1GkCQQDDD6CY2zzKYSus2WSfBnREtcb6ktmo/3nXgmufesR40CVNKaLJB5ej+f6qtMfOdv80d43h1I7HAP9MNKYI7AgHAkBNkwcOYfdFbYZvmpVjq7OKNkeg/Bz1IKPX5FIcBP+B+NkDP/eAi45eAa3KlcKhp0PDRNK0zZ0sjxpJB67WBxixAkA+omH7M0rN4W3YzuWUesoS1hvSkhz6Oy6wmNxeFVnJQWz43gm7a4ixyrCPuAUAsw03l7wja9F87UENA0rdSo05AkEAvMVIUj61Uce6U9Z26YjexBll1DwWS5AMRXgvFiKtaf+DLog1c7c4XS9zxZapzbaRi0WxFX2bz1VLXEbq2ypINg=="
	pubKey2 = "CAASogEwgZ8wDQYJKoZIhvcNAQEBBQADgY0AMIGJAoGBAJXs/ovug1g4gu43I08QiyUSN9E4SSuWqFNe4qYNn6x6PhTTVDW1yatb8uE3aaFB+Jm9Pyh3eADQ9y8EFK9XN5fwJp7y3szeD/xl0HtiNk1xJKmRX+njEPZ3F6XMAL6wA6FFlif6FI9wj4bci0pk4g5xi28vQ6XBO50G71YUIhbfAgMBAAE="
)

func TestNet(t *testing.T) {
	elog.Log.Debug("net test program...")
	p2p.InitNetWork(context.Background(), "")
	toPo := &common.ShardingTopo{}
	errors.CheckErrorPanic(json.Unmarshal([]byte(toPoInfo), toPo))
	event.Send(event.ActorNil, event.ActorP2P, toPo)

	example.Wait()
}

func TestServer(t *testing.T) {
	event.InitMsgDispatcher()
	p2p.InitNetWork(context.Background(), priKey1, "/ip4/0.0.0.0/tcp/9011", "/ip4/0.0.0.0/tcp/9012")
	time.Sleep(time.Second*1)
	m := example.Message("hello client, i am server")
	event.Send(event.ActorNil, event.ActorP2P, message.NetPacket{Address: "127.0.0.1", Port: "9013", PublicKey: pubKey2, Message: &m})
	utils.Pause()
}

func TestClient(t *testing.T) {
	event.InitMsgDispatcher()
	p2p.InitNetWork(context.Background(), priKey2, "/ip4/0.0.0.0/tcp/9013", "/ip4/0.0.0.0/tcp/9014")
	time.Sleep(time.Second*1)
	m := example.Message("hello server, i am client")
	event.Send(event.ActorNil, event.ActorP2P, message.NetPacket{Address: "127.0.0.1", Port: "9012", PublicKey: pubKey1, Message: &m})
	utils.Pause()
}
