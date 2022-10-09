package gateway_handle

import (
	"log"
	"net"
	"github.com/segmentio/ksuid"
	"hbservice/src/util"
	"hbservice/src/net/net_core"
	"hbservice/src/mservice/define"
	"hbservice/src/server/gateway/define"
)

type GatewayHandle struct {}

func (p *GatewayHandle) Init(server net_core.NetServer) error {
	cfg := util.GetConfigValue[gateway_define.GatewayConfig]().Common
	param := net_core.NetTimerParam {
		Id:		"eliminate_user",
		Type:	util.TIMER_PERIOD,
		Delay:	0,
		Period:	cfg.UserHeartbeatSec * 3,
	}
	server.SetTimer(param, p)
	return nil
}

func (p *GatewayHandle) OnAccept(conn net.Conn, handle net_core.PacketHandle, server net_core.NetServer) (string, error) {
	return ksuid.New().String(), nil
}

func (p *GatewayHandle) OnMessage(channel_id string, packet net_core.Packet, server net_core.NetServer) error {
	gw_packet := packet.(*gateway_define.MGatewayPacket)
	var err error
	var resp_packet net_core.Packet
	if gw_packet.Header.Cmd == gateway_define.GW_LOGIN {
		resp_packet, err = p.do_login(gw_packet.Header.Uid)
	} else if gw_packet.Header.Cmd == gateway_define.GW_LOGOUT {
		resp_packet, err = p.do_logout(gw_packet.Header.Uid)
	} else if gw_packet.Header.Cmd == gateway_define.GW_HEARTBEAT {
		resp_packet, err = p.do_heartbeat(gw_packet.Header.Uid)
	} else if gw_packet.Header.Cmd == gateway_define.GW_CMD {
		mpacket := gw_packet.Body.(*mservice_define.MServicePacket)	
		resp_packet, err = p.do_cmd(mpacket)
	}
	if resp_packet != nil {
		server.PushChannel(channel_id, resp_packet)
	}
	return err
}

func (p *GatewayHandle) OnClose(channel_id string, server net_core.NetServer) error {
	return nil
}

func (p *GatewayHandle) OnTimer(timer_id string, value interface {}, server net_core.NetServer) {
	if timer_id == "eliminate_user" {
		p.eliminate_user(server)
	}
}

func (p *GatewayHandle) do_login(uid string) (net_core.Packet, error) {
	return nil, nil
}

func (p *GatewayHandle) do_logout(uid string) (net_core.Packet, error) {
	return nil, nil
}

func (p *GatewayHandle) do_heartbeat(uid string) (net_core.Packet, error) {
	if GetGWUsersInstance().UpdateUser(uid) == true {
		// 填写逻辑代码
		return nil, nil
	}
	return nil, nil
}

func (p *GatewayHandle) do_cmd(packet *mservice_define.MServicePacket) (net_core.Packet, error) {
	return nil, nil
}

func (p *GatewayHandle) eliminate_user(server net_core.NetServer) {
	users := GetGWUsersInstance().GetOverdueUsers()
	for _, x := range users {
		log.Printf("INFO|GatewayHandle|eliminate_user, user:%s, channel_id:%s", x.Uid, x.ChannelId)
		server.CloseChannel(x.ChannelId)
	}
	return
}
