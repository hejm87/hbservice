package container 

import (
	"fmt"
	"log"
	"time"
	"sync"
	"context"
	"hbservice/src/util"
	"hbservice/src/mservice/define"
	"hbservice/src/mservice/obj_client"
	"hbservice/src/net/net_core"
)

const (
	NAMING_HEARTBEAT_TS		int = 10
	NAMING_OVERDUE			int = 30
)

var (
	instance		*Container
	once			sync.Once
)

func GetInstance() *Container {
	once.Do(func() {
		cfg := util.GetConfigValue[mservice_define.MServiceConfig]()
		instance = &Container {
			name:	cfg.Service.Name,
			etcd:	util.NewEtcd(cfg.EtcdCfg.Addrs, cfg.EtcdCfg.Username, cfg.EtcdCfg.Password),
		}
	})
	return instance
}

type Container struct {
	name			string
	etcd			*util.Etcd
	ctx				context.Context
	cfg				mservice_define.MServiceConfig
	net_core.NetServer
}

func (p *Container) Start(server net_core.NetServer) error {
	p.NetServer = server
	p.cfg = *util.GetConfigValue[mservice_define.MServiceConfig]()

	if err := p.register(); err != nil {
		log.Printf("ERROR|container|register error:%#v", err)
		return err
	}
	p.NetServer.Start()
	p.etcd.Close()
	return nil
}

func (p *Container) Shutdown() {
	p.NetServer.Shutdown()
}

func (p *Container) register() error {
	lease_id, err := p.etcd.PutWithTimeout(
		p.get_service_tag(),
		"ok",
		int64(NAMING_OVERDUE),
	)
	if err != nil {
		return err
	}
	go func() {
		for {
			time.Sleep(time.Duration(NAMING_HEARTBEAT_TS) * time.Second)
			if err := p.etcd.KeepAlive(lease_id); err != nil {
				log.Printf("INFO|container|register KeepAlive error:%#v", err)
			}
		}
	} ()
	return nil
}

func (p *Container) get_service_tag() string {
	ips, err := util.GetLocalIp()
	if err != nil || len(ips) == 0 {
		log.Fatalf("FATAL|container|can`t get local ip")
	}
	return fmt.Sprintf("%s/%s/%s_%s:%d_%s",
		obj_client.NameServiceDir, 
		p.name,
		p.name,
		ips[0],
		p.cfg.Service.ListenPort,
		util.GenUuid(),
	)
}
