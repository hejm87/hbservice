package container 

import (
	"os"
	"fmt"
	"log"
	"flag"
	"sync"
	"errors"
	"context"
	"net/http"
	_ "net/http/pprof"
	"hbservice/src/util"
	"hbservice/src/net/tcp"
	"hbservice/src/net/net_core"
	"hbservice/src/mservice/util"
	"hbservice/src/mservice/define"
	"hbservice/src/mservice/naming"
	"hbservice/src/mservice/obj_client"
)

type LauncherParam struct {
	ConfigMode			string
	LogMode				string
}

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
		instance = &Container {
			service_id:	fmt.Sprintf("%s_%s", util.GetExecName(), util.GenUuid()),
		}
		if err := instance.init_container(); err == nil {
			log.Printf("INFO|Container|init success")
		} else {
			panic(fmt.Sprintf("container init error:%#v", err))
		}
	})
	return instance
}

func GetLauncherParam() *LauncherParam {
	return container_instance().param
}

func Run(params []net_core.NetServerParam) error {
	c := container_instance()
	new_params, err := c.set_server_params(params)
	if err != nil {
		return err
	}
	c.NetServer = tcp_server.NetServer(new_params)
	go c.NetServer.Start()
	select {
	case <-c.ctx.Done():
		c.NetServer.Shutdown()
	}
	c.close_container()
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
	service_id		string
	param			*LauncherParam
	ctx				context.Context
	cancel			context.CancelFunc
	LogFilePtr		*os.File
	net_core.NetServer
	sync.Mutex
}

func (p *Container) init_container() error {
	p.init_param()
	if err := p.init_log(); err != nil {
		return err
	}
	if err := p.init_config(); err != nil {
		return err
	}
	cfg := util.GetConfigValue[mservice_define.MServiceConfig]().Service
	if cfg.OpenPerformanceMonitor == true {
		go func() {
			addr := fmt.Sprintf("0.0.0.0:%d", cfg.PerformanceMonitorPort)
			http.ListenAndServe(addr, nil)
		} ()
	}
	p.ctx, p.cancel = context.WithCancel(context.TODO())
	service_info, err := p.get_service_info()
	if err != nil {
		return err
	}
	return naming.GetInstance().Register(service_info)
}

func (p *Container) close_container() {
	if p.LogFilePtr != nil {
		p.LogFilePtr.Close()
	}
}

func (p *Container) init_param() {
	config_mode := flag.String("config_mode", "file", "config_mode")
	log_mode := flag.String("log_mode", "file", "log_mode")
	flag.Parse()
	p.param = &LauncherParam {
		ConfigMode: *config_mode,
		LogMode:	*log_mode,
	}
}

func (p *Container) init_log() error {
	if p.param.LogMode == "file" {
		log_name := fmt.Sprintf("%s.log", util.GetExecName())
		ptr, err := os.OpenFile(log_name, os.O_CREATE | os.O_APPEND | os.O_WRONLY, 0666)
		if err != nil {
			return err
		}
		p.LogFilePtr = ptr
		log.SetOutput(ptr)
	}
	return nil
}

func (p *Container) init_config() error {
	if p.param.ConfigMode == "file" {
		return p.init_config_by_file()
	} else if p.param.ConfigMode == "etcd" {
		return p.init_config_by_etcd()
	}
	return errors.New(fmt.Sprintf("unknow config_mode:%s", p.param.ConfigMode))
}

func (p *Container) init_config_by_file() error {
	path := mservice_util.GetServiceConfigPath()
	return util.SetConfigByFileLoader[mservice_define.MServiceConfig](path)
}

func (p *Container) init_config_by_etcd() error {
	etcd := mservice_util.GetEtcdInstance()
	path := fmt.Sprintf(
		"%s/%s/mservice.cfg",
		mservice_define.SERVICE_ETCD_CONFIG_PATH,
		util.GetExecName(),
	)
	return util.SetConfigByEtcdLoader[mservice_define.MServiceConfig](etcd, path)
}

func (p *Container) get_service_info() (info naming.ServiceInfo, err error) {
	cfg := util.GetConfigValue[mservice_define.MServiceConfig]()
	host, err := util.GetIfaceIpAddr(cfg.Service.ListenIface)
	if err != nil {
		return info, err
	}
	info = naming.ServiceInfo {
		Name:		util.GetExecName(),
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
		name := x.Name
		host := x.Host
		port := x.Port
		if name == "" {
			name = util.GetExecName()
		}
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
			Name:			util.GetExecName(),
			Host:			host,
			Port:			port,
			LogicHandle:	x.LogicHandle,
			PacketHandle:	x.PacketHandle,
		}
		new_params = append(new_params, new_param)
	}
	return new_params, nil
}
