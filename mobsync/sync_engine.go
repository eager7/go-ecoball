package mobsync

import (
	"context"
	"fmt"
	"github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/common/config"
	"github.com/ecoball/go-ecoball/common/elog"
	"github.com/ecoball/go-ecoball/common/errors"
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
	state   bool
	cache   *ChainMap
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
		if engine.state {
			current := ledger.GetCurrentHeader(config.ChainHash)
			if err := event.Send(event.ActorNil, event.ActorP2P, &BlockRequest{ChainId: current.ChainID, BlockHeight: current.Height, Nonce: utils.RandomUint64()}); err != nil {
				log.Error(err)
			}
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
	e.cache = new(ChainMap).Initialize()
	e.state = true //启动后先同步一次
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
				e.state = false //被动同步过程中,暂时关闭主动同步信号
				log.Info("sync engine receive msg:", in.Identify.String())
				switch in.Identify {
				case mpb.Identify_APP_MSG_TRANSACTION:
				case mpb.Identify_APP_MSG_BLOCK:
					if err := e.SyncBlockChain(in); err != nil {
						log.Error("sync block failed:", err)
					}
				case mpb.Identify_APP_MSG_BLOCK_REQUEST:
					if err := e.HandleBlockRequest(in); err != nil {
						log.Error("handle block sync request error:", err)
					} else {
						log.Debug("handle block sync request success")
					}
				case mpb.Identify_APP_MSG_BLOCK_RESPONSE:
					if err := e.HandleBlockResponse(in); err != nil {
						log.Error("handle block sync response error:", err)
					}
				default:
					log.Warn("unsupported sync message:", in.Identify.String())
				}
				e.state = true
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
	if current == nil {
		return errors.New(fmt.Sprintf("can't find the current header:%s", block.ChainID.String()))
	}
	if current.Height+1 < block.Height {
		log.Debug("send block request message:", block.ChainID.String(), current.Height)
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
	if current == nil {
		return errors.New(fmt.Sprintf("can't find the current header:%s", request.ChainId.String()))
	}
	if current.Height <= request.BlockHeight {
		log.Info("our chain block is older than request:", current.Height, request.BlockHeight)
		return nil
	}
	return e.SendBlockResponse(request.ChainId, current.Height, request.BlockHeight)
}

func (e *Engine) SendBlockResponse(chainId common.Hash, current, request uint64) error {
	log.Debug("prepare sync block response message:", current, request)
	var num int
	response := &BlockResponse{ChainId: chainId, Nonce: utils.RandomUint64()}
	for i := request; i < current; i++ {
		block, err := e.ledger.GetTxBlockByHeight(chainId, i+1)
		if err != nil {
			return err
		}
		response.Blocks = append(response.Blocks, block)
		num++
		if num >= 10 {
			log.Debug("send sync block response msg:", response.String())
			if err := event.Send(event.ActorNil, event.ActorP2P, response); err != nil {
				return err
			}
			num = 0
			response = &BlockResponse{ChainId: chainId, Nonce: utils.RandomUint64()}
		}
	}
	if len(response.Blocks) > 0 {
		log.Debug("send sync block response msg:", response.String())
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
	if len(response.Blocks) == 0 {
		return nil
	}
	for _, block := range response.Blocks {
		e.cache.Add(block.ChainID, block)
	}

	chainId := response.Blocks[0].ChainID
	current := e.ledger.GetCurrentHeader(chainId)
	if current == nil {
		return errors.New(fmt.Sprintf("can't find the current header:%s", chainId.String()))
	}
	if chain := e.cache.Get(chainId); chain != nil {
		for b := range chain.IteratorByHeight(chainId) {
			if current.Height+1 < b.Height {
				return errors.New(fmt.Sprintf("the chain:%s maybe still lost blocks, current:%d, recive min:%d", chainId.String(), current.Height, b.Height))
			}
			if b.Height < current.Height+1 {
				continue
			}
			if err := event.Send(event.ActorNil, event.ActorLedger, b); err != nil {
				log.Error("send block to ledger error:", err)
			}
			chain.Del(b.Height)
		}
	}
	return nil
}
