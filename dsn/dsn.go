package dsn

import (
	"github.com/ecoball/go-ecoball/dsn/host"
	"github.com/ecoball/go-ecoball/dsn/renter"
	"github.com/ecoball/go-ecoball/dsn/settlement"
	"context"
	"github.com/ecoball/go-ecoball/account"
	"github.com/ecoball/go-ecoball/core/ledgerimpl/ledger"
	"github.com/ecoball/go-ecoball/core/state"
)

type DsnConf struct {
	hConf host.StorageHostConf
	rConf renter.RenterConf
}

type Dsn struct {
	h    *host.StorageHost
	r    *renter.Renter
	s    *settlement.Settler
	ctx  context.Context
}

var dsn *Dsn

func StartDsn(ctx context.Context, l ledger.Ledger, a account.Account, conf DsnConf)  {
	h := host.NewStorageHost(ctx, l, a, conf.hConf)
	go h.Start()
	r := renter.NewRenter(ctx, l, a, conf.rConf)
	go r.Start()
	s := settlement.NewStorageSettler(ctx, l)
	go s.Start()

	dsn.h = h
	dsn.r = r
	dsn.s = s
	dsn.ctx = ctx
}

func AddFile(file string) (string, error) {
	return dsn.r.AddFile(file)
}

func HandleStoreAnn(para string, st state.InterfaceState)  {
	data := []byte(para)
	dsn.s.HandleHostAnce(data, st)
}

func HandleStorageProof(para string, st state.InterfaceState)  {

}
