package datasync

import (
	"testing"
	"time"
	cs "github.com/ecoball/go-ecoball/core/shard"
	"github.com/ecoball/go-ecoball/core/types"
	"reflect"
)

func TestConversion(t *testing.T) {
	header := cs.MinorBlockHeader {
		Version: 213,
		Height: 21392,
		Timestamp:    time.Now().UnixNano(),

		COSign:       nil,



	}
	cosign := &types.COSign{}
	cosign.Step1 = 1
	cosign.Step2 = 0

	header.COSign = cosign

	minorBlock := cs.MinorBlock {
		MinorBlockHeader: header,
		Transactions: nil  ,
		StateDelta: nil ,
	}
	tmp := interface{}(&minorBlock).(cs.Payload)

	log.Debug("type = ", reflect.TypeOf(tmp))
}
