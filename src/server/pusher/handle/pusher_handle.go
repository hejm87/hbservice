package pusher_handle

import (
	"fmt"
	"log"
	"net"
	"time"
	"sync"
	"errors"
	"github.com/segmentio/ksuid"
	"hbservice/src/util"
	"hbservice/src/mservice"
	"hbservice/src/mservice/define"
	"hbservice/src/mservice/naming"
	"hbservice/src/net/net_core"
	"hbservice/src/server/gateway/define"
	"hbservice/src/server/online/define"
	"hbservice/src/server/pusher/define"
)

type user_info struct {
	addr			string
	expire_ts		int64		// 缓存过期时间
}

type gw_client struct {
	gateway_id				string
	host					string
	port					int
	latest_disconnect_ts	int64
	client					*util.TcpClient
}

type PusherHandle struct {
	cfg						*pusher_define.PusherConfig	
	user_caches				*util.LruCache[string, *user_info]
	get_addr_search			*util.MergeSearch[string, string]

	clients					map[string]*gw_client

	channel_push_small		chan *mservice_define.MServicePacket
	channel_push_large		chan *mservice_define.MServicePacket

	sync.Mutex
}

func (p *PusherHandle) Init(server net_core.NetServer) error {
	conf_path := mservice_define.SERVICE_ETCD_CONFIG_PATH + "/pusher.config"
	err := util.SetConfigByFileLoader[pusher_define.PusherConfig](conf_path)
	if err != nil {
		return err
	}
	p.cfg = util.GetConfigValue[pusher_define.PusherConfig]()

	p.user_caches = util.NewLruCache[string, *user_info](p.cfg.UserCacheSize, nil)
	p.get_addr_search = util.NewMergeSearch[string, string](p.get_remote_user_addr)

	p.channel_push_small = make(chan *mservice_define.MServicePacket, p.cfg.PushChannelCount)
	p.channel_push_large = make(chan *mservice_define.MServicePacket, p.cfg.PushChannelCount)

	if err := p.init_gateway_clients(); err != nil {
		return err
	}
	naming.GetInstance().Subscribe("gateway", p.service_change)

	go func() {
		timer := time.NewTicker(time.Duration(p.cfg.RetryClientConnectTs) * time.Second)
		for {
			select {
			case <-timer.C:
				p.retry_client_connect()
			}
		}
	} ()

	return nil
}

func (p *PusherHandle) OnAccept(conn net.Conn, handle net_core.PacketHandle, server net_core.NetServer) (string, error) {
	return ksuid.New().String(), nil
}

func (p *PusherHandle) OnMessage(channel_id string, packet net_core.Packet, server net_core.NetServer) error {
	mpacket := packet.(*mservice_define.MServicePacket)
	req := (mpacket.Body).(*gateway_define.MIPushReq)
	size := len(req.Uids)
	if size >= p.cfg.BroadcastUidCount {
		p.channel_push_large <-mpacket
	} else {
		p.channel_push_small <-mpacket
	}
	return nil
}

func (p *PusherHandle) OnClose(channel_id string, server net_core.NetServer) error {
	return nil
}

func (p *PusherHandle) OnTimer(timer_id string, value interface {}, server net_core.NetServer) {
	return
}

func (p *PusherHandle) init_gateway_clients() error {
	infos, err := naming.GetInstance().Find("gateway")
	if err != nil {
		return err
	}
	for _, info := range infos {
		client := util.NewTcpClient(
			info.Host, 
			info.Port, 
			&mservice_define.MPacketHandle{}, 
			nil, 
			nil,
			nil,
		)
		if err := client.Connect(500); err != nil {
			log.Printf("ERROR|connect gateway[%s:%d] error:%#v", info.Host, info.Port, err)
		}
		gclient := &gw_client {
			gateway_id:		info.ServiceId,
			host:			info.Host,
			port:			info.Port,
			client:			client,
		}
		addr := fmt.Sprintf("%s:%d", info.Host, info.Port)
		p.clients[addr] = gclient
	}
	return nil
}

