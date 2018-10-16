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
	//"github.com/ecoball/go-ecoball/dsn/api"
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

	//api.DsnHttpServ()

	return nil
}

func AddFile(file string, era int8) (string, error) {
	log.Debug("add file ", file, "era ", era)
	return dsn.r.AddFile(file, era)
}

func CatFile(cid string) (io.Reader, error) {
	r,err := dsn.r.CatFile(cid)
	if err != nil {
		log.Error("cat ", cid, " failed")
		return nil, nil
	}

	//d, err := ioutil.ReadAll(r)
	//ioutil.WriteFile("/root/ecoball/t.toml", d, os.ModePerm)
	return r, err
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
