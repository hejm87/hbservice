package obj_client

import (
	"fmt"
	"sort"
	"time"
	"sync"
	"errors"
	"strconv"
	"context"
	"strings"
	"hbservice/src/util"	
	"hbservice/src/net/net_core"
	"hbservice/src/mservice/define"
)

const (
	kCall			int = 1			// 同步
	kAsyncCall		int = 2			// 异步

	kEventServiceChange		int = 1
	kEventClientChange		int = 2

	kClientExit				int = 1
	kClientTimeout			int = 2

	NameServiceDir			string = "NameService"
)

// event
type EventInfo struct {
	Event		int
	Value		interface {}
}

type EventServiceChange struct {
	IsUpdate		bool
	Service			string
}

type EventClientChange struct {
	Event		int
	Seq			string
	Host		string
	Port		int
}

// call packet
type RespPacketCb func(*mservice_define.MServicePacket, error)

type RespPacketResult struct {
	packet			*mservice_define.MServicePacket
	err				error
}

type CallRecord struct {
	call_type		int			// 调用方式：kCall, kAsyncCall
	hash			uint32
	req				*mservice_define.MServicePacket
	async_cb		RespPacketCb		// 异步方式回调函数
	channel_cb		chan *RespPacketResult
	client_seq		string
	client_host		string
	client_port		int
	send_ts			int64
}

//////////////////////////////////////////////////////////
//					ObjectProxy
//////////////////////////////////////////////////////////
// 名字服务在etcd的存储格式
// 目录: /mservice/{服务名}/{服务名}_{host}:{port}_{seq}
type ObjectAddr struct {
	host		string
	port		int
	seq			string
}

type ObjectProxy struct {
	name				string	
	queue				util.MapQueue[string]	// CallRecord
	active_clients		[]*ObjectClient
	inactive_clients	[]*ObjectClient
	channel_event		chan *EventInfo
	channel_recv		chan *RespPacketResult
	channel_send		chan *CallRecord
	handle				net_core.PacketHandle
	etcd				*util.Etcd
	cfg					mservice_define.ObjClientConfig
	sync.Mutex
}

func NewObjectProxy(name string, handle net_core.PacketHandle) *ObjectProxy {
	etcd_cfg := util.GetConfigValue[mservice_define.MServiceConfig]().EtcdCfg
	oclient_cfg := util.GetConfigValue[mservice_define.MServiceConfig]().ObjClientCfg
	return &ObjectProxy {
		name: 		name,
		handle:		handle,
		active_clients:		make([]*ObjectClient, 5),
		inactive_clients:	make([]*ObjectClient, 5),
		channel_event:		make(chan *EventInfo, 1000),
		etcd:				util.NewEtcd(etcd_cfg.Addrs, etcd_cfg.Username, etcd_cfg.Password),
		cfg:				oclient_cfg,
	}
}

func (p *ObjectProxy) Start() error {
	ctx := context.TODO()
	prefix := NameServiceDir + "/" + p.name + "/" + p.name
	go func(ctx context.Context, prefix string) {
		p.etcd.Watch(ctx, prefix, true, p.do_watch)
	} (ctx, prefix)

	addrs, err := p.get_service_addrs()
	if err != nil {
		return err
	}
	for _, x := range addrs {
		client := NewObjectClient(p.name, x.seq, x.host, x.port, p.handle, p.channel_event, p.channel_recv)
		if err := client.Connect(); err == nil {
			p.active_clients = append(p.active_clients, client)
		} else {
			p.inactive_clients = append(p.inactive_clients, client)
		}
	}
	return nil	
}

// 暴力关闭
func (p *ObjectProxy) Close() {
	for _, x := range p.active_clients	{
		x.Close()
	}
	close(p.channel_event)
	close(p.channel_recv)
	close(p.channel_send)
	p.etcd.Close()
}

