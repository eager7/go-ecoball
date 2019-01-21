package sharding

import (
	"github.com/ecoball/go-ecoball/common/errors"
	"github.com/ecoball/go-ecoball/common/utils"
	"github.com/ecoball/go-ecoball/test/example"
	"testing"
)

func TestStart(t *testing.T) {
	shardActor, err := NewShardingActor(example.Ledger("/tmp/shardActor"))
	errors.CheckErrorPanic(err)
	t.Log(shardActor)
	utils.Pause()
}