package types_test

import (
	"github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/common/config"
	"github.com/ecoball/go-ecoball/common/elog"
	"github.com/ecoball/go-ecoball/common/errors"
	"github.com/ecoball/go-ecoball/core/bloom"
	"github.com/ecoball/go-ecoball/core/types"
	"github.com/ecoball/go-ecoball/test/example"
	"math/big"
	"testing"
	"time"
)

func TestMinorBlockHeader(t *testing.T) {
	header := types.MinorBlockHeader{
		ChainID:           config.ChainHash,
		Version:           1,
		Height:            1,
		Timestamp:         time.Now().UnixNano(),
		PrevHash:          common.Hash{},
		TrxHashRoot:       common.Hash{},
		StateDeltaHash:    common.Hash{},
		CMBlockHash:       common.Hash{},
		ProposalPublicKey: []byte("1234567890"),
		ShardId:           1,
		CMEpochNo:         2,
		Receipt:           types.BlockReceipt{},
		COSign:            &types.COSign{
			Step1: 10,
			Step2: 20,
		},
	}
	errors.CheckErrorPanic(header.ComputeHash())
	data, err := header.Serialize()
	errors.CheckErrorPanic(err)

	headerNew := types.MinorBlockHeader{}
	errors.CheckErrorPanic(headerNew.Deserialize(data))
	errors.CheckEqualPanic(header.JsonString() == headerNew.JsonString())

	block := types.MinorBlock{
		MinorBlockHeader: header,
		Transactions:     []*types.Transaction{example.TestTransfer()},
		StateDelta: []*types.AccountMinor{{
			Balance: new(big.Int).SetUint64(100),
			Nonce:   new(big.Int).SetUint64(2),
		}}}
	data, err = block.Serialize()
	errors.CheckErrorPanic(err)
	blockNew := types.MinorBlock{}
	errors.CheckErrorPanic(blockNew.Deserialize(data))
	errors.CheckEqualPanic(block.JsonString() == blockNew.JsonString())
}

func TestCmBlockHeader(t *testing.T) {
	header := types.CMBlockHeader{
		ChainID:      config.ChainHash,
		Version:      0,
		Height:       10,
		Timestamp:    2340,
		PrevHash:     common.Hash{},
		//ConsData:     example.ConsensusData(),
		LeaderPubKey: []byte("12345678909876554432"),
		Nonce:        23450,
		Candidate: types.NodeInfo{
			PublicKey: config.Root.PublicKey,
			Address:   "1234",
			Port:      "5678",
		},
		ShardsHash: config.ChainHash,
		COSign:            &types.COSign{
			Step1: 10,
			Step2: 20,
		},
	}
	errors.CheckErrorPanic(header.ComputeHash())
	data, err := header.Serialize()
	errors.CheckErrorPanic(err)

	headerNew := types.CMBlockHeader{}
	errors.CheckErrorPanic(headerNew.Deserialize(data))
	elog.Log.Debug(header.JsonString())
	elog.Log.Info(headerNew.JsonString())
	errors.CheckEqualPanic(header.JsonString() == headerNew.JsonString())

	block := types.CMBlock{
		CMBlockHeader: header,
		Shards: []types.Shard{types.Shard{
			Id: 10,
			Member: []types.NodeInfo{
				{
					PublicKey: []byte("12340987"),
					Address:   "ew62",
					Port:      "34523532",
				},
			},
			MemberAddr: []types.NodeAddr{{
				Address: "1234",
				Port:    "5678",
			}},
		}},
	}
	data, err = block.Serialize()
	errors.CheckErrorPanic(err)
	blockNew := types.CMBlock{}
	errors.CheckErrorPanic(blockNew.Deserialize(data))
	elog.Log.Notice(block.JsonString())
	elog.Log.Debug(blockNew.JsonString())
	errors.CheckEqualPanic(block.JsonString() == blockNew.JsonString())
}

