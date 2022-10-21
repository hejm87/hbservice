package obj_client

import (
	"fmt"
	"log"
	"sort"
	"time"
	"sync"
	"errors"
	"hbservice/src/util"	
	"hbservice/src/net/net_core"
	"hbservice/src/mservice/define"
	"hbservice/src/mservice/naming"
)

const (
	kCall			int = 1			// 同步调用
	kCallAsync		int = 2			// 异步调用
	kCast			int = 3			// 投递
	kBroadcast		int = 4			// 异步广播

	kEventServiceChange		int = 1
	kEventClientChange		int = 2

	kClientExit				int = 1
	kClientClose			int = 2
	kClientTimeout			int = 3

	// 调用报错信息
	kErrorCallTimeout			string = "call timeout"
	kErrorCallException			string = "call exception"
	kErrorCallNotExistsClient	string = "call not exists client"
)

// event
type EventInfo struct {
	event		int
	value		interface {}
}

type EventServiceChange struct {
	changes		[]naming.ServiceChange
}

type EventClientChange struct {
	event		int
	service_id	string
	host		string
	port		int
}

// call packet
type RespPacketCb func(*mservice_define.MServicePacket, error)

type RespPacketResult struct {
	packet			*mservice_define.MServicePacket
	err				error
}

type CallRecord struct {
	call_type			int				// 调用方式：kCall, kCallAsync, kCast, kBroadcast
	hash				uint32			// 通过hash定位调用节点
	req					*mservice_define.MServicePacket
	async_cb			RespPacketCb	// 异步方式回调函数
	channel_cb			chan *RespPacketResult
	client_service_id	string
	client_host			string
	client_port			int
	send_ts				int64
}

//////////////////////////////////////////////////////////
//					ObjectProxy
//////////////////////////////////////////////////////////
// 名字服务在etcd的存储格式
// 目录: /mservice/{服务名}/{服务名}_{host}:{port}_{seq}
type ClientStatus struct {
	retry_connect_ts		int64
//	success_count			int
//	fail_count				int
	series_fail_count		int
}

type ClientData struct {
	service_id			string
	status				ClientStatus
}

type ObjectProxy struct {
	name				string
	lru_reqs			*util.LruCache[string, *CallRecord]	// 在途请求
	active_clients		[]*util.TcpClient					// 活跃节点列表
	inactive_clients	[]*util.TcpClient					// 非活跃节点列表
	channel_event		chan *EventInfo
	channel_send		chan *CallRecord
	handle				net_core.PacketHandle
	cfg					mservice_define.ObjClientConfig
	sync.Mutex
}

func NewObjectProxy(name string, handle net_core.PacketHandle) *ObjectProxy {
	cfg := util.GetConfigValue[mservice_define.MServiceConfig]().ObjClientCfg
	proxy := &ObjectProxy {
		name: 				name,
		lru_reqs:			util.NewLruCache[string, *CallRecord](cfg.MaxCacheSize, nil),
		active_clients:		make([]*util.TcpClient, 0),
		inactive_clients:	make([]*util.TcpClient, 0),
		channel_event:		make(chan *EventInfo, 1000),
		channel_send:		make(chan *CallRecord, 1000),
		handle:				handle,
		cfg:				cfg,
	}
	go proxy.init()
	return proxy
}

// 暴力关闭
func (p *ObjectProxy) Close() {
	for _, x := range p.active_clients	{
		x.Close()
	}
	close(p.channel_event)
	close(p.channel_send)
}

func (p *ObjectProxy) Call(hash uint32, req *mservice_define.MServicePacket) (*mservice_define.MServicePacket, error) {
	record := &CallRecord {
		call_type:		kCall,
		hash:			hash,
		req:			req,
		channel_cb:		make(chan *RespPacketResult),
		send_ts:		time.Now().Unix(),
	}

	p.add_to_call_queue(record)
	p.channel_send <-record

	var err error
	var res *mservice_define.MServicePacket = nil
	cfg := util.GetConfigValue[mservice_define.MServiceConfig]().ObjClientCfg
	select {
	case <-time.After(time.Duration(cfg.CallTimeoutSec) * time.Second):
		err = errors.New(kErrorCallTimeout)
	case result, ok := <-record.channel_cb:
		if ok {
			err = result.err
			res = result.packet
		} else {
			err = errors.New(kErrorCallException)
		}
	}
	return res, err
}

