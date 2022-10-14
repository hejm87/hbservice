package online_handle

import (
	"github.com/segmentio/ksuid"
	"hbservice/src/util"
)

type UserInfo {
	host				string	
	port				int
	latest_active_ts	int64
}

type OnlineHandle struct {
	user_caches			util.MapQueue[string]
}

func (p *OnlineHandle) Init(server net_core.NetServer) error {
	return nil
}

func (p *OnlineHandle) OnAccept(conn net.Conn, handle net_core.PacketHandle, server net_core.NetServer) (string, error) {
	return ksuid.New().String(), nil
}

func (p *OnlineHandle) OnMessage(channel_id string, packet net_core.Packet, server net_core.NetServer) error {
	return nil
}

func (p *OnlineHandle) OnClose(channel_id string, server net_core.NetServer) error {
	return nil
}

func (p *OnlineHandle) OnTimer(timer_id string, value interface {}, server net_core.NetServer) {
	return
}