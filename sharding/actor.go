package sharding

import (
	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/ecoball/go-ecoball/common/event"
	"github.com/ecoball/go-ecoball/common/message"
	"github.com/ecoball/go-ecoball/core/ledgerimpl/ledger"
	cs "github.com/ecoball/go-ecoball/core/shard"
	"reflect"
)

type ShardingActor struct {
	instance ShardingInstance
}

func NewShardingActor(l ledger.Ledger) (*ShardingActor, error) {

	shardingActor := &ShardingActor{}

	props := actor.FromProducer(func() actor.Actor { return shardingActor })

	pid, err := actor.SpawnNamed(props, "ShardingActor")
	if err == nil {
		shardingActor.instance = MakeSharding(l)
		shardingActor.instance.Start()

		event.RegisterActor(event.ActorSharding, pid)

		return shardingActor, nil
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
		log.Debug("receive sync complete")
		s.instance.MsgDispatch(msg)

	case *cs.FinalBlock:
		log.Debug("receive final block")
		s.instance.MsgDispatch(msg)

	case *cs.MinorBlock:
		log.Debug("receive minor block")
		s.instance.MsgDispatch(msg)

	default:
		log.Warn("ShardingActor received unknown type message ", msg, " type ", reflect.TypeOf(msg))
	}
}

func (s *ShardingActor) SubscribeShardingTopo() <-chan interface{} {
	return s.instance.SubscribeShardingTopo()
}
