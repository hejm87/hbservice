package container 

import (
	"fmt"
	"log"
	"sync"
	"context"
	"net/http"
	_ "net/http/pprof"
	"hbservice/src/util"
	"hbservice/src/net/tcp"
	"hbservice/src/net/net_core"
	"hbservice/src/mservice/define"
	"hbservice/src/mservice/naming"
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

func container_instance() *Container {
	once.Do(func() {
		if err := util.SetConfigByFileLoader[mservice_define.MServiceConfig]("./mservice.cfg"); err != nil {
			panic(fmt.Sprintf("loader mservice.cfg error:%#v", err))
		}
		cfg := util.GetConfigValue[mservice_define.MServiceConfig]().Service
		instance = &Container {
			name: 		cfg.Name,
			service_id:	fmt.Sprintf("%s_%s", cfg.Name, util.GenUuid()),
		}
		if err := instance.init_container(); err == nil {
			log.Printf("INFO|container|init success")
		} else {
			panic(fmt.Sprintf("container:%s init error:%#v", cfg.Name, err))
		}
	})
	return instance
}

func Run(params []net_core.NetServerParam) error {
	log.Printf("INFO|container|ready to run, net_params:%#v", params)
	c := container_instance()
	new_params, err := c.set_server_params(params)
	if err != nil {
		return err
	}
	log.Printf("INFO|container|ready start NetServer, net_params:%#v", new_params)
	c.NetServer = tcp_server.NetServer(new_params)
	go c.NetServer.Start()
	select {
	case <-c.ctx.Done():
		c.NetServer.Shutdown()
	}
	return nil
}

func Shutdown() {
	container_instance().cancel()
}

func GetServiceId() string {
	return container_instance().service_id
}

func Call[REQ any, RESP any](hash uint32, service string, method string, req REQ) (resp RESP, err error) {
	req_packet := mservice_define.CreateReqPacket(
		service, 
		method,
		mservice_define.MS_CALL,
		req,
	)
	resp_packet, err := obj_client.GetInstance().Call(hash, req_packet)
	if err != nil {
		return resp, err
	}
	return resp_packet.Body.(RESP), nil	
}

func CallAsync[REQ any, RESP any](hash uint32, service string, method string, req REQ, cb obj_client.RespPacketCb) {
	req_packet := mservice_define.CreateReqPacket(
		service, 
		method,
		mservice_define.MS_CALL,
		req,
	)
	obj_client.GetInstance().CallAsync(hash, req_packet, cb)
}

func Cast[REQ any](hash uint32, service string, method string, req REQ) {
	req_packet := mservice_define.CreateReqPacket(
		service, 
		method,
		mservice_define.MS_CAST,
		req,
	)
	obj_client.GetInstance().Cast(hash, req_packet)
}

func CallByPacket(hash uint32, req *mservice_define.MServicePacket) (*mservice_define.MServicePacket, error) {
	return obj_client.GetInstance().Call(hash, req)
}

func CallAsyncByPacket(hash uint32, req *mservice_define.MServicePacket, cb obj_client.RespPacketCb) {
	obj_client.GetInstance().CallAsync(hash, req, cb)
}

func CastByPacket(hash uint32, req *mservice_define.MServicePacket) {
	obj_client.GetInstance().Cast(hash, req)
}


/////////////////////////////////////////////////////////////////////
//						container struct
/////////////////////////////////////////////////////////////////////
type Container struct {
	name			string
	service_id		string
	ctx				context.Context
	cancel			context.CancelFunc
	net_core.NetServer
	sync.Mutex
}

func (p *Container) init_container() error {
	cfg := util.GetConfigValue[mservice_define.MServiceConfig]()
	if cfg.Service.OpenPerformanceMonitor == true {
		go func() {
			http.ListenAndServe("0.0.0.0:8000", nil)
		} ()
	}
	p.ctx, p.cancel = context.WithCancel(context.Background())
	service_info, err := p.get_service_info()
	if err != nil {
		return err
	}
	return naming.GetInstance().Register(service_info)
}

func (p *Container) get_service_info() (info naming.ServiceInfo, err error) {
	cfg := util.GetConfigValue[mservice_define.MServiceConfig]()
	host, err := util.GetIfaceIpAddr(cfg.Service.ListenIface)
	if err != nil {
		return info, err
	}
	info = naming.ServiceInfo {
		Name:		p.name,
		ServiceId:	p.service_id,
		Host:		host,
		Port:		cfg.Service.ListenPort,
	}
	return info, nil
}

func (p *Container) set_server_params(params []net_core.NetServerParam) (result []net_core.NetServerParam, err error) {
	cfg := util.GetConfigValue[mservice_define.MServiceConfig]().Service
	var new_params []net_core.NetServerParam
	for _, x := range params {
		var err  error
		host := x.Host
		port := x.Port
		if host == "" {
			iface := x.Iface
			if iface == "" {
				iface = cfg.ListenIface
			}
			if host, err = util.GetIfaceIpAddr(iface); err != nil {
				continue
			}
		}
		if port == 0 {
			port = cfg.ListenPort
		}
		new_param := net_core.NetServerParam {
			Host:			host,
			Port:			port,
			LogicHandle:	x.LogicHandle,
			PacketHandle:	x.PacketHandle,
		}
		new_params = append(new_params, new_param)
	}
	return new_params, nil
}