func TestFinalBlockHeader(t *testing.T) {
	header := types.FinalBlockHeader{
		ChainID:            config.ChainHash,
		Version:            10,
		Height:             120,
		Timestamp:          3450,
		TrxCount:           670,
		PrevHash:           config.ChainHash,
		//ConsData:           example.ConsensusData(),
		ProposalPubKey:     []byte("123678435634w453226435"),
		EpochNo:            570,
		CMBlockHash:        config.ChainHash,
		TrxRootHash:        config.ChainHash,
		StateDeltaRootHash: config.ChainHash,
		MinorBlocksHash:    config.ChainHash,
		StateHashRoot:      config.ChainHash,
		COSign:            &types.COSign{
			Step1: 10,
			Step2: 20,
		},
	}
	errors.CheckErrorPanic(header.ComputeHash())
	data, err := header.Serialize()
	errors.CheckErrorPanic(err)

	headerNew := types.FinalBlockHeader{}
	errors.CheckErrorPanic(headerNew.Deserialize(data))
	errors.CheckEqualPanic(header.JsonString() == headerNew.JsonString())

	headerMinor := types.MinorBlockHeader{
		ChainID:           config.ChainHash,
		Version:           1,
		Height:            1,
		Timestamp:         time.Now().UnixNano(),
		PrevHash:          common.Hash{},
		TrxHashRoot:       common.Hash{},
		StateDeltaHash:    common.Hash{},
		CMBlockHash:       common.Hash{},
		ProposalPublicKey: []byte("1234567890"),
		//ConsData:          example.ConsensusData(),
		ShardId:           1,
		CMEpochNo:         2,
		Receipt:           types.BlockReceipt{},
		COSign:            &types.COSign{
			Step1: 10,
			Step2: 20,
		},
	}
	block := types.FinalBlock{
		FinalBlockHeader: header,
		MinorBlocks:      []*types.MinorBlockHeader{&headerMinor},
	}
	data, err = block.Serialize()
	errors.CheckErrorPanic(err)
	blockNew := types.FinalBlock{}
	errors.CheckErrorPanic(blockNew.Deserialize(data))
	errors.CheckEqualPanic(block.JsonString() == blockNew.JsonString())
}

func TestHeader(t *testing.T) {
	conData := types.ConsensusData{Type: types.ConSolo, Payload: &types.SoloData{}}
	h, err := types.NewHeader(types.VersionHeader, config.ChainHash, 10, common.Hash{}, common.Hash{}, common.Hash{}, conData, bloom.Bloom{}, types.BlockCpuLimit, types.BlockNetLimit, time.Now().Unix())
	errors.CheckErrorPanic(err)
	errors.CheckErrorPanic(h.SetSignature(&config.Root))

	data, err := h.Serialize()
	errors.CheckErrorPanic(err)

	h2 := new(types.Header)
	errors.CheckErrorPanic(h2.Deserialize(data))

	errors.CheckEqualPanic(h.JsonString() == h2.JsonString())

	//ABA BFT
	sig1 := common.Signature{PubKey: []byte("1234"), SigData: []byte("5678")}
	sig2 := common.Signature{PubKey: []byte("4321"), SigData: []byte("8765")}
	var sigPer []common.Signature
	sigPer = append(sigPer, sig1)
	sigPer = append(sigPer, sig2)
	abaData := types.AbaBftData{NumberRound: 5, PreBlockSignatures: sigPer}
	conData = types.ConsensusData{Type: types.ConABFT, Payload: &abaData}
	h, err = types.NewHeader(types.VersionHeader, config.ChainHash, 10, common.Hash{}, common.Hash{}, common.Hash{}, conData, bloom.Bloom{}, types.BlockCpuLimit, types.BlockNetLimit, time.Now().Unix())
	errors.CheckErrorPanic(err)
	errors.CheckErrorPanic(h.SetSignature(&config.Root))

	data, err = h.Serialize()
	errors.CheckErrorPanic(err)

	h2 = new(types.Header)
	errors.CheckErrorPanic(h2.Deserialize(data))
	errors.CheckEqualPanic(h.JsonString() == h2.JsonString())
}
