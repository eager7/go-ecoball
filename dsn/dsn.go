package dsn

import (
	"github.com/ecoball/go-ecoball/dsn/host"
	"github.com/ecoball/go-ecoball/dsn/renter"
	"github.com/ecoball/go-ecoball/dsn/settlement"
	"context"
	"github.com/ecoball/go-ecoball/account"
	"github.com/ecoball/go-ecoball/core/ledgerimpl/ledger"
	"github.com/ecoball/go-ecoball/core/state"
	"github.com/ecoball/go-ecoball/common/elog"
	"github.com/ecoball/go-ecoball/common/config"
	"github.com/ecoball/go-ecoball/common"
	"io"
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

var (
	dsn *Dsn
	log = elog.NewLogger("dsn", elog.DebugLog)
)
func InitDefaultConf() DsnConf {
	return DsnConf{
		hConf: host.InitDefaultConf(),
		rConf: renter.InitDefaultConf(),
	}
}

func StartDsn(ctx context.Context, l ledger.Ledger) error {
	dsn = new(Dsn)
	//TODO ha should be host's account
	ha, err := account.NewAccount(0)
	if err != nil {
		return err
	}
	//TODO ra should be renter's account
	ra, err := account.NewAccount(0)
	if err != nil {
		return err
	}
	//TODO conf should be user's config
	conf := InitDefaultConf()
	h := host.NewStorageHost(ctx, l, ha, conf.hConf)
	go h.Start()
	r := renter.NewRenter(ctx, l, ra, conf.rConf)
	//go r.Start()
	s, _ := settlement.NewStorageSettler(ctx, l, common.ToHex(config.ChainHash[:]))
	//go s.Start()

	dsn.h = h
	dsn.r = r
	dsn.s = s
	dsn.ctx = ctx
	return nil
}

func AddFile(file string, era int8) (string, error) {
	era = 2
	return dsn.r.AddFile(file, era)
}

func CatFile(cid string) (io.Reader, error) {
	dsn.r.CatFile(cid)
	return nil, nil
}

func PinToServer(cid string) error {
	//TODO
	return nil
}

func HandleStoreAnn(para string, st state.InterfaceState)  {
	data := []byte(para)
	dsn.s.HandleHostAnce(data, st)
}

func HandleStorageProof(para string, st state.InterfaceState)  {
	data := []byte(para)
	dsn.s.HandleStorageProof(data, st)
}

func HandleFileContract(para string, st state.InterfaceState)  {
	data := []byte(para)
	dsn.s.HandleFileContract(data, st)
}
