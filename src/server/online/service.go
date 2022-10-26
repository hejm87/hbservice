package main

import (
	"fmt"
	"log"
	"errors"
	"hbservice/src/util"
	"hbservice/src/net/net_core"
	"hbservice/src/mservice"
	"hbservice/src/mservice/util"
	"hbservice/src/mservice/define"
	"hbservice/src/server/online/define"
	"hbservice/src/server/online/handle"
)

func main() {
	if err := init_config(); err != nil {
		log.Fatalf("Container.init_config error:%#v", err)
	}
	params := get_server_params()
	if err := container.Run(params); err != nil {
		log.Fatalf("container.Run error:%#v", err)
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
	return util.SetConfigByFileLoader[online_define.OnlineConfig](path)
}

func init_config_by_etcd() error {
	etcd := mservice_util.GetEtcdInstance()
	path := fmt.Sprintf(
		"%s/%s/%s.cfg",
		mservice_define.SERVICE_ETCD_CONFIG_PATH,
		util.GetExecName(),
		util.GetExecName(),
	)
	return util.SetConfigByEtcdLoader[online_define.OnlineConfig](etcd, path)
}

func get_server_params() []net_core.NetServerParam {
	var params []net_core.NetServerParam
	params = append(
		params, 
		net_core.NetServerParam {
			LogicHandle:	&online_handle.OnlineHandle {},
			PacketHandle:	&mservice_define.MPacketHandle {},
		},
	)
	return params
}
