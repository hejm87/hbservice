package container 

import (
	"fmt"
	"log"
	"time"
	"sync"
	"errors"
	"context"
	"net/http"
	_ "net/http/pprof"
	"hbservice/src/util"
	"hbservice/src/net/tcp"
	"hbservice/src/net/net_core"
	"hbservice/src/mservice/define"
	"hbservice/src/mservice/obj_client"
)

const (
	NAMING_HEARTBEAT_TS		int = 10
	NAMING_OVERDUE			int = 30
)

var (
	instance		*Container
	once			sync.Once
)

type Container struct {
	lease_id		int64

	obj_proxys		map[string]*ObjectProxy		

	ctx				context.Context
	cancel			context.CancelFunc
	sync.Mutex
	net_core.NetServer
}

func GetInstance() *Container {
	once.Do(func() {
		instance = &Container {
			obj_mgr: obj_client.NewObjClientMgr(&MPacketHandle{})
		}
	})
	return instance
}

func (p *Container) Run(params []net_core.NetServerParam) error {
	if err := p.init(); err != nil {
		return err
	}
	new_params, err := p.set_server_params(params)
	if err != nil {
		return err
	}

	p.NetServer = tcp_server.NetServer(new_params)
	go p.NetServer.Start()

	p.do_keep_alive_util_shutdown()

	return nil
}

func (p *Container) Shutdown() {
	p.NetServer.Shutdown()
	p.cancel()
}

func (p *Container) Call(hash uint32, req *mserivce_define.MServicePacket) (*mservice_define.MServicePacket, error) {
	proxy, err := p.get_proxy(req.Header.Service)
	if err != nil {
		return nil, err
	}
	return proxy.Call(hash, req)
}

func (p *Container) CallByAddr(addr string, req *mserivce_define.MServicePacket) (*mservice_define.MServicePacket, error) {
	proxy, err := p.get_proxy(req.Header.Service)
	if err != nil {
		return nil, err
	}
	return proxy.Call(addr, req)
}

func (p *Container) CallAsync(hash uint32, req *mserivce_define.MServicePacket) error {
	proxy, err := p.get_proxy(req.Header.Service)
	if err != nil {
		return nil, err
	}
	return proxy.CallAsync(hash, req)
}

func (p *Container) CallAsyncByAddr(addr string, req *mserivce_define.MServicePacket) error {
	proxy, err := p.get_proxy(req.Header.Service)
	if err != nil {
		return nil, err
	}
	return proxy.CallAsyncByAddr(hash, req)
}

func (p *Container) Cast(hash uint32, req *mserivce_define.MServicePacket) error {
	proxy, err := p.get_proxy(req.Header.Service)
	if err != nil {
		return nil, err
	}
	return proxy.Cast(hash, req)
}

func (p *Container) CastByAddr(addr string, req *mserivce_define.MServicePacket) error {
	proxy, err := p.get_proxy(req.Header.Service)
	if err != nil {
		return nil, err
	}
	return proxy.CastByAddr(addr, req)
}

func (p *Container) init() error {
	var err error
	if err = util.SetConfigByFileLoader[mservice_define.MServiceConfig]("./mservice.cfg"); err != nil {
		return err
	}
	p.cfg = util.GetConfigValue[mservice_define.MServiceConfig]()
	if p.cfg.CommonCfg.OpenPerformanceMonitor == true {
		go func() {
			http.ListenAndServe("0.0.0.0:8000", nil)
		} ()
	}

	p.ctx, p.cancel = context.WithCancel(context.Background())

	p.lease_id, err = mservice_util.GetEtcdInstance().PutWithTimeout(p.get_service_tag(), "ok", int64(NAMING_OVERDUE))
	if err != nil {
		return err
	}
	return nil
}

func (p *Container) get_obj_proxy(name string) (*ObjectProxy, error) {
	p.Lock()
	defer p.Unlock()
	proxy, ok := p.obj_proxys[name]
	if ok {
		return proxy, nil
	}
	proxy = NewObjectProxy(name, &MPacketHandle {})
	if err := proxy.Start(); err != nil {
		return nil, err
	}
	p.obj_proxys[name] = proxy
	return proxy, nil
}

func (p *Container) do_keep_alive_util_shutdown() {
	for {
		select {
		case <-p.ctx.Done():
			break
		case <-time.After(time.Duration(NAMING_HEARTBEAT_TS) * time.Second):
			if err := mservice_util.GetEtcdInstance().KeepAlive(p.lease_id); err != nil {
				log.Printf("ERROR|container|KeepAlive error:%#v", err)
			}
		}
	}
}

func (p *Container) set_server_params(params []net_core.NetServerParam) ([]net_core.NetServerParam, error) {
	var new_params []net_core.NetServerParam
	for _, x := range params {
		if x.Name == "" && x.Host == "" && x.Port == 0 {
			// 默认服务重置服务信息
			cfg := util.GetConfigValue[mservice_define.MServiceConfig]().Service
			param := net_core.NetServerParam {
				Name:	cfg.Name,
				Host:	cfg.ListenHost,
				Port:	cfg.ListenPort,
				LogicHandle:	x.LogicHandle,
				PacketHandle:	x.PacketHandle,
			}
			new_params = append(new_params, param)
		} else if x.Name != "" && x.Host != "" && x.Port > 0 {
			new_params = append(new_params, x)
			continue
		} else {
			return new_params, errors.New("server_params exception")
		}
	}
	return new_params, nil
}

func (p *Container) get_service_tag() string {
	ips, err := util.GetLocalIp()
	if err != nil || len(ips) == 0 {
		log.Fatalf("FATAL|container|can`t get local ip")
	}
	return fmt.Sprintf("%s/%s/%s_%s:%d_%s",
		obj_client.NameServiceDir, 
		p.cfg.Service.Name,
		p.cfg.Service.Name,
		ips[0],
		p.cfg.Service.ListenPort,
		util.GenUuid(),
	)
}
