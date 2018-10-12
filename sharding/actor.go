package sharding

import (
	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/ecoball/go-ecoball/common/event"
	"github.com/ecoball/go-ecoball/common/message"
	"github.com/ecoball/go-ecoball/core/ledgerimpl/ledger"
	"reflect"
	"github.com/ecoball/go-ecoball/sharding/cell"
	"github.com/ecoball/go-ecoball/common/errors"
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
		//shardingActor.instance.Start()

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
		s.instance.MsgDispatch(msg)

	default:
		log.Warn("ShardingActor received unknown type message ", msg, " type ", reflect.TypeOf(msg))
	}
}

func (s *ShardingActor) GetCell() (*cell.Cell, error) {
	shard, ok := s.instance.(*Sharding)
	if !ok {
		return nil, errors.New(log, "failed to get sharding cell")
	}
	return shard.GetCell(), nil
}

func (s *ShardingActor) Start() {
	s.instance.Start()
}

func SetActor() {

}
