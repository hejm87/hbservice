package gateway_handle

import (
	"net"
	"github.com/segmentio/ksuid"
	"hbservice/src/net/net_core"
	"hbservice/src/mservice/define"
)

type ServiceHandle struct {}

func (p *ServiceHandle) Init(server net_core.NetServer) error {
	return nil
}

func (p *ServiceHandle) OnAccept(conn net.Conn, handle net_core.PacketHandle, server net_core.NetServer) (string, error) {
	return ksuid.New().String(), nil
}

func (p *ServiceHandle) OnMessage(channel_id string, packet net_core.Packet, server net_core.NetServer) error {
	mpacket := packet.(*mservice_define.MServicePacket)
	if mpacket.Header.Service != "Pusher" {
		return nil
	}
	if mpacket.Header.Uid == "" {
		return nil
	}
	if gw_channel_id, ok := GetGWUsersInstance().GetUserChannelId(mpacket.Header.Uid); ok {
		server.PushChannel(gw_channel_id, packet)
	}
	return nil
}

func (p *ServiceHandle) OnClose(channel_id string, server net_core.NetServer) error {
	return nil
}

func (p *ServiceHandle) OnTimer(timer_id string, value interface {}, server net_core.NetServer) {
	return
}