package types_test

import (
	"fmt"
	"github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/common/config"
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
		ConsData:          example.ConsensusData(),
		ShardId:           1,
		CMEpochNo:         2,
		Receipt:           types.BlockReceipt{},
		Hash:              common.Hash{},
	}
	errors.CheckErrorPanic(header.ComputeHash())
	data, err := header.Serialize()
	errors.CheckErrorPanic(err)

	headerNew := types.MinorBlockHeader{}
	errors.CheckErrorPanic(headerNew.Deserialize(data))
	errors.CheckEqualPanic(header.JsonString() == headerNew.JsonString())

	block := types.MinorBlock{
		Header:       &header,
		Transactions: []*types.Transaction{example.TestTransfer()},
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
		LeaderPubKey:    []byte("1234567890"),
		CandidatePubKey: []byte("CandidatePubKey"),
		Nonce:           110,
		ShardsHash:      common.SingleHash([]byte("ShardsHash")),
	}
	data, err := header.Serialize()
	errors.CheckErrorPanic(err)

	headerNew := types.CMBlockHeader{}
	errors.CheckErrorPanic(headerNew.Deserialize(data))
	headerNew.Show()
	errors.CheckEqualPanic(header.JsonString() == headerNew.JsonString())
}

func TestHeader(t *testing.T) {
	payload := &types.MinorBlockHeader{
		ProposalPublicKey: []byte("1234567890")}
	conData := types.ConsensusData{Type: types.ConSolo, Payload: &types.SoloData{}}
	h, err := types.NewHeader(payload, types.VersionHeader, config.ChainHash, 10, common.Hash{}, common.Hash{}, common.Hash{}, conData, bloom.Bloom{}, types.BlockCpuLimit, types.BlockNetLimit, time.Now().Unix())
	errors.CheckErrorPanic(err)
	errors.CheckErrorPanic(h.SetSignature(&config.Root))
	h.Show()

	data, err := h.Serialize()
	errors.CheckErrorPanic(err)

	h2 := new(types.Header)
	errors.CheckErrorPanic(h2.Deserialize(data))

	h2.Show()
	errors.CheckEqualPanic(h.JsonString() == h2.JsonString())

	//ABA BFT
	sig1 := common.Signature{PubKey: []byte("1234"), SigData: []byte("5678")}
	sig2 := common.Signature{PubKey: []byte("4321"), SigData: []byte("8765")}
	var sigPer []common.Signature
	sigPer = append(sigPer, sig1)
	sigPer = append(sigPer, sig2)
	abaData := types.AbaBftData{NumberRound: 5, PreBlockSignatures: sigPer}
	conData = types.ConsensusData{Type: types.ConABFT, Payload: &abaData}
	h, err = types.NewHeader(payload, types.VersionHeader, config.ChainHash, 10, common.Hash{}, common.Hash{}, common.Hash{}, conData, bloom.Bloom{}, types.BlockCpuLimit, types.BlockNetLimit, time.Now().Unix())
	errors.CheckErrorPanic(err)
	errors.CheckErrorPanic(h.SetSignature(&config.Root))

	data, err = h.Serialize()
	errors.CheckErrorPanic(err)

	h2 = new(types.Header)
	errors.CheckErrorPanic(h2.Deserialize(data))
	errors.CheckEqualPanic(h.JsonString() == h2.JsonString())
	h2.Show()
}
