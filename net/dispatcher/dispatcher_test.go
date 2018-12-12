package dispatcher

import (
	"fmt"
	"github.com/ecoball/go-ecoball/net/message/pb"
	"io"
	"runtime"
	"sync"
	"testing"
	"time"
)

type Msg string

func (m *Msg) ChainID() uint32           { return 0 }
func (m *Msg) Type() pb.MsgType          { return pb.MsgType_APP_MSG_TRN }
func (m *Msg) Nonce() uint64             { return 0 }
func (m Msg) Data() []byte               { return []byte(m) }
func (m *Msg) ToProtoV1() *pb.Message    { return nil }
func (m *Msg) ToNetV1(w io.Writer) error { return nil }

func TestSubPubBenchMark(t *testing.T) {
	runtime.GOMAXPROCS(4)

	wg := sync.WaitGroup{}
	wg.Add(1)
	times := 10000000
	InitMsgDispatcher()

	ch, _ := Subscribe(pb.MsgType_APP_MSG_TRN)

	start := time.Now().UnixNano()
	var end int64
	go func() {
		for i := 0; i <= times; i++ {
			var m = Msg(fmt.Sprintf("%d", i))
			Publish(&m)
		}
	}()
	for m := range ch {
		s := string(m.(*Msg).Data())
		if s == "10000000" {
			end = time.Now().UnixNano()
			break
		}
	}
	fmt.Println("run times:", times, "elapsed time:", (end-start)/1000000, "ms")
}
