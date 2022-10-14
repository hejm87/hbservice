package gateway_define

import (
	"net"
	"encoding/json"
	"hbservice/src/net/net_core"
	"hbservice/src/mservice/define"
)

const (
	GW_LOGIN		string = "GW_LOGIN"
	GW_LOGOUT		string = "GW_LOGOUT"
	GW_HEARTBEAT	string = "GW_HEARTBEAT"
	GW_CMD			string = "GW_CMD"
	GW_PUSH			string = "GW_PUSH"

	GW_DIRECT_REQ	int = 0
	GW_DIRECT_RESP	int = 1
)

//////////////////////////////////////////////////////
//					对外网关协议
//////////////////////////////////////////////////////
type MGatewayHeader struct {
	Uid				string
	Cmd				string
	Direct			int		// 当Type为GW_REQUEST时有效，取值GW_DIRECT_REQ/GW_DIRECT_RESP
}

type MGatewayPacket struct {
	Header			MGatewayHeader
	Body			interface {}
}

func (p *MGatewayPacket) ToStr() string {
	return ""
}

//////////////////////////////////////////////////////
//					对外网关协议内容
//////////////////////////////////////////////////////
type GWLoginReq struct {
	Uid				string
}

type GWLoginResp struct {
	Result			int		// 登录结果，0：成功，其他：失败
	ErrMsg			string
}

type GWLogoutReq struct {
	Uid				string
}

type GWLogoutResp struct {
	Result			int		// 登出结果，0：成功，其他：失败
	ErrMsg			string
}

type GWHeartbeatReq struct {
	Uid				string
}

type GWHeartbeatResp struct {
	Result			int
	ErrMsg			string
}

type GWCmdReq struct {
	Service			string
	Method			string
	Body			interface {}
}

type GWCmdResp struct {
	Result			int
	ErrMsg			string
	Body			interface {}
}

// 推送协议
type MPushItem struct {
	Cmd				string
	Body			interface {}
}

type MPushPacket struct {
	Count			int
	PushItems		[]MPushItem
}

//////////////////////////////////////////////////////
//					对外网关PacketHandle
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

//////////////////////////////////////////////////////
//					内部推送协议
//////////////////////////////////////////////////////
type MIPushReq struct {
	Uids			[]string
	Cmd				string
	Body			interface {}
}

//////////////////////////////////////////////////////
//					内部推送PacketHandle
//////////////////////////////////////////////////////
type MIPushPacketHandle struct {}

func (p *MIPushPacketHandle) RecvPacket(conn net.Conn, timeout_ms int) (net_core.Packet, error) {
	var packet MInnerPushPacket
	if bytes, err := mservice_define.CommonRecvPacket(conn, timeout_ms); err == nil {
		if err := json.Unmarshal(bytes, &packet); err != nil {
			return nil, err
		}
	}
	return &packet, nil
}

func (p *MIPushPacketHandle) SendPacket(conn net.Conn, packet net_core.Packet) error {
	return mservice_define.CommonSendPacket(conn, packet)
}
