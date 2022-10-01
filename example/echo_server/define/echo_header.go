package define

import (
	"net"
	"time"
	"encoding/json"
	"encoding/binary"
	"hbservice/src/util"
	"hbservice/src/net/net_core"
)

////////////////////////////////////////////////////
//					Packet
////////////////////////////////////////////////////
type EchoHeader struct {
	Method		string
}

type EchoPacket struct {
	Header		EchoHeader
	Body		interface {}
}

func (p *EchoPacket) ToStr() string {
	return ""
}

type EchoReq struct {
	Content		string
}

type EchoResp struct {
	Content		string
}

////////////////////////////////////////////////////
//					PacketHandle
////////////////////////////////////////////////////
type EchoPacketHandle struct {}

func (p *EchoPacketHandle) RecvPacket(conn net.Conn, timeout_ms int) (net_core.Packet, error) {
	var timeout *time.Time = nil	
	if timeout_ms > 0 {
		duration := time.Duration(timeout_ms)
		timeout = new(time.Time)
		*timeout = time.Now().Add(duration * time.Millisecond)
	}
	// read body size
	var body_size int32
	util.SetTimeout(conn, timeout)
	if err := binary.Read(conn, binary.BigEndian, &body_size); err != nil {
		return nil, err
	}

	// read body content
	buf := make([]byte, body_size)
	if _, err := util.NetRecvTimeout(conn, buf, int(body_size), timeout); err != nil {
		return nil, err
	}

	// convert buffer to packet
	var packet EchoPacket
	if err := json.Unmarshal(buf, &packet); err != nil {
		return nil, err
	}
	return &packet, nil
}

func (p *EchoPacketHandle) SendPacket(conn net.Conn, packet net_core.Packet) error {
	jsonByte, err := json.Marshal(packet)
	if err != nil {
		return err
	}
	size := int32(len(jsonByte))
	if err := binary.Write(conn, binary.BigEndian, &size); err != nil {
		return err
	}
	if err := binary.Write(conn, binary.BigEndian, jsonByte); err != nil {
		return err
	}
	return nil
}
