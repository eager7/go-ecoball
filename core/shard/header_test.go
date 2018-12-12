package shard_test

import (
	"github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/common/config"
	"github.com/ecoball/go-ecoball/common/elog"
	"github.com/ecoball/go-ecoball/common/errors"
	"github.com/ecoball/go-ecoball/core/bloom"
	"github.com/ecoball/go-ecoball/core/shard"
	"github.com/ecoball/go-ecoball/core/types"
	"github.com/ecoball/go-ecoball/test/example"
	"testing"
	"time"
)

func TestMinorBlockHeader(t *testing.T) {
	header := shard.MinorBlockHeader{
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
		COSign: &types.COSign{
			TPubKey: []byte("tPubKey"),
			Step1:   10,
			Sign1:   [][]byte{[]byte("sign1"), []byte("sign11")},
			Step2:   20,
			Sign2:   [][]byte{[]byte("sign2"), []byte("sign22")},
		},
	}
	errors.CheckErrorPanic(header.ComputeHash())
	data, err := header.Serialize()
	errors.CheckErrorPanic(err)

	headerNew := shard.MinorBlockHeader{}
	errors.CheckErrorPanic(headerNew.Deserialize(data))
	errors.CheckEqualPanic(header.JsonString() == headerNew.JsonString())

	block, err := shard.NewMinorBlock(header, nil, []*types.Transaction{example.TestTransfer()}, 0, 0)
	data, err = block.Serialize()
	errors.CheckErrorPanic(err)
	blockNew := shard.MinorBlock{}
	errors.CheckErrorPanic(blockNew.Deserialize(data))
	elog.Log.Debug(block.JsonString())
	elog.Log.Info(blockNew.JsonString())
	errors.CheckEqualPanic(block.JsonString() == blockNew.JsonString())
}

func TestCmBlockHeader(t *testing.T) {
	header := shard.CMBlockHeader{
		ChainID:   config.ChainHash,
		Version:   0,
		Height:    10,
		Timestamp: 2340,
		PrevHash:  common.Hash{},
		//ConsData:     example.ConsensusData(),
		LeaderPubKey: []byte("12345678909876554432"),
		Nonce:        23450,
		Candidate: shard.NodeInfo{
			PublicKey: config.Root.PublicKey,
			Address:   "1234",
			Port:      "5678",
		},
		ShardsHash: config.ChainHash,
		COSign: &types.COSign{
			TPubKey: []byte("tPubKey"),
			Step1:   10,
			Sign1:   [][]byte{[]byte("sign1"), []byte("sign11")},
			Step2:   20,
			Sign2:   [][]byte{[]byte("sign2"), []byte("sign22")},
		},
	}
	errors.CheckErrorPanic(header.ComputeHash())
	data, err := header.Serialize()
	errors.CheckErrorPanic(err)

	headerNew := shard.CMBlockHeader{}
	errors.CheckErrorPanic(headerNew.Deserialize(data))
	elog.Log.Debug(header.JsonString())
	elog.Log.Info(headerNew.JsonString())
	errors.CheckEqualPanic(header.JsonString() == headerNew.JsonString())

	Shards := []shard.Shard{shard.Shard{
		Member: []shard.NodeInfo{
			{
				PublicKey: []byte("12340987"),
				Address:   "ew62",
				Port:      "34523532",
			},
		},
		MemberAddr: []shard.NodeAddr{{
			Address: "1234",
			Port:    "5678",
		}},
	}}

	block, err := shard.NewCmBlock(header, Shards)
	errors.CheckErrorPanic(err)
	data, err = block.Serialize()
	errors.CheckErrorPanic(err)
	blockNew := shard.CMBlock{}
	errors.CheckErrorPanic(blockNew.Deserialize(data))
	elog.Log.Notice(block.JsonString())
	elog.Log.Debug(blockNew.JsonString())
	errors.CheckEqualPanic(block.JsonString() == blockNew.JsonString())
}

