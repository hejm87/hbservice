package mservice_util

import (
	"sync"
	"gopkg.in/redis.v5"
	"hbservice/src/util"
	"hbservice/src/mservice/define"
)

var (
	redis_instance		*redis.Client
	etcd_instance		*util.Etcd

	redis_once			sync.Once
	etcd_once			sync.Once
)

func GetRedisInstance() *redis.Client {
	redis_once.Do(func() {
		cfg := util.GetConfigValue[mservice_define.MServiceConfig]().RedisCfg
		redis_instance = redis.NewClient(&redis.Options {
			Addr:		cfg.Addr,
			Password:	cfg.Password,
			DB:			cfg.DbIndex,
		})
	})
	return redis_instance
}

func GetEtcdInstance() *util.Etcd {
	etcd_once.Do(func() {
		cfg := util.GetConfigValue[mservice_define.MServiceConfig]().EtcdCfg
		etcd_instance = util.NewEtcd(cfg.Addrs, cfg.Username, cfg.Password)
	})
	return etcd_instance
}