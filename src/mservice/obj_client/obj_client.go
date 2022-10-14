package obj_client

import (
	"fmt"
	"log"
	"net"
	"time"
	"hbservice/src/util"
	"hbservice/src/net/net_core"
	"hbservice/src/mservice/define"
)

const (
	kClose		int = 0
	kConnected	int = 1

	kConnectMs			int = 100
	kMaxSendQueueSize	int = 1000
)

//type ObjectClientStatus struct {
//	LatestActiveTs		int64
//	LatestInactiveTs	int64
//	SuccessCount		int
//	FailCount			int
//	SeriesFailCount		int
//}
//
//type ObjectClientData struct {
//	Status				ObjectClientStatus
//	ObjProxy			*ObjectProxy
//}

type ObjectClient struct {
	Name		string
	Seq			string
	Host		string
	Port		int
	State		int		// 连接状态：kClose, kConnected
	Status		ObjectClientStatus

	handle				net_core.PacketHandle
	channel_event		chan *EventInfo
	channel_send		chan net_core.Packet
	channel_recv_notify	chan *RespPacketResult

	net.Conn
}

func NewObjectClient(name string, seq string, host string, port int, handle net_core.PacketHandle, channel_event chan *EventInfo, channel_recv_notify chan *RespPacketResult) *ObjectClient {
	return &ObjectClient {
		Name:		name,
		Seq:		seq,
		Host:		host,
		Port:		port,
		State:		kClose,
		handle:		handle,
		channel_event:			channel_event,
		channel_send:			make(chan net_core.Packet, kMaxSendQueueSize),
		channel_recv_notify:	channel_recv_notify,
	}
}

func (p *ObjectClient) IsConnected() bool {
	if p.State == kClose {
		return false
	}
	return true
}

func (p *ObjectClient) Connect() error {
	if p.State == kClose {
		addr := fmt.Sprintf("%s:%d", p.Host, p.Port)
		timeout := time.Duration(kConnectMs) * time.Millisecond
		if conn, err := net.DialTimeout("tcp", addr, timeout); err == nil {
			p.Conn = conn
			p.State = kConnected
			go p.send_loop()
			go p.recv_loop()
		}
	}
	return nil
}

func (p *ObjectClient) Send(packet net_core.Packet) error {
	return p.handle.SendPacket(p.Conn, packet)
}

func (p *ObjectClient) Close() {
	defer func() {
		recover()
	} ()
	if p.State == kConnected {
		p.State = kClose
		close(p.channel_send)
		p.Conn.Close()
		p.channel_event <-&EventInfo {
			Event:	kEventClientChange,
			Value:	&EventClientChange {
				Event:	kClientExit,
				Seq:	p.Seq,
				Host:	p.Host,
				Port:	p.Port,
			},
		}
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
			p.channel_recv_notify <-&RespPacketResult {
				packet:		packet.(*mservice_define.MServicePacket),
				err:		err,
			}
		} else {
			log.Printf("sclient|name:%s, host:%s, port:%d RecvPacket fail, error:%#v\n", p.Name, p.Host, p.Port, err)
			p.Close()
			break
		}
	}
}