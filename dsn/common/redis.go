package common

import (
	"github.com/go-redis/redis"
	"time"
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