//func (p *ObjectProxy) Call(hash uint32, service string, method string, req interface {}) (resp interface {}, err error) {
//	req_packet := mservice_define.CreateReqPacket(
//		service, 
//		method,
//		util.GenUuid(),
//		mservice_define.MS_CALL,
//		mservice_define.MS_REQUEST,
//		body,
//	)
//
//	record := &CallRecord {
//		call_type:		kCall,
//		hash:			hash,
//		req:			req_packet,
//		channel_cb:		make(chan *RespPacketResult),
//		send_ts:		time.Now().Unix(),
//	}
//
//	p.add_to_call_queue(record)
//	p.channel_send <-record
//
//	var err error
//	var res *mservice_define.MServicePacket = nil
//	cfg := util.GetConfigValue[mservice_define.MServiceConfig]().ObjClientCfg
//	select {
//	case <-time.After(time.Duration(cfg.CallTimeoutSec) * time.Second):
//		err = errors.New(kErrorCallTimeout)
//	case result, ok := <-record.channel_cb:
//		if ok {
//			err = result.err
//			res = result.packet
//		} else {
//			err = errors.New(kErrorCallException)
//		}
//	}
//	return res, err
//}

func (p *ObjectProxy) CallAsync(hash uint32, req *mservice_define.MServicePacket, cb RespPacketCb) {
	record := &CallRecord {
		call_type:		kCallAsync,
		hash:			hash,
		req:			req,
		channel_cb:		make(chan *RespPacketResult),
		send_ts:		time.Now().Unix(),
	}
	p.add_to_call_queue(record)
}

//func (p *ObjectProxy) CallAsync(hash uint32, service string, method string, req interface {}, cb RespPacketCb) {
//	req_packet := mservice_define.CreateReqPacket(
//		service, 
//		method,
//		util.GenUuid(),
//		mservice_define.MS_CALL,
//		mservice_define.MS_REQUEST,
//		body,
//	)
//	record := &CallRecord {
//		call_type:		kCallAsync,
//		hash:			hash,
//		req:			req,
//		channel_cb:		make(chan *RespPacketResult),
//		send_ts:		time.Now().Unix(),
//	}
//	p.add_to_call_queue(record)
//}

func (p *ObjectProxy) Cast(hash uint32, req *mservice_define.MServicePacket) {
	record := &CallRecord {
		call_type:		kCast,
		hash:			hash,
		req:			req, 
		send_ts:		time.Now().Unix(),
	}
	p.add_to_call_queue(record)
}

//func (p *ObjectProxy) Cast(hash uint32, service string, method string, req interface {}) {
//	req_packet := mservice_define.CreateReqPacket(
//		service, 
//		method,
//		util.GenUuid(),
//		mservice_define.MS_CAST,
//		mservice_define.MS_REQUEST,
//		body,
//	)
//	record := &CallRecord {
//		call_type:		kCast,
//		hash:			hash,
//		req:			req,
//		send_ts:		time.Now().Unix(),
//	}
//	p.add_to_call_queue(record)
//}

func (p *ObjectProxy) new_client(host string, port int, service_id string) *util.TcpClient {
	client := util.NewTcpClient(
		host, 
		port, 
		p.handle,
		p.on_client_recv,
		p.on_client_close,
		ClientData {service_id: service_id},
	)
	return client 
}

func (p *ObjectProxy) add_to_call_queue(record *CallRecord) {
	p.Lock()
	defer p.Unlock()
	p.lru_reqs.Set(record.req.Header.TraceId, record)
}

func (p *ObjectProxy) notify_service_change(changes []naming.ServiceChange) {
	p.channel_event <-&EventInfo {
		event:		kEventServiceChange,
		value:		&EventServiceChange {changes: changes},
	}
}

////////////////////////////////////////////////////////////////
//				loop_main协程内处理，无竞争无需加锁
////////////////////////////////////////////////////////////////
func (p *ObjectProxy) loop_main() {
	p.init()
	timer := time.NewTicker(1 * time.Second)
	for {
		select {
		case <-timer.C:
			p.retry_client_connect()
		case record, ok := <-p.channel_send:
			if !ok {
				break
			}
			p.do_send(record)
		case ev, ok := <-p.channel_event:
			if !ok {
				break
			}
			p.do_event(ev)
		}
	}
}

