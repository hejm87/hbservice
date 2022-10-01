package net_core

import (
	"fmt"
	"log"
	"net"
	"sync"
)

////////////////////////////////////////////////////////////
//						channel
////////////////////////////////////////////////////////////
type Channel struct {
	LogicHandle
	PacketHandle
	channelId	string
	chSend		chan Packet
	server		NetServer
	net.Conn
}

func NewChannel(id string, conn net.Conn, logic_handle LogicHandle, packet_handle PacketHandle, server NetServer) *Channel {
	return &Channel {
		LogicHandle:	logic_handle,
		PacketHandle:	packet_handle,
		channelId:		id,
		chSend:			make(chan Packet, 100),
		server:			server,
		Conn:			conn,
	}
}

func (p *Channel) ID() string {
	return p.channelId
}

func (p *Channel) Push(packet Packet) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("internal error: %#v", r)
		}
	} ()
	p.chSend <-packet
	return nil
}

func (p *Channel) Close() (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("internal error: %#v", r)
		}
	} ()
	close(p.chSend)
	p.Conn.Close()
	return nil
}

func (p *Channel) SendLoop() {
	for packet := range p.chSend {
		if err := p.PacketHandle.SendPacket(p.Conn, packet); err != nil {
			log.Printf("channel|id:%s SendPacket fail, error:%#v", p.channelId, err)
			return
		}
	}
}

func (p *Channel) RecvLoop() error {
	var err error = nil
	for {
		var packet Packet
		if packet, err = p.PacketHandle.RecvPacket(p.Conn, -1); err != nil {
			log.Printf("channel|id:%s GetPacket fail, error:%#v", p.channelId, err)
			break
		}
		if err = p.LogicHandle.OnMessage(p.channelId, packet, p.server); err != nil {
			log.Printf("channel|id:%s OnMessage fail, error:%#v", p.channelId, err)
			break
		}
	}
	log.Printf("channel|id:%s finish RecvLoop\n", p.channelId)
	return err
}

////////////////////////////////////////////////////////////
//						channels
////////////////////////////////////////////////////////////
type ChannelMap struct {
	channels	*sync.Map
}

func NewChannels() *ChannelMap {
	return &ChannelMap {
		channels: new(sync.Map),
	}
}

func (p *ChannelMap) Add(channel *Channel) {
	if channel.ID() == "" {
		return
	}
	p.channels.Store(channel.ID(), channel)
}

func (p *ChannelMap) Remove(id string) {
	p.channels.Delete(id)
}

func (p *ChannelMap) Get(id string) (*Channel, bool) {
	if id == "" {
		return nil, false
	}
	v, ok := p.channels.Load(id)
	if !ok {
		return nil, false
	}
	return v.(*Channel), true
}

func (p *ChannelMap) All() []Channel {
	arr := make([]Channel, 0)
	p.channels.Range(func(key, value interface{}) bool {
		arr = append(arr, value.(Channel))
		return true
	})
	return arr
}
