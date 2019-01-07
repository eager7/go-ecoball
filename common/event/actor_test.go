package event_test

import (
	"fmt"
	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/ecoball/go-ecoball/common/errors"
	"github.com/ecoball/go-ecoball/common/event"
	"github.com/ecoball/go-ecoball/common/message/mpb"
	"testing"
	"time"
)

type Data struct {
	val int
}

func TestActorRegister(t *testing.T) {
	props := actor.FromFunc(func(c actor.Context) {
		switch msg := c.Message().(type) {
		case int32:
			fmt.Println(msg)
		case *Data:
			fmt.Println(msg.val)
		default:
			fmt.Println("unknown type")
		}
	})
	actorA, _ := actor.SpawnNamed(props, "actorA")
	actorB, _ := actor.SpawnNamed(props, "actorB")

	if err := event.RegisterActor(event.ActorTxPool, actorA); err != nil {
		t.Fatal(err)
	}
	if err := event.RegisterActor(event.ActorLedger, actorB); err != nil {
		t.Fatal(err)
	}

	actorTxPool, _ := event.GetActor(event.ActorTxPool)
	actorLedger, _ := event.GetActor(event.ActorLedger)
	var i int32 = 1000
	actorTxPool.Request(i, actorLedger)
	actorTxPool.Request(&Data{val: 99}, actorLedger)
	time.Sleep(1 * time.Second)
}

func TestPublish(t *testing.T) {
	event.InitMsgDispatcher()
	channel, err := event.Subscribe([]mpb.Identify{mpb.Identify_APP_MSG_STRING}...)
	errors.CheckErrorPanic(err)
	go func() {
		time.Sleep(time.Millisecond * 500)
		for {
			select {
			case <-time.After(time.Second * 1):
				break
			case in := <-channel:
				fmt.Println(in.(*mpb.Message))
			}
		}
	}()
	fmt.Println("send1")
	_ = event.Publish(&mpb.Message{Nonce: 0, Identify: 0, Payload: []byte("test1")}, mpb.Identify_APP_MSG_STRING)
	fmt.Println("send2")
	_ = event.Publish(&mpb.Message{Nonce: 0, Identify: 0, Payload: []byte("test2")}, mpb.Identify_APP_MSG_STRING)
	fmt.Println("send3")
	_ = event.Publish(&mpb.Message{Nonce: 0, Identify: 0, Payload: []byte("test3")}, mpb.Identify_APP_MSG_STRING)
	fmt.Println("send4")
	_ = event.Publish(&mpb.Message{Nonce: 0, Identify: 0, Payload: []byte("test4")}, mpb.Identify_APP_MSG_STRING)
	time.Sleep(time.Second * 1)
}
