package util

import (
	"fmt"
	"log"
	"net"
	"time"
	"sync"
	"errors"
	"hbservice/src/net/net_core"
)

const (
	stClose			int = 0
	stConneted		int = 1

	kMaxSendChannelSize		int = 1000
)

type RecvCallback		func(*TcpClient, net_core.Packet) error

// 参数说明:
// param[1]: 对象指针
// param[2]: 主动 or 被动关闭
// param[3]: 关闭原因，主动关闭为nil
type CloseCallback		func(*TcpClient, bool, error)

type TcpClient struct {
	net.Conn
	sync.Mutex

	Host			string
	Port			int
	Param			interface {}		// 自定义参数

	ctx				context.Context
	cancel			context.CancelFunc

	state			int		// stClose, stConnected
	channel_send	chan net_core.Packet
	packet_handle	net_core.PacketHandle

	cb_recv			RecvCallback		// 包接收回调函数
	cb_close		CloseCallback		// 连接关闭回调函数
}

func NewTcpClient(host string, port int, handle net_core.PacketHandle, cb_recv RecvCallback, cb_close CloseCallback, param interface {}) *TcpClient {
	return &TcpClient {
		Host:			host,
		Port:			port,
		state:			stClose,
		packet_handle:	handle,
		cb_recv:		cb_recv,
		cb_close:		cb_close,
		param:			param,
	}
}

func (p *TcpClient) Connect(timeout_ms int) error {
	p.Lock()
	defer p.Unlock()
	if p.state == stConnect {
		return nil
	}

	addr := fmt.Sprintf("%s:%d", p.Host, p.Port)
	timeout := time.Duration(timeout_ms) * time.Millisecond
	conn, err := net.DialTimeout("tcp", addr, timeout)
	if err != nil {
		return err
	}
	p.ctx, p.cancel = context.WithCancel(context.TODO())
	p.channel_send := make(chan net_core.Packet, kMaxSendChannelSize)

	p.state = stConnected
	p.Conn = conn
	return nil
}

func (p *TcpClient) Send(packet net_core.Packet) (err error) {
	defer func() {
		recover()
		err = errors.New("channel close")
	} ()
	p.channel_send <-packet
}

func (p *TcpClient) Close() {
	close(true)
}

func (p *TcpClient) GetState() {
	p.Lock()
	defer p.Unlock()
	return p.state
}

func (p *TcpClient) close(is_active_close bool) {
	defer func() {
		recover()
	} ()
	p.Lock()
	if p.state == stClose {
		return
	}
	p.state = stClose
	p.Unlock()

	if is_active_close == true {
		p.cancel()
	}
	p.Conn.Close()
	close(p.channel_send)
	if p.cb_close != nil {
		p.cb_close(is_active_close)
	}
}

func (p *TcpClient) do_send_loop() {
	for {
		select {
		case <-p.ctx.Done():
			break
		case packet, ok := <-p.channel_send:
			if !ok {
				break
			}
			if err := p.handle.SendPacket(p.Conn, packet); err != nil {
				log.Printf("TcpClient|host:%s, port:%d, SendPacket error:%#v", p.host, p.port, err)
				break
			}
		}
	}
	if p.cb_close != nil {
		p.cb_close(false)
	}
}

func (p *TcpClient) do_recv_loop() {
	for {
		packet, err := p.handle.RecvPacket(p.Conn, -1)
		if err != nil {
			log.Printf("TcpClient|host:%s, port:%d, RecvPacket error:%#v", p.host, p.port, err)
			break
		}
		if p.cb_recv != nil {
			if err := p.cb_recv(p, packet); err != nil {
				log.Printf("TcpClient|host:%s, port:%d, cb_recv error:%#v", p.host, p.port, err)
				break
			}
		}
	}
	if p.cb_close != nil {
		p.cb_close(false)
	}
}
