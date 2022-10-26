package tcp_server

import (
	"fmt"
	"log"
	"net"
	"errors"
//	"context"
	"hbservice/src/util"
	"hbservice/src/net/net_core"
)

type CbTimerValue struct {
	handle		net_core.LogicHandle
	value		interface {}
}

type TcpServer struct {
	params			[]net_core.NetServerParam
	srv_socks		map[string]net.Listener
	channels		*net_core.ChannelMap
	channel_exit	chan string
//	ctx			context.Context
//	cancel		context.CancelFunc
}

func NetServer(params []net_core.NetServerParam) net_core.NetServer {
//	ctx, cancel := context.WithCancel(context.Background())
	server := &TcpServer {
		params:			params,
		srv_socks:		make(map[string]net.Listener),
		channels:		net_core.NewChannels(),
		channel_exit:	make(chan string),	
	//	ctx:		ctx,
	//	cancel:		cancel,
	}
	return server
}

// 阻塞知道程序结束
func (p *TcpServer) Start() error {
	defer func() {
		p.release()
	} ()
	for _, x := range p.params {
		addr := fmt.Sprintf("%s:%d", x.Host, x.Port)
		if sock, err := net.Listen("tcp", addr); err == nil {
			log.Printf("INFO|TcpServer|listen addr:%s success", addr)
			p.srv_socks[x.Name] = sock
			go p.start_logic_service(x, sock)
		} else {
			log.Printf("ERROR|TcpServer|listen addr:%s error:%#v", addr, err)
			return err
		}
	}
	p.wait()
	return nil
}

// 考虑优雅关闭
func (p *TcpServer) Shutdown() {
//	p.cancel()
	p.channel_exit <-"application shutdown"
}

func (p *TcpServer) PushChannel(id string, packet net_core.Packet) error {
	if channel, ok := p.channels.Get(id); ok {
		channel.Push(packet)
	} else {
		return errors.New("no channel exists")
	}
	return nil
}

func (p *TcpServer) CloseChannel(id string) error {
	if channel, ok := p.channels.LoadAndDelete(id); ok {
		channel.Close()
	} else {
		return errors.New("no channel exists")
	}
	return nil
}

func (p *TcpServer) SetTimer(param net_core.NetTimerParam, handle net_core.LogicHandle) error {
	var err error
	cb_value := &CbTimerValue {
		handle:		handle,
		value:		param.Value,
	}
	if param.Type == util.TIMER_DELAY {
		err = util.GetTimerMgrInstance().SetTimerDelay(
			param.Id,
			cb_value,
			param.Delay,
			p.TimerCallback,
		)
	} else {
		err = util.GetTimerMgrInstance().SetTimerPeriod(
			param.Id,
			cb_value,
			param.Delay,
			param.Period,
			p.TimerCallback,
		)
	}
	return err
}

func (p *TcpServer) TimerCallback(id string, value interface {}) {
	cb_value := value.(*CbTimerValue)
	cb_value.handle.OnTimer(id, cb_value.value, p)
}

///////////////////////////////////////////////////
//				private function
///////////////////////////////////////////////////
func (p *TcpServer) start_logic_service(param net_core.NetServerParam, srv_conn net.Listener) {
	defer func() {
		p.channel_exit <-"logic_service exit"
	} ()
	fmt.Printf("INFO|TcpServer|start_logic_service, name:%s\n", param.Name)
	if err := param.LogicHandle.Init(p); err != nil {
		return
	}
	for {
		if cli, err := srv_conn.Accept(); err == nil {
			go func(conn net.Conn) {
				var id  string
				var err error
				if id, err = param.LogicHandle.OnAccept(conn, param.PacketHandle, p); err != nil {
					log.Printf("INFO|TcpServer|name:%s OnAccept fail, error:%#v", param.Name, err)
					return
				}
				if _, ok := p.channels.Get(id); ok {
					log.Printf("INFO|TcpServer|name:%s, channel_id:%s exists", param.Name, id)
					return
				}
				channel := net_core.NewChannel(id, cli, param.LogicHandle, param.PacketHandle, p)
				defer func(channel *net_core.Channel) {
					channel.Close()
					p.channels.Remove(channel.ID())
				} (channel)
				p.channels.Add(channel)
				go channel.SendLoop()
				if err = channel.RecvLoop(); err != nil {
					log.Printf("INFO|TcpServer|name:%s RecvLoop fail, error:%#v", param.Name, err)
				}
			} (cli)
		} else {
			// 任意一个logic_service挂掉，都需要将所有的logic_service关掉
			log.Printf("INFO|TcpServer|name:%s Accept fail, error:%#v", param.Name, err)
			break
		}
	}
	return
}

func (p *TcpServer) wait() {
	select {
//	case <-p.ctx.Done():
	case <-p.channel_exit:
		break
	}
}

func (p *TcpServer) release() {
	// 防止新连接
	for k, v := range p.srv_socks {
		v.Close()
		delete(p.srv_socks, k)
	}
	// 删除所有的channel
	for _, ch := range p.channels.All() {
		ch.Close()
		p.channels.Remove(ch.ID())
	}
}
