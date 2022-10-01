package net_core

import (
	"net"
)

type Packet interface {
	ToStr() string
}

//type Agent interface {
//	ID() string
//	Push(Packet) error
//}

// 用于收发包
type PacketHandle interface {
	// param[1]: conn		net.Conn
	// param[2]: timeout_ms	int
	RecvPacket(net.Conn, int) (Packet, error)
	SendPacket(net.Conn, Packet) error
}

// 服务逻辑处理
type LogicHandle interface {
	Init() error
	OnAccept(net.Conn, PacketHandle, NetServer) (string, error)
	// param[1]: channel_id		string
	// param[2]: packet			Packet
	// param[3]: server			NetServer
	OnMessage(string, Packet, NetServer) error

	// param[1]: channel_id		string
	// param[2]: server			NetServer
	OnClose(string, NetServer) error
}

// 网络服务器抽象
type NetServer interface {
	Start() error
	Shutdown()
	PushChannel(string, Packet) error
}

// 网络服务参数
type NetServerParam struct {
	LogicHandle
	PacketHandle
	Name		string
	Host		string
	Port		int
}
