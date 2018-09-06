package settlement

import (
	"github.com/go-redis/redis"
	"time"
	"fmt"
)

type RedisConf struct {
	Addr         string
	DialTimeout  int64
	ReadTimeout  int64
	WriteTimeout int64
	PoolSize     int
	PoolTimeout  int64
}

func DefaultRedisConf() RedisConf {
	return RedisConf{
		Addr: ":6379",
		DialTimeout: 10,
		ReadTimeout: 30,
		WriteTimeout: 30,
		PoolSize: 10,
		PoolTimeout: 30,
	}
}

func InitRedis(conf RedisConf) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr: conf.Addr,
		DialTimeout: time.Duration(conf.DialTimeout) * time.Second,
		ReadTimeout:time.Duration(conf.ReadTimeout) * time.Second,
		WriteTimeout:time.Duration(conf.WriteTimeout) * time.Second,
		PoolSize: conf.PoolSize,
		PoolTimeout: time.Duration(conf.PoolTimeout) * time.Second,
	})
}

func (s *Settler)getEcoballTotalCap() uint64 {
	rKey := "host_*"
	kHosts := s.rClient.Keys(rKey)
	var tSize uint64
	for _, v := range kHosts.Val() {
		ret := s.rClient.HGet(v, "total")
		if ret.Err() != nil {
			continue
		}
		size, err := ret.Int64()
		if err != nil {
			tSize = tSize + uint64(size)
		}
	}
	return tSize
}

func (s *Settler) getHostTotalCap(pk string) uint64 {
	pKey := fmt.Sprintf("host_%s", pk)
	ret := s.rClient.HGet(pKey, "total")
	if ret.Err() != nil {
		return 0
	}
	size, _ := ret.Int64()
	return uint64(size)
}

func (s *Settler)getEcoballRepoSize() uint64 {
	//TODO
	return 0
}

func (s *Settler)getHostRepoSize(pk string) uint64 {
	//TODO
	return 0
}

func (s *Settler)getHostOnlineTime(pk string) uint32 {
	//TODO
	return 0
}

func (s *Settler)getRenterUsedSize(pk string) uint64 {
	//TODO
	return 0
}
