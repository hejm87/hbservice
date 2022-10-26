package main

import (
	"fmt"
	"log"
	"errors"
	"hbservice/src/util"
	"hbservice/src/mservice"
	"hbservice/src/mservice/util"
	"hbservice/src/mservice/define"
	"hbservice/src/net/net_core"
	"hbservice/src/server/gateway/define"
	"hbservice/src/server/gateway/handle"
)

func main() {
	if err := init_config(); err != nil {
		log.Fatalf("container.init_config error:%#v", err)
	}
	params := get_server_params()
	if err := container.Run(params); err != nil {
		log.Fatalf("container.Start error:%#v", err)
	}
}

func init_config() error {
	mode := container.GetLauncherParam().ConfigMode
	if mode == "file" {
		return init_config_by_file()
	} else if mode == "etcd" {
		return init_config_by_etcd()
	}
	return errors.New(fmt.Sprintf("unknow config_mode:%s", mode))
}

func init_config_by_file() error {
	path := mservice_util.GetConfigPath()
	return util.SetConfigByFileLoader[gateway_define.GatewayConfig](path)
}

func init_config_by_etcd() error {
	etcd := mservice_util.GetEtcdInstance()
	path := fmt.Sprintf(
		"%s/%s/%s.cfg",
		mservice_define.SERVICE_ETCD_CONFIG_PATH,
		util.GetExecName(),
		util.GetExecName(),
	)
	return util.SetConfigByEtcdLoader[gateway_define.GatewayConfig](etcd, path)
}

func get_server_params() []net_core.NetServerParam {
	var params []net_core.NetServerParam

	cfg := util.GetConfigValue[gateway_define.GatewayConfig]().Listen
	params = append(
		params,
		net_core.NetServerParam {
			LogicHandle:	&gateway_handle.GatewayHandle {},
			PacketHandle:	&gateway_define.GWPacketHandle {},
			Name:			"gateway_outter",
			Iface:			cfg.Iface,
			Port:			cfg.Port,
		},
	)
	params = append(
		params,
		net_core.NetServerParam {
			LogicHandle:	&gateway_handle.ServiceHandle {},
			PacketHandle:	&mservice_define.MPacketHandle {},
		},
	)
	return params
}
