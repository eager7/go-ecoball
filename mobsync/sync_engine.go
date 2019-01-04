package mobsync

import (
	"context"
	"github.com/ecoball/go-ecoball/common/config"
	"github.com/ecoball/go-ecoball/common/elog"
	"github.com/ecoball/go-ecoball/common/event"
	"github.com/ecoball/go-ecoball/common/message/mpb"
	"github.com/ecoball/go-ecoball/common/utils"
	"github.com/ecoball/go-ecoball/core/ledgerimpl/ledger"
	"github.com/ecoball/go-ecoball/core/types"
	"gx/ipfs/QmSF8fPo3jgVBAy8fpdjjYqgG87dkJgUprRBHRd2tmfgpP/goprocess"
	//ptx "gx/ipfs/QmSF8fPo3jgVBAy8fpdjjYqgG87dkJgUprRBHRd2tmfgpP/goprocess/context"
	"gx/ipfs/QmSF8fPo3jgVBAy8fpdjjYqgG87dkJgUprRBHRd2tmfgpP/goprocess/periodic"
	"time"
)

var log = elog.Log

type Engine struct {
	ledger  ledger.Ledger
	stop    chan struct{}
	message <-chan interface{}
	process goprocess.Process
}

func NewSyncEngine(ctx context.Context, ledger ledger.Ledger) (err error) {
	engine := new(Engine).Initialize()
	engine.ledger = ledger
	msg := []mpb.Identify{
		mpb.Identify_APP_MSG_BLOCK,
		mpb.Identify_APP_MSG_TRANSACTION,
		mpb.Identify_APP_MSG_BLOCK_REQUEST,
		mpb.Identify_APP_MSG_BLOCK_RESPONSE,
	}
	if engine.message, err = event.Subscribe(msg...); err != nil {
		log.Error(err)
		return err
	}

	doneWithRound := make(chan struct{})
	periodic := func(worker goprocess.Process) {
		//ctx := ptx.OnClosedContext(worker)
		current := ledger.GetCurrentHeader(config.ChainHash)
		if err := event.Send(event.ActorNil, event.ActorP2P, &BlockRequest{ChainId: current.Hash, BlockHeight: current.Height, Nonce: utils.RandomUint64()}); err != nil {
			log.Error(err)
		}
		<-doneWithRound
	}
	engine.process = periodicproc.Tick(time.Millisecond*time.Duration(config.TimeSlot*10), periodic)
	engine.process.Go(periodic)
	doneWithRound <- struct{}{}
	close(doneWithRound)

	return nil
}

func (e *Engine) Initialize() *Engine {
	e.stop = make(chan struct{}, 1)
	e.message = make(chan interface{}, 100)
	return e
}

func (e *Engine) handlerThread() {
	for {
		select {
		case msg := <-e.message:
			if in, ok := msg.(*mpb.Message); !ok {
				log.Error("can't parse msg")
				continue
			} else {
				log.Info("receive msg:", in.Identify.String())
				switch in.Identify {
				case mpb.Identify_APP_MSG_TRANSACTION:
				case mpb.Identify_APP_MSG_BLOCK:
					if err := e.SyncBlockChain(in); err != nil {
						log.Error("sync block failed:", err)
					}
				case mpb.Identify_APP_MSG_BLOCK_REQUEST:
					//TODO
				case mpb.Identify_APP_MSG_BLOCK_RESPONSE:
					//TODO
				default:
					log.Warn("unsupported sync message:", in.Identify.String())
				}
			}

		case <-e.stop:
			log.Info(e.process.Close())
			log.Info("Stop Solo Mode")
			return
		}
	}
}

func (e *Engine) SyncBlockChain(msg *mpb.Message) error {
	block := new(types.Block)
	if err := block.Deserialize(msg.Payload); err != nil {
		return err
	}
	headerCurrent := e.ledger.GetCurrentHeader(block.ChainID)
	if headerCurrent.Height == block.Height {
		return nil
	}
	if headerCurrent.Height >= block.Height {

	}
	//if err := event.Send(event.ActorConsensusSolo, event.ActorLedger, block); err != nil {
	//	return err
	//}
	return nil
}

func (e *Engine) SyncTransaction(msg *mpb.Message) error {
	return nil
}