func (p *PusherHandle) retry_client_connect() {
	var retry_clients []*gw_client
	p.Lock()
	for _, gclient := range p.clients {
		if gclient.client.GetState() == util.StateClose {
			retry_clients = append(retry_clients, gclient)
		}
	}
	p.Unlock()

	for _, gclient := range retry_clients {
		gclient.client.Connect(500)
	}
}

func (p *PusherHandle) service_change(changes []naming.ServiceChange) {
	var new_clients []*gw_client
	p.Lock()
	for _, change := range changes {
		addr := fmt.Sprintf("%s:%d", change.Service.Host, change.Service.Port)
		if change.Op == naming.ServicePut {
			if gclient, ok := p.clients[addr]; ok {
				if change.Service.ServiceId != gclient.gateway_id {
					gclient.client.Close()
					new_gclient := &gw_client {
						gateway_id:		change.Service.ServiceId,
						host:			change.Service.Host,
						port:			change.Service.Port,
					}
					new_clients = append(new_clients, new_gclient)
				}
			} else {

			}
		} else {
			if gclient, ok := p.clients[addr]; ok {
				if change.Service.ServiceId == gclient.gateway_id {
					gclient.client.Close()
					delete(p.clients, addr)
				}
			}
		}
	}
	defer p.Unlock()

	for _, gclient := range new_clients {
		client := util.NewTcpClient(
			gclient.host,
			gclient.port,
			&mservice_define.MPacketHandle{}, 
			nil, 
			nil,
			nil,
		)
		client.Connect(500)
		gclient.client = client
	}

	p.Lock()
	for _, gclient := range new_clients {
		addr := fmt.Sprintf("%s:%d", gclient.host, gclient.port)
		p.clients[addr] = gclient
	}
	p.Unlock()
}

func (p *PusherHandle) do_pusher_small() {
	for packet := range p.channel_push_small {
		req := (packet.Body).(*gateway_define.MIPushReq)
		uids := req.Uids
		for _, uid := range uids {
			addr, err := p.get_uid_addr(uid)
			if err != nil {
				continue
			}
			req.Uids = []string{uid}
			if gclient, ok := p.clients[addr]; ok {
				gclient.client.Send(packet)
			}
		}
	}
}

func (p *PusherHandle) do_pusher_large() {
	for req := range p.channel_push_large {
		for _, gclient := range p.get_active_clients() {
			gclient.client.Send(req)
		}
	}
}

func (p *PusherHandle) get_active_clients() (clients []*gw_client) {
	p.Lock()
	defer p.Unlock()
	for _, gclient := range p.clients {
		if gclient.client.GetState() == util.StateConnected {
			clients = append(clients, gclient)
		}
	}
	return clients
}

func (p *PusherHandle) get_uid_addr(uid string) (string, error) {
	p.Lock()
	if user, ok := p.user_caches.Get(uid); ok {
		if user.expire_ts > time.Now().Unix() {
			return user.addr, nil
		}
	}
	p.Unlock()
	addr, err := p.get_addr_search.Call(uid)
	if err != nil {
		return "", err
	}
	p.Lock()
	user := &user_info {
		addr:		addr,
		expire_ts:	time.Now().Unix() + int64(p.cfg.UserCacheTs),
	}
	p.user_caches.Set(uid, user)
	p.Unlock()
	return addr, nil
}

func (p *PusherHandle) get_remote_user_addr(uid string) (string, error) {
	req := mservice_define.CreateReqPacket(
		online_define.SERVICE_NAME,
		online_define.MS_ONLINE_GET_USER_NODE,
		mservice_define.MS_CALL,
		&online_define.OLGetUserNodeReq {Uid: uid},
	)
	resp, err := container.CallByPacket(util.GenHash(uid), req)
	if err != nil {
		return "", err
	}
	body := resp.Body.(*online_define.OLGetUserNodeResp)
	if body.Err != "" {
		return "", errors.New(body.Err)
	}
	addr := fmt.Sprintf("%s:%d", body.User.Host, body.User.Port)
	return addr, nil
}
