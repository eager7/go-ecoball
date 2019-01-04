package mobsync

import (
	"context"
	"github.com/ecoball/go-ecoball/common"
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
	ctx     context.Context
	ledger  ledger.Ledger
	stop    chan struct{}
	message <-chan interface{}
	process goprocess.Process
}

func NewSyncEngine(ctx context.Context, ledger ledger.Ledger) (err error) {
	engine := new(Engine).Initialize()
	engine.ledger = ledger
	engine.ctx = ctx
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

	go engine.handlerThread()

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
				log.Info("sync engine receive msg:", in.Identify.String())
				switch in.Identify {
				case mpb.Identify_APP_MSG_TRANSACTION:
				case mpb.Identify_APP_MSG_BLOCK:
					if err := e.SyncBlockChain(in); err != nil {
						log.Error("sync block failed:", err)
					}
				case mpb.Identify_APP_MSG_BLOCK_REQUEST:
					if err := e.HandleBlockRequest(in); err != nil {
						log.Error("handle block request error:", err)
					}
				case mpb.Identify_APP_MSG_BLOCK_RESPONSE:

				default:
					log.Warn("unsupported sync message:", in.Identify.String())
				}
			}

		case <-e.stop:
			log.Info(e.process.Close())
			log.Info("Stop Solo Mode")
			return
		case <-e.ctx.Done():
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
	current := e.ledger.GetCurrentHeader(block.ChainID)
	if current != nil && current.Height < block.Height {
		return event.Send(event.ActorNil, event.ActorP2P, &BlockRequest{ChainId: block.ChainID, BlockHeight: current.Height, Nonce: utils.RandomUint64()})
	}
	return nil
}

func (e *Engine) SyncTransaction(msg *mpb.Message) error {
	return nil
}

func (e *Engine) HandleBlockRequest(msg *mpb.Message) error {
	request := new(BlockRequest)
	if err := request.Deserialize(msg.Payload); err != nil {
		return err
	}
	log.Debug("handle block request message:", request.ChainId.String(), request.BlockHeight)
	current := e.ledger.GetCurrentHeader(request.ChainId)
	if current != nil && current.Height <= request.BlockHeight {
		log.Info("our chain block is older than request:", current.Height, request.BlockHeight)
		return nil
	}
	return e.SendBlockResponse(request.ChainId, current.Height, request.BlockHeight)
}

func (e *Engine) SendBlockResponse(chainId common.Hash, current, request uint64) error {
	var num int
	response := &BlockResponse{ChainId: chainId, Nonce: utils.RandomUint64()}
	for i := current; i < request; i++ {
		block, err := e.ledger.GetTxBlockByHeight(chainId, i+1)
		if err != nil {
			return err
		}
		response.Blocks = append(response.Blocks, block)
		num++
		if num >= 10 {
			if err := event.Send(event.ActorNil, event.ActorP2P, response); err != nil {
				return err
			}
			num = 0
			response = &BlockResponse{ChainId: chainId, Nonce: utils.RandomUint64()}
		}
	}
	if len(response.Blocks) > 0 {
		if err := event.Send(event.ActorNil, event.ActorP2P, response); err != nil {
			return err
		}
	}
	return nil
}

func (e *Engine) HandleBlockResponse(msg *mpb.Message) error {
	response := new(BlockResponse)
	if err := response.Deserialize(msg.Payload); err != nil {
		return err
	}
	for _, block := range response.Blocks {
		if err := event.Send(event.ActorNil, event.ActorLedger, block); err != nil {
			log.Error("send block to ledger error:", err)
		}
	}
	return nil
}
