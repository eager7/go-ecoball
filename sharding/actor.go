package sharding

import (
	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/ecoball/go-ecoball/common/message"
	"reflect"
)

type ShardingActor struct {
	instance ShardingInstance
}

func NewShardingActor() (pid *actor.PID, err error) {

	shardingActor := &ShardingActor{}

	props := actor.FromProducer(func() actor.Actor { return shardingActor })

	pid, err = actor.SpawnNamed(props, "ShardingActor")
	if err == nil {
		shardingActor.instance = MakeSharding()
		shardingActor.instance.Start()

		return pid, nil
	} else {
		return nil, err
	}
}

func (s *ShardingActor) Receive(ctx actor.Context) {
	switch msg := ctx.Message().(type) {
	case *actor.Started:
		log.Info("ShardingActor received started msg")

	case *actor.Stopping:
		log.Info("ShardingActor received stopping msg")

	case *actor.Restart:
		log.Info("ShardingActor received restart msg")

	case *actor.Restarting:
		log.Info("ShardingActor received restarting msg")

	case *actor.Stop:
		log.Info("ShardingActor received Stop msg")

	case *actor.Stopped:
		log.Info("ShardingActor received Stopped msg")

	case *message.SyncComplete:
		s.instance.MsgDispatch(msg)

	default:
		log.Warn("ShardingActor received unknown type message ", msg, " type ", reflect.TypeOf(msg))
	}
}

func SetActor() {

}
