package audit

import (
	"context"
	"github.com/ecoball/go-ecoball/core/ledgerimpl/ledger"
	"github.com/ecoball/go-ecoball/core/state"
	"github.com/ecoball/go-ecoball/common/elog"
	"github.com/ecoball/go-ecoball/common/config"
	"github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/dsn/host"
)

type DsnConf struct {
}

type Dsn struct {
	h    *host.StorageHost
	s    *Settler
	ctx  context.Context
}

var (
	dsn *Dsn
	log = elog.NewLogger("dsn", elog.DebugLog)
)
func InitDefaultConf() DsnConf {
	return DsnConf{
	}
}

func StartDsn(ctx context.Context, l ledger.Ledger) error {
	dsn = new(Dsn)
	h := host.NewHostWithDefaultConf()
	h.Start()

	//r := renter.NewRenter(ctx, l, ra, conf.rConf)
	//go r.Start()
	s, _ := NewStorageSettler(ctx, l, common.ToHex(config.ChainHash[:]))
	//go s.Start()

	dsn.h = h
	//dsn.r = r
	dsn.s = s
	dsn.ctx = ctx

	//go DsnHttpServ()

	return nil
}

func HandleStoreAnn(para string, st state.InterfaceState)  {
	log.Debug("Handle storage announce...")
	data := []byte(para)
	dsn.s.HandleHostAnce(data, st)
}

func HandleStorageProof(para string, st state.InterfaceState)  {
	log.Debug("Handle storage proof...")
	data := []byte(para)
	dsn.s.HandleStorageProof(data, st)
}

func HandleFileContract(para string, st state.InterfaceState)  {
	log.Debug("Handle file contract...")
	data := []byte(para)
	dsn.s.HandleFileContract(data, st)
}
