package pusher_handle

import (
	"sync"
	"github.com/segmentio/ksuid"
	"hbservice/src/util"
	"hbservice/src/mservice"
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
	client					*TcpClient
}

type PusherHandle struct {
	cfg						pusher_define.PusherConfig						
	user_caches				util.LruCache[string, *user_info]
	get_addr_search			util.MergeSearch[string, *UserInfo]

	clients					map[string]*gw_client

	channel_push_small		chan *mservice_define.MServicePacket
	channel_push_large		chan *mservice_define.MServicePacket

	sync.Mutex
}

func (p *PusherHandle) Init(server net_core.NetServer) error {
	util.SetConfigByFileLoader[pusher_define.PusherConfig](
		mservice_util.GetEtcdInstance(),
		mservice_define.SERVICE_ETCD_CONFIG_PATH + "/pusher.config"
	)
	p.cfg := util.GetConfigValue[pusher_define.PusherConfig]()

	p.user_caches = util.NewLruCache[string, *UserInfo]
	p.get_addr_search = util.NewMergeSearch[string, string](p.get_remote_user_addr)

	p.channel_push_small = make(chan *service_define.MServicePacket, cfg.PushChannelCount)
	p.channel_push_large = make(chan *service_define.MServicePacket, cfg.PushChannelCount)

	if err := p.init_gateway_clients(); err != nil {
		return err
	}
	naming.GetInstance().Subscribe("gateway", p.service_change)

	go func() {
		timer := time.NewTricker(time.Duration(cfg.RetryClientConnectTs) * time.Second)
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
	req := (packet.Body).(gateway_define.MIPushReq)
	size := len(req.Uids)
	if size >= cfg.BroadcastUidCount {
		p.channel_push_large <-packet
	} else {
		p.channel_push_small <-packet
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
	infos, err := naming.Find("gateway")
	if err != nil {
		return err
	}
	for _, info := range infos {
		client := util.NewTcpClient(
			info.Host, 
			info.Port, 
			mservice_define.MPacketHandle{}, 
			nil, 
			nil
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
		p.clients = append(p.clients, gclient)
	}
	return nil
}

func (p *PusherHandle) retry_client_connect() {
	var retry_clients []*gw_client
	p.Lock()
	for _, gclient := range p.clients {
		if gclient.client.GetState() == util.stClose {
			retry_clients = append(retry_clients, gclient)
		}
	}
	p.Unlock()

	for _, gclient := range retry_clients {
		gclient.client.Connect(500)
	}
}

func (p *PusherHandle) service_change(changes []naming.ServiceChange) {
	var new_clients []*TcpClient
	p.Lock()
	for _, change := range changes {
		addr := fmt.Sprintf("%s:%d", change.Host, change.Port)
		if change.Op == naming.kServicePut {
			if client, ok := p.clients[addr]; ok {
				if change.ServiceId != client.gateway_id {
					client.Close()
					gclient := &gw_client {
						gateway_id:		info.ServiceId,
						host:			info.Host,
						port:			info.Port,
					}
					new_clients = append(new_clients, gclient)
				}
			} else {

			}
		} else {
			if client, ok := p.clients[addr]; ok {
				if change.ServiceId == client.gateway_id {
					client.Close()
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
			mservice_define.MPacketHandle{}, 
			nil, 
			nil
		)
		client.Connect(500)
		gclient.client = client
	}

	p.Lock()
	p.clients = append(p.clients, new_clients...)
	p.Unlock()
}

func (p *PusherHandle) do_pusher_small() {
	for packet := range p.channel_push_small {
		req := (packet.Body).(*MIPushReq)
		uids := req.Uids
		for _, uid := range Uids {
			addr, err := p.get_uid_addr(uid)
			if err != nil {
				continue
			}
			req.Uids = []string{uid}
			if client, ok := p.gw_clients[addr]; ok {
				client.Send(packet)
			}
		}
	}
}

func (p *PusherHandle) do_pusher_large() {
	for req := range p.channel_push_large {
		for _, c := range p.get_active_clients() {
			c.Send(packet)
		}
	}
}

func (p *PusherHandle) get_client(addr string) (*TcpClient, bool) {
	p.Lock()
	defer p.Unlock()
	client, ok := p.gw_clients[addr]
	if ok && client.GetState() {
		return client, true
	}
	return nil, false
}

func (p *PusherHandle) get_active_clients() (clients []*TcpClient) {
	p.Lock()
	defer p.Unlock()
	for _, c := range p.gw_clients {
		if c.GetState() == util.stConnected {
			clients = append(clients, c)
		}
	}
	return clients
}

func (p *PusherHandle) get_uid_addr(uid string) (string, error) {
	p.Lock()
	if user, ok := user_caches.Get(uid); ok {
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
	user := &UserInfo {
		addr:		addr,
		expire_ts:	time.Now().Unix() + p.cfg.UserCacheTs,
	}
	p.Unlock()
	return 
}

func (p *PusherHandle) get_remote_user_addr(uid string) (string, error) {
	req := &mservice_define.MServicePacket {
		Header:	mservice_define.MServiceHeader {
			Service:	online_define.SERVICE_NAME,
			Mechod:		online_define.OL_GET_UID_ADDR,
			Direction:	mservice_define.DIRECT_REQUEST,
		},
		Body:	&online_define.OLGetUidAddrReq {Uid: uid},
	}	
	resp, err := Container.GetInstance().Call(util.GenHash(uid), req)
	if err != nil {
		return "", err
	}
	body := mservice_define.GetPacketBody[online_define.OLGetUidAddrResp](resp)
	if body.Err != nil {
		return "", body.Err
	}
	return body.Addr, nil
}
