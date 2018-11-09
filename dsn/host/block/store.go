package block

import (
	"github.com/syndtr/goleveldb/leveldb/iterator"
	"context"
	"github.com/ecoball/go-ecoball/dsn/host/ipfs/api"
	"io/ioutil"
	"github.com/ecoball/go-ecoball/core/store"
)

type DsnStore struct {
	ldb		 store.Storage
	ctx      context.Context
}

func NewDsnStore(path string) (store.Storage, error) {
	db, err := store.NewBlockStore(path)
	if err != nil {
		return nil, err
	}
	ctx := context.Background()
	return &DsnStore{
		ldb: db,
		ctx:ctx,
	}, nil
}

func (ds *DsnStore) Put(key, value []byte) error  {
	cid, err := api.IpfsAbaBlkPut(ds.ctx, value)
	if err != nil {
		return err
	}
	return ds.ldb.Put(key, []byte(cid))
}

func (ds *DsnStore) Get(key []byte) ([]byte, error)  {
	cid, err := ds.ldb.Get(key)
	if err != nil {
		return nil, err
	}
	scid := string(cid)
	blk, err := api.IpfsBlockGet(ds.ctx, scid)
	if err != nil {
		return nil, nil
	}
	return ioutil.ReadAll(blk)
}

func (ds *DsnStore) Has(key []byte) (bool, error) {
	return  ds.ldb.Has(key)
}

func (ds *DsnStore) Delete(key []byte) error {
	ds.ldb.Delete(key)
	return api.IpfsBlockDel(ds.ctx, string(key))
}

func (ds *DsnStore) BatchPut(key, value []byte) {
	cid, _ := api.IpfsAbaBlkPut(ds.ctx, value)
	ds.ldb.BatchPut(key, []byte(cid))
}

func (ds *DsnStore) BatchCommit() error {
	return ds.ldb.BatchCommit()
}

func (ds *DsnStore) SearchAll() (result map[string]string, err error) {
	ret := make(map[string]string, 0)
	keys, err := ds.ldb.SearchAll()
	for k, v := range keys {
		r, err := api.IpfsBlockGet(ds.ctx, string(v))
		if err != nil {
			ret[string(k)] = ""
		} else {
			data, _ := ioutil.ReadAll(r)
			ret[string(k)] = string(data)
		}
	}
	return ret, nil
}

func (ds *DsnStore) DeleteAll() error {
	keys, _ := ds.ldb.SearchAll()
	for _, v := range keys {
		api.IpfsBlockDel(ds.ctx, string(v))
	}
	return ds.ldb.DeleteAll()
}

func (ds *DsnStore) NewIterator() iterator.Iterator {
	return ds.ldb.NewIterator()
}
