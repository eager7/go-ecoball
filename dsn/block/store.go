package block

import (
	"github.com/go-redis/redis"
	"github.com/syndtr/goleveldb/leveldb/iterator"
	"github.com/ecoball/go-ecoball/dsn/common"
	"context"
	"github.com/ecoball/go-ecoball/dsn/ipfs/api"
	"io/ioutil"
)

type DsnStore struct {
	rClient  *redis.Client
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
	cid, err := api.IpfsAbaBlkPut(ds.ctx, value)
	if err != nil {
		return err
	}
	ret := ds.rClient.Set(skey, cid, -1)
	return ret.Err()
}

func (ds *DsnStore) Get(key []byte) ([]byte, error)  {
	skey := string(key)
	cid := ds.rClient.Get(skey).String()
	blk, err := api.IpfsBlockGet(ds.ctx, cid)
	if err != nil {
		return nil, nil
	}
	return ioutil.ReadAll(blk)
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
