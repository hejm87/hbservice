package main

import (
	"log"
	"hbservice/src/mservice"
	"hbservice/src/mservice/define"
	"hbservice/src/net/net_core"
	"hbservice/src/server/gateway/define"
	"hbservice/src/server/gateway/handle"
)

func main() {
	params := get_server_params()
	if err := container.GetInstance().Run(params); err != nil {
		log.Fatalf("container.Start error:%#v", err)
	}
}

func get_server_params() []net_core.NetServerParam {
	var params []net_core.NetServerParam
	params = append(
		params, 
		net_core.NetServerParam {
			LogicHandle:	&gateway_handle.GatewayHandle {},
			PacketHandle:	&gateway_define.GWPacketHandle {},
		},
	)
	params = append(
		params,
		net_core.NetServerParam {
			LogicHandle:	&gateway_handle.ServiceHandle {},
			PacketHandle:	&mservice_define.MIPushPacketHandle {},
		},
	)
	return params
}