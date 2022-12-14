package mservice_define

import (
	"net"
	"time"
	"errors"
	"encoding/json"
	"encoding/binary"
	"hbservice/src/util"
	"hbservice/src/net/net_core"
	"github.com/mitchellh/mapstructure"
)

const (
	MS_REQUEST		int = 1		// 请求
	MS_RESPONSE		int = 2		// 回复

	MS_CALL			int = 1
	MS_CAST			int = 2
)

type MServiceHeader struct {
	Service			string		// 服务名
	Method			string		// 方法名
	TraceId			string		// 追踪id
	CallType		int			// 调用方式, 取值: MS_CALL, MS_CAST
	Direction		int			// 请求/回复方向, 取值:MS_REQUEST, MS_RESPONSE
}

type MServicePacket struct {
	Header			MServiceHeader
	Body			interface {}
}

func (p *MServicePacket) ToStr() string {
	return ""
}

func CreateReqPacket(service string, method string, call_type int, body interface {}) *MServicePacket {
	return &MServicePacket {
		Header:		MServiceHeader {
			Service:	service,
			Method:		method,
			TraceId:	util.GenUuid(),
			CallType:	call_type,
			Direction:	MS_REQUEST,
		},
		Body:		body,
	}
}

func CreateRespPacket(service string, method string, trace_id string, call_type int, body interface {}) *MServicePacket {
	return &MServicePacket {
		Header:		MServiceHeader {
			Service:	service,
			Method:		method,
			TraceId:	trace_id,
			CallType:	call_type,
			Direction:	MS_RESPONSE,
		},
		Body:		body,
	}
}

type MPacketHandle struct {}

func (p *MPacketHandle) RecvPacket(conn net.Conn, timeout_ms int) (net_core.Packet, error) {
	var packet MServicePacket
	if bytes, err := CommonRecvPacket(conn, timeout_ms); err == nil {
		if err := json.Unmarshal(bytes, &packet); err != nil {
			return nil, err
		}
	} else {
		return nil, err
	}
	return &packet, nil
}

func (p *MPacketHandle) SendPacket(conn net.Conn, packet net_core.Packet) error {
	return CommonSendPacket(conn, packet)
}

///////////////////////////////////////////////////////////////////////
//						通用转包函数
///////////////////////////////////////////////////////////////////////
func GetPacketBody[T any](packet *MServicePacket) (res T, err error) {
	 err = mapstructure.Decode(packet.Body, &res); 
	 return res, err
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
	n, err := util.NetRecvTimeout(conn, bytes, int(body_size), timeout)
	if err != nil {
		return nil, err
	}
	if n == 0 {
		return nil, errors.New("socket peer close")
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
