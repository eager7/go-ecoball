package net_test

import (
	"context"
	"fmt"

	"os"
	"testing"
	"time"
	"github.com/ecoball/go-ecoball/lib-p2p/net"
	"github.com/ecoball/go-ecoball/common/elog"
	"github.com/ecoball/go-ecoball/common/utils"
	"github.com/ecoball/go-ecoball/common/message/mpb"
	"github.com/ecoball/go-ecoball/common/event"
)

const (
	priKey1 = "CAAS4QQwggJdAgEAAoGBANqQleuG0BmzpttZ1lfkGmxyKILudJEFLgFcnguSllgdN+6GoeZmByZLoiioTTVgexmXcLGDUdHz5wREhaEo/cx2RwdaUZES6Lewzc82vkmPmp1HMQB3d5s45SMuwqDVSgfvlzdUOXu9629hTgDE//wlq47Kgk6aDCyuLA7jlLGzAgMBAAECgYB96Yukuu6Jz/hRJ6kWyx752K5D95GJth0xxaR68EDSlEqTjFYawC5gPnQ1zfdkx6dDL/5JFWj+de9hgwQkutOydDB8c6HVweTVBrPMB2qIwkWxqofSsHzELP6tF9SuS7tz0ZTmgzkXIcK69nQt/Jlwg+3ronTfkkXCs38sjqA1EQJBAP5xndgg/CPjwwbkF3uaLkz2OytGd445BhqUByK/Ptnz4w+IJ8xMg16uCgglTDIz9454Grc7DpPD3Q1c8XI9UTkCQQDb5ssLzJ0El1JHfo2DiWE1upcJXHlM10vpDL2XHi94eTIfzEj7VxqYMoyC9BJZnRUGMh7gAc9petOORZdiuxZLAkAl825WoTzaYYtiSL0T64BCbGuQ3dbROMInTrLtxNasDYttcqJ0/2iMw6qtYlrGFigzcMiTUdSvx4P+DUHaBzlJAkEAjp0cXBekUaDt3K4niwIiyFytrYWKqZoLgiYgIwyRjtlS96pePpscBU7rL9aou/OS+gSxX2ftIyRkZaWea4qYBwJBAMmHnCCfH87KQY+OwERJHb/z5g4skfLZLKBK1x2bMs2uI14Q5keDRTrb/B6cZzeKsViWK3hvFdXMq5Uc8i5uDyQ="
	pubKey1 = "CAASogEwgZ8wDQYJKoZIhvcNAQEBBQADgY0AMIGJAoGBANqQleuG0BmzpttZ1lfkGmxyKILudJEFLgFcnguSllgdN+6GoeZmByZLoiioTTVgexmXcLGDUdHz5wREhaEo/cx2RwdaUZES6Lewzc82vkmPmp1HMQB3d5s45SMuwqDVSgfvlzdUOXu9629hTgDE//wlq47Kgk6aDCyuLA7jlLGzAgMBAAE="
	priKey2 = "CAAS4AQwggJcAgEAAoGBAJXs/ovug1g4gu43I08QiyUSN9E4SSuWqFNe4qYNn6x6PhTTVDW1yatb8uE3aaFB+Jm9Pyh3eADQ9y8EFK9XN5fwJp7y3szeD/xl0HtiNk1xJKmRX+njEPZ3F6XMAL6wA6FFlif6FI9wj4bci0pk4g5xi28vQ6XBO50G71YUIhbfAgMBAAECgYA6mk2RQuTiSgybsr/BevT4w5s/06F+QUCAfhlX0QF1+L5lg4lqCSnQKnvQnslSOChFZ9zVI4WrxAKqxQyU0SGwUA0yDGIQ+MKcr85+vhrPB9qlA6+/Ruy7cqQ8ZF38Y57KSAC7jXLiuOfm580bHHWd1k0ijgR/7j7FLvjF6JChcQJBAMTDloPI99mGkUzqRZ2Gwl9ArVdTWDZZxmuuOGYpSpif5zszDYoME6w4J+ldrmSQZEr9G01sZF5djwMC/air1GkCQQDDD6CY2zzKYSus2WSfBnREtcb6ktmo/3nXgmufesR40CVNKaLJB5ej+f6qtMfOdv80d43h1I7HAP9MNKYI7AgHAkBNkwcOYfdFbYZvmpVjq7OKNkeg/Bz1IKPX5FIcBP+B+NkDP/eAi45eAa3KlcKhp0PDRNK0zZ0sjxpJB67WBxixAkA+omH7M0rN4W3YzuWUesoS1hvSkhz6Oy6wmNxeFVnJQWz43gm7a4ixyrCPuAUAsw03l7wja9F87UENA0rdSo05AkEAvMVIUj61Uce6U9Z26YjexBll1DwWS5AMRXgvFiKtaf+DLog1c7c4XS9zxZapzbaRi0WxFX2bz1VLXEbq2ypINg=="
	pubKey2 = "CAASogEwgZ8wDQYJKoZIhvcNAQEBBQADgY0AMIGJAoGBAJXs/ovug1g4gu43I08QiyUSN9E4SSuWqFNe4qYNn6x6PhTTVDW1yatb8uE3aaFB+Jm9Pyh3eADQ9y8EFK9XN5fwJp7y3szeD/xl0HtiNk1xJKmRX+njEPZ3F6XMAL6wA6FFlif6FI9wj4bci0pk4g5xi28vQ6XBO50G71YUIhbfAgMBAAE="
)

func TestNode1(t *testing.T) {
	n, err := net.NewInstance(context.Background(), priKey1, "/ip4/0.0.0.0/tcp/9011","/ip4/0.0.0.0/tcp/9012")
	CheckErrorPanic(err)

	fmt.Println("wait...")
	time.Sleep(time.Second * 3)

	{
		CheckErrorPanic(n.SendMessage(pubKey2, "127.0.0.1", "9014", &mpb.Message{Identify: mpb.Identify_APP_MSG_STRING, Payload: []byte(fmt.Sprintf("node1111111111:%d", 9001))}))
	}
	utils.Pause()
}

func TestNode2(t *testing.T) {
	n, err := net.NewInstance(context.Background(), priKey2, "/ip4/0.0.0.0/tcp/9013","/ip4/0.0.0.0/tcp/9014")
	CheckErrorPanic(err)

	fmt.Println("wait...")
	time.Sleep(time.Second * 3)

	{
		CheckErrorPanic(n.SendMessage(pubKey1, "127.0.0.1", "9012", &mpb.Message{Identify: mpb.Identify_APP_MSG_STRING, Payload: []byte(fmt.Sprintf("node222222222222:%d", 9002))}))
	}
	utils.Pause()
}

func CheckErrorPanic(err error) {
	if err != nil {
		elog.Log.Error(err)
		os.Exit(-1)
	}
}

func init() {
	event.InitMsgDispatcher()
}