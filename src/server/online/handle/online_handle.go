package online_handle

import (
	"log"
	"net"
	"time"
	"sync"
	"github.com/segmentio/ksuid"
	"hbservice/src/util"
	"hbservice/src/net/net_core"
	"hbservice/src/mservice/define"
	"hbservice/src/server/online/define"
)

type OnlineHandle struct {
	user_cache		*UserNodeCache
	sync.RWMutex
}

func (p *OnlineHandle) Init(server net_core.NetServer) error {
	cfg := util.GetConfigValue[online_define.OnlineConfig]()
	p.user_cache = NewUserNodeCache(cfg.MaxCacheSize)
	return nil
}

func (p *OnlineHandle) OnAccept(conn net.Conn, handle net_core.PacketHandle, server net_core.NetServer) (string, error) {
	return ksuid.New().String(), nil
}

func (p *OnlineHandle) OnMessage(channel_id string, packet net_core.Packet, server net_core.NetServer) error {
	var err error
	var resp_packet *mservice_define.MServicePacket
	req_packet := packet.(*mservice_define.MServicePacket)
	if req_packet.Header.Method == online_define.MS_ONLINE_LOGIN {
		resp_packet, err = p.do_login(req_packet)
	} else if req_packet.Header.Method == online_define.MS_ONLINE_LOGOUT {
		resp_packet, err = p.do_logout(req_packet)
	} else if req_packet.Header.Method == online_define.MS_ONLINE_HEARTBEAT {
		resp_packet, err = p.do_heartbeat(req_packet)
	} else if req_packet.Header.Method == online_define.MS_ONLINE_GET_USER_NODE {
		resp_packet, err = p.do_get_user_node(req_packet)
	} else if req_packet.Header.Method == online_define.MS_ONLINE_GET_USER_NODE_BATCH {
		resp_packet, err = p.do_get_user_node_batch(req_packet)
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

func (p *OnlineHandle) do_login(req_packet *mservice_define.MServicePacket) (*mservice_define.MServicePacket, error) {
	body, err := mservice_define.GetPacketBody[online_define.OLLoginReq](req_packet)
	if err != nil {
		return nil, err
	}
	var err_msg string
	p.Lock()
	if _, ok := p.get_user(body.User.Uid); !ok {
		if err := p.set_user(&body.User); err != nil {
			err_msg = err.Error()
		}
	} else {
		err_msg = online_define.ERR_USER_ALREADY_LOGIN
	}
	p.Unlock()

	resp_packet := mservice_define.CreateRespPacket(
		online_define.SERVICE_NAME,
		online_define.MS_ONLINE_LOGIN,
		req_packet.Header.TraceId,
		req_packet.Header.CallType,
		&online_define.OLLoginResp {Err: err_msg},
	)
	return resp_packet, nil
}

func (p *OnlineHandle) do_logout(req_packet *mservice_define.MServicePacket) (*mservice_define.MServicePacket, error) {
	body, err := mservice_define.GetPacketBody[online_define.OLLogoutReq](req_packet)
	if err != nil {
		return nil, err
	}
	var err_msg string
	p.Lock()
	if _, ok := p.get_user(body.Uid); ok {
		p.del_user(body.Uid)
	} else {
		err_msg = online_define.ERR_USER_NOT_EXISTS
	}
	p.Unlock()

	resp_packet := mservice_define.CreateRespPacket(
		online_define.SERVICE_NAME,
		online_define.MS_ONLINE_LOGOUT,
		req_packet.Header.TraceId,
		req_packet.Header.CallType,
		&online_define.OLLoginResp {Err: err_msg},
	)
	return resp_packet, nil
}

func (p *OnlineHandle) do_heartbeat(req_packet *mservice_define.MServicePacket) (*mservice_define.MServicePacket, error) {
	var err_msg string
	body, err := mservice_define.GetPacketBody[online_define.OLHeartbeatReq](req_packet)
	if err != nil {
		return nil, err
	}
	p.Lock()
	if err := p.set_user(&body.User); err != nil {
		err_msg = err.Error()
	}
	p.Unlock()

	resp_packet := mservice_define.CreateRespPacket(
		online_define.SERVICE_NAME,
		online_define.MS_ONLINE_HEARTBEAT,
		req_packet.Header.TraceId,
		req_packet.Header.CallType,
		&online_define.OLHeartbeatResp {Err: err_msg},
	)	
	return resp_packet, nil
}

func (p *OnlineHandle) do_get_user_node(req_packet *mservice_define.MServicePacket) (*mservice_define.MServicePacket, error) {
	var err_msg string
	var ok bool
	var user *online_define.UserNode
	body, err := mservice_define.GetPacketBody[online_define.OLGetUserNodeReq](req_packet)
	if err != nil {
		return nil, err
	}
	p.RLock()
	user, ok = p.get_user(body.Uid)
	if !ok {
		err_msg = online_define.ERR_USER_NOT_EXISTS
	}
	p.Unlock()

	resp := &online_define.OLGetUserNodeResp {
		Err:	err_msg,
		User:	*user,
	}
	resp_packet := mservice_define.CreateRespPacket(
		online_define.SERVICE_NAME,
		online_define.MS_ONLINE_GET_USER_NODE,
		req_packet.Header.TraceId,
		req_packet.Header.CallType,
		&resp,
	)
	return resp_packet, nil
}

func (p *OnlineHandle) do_get_user_node_batch(req_packet *mservice_define.MServicePacket) (*mservice_define.MServicePacket, error) {
	var resp_users []online_define.UserNode
	body, err := mservice_define.GetPacketBody[online_define.OLGetUserNodeBatchReq](req_packet)
	if err != nil {
		return nil, err
	}
	p.RLock()
	for _, uid := range body.Uids {
		if user, ok := p.get_user(uid); ok {
			resp_users = append(resp_users, *user)
		}
	}
	p.Unlock()

	resp := &online_define.OLGetUserNodeBatchResp{Users: resp_users}
	resp_packet := mservice_define.CreateRespPacket(
		online_define.SERVICE_NAME,
		online_define.MS_ONLINE_GET_USER_NODE_BATCH,
		req_packet.Header.TraceId,
		req_packet.Header.CallType,
		&resp,
	)
	return resp_packet, nil
}

//////////////////////////////////////////////////////////////
//					后续改成读/写redis
//////////////////////////////////////////////////////////////
func (p *OnlineHandle) get_user(uid string) (*online_define.UserNode, bool) {
	now := time.Now().Unix()
	user, ok := p.user_cache.Get(uid)
	if !ok || user.ExpireTs > now {
		return nil, false
	}
	return &user.UserNode, true
}

func (p *OnlineHandle) set_user(user *online_define.UserNode) error {
	cfg := util.GetConfigValue[online_define.OnlineConfig]()
	user_node := &UserNodeInfo {
		UserNode:	*user,
		ExpireTs:	time.Now().Unix() + int64(cfg.CacheExpireSec),
	}
	return p.user_cache.Set(user.Uid, user_node)
}

func (p *OnlineHandle) del_user(uid string) bool {
	return p.user_cache.Del(uid)
}