func TestFinalBlockHeader(t *testing.T) {
	header := shard.FinalBlockHeader{
		ChainID:   config.ChainHash,
		Version:   10,
		Height:    120,
		Timestamp: 3450,
		TrxCount:  670,
		PrevHash:  config.ChainHash,
		//ConsData:           example.ConsensusData(),
		ProposalPubKey:     []byte("123678435634w453226435"),
		EpochNo:            570,
		CMBlockHash:        config.ChainHash,
		TrxRootHash:        config.ChainHash,
		StateDeltaRootHash: config.ChainHash,
		MinorBlocksHash:    config.ChainHash,
		StateHashRoot:      config.ChainHash,
		COSign: &types.COSign{
			TPubKey: []byte("tPubKey"),
			Step1:   10,
			Sign1:   [][]byte{[]byte("sign1"), []byte("sign11")},
			Step2:   20,
			Sign2:   [][]byte{[]byte("sign2"), []byte("sign22")},
		},
	}
	errors.CheckErrorPanic(header.ComputeHash())
	data, err := header.Serialize()
	errors.CheckErrorPanic(err)

	headerNew := shard.FinalBlockHeader{}
	errors.CheckErrorPanic(headerNew.Deserialize(data))
	errors.CheckEqualPanic(header.JsonString() == headerNew.JsonString())

	headerMinor := shard.MinorBlockHeader{
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
		COSign: &types.COSign{
			TPubKey: []byte("tPubKey"),
			Step1:   10,
			Sign1:   [][]byte{[]byte("sign1"), []byte("sign11")},
			Step2:   20,
			Sign2:   [][]byte{[]byte("sign2"), []byte("sign22")},
		},
	}
	block := shard.FinalBlock{
		FinalBlockHeader: header,
		MinorBlocks:      []*shard.MinorBlockHeader{&headerMinor},
	}
	data, err = block.Serialize()
	errors.CheckErrorPanic(err)
	blockNew := shard.FinalBlock{}
	errors.CheckErrorPanic(blockNew.Deserialize(data))
	errors.CheckEqualPanic(block.JsonString() == blockNew.JsonString())
}

func TestVCBlockHeader(t *testing.T) {
	//Init ViewChange Block
	headerVC := shard.ViewChangeBlockHeader{
		ChainID:          config.ChainHash,
		Version:          types.VersionHeader,
		Height:           1,
		Timestamp:        time.Now().UnixNano(),
		PrevHash:         common.Hash{},
		CMEpochNo:        1,
		FinalBlockHeight: 1,
		Round:            0,
		Candidate:        shard.NodeInfo{},
		COSign: &types.COSign{
			TPubKey: []byte("tPubKey"),
			Step1:   10,
			Sign1:   [][]byte{[]byte("sign1"), []byte("sign11")},
			Step2:   20,
			Sign2:   [][]byte{[]byte("sign2"), []byte("sign22")},
		},
	}
	data, err := headerVC.Serialize()
	headerVC2 := new(shard.ViewChangeBlockHeader)
	errors.CheckErrorPanic(headerVC2.Deserialize(data))
	errors.CheckEqualPanic(headerVC.JsonString() == headerVC2.JsonString())

	blockVC, err := shard.NewVCBlock(headerVC)
	errors.CheckErrorPanic(err)
	data, err = blockVC.Serialize()

	blockVC2 := new(shard.ViewChangeBlock)
	errors.CheckErrorPanic(blockVC2.Deserialize(data))

	errors.CheckEqualPanic(blockVC.JsonString() == blockVC2.JsonString())
}

func TestHeader(t *testing.T) {
	conData := types.ConsensusData{Type: types.ConSolo, Payload: &types.SoloData{}}
	h, err := types.NewHeader(types.VersionHeader, config.ChainHash, 10, common.Hash{}, common.Hash{}, common.Hash{}, conData, bloom.Bloom{}, config.BlockCpuLimit, config.BlockNetLimit, time.Now().Unix())
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
	h, err = types.NewHeader(types.VersionHeader, config.ChainHash, 10, common.Hash{}, common.Hash{}, common.Hash{}, conData, bloom.Bloom{}, config.BlockCpuLimit, config.BlockNetLimit, time.Now().Unix())
	errors.CheckErrorPanic(err)
	errors.CheckErrorPanic(h.SetSignature(&config.Root))

	data, err = h.Serialize()
	errors.CheckErrorPanic(err)

	h2 = new(types.Header)
	errors.CheckErrorPanic(h2.Deserialize(data))
	errors.CheckEqualPanic(h.JsonString() == h2.JsonString())
}