func (p *ObjectProxy) Call(hash uint32, req *mservice_define.MServicePacket) (*mservice_define.MServicePacket, error) {
	record := &CallRecord {
		call_type:		kCall,
		hash:			hash,
		req:			req,
		channel_cb:		make(chan *RespPacketResult),
		send_ts:		time.Now().Unix(),
	}

	p.Lock()
	p.queue.Push(util.GenUuid(), record, true)
	p.Unlock()

	p.channel_send <-record

	var err error
	var res *mservice_define.MServicePacket = nil
	cfg := util.GetConfigValue[mservice_define.MServiceConfig]().ObjClientCfg
	select {
	case <-time.After(time.Duration(cfg.CallTimeoutSec) * time.Second):
		err = errors.New("timeout")
	case result, ok := <-record.channel_cb:
		if ok {
			err = result.err
			res = result.packet
		} else {
			err = errors.New("call exception")
		}
	}
	return res, err
}

func (p *ObjectProxy) AsyncCall(hash uint32, req *mservice_define.MServicePacket, cb RespPacketCb) {
	record := &CallRecord {
		call_type:		kAsyncCall,
		hash:			hash,
		req:			req,
		channel_cb:		make(chan *RespPacketResult),
		send_ts:		time.Now().Unix(),
	}

	p.Lock()
	p.queue.Push(util.GenUuid(), record, true)
	p.Unlock()
}

func (p *ObjectProxy) do_watch(results []util.WatchResult) {
	for _, x := range results {
		var change *EventServiceChange
		if x.Type == util.ETCD_PUT {
			change = &EventServiceChange {
				IsUpdate:	true,
				Service:	x.Key,
			}
		} else {
			change = &EventServiceChange {
				IsUpdate:	false,
				Service:	x.Key,
			}
		}
		p.channel_event <-&EventInfo {
			Event:	kEventServiceChange,
			Value:	change,
		}
	}
}

func (p *ObjectProxy) loop_main() {
	for {
		select {
		case record, ok := <-p.channel_send:
			if !ok {
				break
			}
			p.do_send(record)
		case event, ok := <-p.channel_event:
			if !ok {
				break
			}
			p.do_event(event)
		}
	}
}

func (p *ObjectProxy) loop_recv() {
	for recv := range p.channel_recv {
		var exists bool
		var record *CallRecord = nil
		p.Lock()
		if res, exists := p.queue.Get(recv.packet.Header.TraceID, true); exists {
			record = res.(*CallRecord)
		}
		p.Unlock()

		if exists == true {
			if record.call_type == kCall {
				record.channel_cb <-recv
			} else {
				record.async_cb(recv.packet, recv.err)
			}
		} else {
			// 没有找到发送包，记为超时处理
			p.channel_event <-&EventInfo {
				Event:	kEventClientChange,
				Value:	&EventClientChange {
					Event:	kClientTimeout,
					Seq:	record.client_seq,
					Host:	record.client_host,
					Port:	record.client_port,
				},
			}
		}
	}
}

func (p *ObjectProxy) do_send(record *CallRecord) {
	index := int(record.hash % uint32(len(p.active_clients)))
	client := p.active_clients[index]
	if err := client.Send(record.req); err != nil {
		p.remove_active_client_by_index(index)
	}
}

func (p *ObjectProxy) do_event(ev *EventInfo) {
	if ev.Event == kEventServiceChange {
		p.do_event_service_change(ev.Value.(*EventServiceChange))
	} else if ev.Event == kEventClientChange {
		p.do_event_client_change(ev.Value.(*EventClientChange))	
	}
}

func (p *ObjectProxy) do_event_service_change(change *EventServiceChange) {
	if oaddr, ok := convert_service_addr(change.Service); ok {
		if change.IsUpdate == true {
			p.do_service_modify(oaddr.host, oaddr.port, oaddr.seq)
		} else {
			p.do_service_remove(oaddr.host, oaddr.port, oaddr.seq)
		}
	}
}

func (p *ObjectProxy) do_event_client_change(change *EventClientChange) {
	if change.Event == kClientExit {
		p.do_event_client_exit(change)
	} else {
		p.do_event_client_timeout(change)
	}
}

func (p *ObjectProxy) do_event_client_exit(change *EventClientChange) {
	p.do_client_exit(change.Host, change.Port, change.Seq)
}

