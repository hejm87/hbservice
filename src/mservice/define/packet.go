package mservice_define

import (
	"net"
	"time"
	"encoding/json"
	"encoding/binary"
	"hbservice/src/util"
	"hbservice/src/net/net_core"
)

// 对于推送协议：
// Service设置为Pusher，Method为推送的命令
type MServiceHeader struct {
	Uid				string		// 只有推送时用于查找uid对应channel_id
	Service			string
	Method			string
	TraceID			string
	Direction		int			// 0：请求，1：回复
}

type MServicePacket struct {
	Header			MServiceHeader
	Body			interface {}
}

func (p *MServicePacket) ToStr() string {
	return ""
}

type MPacketHandle struct {}

func (p *MPacketHandle) RecvPacket(conn net.Conn, timeout_ms int) (net_core.Packet, error) {
	var packet MServicePacket
	if bytes, err := CommonRecvPacket(conn, timeout_ms); err == nil {
		if err := json.Unmarshal(bytes, &packet); err != nil {
			return nil, err
		}
	}
	return &packet, nil
}

func (p *MPacketHandle) SendPacket(conn net.Conn, packet net_core.Packet) error {
	return CommonSendPacket(conn, packet)
}

///////////////////////////////////////////////////////////////////////
//						通用收/发包函数
///////////////////////////////////////////////////////////////////////
func CommonRecvPacket(conn net.Conn, timeout_ms int) ([]byte, error) {
	var timeout *time.Time = nil
	if timeout_ms > 0 {
		timeout = new(time.Time)
		*timeout = time.Now().Add(time.Duration(timeout_ms) * time.Millisecond)
	}
	// read body size
	var body_size int32
	util.SetTimeout(conn, timeout)
	if err := binary.Read(conn, binary.BigEndian, &body_size); err != nil {
		return nil, err
	}

	// read body content
	bytes := make([]byte, body_size)
	if _, err := util.NetRecvTimeout(conn, bytes, int(body_size), timeout); err != nil {
		return nil, err
	}
	return bytes, nil
}

func CommonSendPacket(conn net.Conn, packet net_core.Packet) error {
	bytes, err := json.Marshal(packet)
	if err != nil {
		return err
	}
	size := int32(len(bytes))
	if err := binary.Write(conn, binary.BigEndian, &size); err != nil {
		return err
	}
	if err := binary.Write(conn, binary.BigEndian, bytes); err != nil {
		return err
	}
	return nil
}

///////////////////////////////////////////////
//				通用微服务包处理
///////////////////////////////////////////////
//type MCommonPacketHandle[T any] struct {}
//
//func (p *MCommonPacketHandle[T]) RecvPacket(conn net.Conn, timeout_ms int) (net_core.Packet, error) {
//	var timeout *time.Time = nil
//	if timeout_ms > 0 {
//		timeout = new(time.Time)
//		*timeout = time.Now().Add(time.Duration(timeout_ms) * time.Millisecond)
//	}
//
//	// read body size
//	var body_size int32
//	util.SetTimeout(conn, timeout)
//	if err := binary.Read(conn, binary.BigEndian, &body_size); err != nil {
//		return nil, err
//	}
//
//	// read body content
//	bytes := make([]byte, body_size)
//	if _, err := util.NetRecvTimeout(conn, bytes, int(body_size), timeout); err != nil {
//		return nil, err
//	}
//
//	// convert buffer to packet
//	var packet T
//	if err := json.Unmarshal(bytes, &packet); err != nil {
//		return nil, err
//	}
//	return packet, nil
//}
//
//func (p *MCommonPacketHandle[T]) SendPacket(conn net.Conn, packet net_core.Packet) error {
//	bytes, err := json.Marshal(packet)
//	if err != nil {
//		return err
//	}
//	size := int32(len(bytes))
//	if err := binary.Write(conn, binary.BigEndian, &size); err != nil {
//		return err
//	}
//	if err := binary.Write(conn, binary.BigEndian, bytes); err != nil {
//		return err
//	}
//	return nil
//}