package example

import (
	"testing"
	"github.com/ecoball/go-ecoball/core/types"
	"github.com/ecoball/go-ecoball/common/config"
	"time"
	"github.com/ecoball/go-ecoball/common/errors"
	"fmt"
	"github.com/ecoball/go-ecoball/common/utils"
	"github.com/ecoball/go-ecoball/common"
)

func TestTransferBlock(t *testing.T) {
	l := ShardLedger("/tmp/example/")
	r, err := l.AccountGet(config.ChainHash, common.NameToIndex("root"))
	errors.CheckErrorPanic(err)
	fmt.Println(r.JsonString(false))
	fmt.Println(l.RequireResources(config.ChainHash, common.NameToIndex("root"), time.Now().UnixNano()))

	var txs []*types.Transaction
	for i := 0; i < 2000; i ++ {
		tx := TestTransfer()
		txs = append(txs, tx)
	}
	block, _, err := l.NewMinorBlock(config.ChainHash, txs, time.Now().UnixNano())
	errors.CheckErrorPanic(err)
	data, err := block.Serialize()
	errors.CheckErrorPanic(err)
	fmt.Println("len:", len(data))

	utils.FileWrite("test.dat", data)
}