func (p *ObjectProxy) do_event_client_timeout(change *EventClientChange) {
	index := p.find_client_index(p.active_clients, change.Host, change.Port, change.Seq)
	if index < 0 {
		return
	}
	client := p.active_clients[index]
	client.Status.FailCount++
	client.Status.SeriesFailCount++
	if client.Status.SeriesFailCount >= p.cfg.InvalidSeriesFailCount {
		p.remove_active_client_by_index(index)
	}
}

func (p *ObjectProxy) do_service_modify(host string, port int, seq string) {
	if index := p.find_client_index(p.active_clients, host, port, seq); index >= 0 {
		client := p.active_clients[index]
		if client.Seq == seq {
			return
		}
		// 替换service
		client.Seq = seq
		p.remove_active_client_by_index(index)
	} else {
		// 新增service
		client := NewObjectClient(p.name, seq, host, port, p.handle, p.channel_event, p.channel_recv)
		p.active_clients = append(p.active_clients, client)
		sort_clients(p.active_clients)
	}
}

func (p *ObjectProxy) do_service_remove(host string, port int, seq string) {
	if index := p.find_client_index(p.active_clients, host, port, seq); index >= 0 {
		client := p.active_clients[index]
		client.Close()
		p.active_clients = append(p.active_clients[0:index], p.active_clients[index + 1:]...)
		return
	}
	if index := p.find_client_index(p.inactive_clients, host, port, seq); index >= 0 {
		p.inactive_clients = append(p.inactive_clients, p.inactive_clients[index + 1:]...)
	}
}

func (p *ObjectProxy) do_client_exit(host string, port int, seq string) {
	p.remove_active_client(host, port, seq)
}

func (p *ObjectProxy) remove_active_client(host string, port int, seq string) bool {
	index := p.find_client_index(p.active_clients, host, port, seq)
	if index < 0 {
		return false
	}
	return p.remove_active_client_by_index(index)
}

func (p *ObjectProxy) remove_active_client_by_index(index int) bool {
	if len(p.active_clients) >= index {
		return false
	}
	client := p.active_clients[index]
	client.Close()
	client.Status.LatestInactiveTs = time.Now().Unix()
	client.Status.SuccessCount = 0
	client.Status.FailCount = 0
	client.Status.SeriesFailCount = 0
	p.active_clients = append(p.active_clients[0:index], p.active_clients[index + 1:]...)
	p.inactive_clients = append(p.inactive_clients, client)
	sort_clients(p.inactive_clients)
	return true
}

func (p *ObjectProxy) find_client_index(clients []*ObjectClient, host string, port int, seq string) int {
	for index, client := range clients {
		if client.Host == host && client.Port == port {
			if seq != "" && seq != client.Seq {
				break
			}
			return index
		}	
	}
	return -1
}

func (p *ObjectProxy) get_service_addrs() ([]ObjectAddr, error) {
	var addrs []ObjectAddr
	prefix := NameServiceDir + "/" + p.name + "/" + p.name
	kvs, err := p.etcd.GetWithPrefix(prefix)
	if err != nil {
		return addrs, err
	}
	for k, _ := range kvs {
		if oaddr, ok := convert_service_addr(k); ok {
			addrs = append(addrs, oaddr)
		}
	}
	return addrs, nil
}

////////////////////////////////////////////////////////
//					工具函数
////////////////////////////////////////////////////////
func sort_clients(clients []*ObjectClient) {
	sort.SliceStable(clients, func(i, j int) bool {
		addr1 := fmt.Sprintf("%s:%d", clients[i].Host, clients[i].Port)
		addr2 := fmt.Sprintf("%s:%d", clients[j].Host, clients[j].Port)
		if addr1 < addr2 {
			return true
		}
		return false
	})
}

func convert_service_addr(addr string) (ObjectAddr, bool) {
	var oaddr ObjectAddr
	arr := strings.Split(addr, "_")
	if len(arr) != 3 {
		return oaddr, false
	}
	arr1 := strings.Split(arr[1], ":")
	port, err := strconv.Atoi(arr1[1])
	if err != nil {
		return oaddr, false
	}
	oaddr = ObjectAddr {
		host:	arr1[0],
		port:	port,
		seq:	arr[2],
	}
	return oaddr, true
}
