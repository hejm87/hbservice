package obj_client

import (
	"fmt"
	"net"
)

const (
	kClose		int = 0
	kConnected	int = 1

	kConnectMs			int = 100
	kMaxSendQueueSize	int = 1000
)

type ObjectClient struct {
	name		string
	host		string
	port		int

	handle				net_core.PacketHandle
	channel_send		chan net_core.Packet
	channel_recv_notify	chan net_core.Packet	

	State		int		// 连接状态：kClose, kConnected

	net.Conn
}

func NewObjectClient(name string, host string, port int, handle net_core.PacketHandle, recv_notity chan) *ObjectClient {
	return &ObjectClient {
		name:	name,
		host:	host,
		port:	port,
		Status:	kClose,
		handle:	handle,
		channel_send:	make(net_core.Packet, kMaxSendQueueSize),
		channel_recv_notify:	recv_notify,
	}
}

func (p *ObjectClient) IsConnected() bool {
	if p.Status == kClose {
		return false
	}
	return true
}

func (p *ObjectClient) Connect() error {
	if p.Status == kClose {
		addr := fmt.Sprintf("%s:%d", p.host, p.port)
		timeout := time.Duration(kConnectMs) * time.MilliSecond
		if conn, err := net.DialTimeout("tcp", addr, timeout); err == nil {
			p.Conn = conn
			p.Status = kConnected
			go send_loop()
			go recv_loop()
		}
	}
	return nil
}

func (p *ObjectClient) Close() {
	if p.Status == kConnected {
		p.Status = kClose
		close(p.channel_send)
		p.Conn.Close()
	}
}

func (p *ObjectClient) send_loop() {
	for packet := range p.channel_send {
		if err := p.handle.SendPacket(p.Conn, packet); err != nil {
			p.Close()
			break
		}
	}
}

func (p *ObjectClient) recv_loop() {
	for {
		if packet, err := p.handle.RecvPacket(p.Conn, -1); err == nil {
			p.channel_recv_notify <-packet
		} else {
			log.Printf("sclient|name:%s, host:%s, port:%d RecvPacket fail, error:%#v\n", p.Name, p.Host, p.Port, err)
			p.Close()
			break
		}
	}
}