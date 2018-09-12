package block

import (
	"github.com/go-redis/redis"
	"github.com/syndtr/goleveldb/leveldb/iterator"
	"github.com/ecoball/go-ecoball/dsn/common"
	"context"
	//"github.com/ecoball/go-ecoball/dsn/ipfs/ipld"
	coreiface "github.com/ipfs/go-ipfs/core/coreapi/interface"
	opt "github.com/ipfs/go-ipfs/core/coreapi/interface/options"
	"bytes"
	"gx/ipfs/QmYVNvtQkeZ6AKSwDrjQTs432QtL6umrrK41EBq3cu7iSP/go-cid"
)

type DsnStore struct {
	rClient  *redis.Client
	api      coreiface.CoreAPI
	ctx      context.Context
}

func NewDsnStore(ctx context.Context) *DsnStore {
	return &DsnStore{
		rClient:common.InitRedis(common.DefaultRedisConf()),
		ctx:ctx,
	}
}

func (ds *DsnStore) Put(key, value []byte) error  {
	skey := string(key)
	r := bytes.NewReader(value)
	rp, err := ds.api.Dag().Put(ds.ctx, r, opt.Dag.InputEnc("raw"), opt.Dag.Codec(cid.EcoballRawData))
	if err != nil {
		return err
	}
	cidValue := rp.Root().String()
	ret := ds.rClient.Set(skey, cidValue, -1)
	return ret.Err()
}

func (ds *DsnStore) Get(key []byte) ([]byte, error)  {
	skey := string(key)
	cid := ds.rClient.Get(skey).String()
	path, err := coreiface.ParsePath(cid)
	if err != nil {
		return nil, nil
	}
	node, err := ds.api.Dag().Get(ds.ctx, path)
	if err != nil {
		return nil, err
	}
	return node.RawData(), nil
}

func (ds *DsnStore) Has(key []byte) (bool, error) {
	skey := string(key)
	ret := ds.rClient.Exists(skey)
	retInt, _ := ret.Result()
	if retInt == 1 {
		return true, nil
	}
	return false, ret.Err()
}

func (ds *DsnStore) Delete(key []byte) error {
	skey := string(key)
	ret := ds.rClient.Del(skey)
	return ret.Err()
}

func (ds *DsnStore) BatchPut(key, value []byte) {

}

func (ds *DsnStore) BatchCommit() error {
	return nil
}

func (ds *DsnStore) SearchAll() (result map[string]string, err error) {
	return nil, nil
}

func (ds *DsnStore) DeleteAll() error {
	return nil
}

func (ds *DsnStore) NewIterator() iterator.Iterator {
	return nil
}
