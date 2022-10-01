package main

import (
	"fmt"
	"net"
	"time"
	"flag"
	"sync/atomic"
	"encoding/binary"
	"github.com/mitchellh/mapstructure"
	"hbservice/example/echo_server/define"
	"hbservice/src/util"
)

var g_count int32 = 0

func main() {
	host := flag.String("h", "", "host")	
	port := flag.Int("p", 8080, "port")
	clients := flag.Int("clients", 1, "clients")
	flag.Parse()
	fmt.Printf("param|host:%s\n", *host)
	fmt.Printf("param|port:%d\n", *port)
	fmt.Printf("param|clients:%d\n", *clients)
	for i := 0; i < *clients; i++ {
		go do_client(*host, *port)
	}
	flag.Parse()

	var latest int32 = atomic.LoadInt32(&g_count)
	for {
		select {
		case <-time.After(1 * time.Second):
			new_latest := atomic.LoadInt32(&g_count)
			fmt.Printf("################### delta: %d/s\n", new_latest - latest)
			latest = new_latest
		}
	}
}

func do_client(host string, port int) {
	addr := fmt.Sprintf("%s:%d", host, port)
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		fmt.Printf("client dial addr:%s error:%#v\n", addr, err)
		return
	}
	fmt.Printf("client.tcp dial addr:%s succeed\n", addr)
	packet_handle := define.EchoPacketHandle {}

	for {
		if err := send_echo_packet_quick(conn, &packet_handle); err != nil {
			fmt.Printf("send_echo_packet error:%#v\n", err)
			return
		}
		if err := recv_echo_packet_quick(conn, &packet_handle); err != nil {
			fmt.Printf("recv_echo_packet error:%#v\n", err)
			return
		}
		atomic.AddInt32(&g_count, 1)
	}
}

func send_echo_packet(conn net.Conn, handle *define.EchoPacketHandle) error {
	packet := &define.EchoPacket {
		Header: define.EchoHeader {Method: "echo"},
		Body:	define.EchoReq {Content: "hello"},
	}
	return handle.SendPacket(conn, packet)
}

func recv_echo_packet(conn net.Conn, handle *define.EchoPacketHandle) error {
	packet, err := handle.RecvPacket(conn, -1)
	if err != nil {
		return err
	}
	echo_packet := packet.(*define.EchoPacket)
	var resp define.EchoResp
	err = mapstructure.Decode(echo_packet.Body, &resp)
	if err != nil {
		return err
	}
	return nil
}

func send_echo_packet_quick(conn net.Conn, handle *define.EchoPacketHandle) error {
	jsonByte := []byte("{\"Header\":{\"Method\":\"echo\"},\"Body\":{\"Content\":\"hello\"}}")
	var size int32 = 55
	if err := binary.Write(conn, binary.BigEndian, &size); err != nil {
		return err
	}
	if err := binary.Write(conn, binary.BigEndian, jsonByte); err != nil {
		return err
	}
	return nil
}

func recv_echo_packet_quick(conn net.Conn, handle *define.EchoPacketHandle) error {
	var body_size int32
	if err := binary.Read(conn, binary.BigEndian, &body_size); err != nil {
		return err
	}

	// read body content
	buf := make([]byte, body_size)
	if _, err := util.NetRecv(conn, buf, int(body_size), 3000); err != nil {
		return err
	}
	return nil
}