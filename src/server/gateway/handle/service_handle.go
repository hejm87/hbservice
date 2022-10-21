package gateway_handle

import (
	"net"
	"github.com/segmentio/ksuid"
	"hbservice/src/net/net_core"
	"hbservice/src/mservice/define"
	"hbservice/src/server/gateway/define"
)

type ServiceHandle struct {}

func (p *ServiceHandle) Init(server net_core.NetServer) error {
	return nil
}

func (p *ServiceHandle) OnAccept(conn net.Conn, handle net_core.PacketHandle, server net_core.NetServer) (string, error) {
	return ksuid.New().String(), nil
}

//func (p *ServiceHandle) OnMessage(channel_id string, packet net_core.Packet, server net_core.NetServer) error {
//	mpacket := packet.(*mservice_define.MServicePacket)
//	if mpacket.Header.Service != gateway_define.SERVICE_NAME {
//		return nil
//	}
//	push_req := mpakcet.Body.(MIPushReq)
//	push_item := gateway_define.MPushItem {
//		Cmd:	push_req.Cmd,
//		Body:	push_req.Body,
//	}
//	resp_packet := gateway_define.CreateRespPacket(
//		req.Header.Uid,
//		req.Header.Cmd,
//		gateway_define.MPushPacket {
//			Count:		1,
//			PushItems:	[]gateway_define.MPushPacket {push_item},
//		},
//	)
//	return resp_packet, nil
//}

func (p *ServiceHandle) OnMessage(channel_id string, packet net_core.Packet, server net_core.NetServer) error {
	mpacket := packet.(*mservice_define.MServicePacket)
	if mpacket.Header.Service != gateway_define.SERVICE_NAME {
		return nil
	}
	push_req := mpacket.Body.(gateway_define.MIPushReq)
	push_packet := p.convert_push_packets(&push_req)
	for _, uid := range push_req.Uids {
		if cid, ok := GetUserConnsInstance().GetChannelId(uid); ok {
			gw_packet := &gateway_define.MGatewayPacket {
				Header:		gateway_define.MGatewayHeader {
					Uid:	uid,
					Cmd:	gateway_define.GW_PUSH,
				},
				Body:		push_packet,
			}	
			server.PushChannel(cid, gw_packet)
		}
	}
	return nil
}

func (p *ServiceHandle) OnClose(channel_id string, server net_core.NetServer) error {
	return nil
}

func (p *ServiceHandle) OnTimer(timer_id string, value interface {}, server net_core.NetServer) {
	return
}

func (p *ServiceHandle) convert_push_packets(req *gateway_define.MIPushReq) *gateway_define.MPushPacket {
	push_item := gateway_define.MPushItem {
		Cmd:	req.Cmd,
		Body:	req.Body,
	}
	push_packet := &gateway_define.MPushPacket {
		Count:		1,
		PushItems:	[]gateway_define.MPushItem {push_item},
	}
	return push_packet
}