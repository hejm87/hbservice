package gateway_handle

import (
	"log"
	"net"
	"github.com/segmentio/ksuid"
	"hbservice/src/util"
	"hbservice/src/net/net_core"
	"hbservice/src/mservice"
	"hbservice/src/mservice/define"
	"hbservice/src/server/gateway/define"
)

const (
	GATEWAY_CONF
)

type GatewayHandle struct {}

func (p *GatewayHandle) Init(server net_core.NetServer) error {
	cfg := util.GetConfigValue[gateway_define.GatewayConfig]()
	param := net_core.NetTimerParam {
		Id:		"eliminate_user",
		Type:	util.TIMER_PERIOD,
		Delay:	0,
		Period:	cfg.Common.UserHeartbeatSec * 3,
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
		resp_packet, err = p.do_cmd(gw_packet.Header.Uid, mpacket)
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

// 对于登录/登出目前没有具体操作，返回成功即可
func (p *GatewayHandle) do_login(req *MGatewayPacket) (*MGatewayPacket, error) {
	resp := &gateway_define.GWLoginResp {}
	err := GetUserConnsInstance().AddUserConn(uid, channel_id)
	if err != nil {
		resp.Err = err.Error()
	}
	resp_packet := gateway_define.CreateRespPacket(
		req.Header.Uid,
		req.Header.Cmd,
		resp,
	)
	return resp_packet, nil
}

func (p *GatewayHandle) do_logout(req *MGatewayPacket) (*MGatewayPacket, error) {
	resp := &gateway_define.GWLogoutResp {}
	ok := GetUserConnsInstance().RemoveUserConn()
	if !ok {
		resp.Err = "user conn not exists"
	}
	resp_packet := gateway_define.CreateRespPacket(
		req.Header.Uid,
		req.Header.Cmd,
		resp,
	)
	return resp_packet, nil
}

func (p *GatewayHandle) do_heartbeat(req *MGatewayPacket) (*MGatewayPacket, error) {
	hash := util.GenHash(req.Header.Uid)
	req := online_define.OLHeartbeatReq {
		User:	UserNode {
			Uid:		req.Header.Uid,
			ServiceId:	container.GetServiceId(),
		},
	}
	resp, err := container.Cast[mservice_define.OLHeartbeatReq](hash, req)
	if err != nil {
		return nil, err
	}
	resp_packet := gateway_define.CreateRespPacket(
		req.Header.Uid,
		req.Header.Cmd,
		gateway_define.GWHeartbeatResp {Err: ""},
	)
	return resp_packet, nil
}

func (p *GatewayHandle) do_cmd(req *MGatewayPacket) (*MGatewayPacket, error) {
	req_packet := req.Body.(*MServicePacket)
	return container.CallByPacket(util.GenHash(req.Header.Uid), req_packet)
}

func (p *GatewayHandle) eliminate_user(server net_core.NetServer) {
	users := GetUserConnsInstance().GetTimeoutUserConns()
	for _, x := range users {
		service.CloseChannel(x.ChannelId)
	}
}
