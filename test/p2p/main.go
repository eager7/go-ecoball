package main

import (
	"context"
	"encoding/json"
	"flag"
	"github.com/ecoball/go-ecoball/common/elog"
	"github.com/ecoball/go-ecoball/common/errors"
	"github.com/ecoball/go-ecoball/net"
	"github.com/ecoball/go-ecoball/net/message"
	"github.com/ecoball/go-ecoball/net/message/pb"
	"github.com/ecoball/go-ecoball/net/network"
	"github.com/ecoball/go-ecoball/sharding/common"
	"github.com/ecoball/go-ecoball/test/example"
	"time"
)

const (
	pubKey0    = "CAASogEwgZ8wDQYJKoZIhvcNAQEBBQADgY0AMIGJAoGBAMA7hQfC/uYM1M9FrOdK2mHkEchOW0ayEEa1C4vGnldWSF8dWPfFWSwSdF1xeUVm90BegrmM76U/wJecCtNwKu+kzZ9s3UacA4TouKnhdhOEbKIPknVwtOdUw578/fyAl0RVk8Gr0/MY2CX4xdB+bbjNvE+9vas4uFdH75LcJ5MTAgMBAAE="
	pubKey1    = "CAASogEwgZ8wDQYJKoZIhvcNAQEBBQADgY0AMIGJAoGBAMJievBIrNu158t+fzzfnuLLBP/bToApftfI56OZIwvhglY82fGsGqm6RGfg+/8fC2noBaB8fw9Tsf4LRH/8F3tKMjIQ4alzV4jwxTbdbkrLavgr5XNNce666M/HNmgl7X/USLolGmL9HEiBS8U2Vg/0OTVj9yNzSplsHV8sb6htAgMBAAE="
	pubKey2    = "CAASogEwgZ8wDQYJKoZIhvcNAQEBBQADgY0AMIGJAoGBAJ53lGZe0yul67lpH0DlD5VgHd5otmxfvDX9nzgk0JTN/HGXOTS/Kcf6f70UrpTi7Yk5WUpeFrLAuwr+Inulco8+Rb9Cm/giUtQ94ZpjQvx7AOcEU2qy75Is1teF4jFRM2EpqyJnrDw0TGYB/C5lXzfExCViojX15XURKG7ziPFfAgMBAAE="
	pubKey3    = "CAASogEwgZ8wDQYJKoZIhvcNAQEBBQADgY0AMIGJAoGBAMMg1HzNHAa/ay0XpnsXyjUdudMzldk/dA5ar9+hf0oPbJdDWIH9Pn7h3eE7EcZ8HDYklqiquUvamIfnOjNk0oTiRHm+KiKCXrVBAbeYDJAz4GYFTRiCcfpywpHFBTWZiqD9YuzOz1jErjAwZv9AeO6iKT1XOqiySKPlzPN0gqnnAgMBAAE="
	toPoShard1 = `{
	"ShardId": 1,
	"Pubkey": "CAASogEwgZ8wDQYJKoZIhvcNAQEBBQADgY0AMIGJAoGBAMJievBIrNu158t+fzzfnuLLBP/bToApftfI56OZIwvhglY82fGsGqm6RGfg+/8fC2noBaB8fw9Tsf4LRH/8F3tKMjIQ4alzV4jwxTbdbkrLavgr5XNNce666M/HNmgl7X/USLolGmL9HEiBS8U2Vg/0OTVj9yNzSplsHV8sb6htAgMBAAE=",
	"ShardingInfo": [
		[{
			"Pubkey": "CAASogEwgZ8wDQYJKoZIhvcNAQEBBQADgY0AMIGJAoGBAMA7hQfC/uYM1M9FrOdK2mHkEchOW0ayEEa1C4vGnldWSF8dWPfFWSwSdF1xeUVm90BegrmM76U/wJecCtNwKu+kzZ9s3UacA4TouKnhdhOEbKIPknVwtOdUw578/fyAl0RVk8Gr0/MY2CX4xdB+bbjNvE+9vas4uFdH75LcJ5MTAgMBAAE=",
			"Address": "192.168.8.35",
			"Port": "9901"
		}],
		[{
			"Pubkey": "CAASogEwgZ8wDQYJKoZIhvcNAQEBBQADgY0AMIGJAoGBAMJievBIrNu158t+fzzfnuLLBP/bToApftfI56OZIwvhglY82fGsGqm6RGfg+/8fC2noBaB8fw9Tsf4LRH/8F3tKMjIQ4alzV4jwxTbdbkrLavgr5XNNce666M/HNmgl7X/USLolGmL9HEiBS8U2Vg/0OTVj9yNzSplsHV8sb6htAgMBAAE=",
			"Address": "192.168.8.35",
			"Port": "9902"
		}],
		[{
			"Pubkey": "CAASogEwgZ8wDQYJKoZIhvcNAQEBBQADgY0AMIGJAoGBAJ53lGZe0yul67lpH0DlD5VgHd5otmxfvDX9nzgk0JTN/HGXOTS/Kcf6f70UrpTi7Yk5WUpeFrLAuwr+Inulco8+Rb9Cm/giUtQ94ZpjQvx7AOcEU2qy75Is1teF4jFRM2EpqyJnrDw0TGYB/C5lXzfExCViojX15XURKG7ziPFfAgMBAAE=",
			"Address": "192.168.8.35",
			"Port": "9903"
		}],
		[{
			"Pubkey": "CAASogEwgZ8wDQYJKoZIhvcNAQEBBQADgY0AMIGJAoGBAMMg1HzNHAa/ay0XpnsXyjUdudMzldk/dA5ar9+hf0oPbJdDWIH9Pn7h3eE7EcZ8HDYklqiquUvamIfnOjNk0oTiRHm+KiKCXrVBAbeYDJAz4GYFTRiCcfpywpHFBTWZiqD9YuzOz1jErjAwZv9AeO6iKT1XOqiySKPlzPN0gqnnAgMBAAE=",
			"Address": "192.168.8.35",
			"Port": "9904"
		}], null
	]
}`

	toPoShard2 = `{
	"ShardId": 2,
	"Pubkey": "CAASogEwgZ8wDQYJKoZIhvcNAQEBBQADgY0AMIGJAoGBAJ53lGZe0yul67lpH0DlD5VgHd5otmxfvDX9nzgk0JTN/HGXOTS/Kcf6f70UrpTi7Yk5WUpeFrLAuwr+Inulco8+Rb9Cm/giUtQ94ZpjQvx7AOcEU2qy75Is1teF4jFRM2EpqyJnrDw0TGYB/C5lXzfExCViojX15XURKG7ziPFfAgMBAAE=",
	"ShardingInfo": [
		[{
			"Pubkey": "CAASogEwgZ8wDQYJKoZIhvcNAQEBBQADgY0AMIGJAoGBAMA7hQfC/uYM1M9FrOdK2mHkEchOW0ayEEa1C4vGnldWSF8dWPfFWSwSdF1xeUVm90BegrmM76U/wJecCtNwKu+kzZ9s3UacA4TouKnhdhOEbKIPknVwtOdUw578/fyAl0RVk8Gr0/MY2CX4xdB+bbjNvE+9vas4uFdH75LcJ5MTAgMBAAE=",
			"Address": "192.168.8.35",
			"Port": "9901"
		}],
		[{
			"Pubkey": "CAASogEwgZ8wDQYJKoZIhvcNAQEBBQADgY0AMIGJAoGBAMJievBIrNu158t+fzzfnuLLBP/bToApftfI56OZIwvhglY82fGsGqm6RGfg+/8fC2noBaB8fw9Tsf4LRH/8F3tKMjIQ4alzV4jwxTbdbkrLavgr5XNNce666M/HNmgl7X/USLolGmL9HEiBS8U2Vg/0OTVj9yNzSplsHV8sb6htAgMBAAE=",
			"Address": "192.168.8.35",
			"Port": "9902"
		}],
		[{
			"Pubkey": "CAASogEwgZ8wDQYJKoZIhvcNAQEBBQADgY0AMIGJAoGBAJ53lGZe0yul67lpH0DlD5VgHd5otmxfvDX9nzgk0JTN/HGXOTS/Kcf6f70UrpTi7Yk5WUpeFrLAuwr+Inulco8+Rb9Cm/giUtQ94ZpjQvx7AOcEU2qy75Is1teF4jFRM2EpqyJnrDw0TGYB/C5lXzfExCViojX15XURKG7ziPFfAgMBAAE=",
			"Address": "192.168.8.35",
			"Port": "9903"
		}],
		[{
			"Pubkey": "CAASogEwgZ8wDQYJKoZIhvcNAQEBBQADgY0AMIGJAoGBAMMg1HzNHAa/ay0XpnsXyjUdudMzldk/dA5ar9+hf0oPbJdDWIH9Pn7h3eE7EcZ8HDYklqiquUvamIfnOjNk0oTiRHm+KiKCXrVBAbeYDJAz4GYFTRiCcfpywpHFBTWZiqD9YuzOz1jErjAwZv9AeO6iKT1XOqiySKPlzPN0gqnnAgMBAAE=",
			"Address": "192.168.8.35",
			"Port": "9904"
		}], null
	]
}`
)

