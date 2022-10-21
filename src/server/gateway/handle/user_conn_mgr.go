package gateway_handle

import (
	"sync"
	"time"
	"errors"
	"hbservice/src/util"
	"hbservice/src/server/gateway/define"
)

var (
	instance		*GWUserConnMgr
	once			sync.Once
)

func GetUserConnsInstance() *GWUserConnMgr {
	once.Do(func() {
		cfg := util.GetConfigValue[gateway_define.GatewayConfig]().Common
		instance = &GWUserConnMgr {
			zset_uid:			util.MakeSortedSet(),
			user_conns:			make(map[string]*UserConnInfo),
			max_conn_size:		cfg.MaxUserCount,
			conn_timeout_sec:	cfg.UserHeartbeatSec * 3,
		}
	})
	return instance
}

////////////////////////////////////////////////////////////////
//						连接管理
////////////////////////////////////////////////////////////////
const (
	eUidAlreadyExists			string = "uid already exists"
	eExceedMaxConnLimit			string = "exceed max conn limit"
	eAddUserConnFail			string = "add user conn fail"
)

type UserConnInfo struct {
	Uid				string
	ChannelId		string
	ConnectTs		int64
	ExpireTs		int64
}

type GWUserConnMgr struct {
	zset_uid			*util.SortedSet
	user_conns			map[string]*UserConnInfo
	max_conn_size		int		// 最大连接限制数
	conn_timeout_sec	int		// 连接超时时长
	sync.Mutex
}

func (p *GWUserConnMgr) AddUserConn(uid string, channel_id string) error {
	p.Lock()
	defer p.Unlock()
	if _, ok := p.user_conns[uid]; ok {
		return errors.New(eUidAlreadyExists)
	}
	if p.zset_uid.Len() >= int64(p.max_conn_size) {
		return errors.New(eExceedMaxConnLimit)
	}

	now := time.Now().Unix()
	if ok := p.zset_uid.Add(uid, float64(now)); !ok {
		return errors.New(eAddUserConnFail)
	}

	user_conn := &UserConnInfo {
		Uid:			uid,
		ChannelId:		channel_id,
		ConnectTs:		now,
		ExpireTs:		now + int64(p.conn_timeout_sec),
	}
	p.user_conns[uid] = user_conn
	return nil
}

func (p *GWUserConnMgr) RemoveUserConn(uid string) bool {
	p.Lock()
	defer p.Unlock()
	if ok := p.zset_uid.Remove(uid); !ok {
		return false
	}
	delete(p.user_conns, uid)
	return true
}

func (p *GWUserConnMgr) RefreshUserConn(uid string) bool {
	p.Lock()
	defer p.Unlock()
	uc, ok := p.user_conns[uid]
	if !ok {
		return false
	}
	expire_ts := time.Now().Unix() + int64(p.conn_timeout_sec)
	if ok := p.zset_uid.Add(uid, float64(expire_ts)); !ok {
		return false
	}
	uc.ExpireTs = expire_ts
	return true
}

func (p *GWUserConnMgr) GetChannelId(uid string) (string, bool) {
	p.Lock()
	defer p.Unlock()
	if uc, ok := p.user_conns[uid]; ok {
		return uc.ChannelId, true
	}
	return "", false
}

func (p *GWUserConnMgr) GetTimeoutUserConns() []string {
	p.Lock()
	defer p.Unlock()
	var users []string
	min := &util.ScoreBorder {Value: 0}
	max := &util.ScoreBorder {Value: float64(time.Now().Unix())}
	elems := p.zset_uid.RangeByScore(min, max, 0, -1, false)
	for _, x := range elems {
		users = append(users, x.Member)
	}
	return users
}
