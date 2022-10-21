package online_handle

import (
	"sync"
	"github.com/segmentio/ksuid"
	"hbservice/src/util"
	"hbservice/src/mservice/define"
)

type OnlineHandle struct {
	cfg				*OnlineConfig
	user_cache		*UserNodeCache
	sync.RWMutex
}

func (p *OnlineHandle) Init(server net_core.NetServer) error {
	err := util.SetConfigByEtcdLoader[OnlineConfig](
		mservice_util.GetEtcdInstance(),
		mservice_define.SERVICE_ETCD_CONFIG_PATH + "/online.config"
	)
	if err != nil {
		return err
	}
	cfg := util.GetConfigValue[online_define.OnlineConfig]()
	p.user_cache = NewUserNodeCache(cfg.MaxCacheSize)
	return nil
}

func (p *OnlineHandle) OnAccept(conn net.Conn, handle net_core.PacketHandle, server net_core.NetServer) (string, error) {
	return ksuid.New().String(), nil
}

func (p *OnlineHandle) OnMessage(channel_id string, packet net_core.Packet, server net_core.NetServer) error {
	var err error
	var resp_packet *MServicePacket
	req_packet := packet.(*MServicePacket)
	if mpacket.Header.Method == online_define.MS_ONLINE_HEARTBEAT {
		resp_packet, err = p.do_heartbeat(packet)
	} else if mpacket.Header.Method == online_define.MS_ONLINE_GET_USER_NODE {
		resp_packet, err = p.do_get_user_node(packet)
	} else if mpacket.Header.Method == online_define.MS_ONLINE_GET_USER_NODE_BATCH {
		resp_packet, err = p.do_get_user_node_batch(packet)
	}
	if resp_packet != nil {
		server.PushChannel(channel_id, resp_packet)
	}
	return err
}

func (p *OnlineHandle) OnClose(channel_id string, server net_core.NetServer) error {
	return nil
}

func (p *OnlineHandle) OnTimer(timer_id string, value interface {}, server net_core.NetServer) {
	return
}

func (p *OnlineHandle) do_heartbeat(req_packet *mservice_define.MServicePacket) (*MServicePacket, error) {
	var err_msg string
	body := req_packet.Body.(*online_define.OLHeartbeatReq)
	p.WLock()
	user_node := &UserNodeInfo {
		UserNode:	body.User,
		ExpireTs:	time.Now().Unix() + p.cfg.CacheExpireSec,
	}
	if err = p.user_cache.Set(uid, user_node); err != nil {
		err_msg = err.Error()
	}
	p.Unlock()

	resp_packet := mservice_define.CreateReqPacket(
		online_define.SERVICE_NAME,
		online_define.MS_ONLINE_HEARTBEAT,
		req_packet.Header.TraceId,
		req_packet.Header.CallType,
		mservice_define.MS_RESPONSE,
		&online_define.OLHeartbeatResp {Err: err_msg},
	)	
	return resp_packet
}

func (p *OnlineHandle) do_get_user_node(req_packet *mservice_define.MServicePacket) (*MServicePacket, error) {
	var err_msg	string
	var resp_user online_define.UserNode
	p.RLock()
	user, ok := p.user_cache.Get(uid)
	if !ok || user.ExpireTs > time.Now().Unix() {
		resp_err = errors.New(online_define.ERR_NOT_EXIST_USER)
	} else {
		resp_user = user.UserNode
	}
	p.Unlock()

	resp := &online_define.OLGetUserNodeResp {
		Err:	err_msg,
		User:	resp_user,
	}
	resp_packet := mservice_define.CreateRespPacket(
		online_define.SERVICE_NAME,
		online_define.MS_ONLINE_GET_USER_NODE,
		req_packet.Header.TraceId,
		req_packet.Header.CallType,
		mservice_define.MS_RESPONSE,
		&resp
	)
	return resp_packet
}

func (p *OnlineHandle) do_get_user_node_batch(req_packet *mservice_define.MServicePacket) (*MServicePacket, error) {
	var resp_users []online_define.UserNode
	p.RLock()
	now := time.Now().Unix()
	req := req_packet.Body.(*OLGetUserNodeBatchReq)
	for _, uid := range req.Uids {
		user, ok := p.user_cache.Get(uid)
		if ok && user.ExpireTs > now {
			resp_users = append(resp_users, user.UserNode)
		}
	}
	p.Unlock()

	resp := &online_define.OLGetUserNodeResp {
		Err:	"",
		Users:	resp_users,
	}
	resp_packet := mservice_define.CreateRespPacket(
		online_define.SERVICE_NAME,
		online_define.MS_ONLINE_GET_USER_NODE_BATCH,
		req_packet.Header.TraceId,
		req_packet.Header.CallType,
		mservice_define.MS_RESPONSE,
		&resp,
	)
	return resp_packet
}
