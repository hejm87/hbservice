package main

import (
	"fmt"
	"log"
	"net"
	"errors"
	"net/http"
	_ "net/http/pprof"
	"github.com/segmentio/ksuid"
	"github.com/mitchellh/mapstructure"
	"hbservice/src/net/net_core"
	"hbservice/src/mservice"
	"hbservice/example/echo_server/define"
)

////////////////////////////////////////////////////
//					LogicHandle
////////////////////////////////////////////////////
type EchoLogicHandle struct {}

func (p *EchoLogicHandle) Init(server net_core.NetServer) error {
	return nil
}
       	
func (p *EchoLogicHandle) OnAccept(conn net.Conn, handle net_core.PacketHandle, server net_core.NetServer) (string, error) {
	return ksuid.New().String(), nil
}
       	
func (p *EchoLogicHandle) OnMessage(channel_id string, packet net_core.Packet, server net_core.NetServer) error {
	req_packet := packet.(*define.EchoPacket)

	var resp_body interface {}
	var resp_packet net_core.Packet
	var err error

	resp_body, err = p.get_body(req_packet)
	if err != nil {
		return err
	}

	resp_packet, err = p.do_method(req_packet.Header.Method, resp_body)
	if err != nil {
		return err
	}

	return server.PushChannel(channel_id, resp_packet)
}

func (p *EchoLogicHandle) OnClose(channel_id string, server net_core.NetServer) error {
	return nil
}

func (p *EchoLogicHandle) OnTimer(timer_id string, server net_core.NetServer) {
	return
}

func (p *EchoLogicHandle) get_body(packet *define.EchoPacket) (interface {}, error) {
	var resp interface {}
	switch packet.Header.Method {
	case "echo":
		body := define.EchoReq {}
		if err := mapstructure.Decode(packet.Body, &body); err == nil {
			resp = body
		} else {
			return nil, err
		}
	default:
		return nil, errors.New("no method exists")
	}
	return resp, nil
}

func (p *EchoLogicHandle) do_method(method string, body interface {}) (net_core.Packet, error) {
	var resp_packet net_core.Packet = nil
	var resp_err error = nil
	switch method {
	case "echo":
		resp_packet, resp_err = p.do_echo(body)
	default:
		resp_err = errors.New("no method exists")
	}
	return resp_packet, resp_err
}

func (p *EchoLogicHandle) do_echo(body interface{}) (net_core.Packet, error) {
	req := body.(define.EchoReq)
	resp_packet := &define.EchoPacket {
		Header: define.EchoHeader {Method: "echo"},
		Body:	define.EchoResp {Content: req.Content},
	}
	return resp_packet, nil
}

///////////////////////////////////////////////////////////
//func main() {
//	go func() {
//		fmt.Printf("ready to listen 8000\n")
//		http.ListenAndServe("0.0.0.0:8000", nil)
//	} ()
//
//	var server_params []net_core.NetServerParam
//	server_params = append(
//		server_params, 
//		net_core.NetServerParam {
//			Name:	"echo",
//			Host:	"",
//			Port:	8080,
//			LogicHandle:	&EchoLogicHandle {},
//			PacketHandle:	&define.EchoPacketHandle {},
//		},
//	)
//
//	server := tcp_server.NetServer(server_params)
//	go func(server net_core.NetServer) {
//		server.Start()
//	} (server)
//
//	select {
//	case <- time.After(time.Second * 3600):
//		fmt.Printf("tcp_server shutdown\n")
//		server.Shutdown()
//	}
//}

func main() {
	server_params := get_server_params()
	if err := container.GetInstance().Start(server_params); err != nil {
		log.Fatalf("container.Start error:%#v", err)
	}
}

func get_server_params() []net_core.NetServerParam {
	var server_params []net_core.NetServerParam
	server_params = append(
		server_params, 
		net_core.NetServerParam {
			LogicHandle:	&EchoLogicHandle {},
			PacketHandle:	&define.EchoPacketHandle {},
		},
	)
	return server_params
}