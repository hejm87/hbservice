package gateway_define

import (
	"net"
	"encoding/json"
	"hbservice/src/net/net_core"
	"hbservice/src/mservice/define"
)

const (
	GW_LOGIN		string = "LOGIN"
	GW_LOGOUT		string = "LOGOUT"
	GW_HEARTBEAT	string = "HEARTBEAT"
	GW_CMD			string = "CMD"

	GW_REQUEST		int = 0
	GW_RESPONSE		int = 1
)

type MGatewayHeader struct {
	Uid				string
	Cmd				string
	Direction		int		// 方向，取值：GW_REQUEST, GW_RESPONSE
}

type MGatewayPacket struct {
	Header			MGatewayHeader
	Body			interface {}
}

func (p *MGatewayPacket) ToStr() string {
	return ""
}

//////////////////////////////////////////////////////
//					具体网关协议
//////////////////////////////////////////////////////
type GWLoginResponse struct {
	Result			int		// 登录结果，0：成功，其他：失败
	ErrMsg			string
}

type GWLogoutResponse struct {
	Result			int		// 登出结果，0：成功，其他：失败
	ErrMsg			string
}

//////////////////////////////////////////////////////
//					网关PacketHandle
//////////////////////////////////////////////////////
type GWPacketHandle struct {}

func (p *GWPacketHandle) RecvPacket(conn net.Conn, timeout_ms int) (net_core.Packet, error) {
	var packet MGatewayPacket
	if bytes, err := mservice_define.CommonRecvPacket(conn, timeout_ms); err == nil {
		if err := json.Unmarshal(bytes, &packet); err != nil {
			return nil, err
		}
	}
	return &packet, nil
}

func (p *GWPacketHandle) SendPacket(conn net.Conn, packet net_core.Packet) error {
	return mservice_define.CommonSendPacket(conn, packet)
}