func (p *ObjectProxy) init() {
	if err := naming.GetInstance().Subscribe(p.name, p.notify_service_change); err != nil {
		err_msg := fmt.Sprintf("ObjectProxy|naming.Subscribe name:%s error:%#v", p.name, err)
		panic(err_msg)
	}

	infos, err := naming.GetInstance().Find(p.name)
	if err != nil {
		naming.GetInstance().Unsubscribe(p.name)
		err_msg := fmt.Sprintf("ObjectProxy|naming.Find name:%s error:%#v", p.name, err)
		panic(err_msg)
	}

	for _, info := range infos {
		client := p.new_client(info.Host, info.Port, info.ServiceId)
		if err := client.Connect(500); err != nil {
			p.inactive_clients = append(p.inactive_clients, client)
		} else {
			p.active_clients = append(p.active_clients, client)
		}
	}
}

func (p *ObjectProxy) retry_client_connect() {
	now := time.Now().Unix()
	for {
		connect_ok := false
		size := len(p.inactive_clients)
		for i := 0; i < size; i++ {
			client := p.inactive_clients[i]
			cdata := (client.Param).(ClientData)
			if now < cdata.status.retry_connect_ts {
				continue
			}
			if err := client.Connect(100); err == nil {
				connect_ok = true
				cdata := (client.Param).(ClientData)
				cdata.status.retry_connect_ts = 0
				cdata.status.series_fail_count = 0
				p.active_clients = append(p.active_clients, client)
				p.inactive_clients = append(p.inactive_clients[0:i], p.inactive_clients[i+1:]...)
				break
			} else {
				log.Printf("ERROR|obj_proxy, client[%s:%d] reconnect error:%#v", client.Host, client.Port, err)
			}
		}
		if !connect_ok {
			break
		}
	}
	sort_clients(p.active_clients)
}

func (p *ObjectProxy) do_send(record *CallRecord) {
	index := int(record.hash % uint32(len(p.active_clients)))
	client := p.active_clients[index]
	if err := client.Send(record.req); err != nil {
		p.remove_active_client_by_index(index)
	}
}

func (p *ObjectProxy) do_event(ev *EventInfo) {
	if ev.event == kEventServiceChange {
		p.do_event_service_change(ev.value.(*EventServiceChange))
	} else if ev.event == kEventClientChange {
		p.do_event_client_change(ev.value.(*EventClientChange))	
	}
}

func (p *ObjectProxy) do_event_service_change(change *EventServiceChange) {
	for _, x := range change.changes {
		if x.Op == naming.ServicePut {
			is_new_client := false
			client := p.find_client(x.Service.Host, x.Service.Port, x.Service.ServiceId)
			if client == nil {
				is_new_client = true
			} else {
				cdata := (client.Param).(ClientData)
				if cdata.service_id != x.Service.ServiceId {
					is_new_client = true
					p.do_client_remove(x.Service.Host, x.Service.Port, x.Service.ServiceId)
				}
			}
			if is_new_client {
				new_client := p.new_client(x.Service.Host, x.Service.Port, x.Service.ServiceId)
				cdata := (new_client.Param).(ClientData)
				cdata.status.retry_connect_ts = time.Now().Unix()
				p.inactive_clients = append(p.inactive_clients, new_client)
			}
		} else {
			p.do_client_remove(x.Service.Host, x.Service.Port, x.Service.ServiceId)
		}
	}
}

func (p *ObjectProxy) do_event_client_change(change *EventClientChange) {
	if change.event == kClientExit {
		p.do_client_exit(change)
	} else if change.event == kClientClose {
		p.do_client_close(change)
	} else if change.event == kClientTimeout {
		p.do_client_timeout(change)
	}
}

func (p *ObjectProxy) do_client_exit(change *EventClientChange) {
	p.do_client_remove(change.host, change.port, change.service_id)
}

func (p *ObjectProxy) do_client_close(change *EventClientChange) {
	if index := p.find_client_index(p.active_clients, change.host, change.port, change.service_id); index >= 0 {
		client := p.active_clients[index]
		p.inactive_clients = append(p.inactive_clients, client)
	}
}