type shardInstance struct {
	c chan interface{}
}

func (s *shardInstance) Start()                                    {}
func (s *shardInstance) MsgDispatch(msg interface{})               {}
func (s *shardInstance) SubscribeShardingTopo() <-chan interface{} { return s.c }
func (s *shardInstance) SetNet(n network.EcoballNetwork)           {}

var bVal = flag.Bool("bool", false, "bool value for test")

func main() {
	flag.Parse()
	elog.Log.Debug("net test program...", *bVal)

	c := make(chan interface{})
	ctx, cancel := context.WithCancel(context.Background())
	net.InitNetWork(ctx)
	net.StartNetWork(c)

	toPo := &common.ShardingTopo{}
	if !*bVal {
		errors.CheckErrorPanic(json.Unmarshal([]byte(toPoShard1), toPo)) //set node as shard 1
		c <- toPo

		instance, err := network.GetNetInstance()
		errors.CheckErrorPanic(err)
		msg := message.New(pb.MsgType_APP_MSG_STRING, []byte("my name is shard1"))
		for {
			instance.SendMsgToPeer("192.168.8.35", "9003", pubKey2, msg) //send msg to shard2
			time.Sleep(time.Second * 2)
		}
	} else {
		errors.CheckErrorPanic(json.Unmarshal([]byte(toPoShard2), toPo)) //set node as shard 1
		c <- toPo
		if false {
			instance, err := network.GetNetInstance()
			errors.CheckErrorPanic(err)
			msg := message.New(pb.MsgType_APP_MSG_STRING, []byte("this is shard 2"))
			for {
				instance.SendMsgToPeer("192.168.8.35", "9002", pubKey1, msg) //send msg to shard1
				time.Sleep(time.Second * 2)
			}
		}
	}

	example.Wait()
	cancel()
}
