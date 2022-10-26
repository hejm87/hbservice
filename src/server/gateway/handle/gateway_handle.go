package gateway_handle

import (
	"log"
	"net"
	"time"
	"strings"
	"net/http"
	"encoding/json"
	"github.com/segmentio/ksuid"
	"hbservice/src/util"
	"hbservice/src/net/net_core"
	"hbservice/src/mservice"
	"hbservice/src/mservice/define"
	"hbservice/src/server/online/define"
	"hbservice/src/server/gateway/define"
	"hbservice/src/server/httpdns/define"
)

type GatewayHandle struct {
	host			string
	port			int
}

func (p *GatewayHandle) Init(server net_core.NetServer) error {
	cfg := util.GetConfigValue[gateway_define.GatewayConfig]()
	param := net_core.NetTimerParam {
		Id:		"eliminate_user",
		Type:	util.TIMER_PERIOD,
		Delay:	0,
		Period:	cfg.Common.UserHeartbeatSec * 3,
	}

	ip, err := util.GetIfaceIpAddr(cfg.Listen.Iface)
	if err != nil {
		return err
	}
	p.host = ip
	p.port = cfg.Listen.Port

	server.SetTimer(param, p)
	go p.report_dns()
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
		resp_packet, err = p.do_login(channel_id, gw_packet)
	} else if gw_packet.Header.Cmd == gateway_define.GW_LOGOUT {
		resp_packet, err = p.do_logout(gw_packet)
	} else if gw_packet.Header.Cmd == gateway_define.GW_HEARTBEAT {
		resp_packet, err = p.do_heartbeat(gw_packet)
	} else if gw_packet.Header.Cmd == gateway_define.GW_CMD {
		resp_packet, err = p.do_cmd(gw_packet)
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
func (p *GatewayHandle) do_login(channel_id string, req *gateway_define.MGatewayPacket) (*gateway_define.MGatewayPacket, error) {
	login_resp := &gateway_define.GWLoginResp {}
	err := GetUserConnsInstance().AddUserConn(req.Header.Uid, channel_id)
	if err == nil {
		login_req := online_define.OLLoginReq {
			User:	online_define.UserNode {
				Uid:		req.Header.Uid,
				ServiceId:	container.GetServiceId(),	
				Host:		p.host,
				Port:		p.port,
			},
		}
		container.Cast[online_define.OLLoginReq](
			util.GenHash(req.Header.Uid),
			online_define.SERVICE_NAME,
			online_define.MS_ONLINE_LOGIN,
			login_req,
		)
	} else {
		login_resp.Err = err.Error()
	}
	resp_packet := gateway_define.CreateRespPacket(
		req.Header.Uid,
		req.Header.Cmd,
		login_resp,
	)
	return resp_packet, nil
}

func (p *GatewayHandle) do_logout(req *gateway_define.MGatewayPacket) (*gateway_define.MGatewayPacket, error) {
	login_resp := &gateway_define.GWLogoutResp {}
	ok := GetUserConnsInstance().RemoveUserConn(req.Header.Uid)
	if ok {
		login_req := online_define.OLLogoutReq {Uid: req.Header.Uid}	
		container.Cast[online_define.OLLogoutReq](
			util.GenHash(req.Header.Uid),
			online_define.SERVICE_NAME,
			online_define.MS_ONLINE_LOGOUT,
			login_req,
		)
	} else {
		login_resp.Err = gateway_define.ERR_USER_CONN_NOT_EXISTS
	}
	resp_packet := gateway_define.CreateRespPacket(
		req.Header.Uid,
		req.Header.Cmd,
		login_resp,
	)
	return resp_packet, nil
}

func (p *GatewayHandle) do_heartbeat(req *gateway_define.MGatewayPacket) (*gateway_define.MGatewayPacket, error) {
	GetUserConnsInstance().RefreshUserConn(req.Header.Uid)
	ol_req := online_define.OLHeartbeatReq {
		User:	online_define.UserNode {
			Uid:		req.Header.Uid,
			ServiceId:	container.GetServiceId(),
		},
	}
	container.Cast[online_define.OLHeartbeatReq](
		util.GenHash(req.Header.Uid),
		online_define.SERVICE_NAME,
		online_define.MS_ONLINE_HEARTBEAT,
		ol_req,
	)
	resp_packet := gateway_define.CreateRespPacket(
		req.Header.Uid,
		req.Header.Cmd,
		gateway_define.GWHeartbeatResp {Err: ""},
	)
	return resp_packet, nil
}

func (p *GatewayHandle) do_cmd(req *gateway_define.MGatewayPacket) (*gateway_define.MGatewayPacket, error) {
	req_packet := req.Body.(*mservice_define.MServicePacket)
	resp_packet, err := container.CallByPacket(util.GenHash(req.Header.Uid), req_packet)
	if err != nil {
		return nil, err
	}
	gw_packet := gateway_define.CreateRespPacket(
		req.Header.Uid,
		req.Header.Cmd,
		resp_packet,
	)
	return gw_packet, nil
}

func (p *GatewayHandle) eliminate_user(server net_core.NetServer) {
	users := GetUserConnsInstance().EliminateTimeoutUserConns()
	for _, user := range users {
		container.Cast[online_define.OLLogoutReq](
			util.GenHash(user.Uid),
			online_define.SERVICE_NAME,
			online_define.MS_ONLINE_LOGOUT,
			online_define.OLLogoutReq {Uid: user.Uid},
		)
		server.CloseChannel(user.ChannelId)
	}
}

func (p *GatewayHandle) report_dns() {
	info := util.GetConfigValue[gateway_define.GatewayConfig]().Report
	timer := time.NewTicker(time.Duration(info.Period) * time.Second)
	for {
		select {
		case <-timer.C:
			cfg := util.GetConfigValue[gateway_define.GatewayConfig]()
			ip, err := util.GetIfaceIpAddr(cfg.Listen.Iface)
			if err != nil {
				log.Printf("ERROR|GetIfaceIpAddr error:%#v", err)
				continue
			}
			report := httpdns_define.NodeReport {
				Info:	httpdns_define.NodeInfo {
					Name:		"gateway",
					Addr:		httpdns_define.NodeAddr {
						Host:	ip,
						Port:	cfg.Listen.Port,
					},
				},
				Status:	httpdns_define.NodeStatus {
					ConnCount:	GetUserConnsInstance().GetCount(),
				},
			}
			content,err := json.Marshal(report)
			if err != nil {
				log.Printf("ERROR|json.Marshal struct:%#v error:%#v", report, err)
				continue
			}
			if _, err := http.Post(info.Url, "application/json", strings.NewReader(string(content))); err != nil {
				log.Printf("ERROR|report_dns error:%#v", err)
			} else {
				log.Printf("INFO|report_dns succeed")
			}
		}
	}
}
