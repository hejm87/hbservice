package main

import (
	"log"
	"hbservice/src/mservice"
	"hbservice/src/mservice/define"
	"hbservice/src/net/net_core"
	"hbservice/src/server/pusher/handle"
)

func main() {
	params := get_server_params()
	if err := container.Run(params); err != nil {
		log.Fatalf("container.Start error:%#v", err)
	}
}

func get_server_params() []net_core.NetServerParam {
	var params []net_core.NetServerParam
	params = append(
		params, 
		net_core.NetServerParam {
			LogicHandle:	&pusher_handle.PusherHandle {},
			PacketHandle:	&mservice_define.MPacketHandle {},
		},
	)
	return params
}