func (p *ObjectProxy) do_client_timeout(change *EventClientChange) {
	if index := p.find_client_index(p.active_clients, change.host, change.port, change.service_id); index >= 0 {
		client := p.active_clients[index]
		cdata := (client.Param).(ClientData)
	//	cdata.status.fail_count++
		cdata.status.series_fail_count++
		if cdata.status.series_fail_count >= p.cfg.InvalidSeriesFailCount {
			p.remove_active_client_by_index(index)
		}
	}
}

func (p *ObjectProxy) do_client_remove(host string, port int, service_id string) {
	if index := p.find_client_index(p.active_clients, host, port, service_id); index >= 0 {
		p.active_clients[index].Close()
		p.active_clients = append(p.active_clients[0:index], p.active_clients[index+1:]...)
		return
	}
	if index := p.find_client_index(p.inactive_clients, host, port, service_id); index >= 0 {
		p.inactive_clients = append(p.inactive_clients[0:index], p.inactive_clients[index+1:]...)
	}
}

func (p *ObjectProxy) remove_active_client_by_index(index int) bool {
	if len(p.active_clients) >= index {
		return false
	}

	client := p.active_clients[index]
	client.Close()

	cdata := (client.Param).(ClientData)
	cdata.status.series_fail_count = 0
	cdata.status.retry_connect_ts = time.Now().Unix() + int64(p.cfg.RetryConnectSec)

	p.active_clients = append(p.active_clients[0:index], p.active_clients[index+1:]...)
	p.inactive_clients = append(p.inactive_clients, client)
	sort_clients(p.active_clients)

	return true
}

func (p *ObjectProxy) find_client_index(clients []*util.TcpClient, host string, port int, service_id string) int {
	for index, client := range clients {
		if client.Host == host && client.Port == port {
			cdata := (client.Param).(ClientData)
			if service_id != "" && service_id != cdata.service_id {
				break
			}
			return index
		}	
	}
	return -1
}

func (p *ObjectProxy) find_client(host string, port int, service_id string) *util.TcpClient {
	if index := p.find_client_index(p.active_clients, host, port, service_id); index >= 0 {
		return p.active_clients[index]
	}
	if index := p.find_client_index(p.inactive_clients, host, port, service_id); index >= 0 {
		return p.inactive_clients[index]
	}
	return nil
}

/////////////////////////////////////////////////////////////
//					TcpClient回调函数
/////////////////////////////////////////////////////////////
func (p *ObjectProxy) on_client_recv(client *util.TcpClient, packet net_core.Packet) error {
	mpacket := packet.(*mservice_define.MServicePacket)

	var exists bool
	var record *CallRecord
	p.Lock()
	if record, exists = p.lru_reqs.Get(mpacket.Header.TraceId); !exists {
		return nil
	}
	p.lru_reqs.Remove(mpacket.Header.TraceId)
	p.Unlock()

	if record.call_type == kCall {
		record.channel_cb <-&RespPacketResult {
			packet:	mpacket,
			err:	nil,
		}
	} else if record.call_type == kCallAsync {
		record.async_cb(mpacket, nil)
	} else {
		return nil
	}

	cdata := (client.Param).(ClientData)
	if !exists || time.Now().Unix() - record.send_ts >= int64(p.cfg.CallTimeoutSec) {
		p.channel_event <-&EventInfo {
			event:		kEventClientChange,
			value:		&EventClientChange {
				event:		kClientTimeout,
				service_id:	cdata.service_id,
				host:		client.Host,
				port:		client.Port,
			},
		}
	}
	return nil
}

func (p *ObjectProxy) on_client_close(client *util.TcpClient, is_active_close bool , err error) {
	cdata := (client.Param).(ClientData)
	p.channel_event <-&EventInfo {
		event:		kEventClientChange,
		value:		&EventClientChange {
			event:		kClientClose,
			service_id:	cdata.service_id,
			host:		client.Host,
			port:		client.Port,
		},
	}
}

////////////////////////////////////////////////////////
//					工具函数
////////////////////////////////////////////////////////
func sort_clients(clients []*util.TcpClient) {
	sort.SliceStable(clients, func(i, j int) bool {
		addr1 := fmt.Sprintf("%s:%d", clients[i].Host, clients[i].Port)
		addr2 := fmt.Sprintf("%s:%d", clients[j].Host, clients[j].Port)
		if addr1 < addr2 {
			return true
		}
		return false
	})
}
