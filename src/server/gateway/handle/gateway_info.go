package gateway_handle

import (
	"sync"
	"time"
	"hbservice/src/util"
)

var (
	instance		*GWUsersMgr
	once			sync.Once
)

func GetGWUsersInstance() *GWUsersMgr {
	once.Do(func() {
		instance = new(GWUsersMgr)
	})
	return instance
}

type UserInfo struct {
	Uid					string
	ChannelId			string		// 网关对应channel_id
	ConnectTs			int64		// 连接时间
	LatestActiveTs		int64		// 最近通信时间
}

type GWUsersMgr struct {
	users				util.MapQueue[string]
	timeout_sec			int
	sync.Mutex
}

func (p *GWUsersMgr) NewUser(uid string, channel_id string) bool {
	p.Lock()
	defer p.Unlock()
	now := time.Now().Unix()
	user := &UserInfo {
		Uid:			uid,
		ChannelId:		channel_id,
		ConnectTs:		now,
		LatestActiveTs:	now,
	}
	if err := p.users.Push(uid, user, false); err != nil {
		return false
	}
	return true
}

func (p *GWUsersMgr) GetUserChannelId(uid string) (string, bool) {
	p.Lock()
	defer p.Unlock()
	if obj, ok := p.users.Get(uid); ok {
		user_info := obj.(*UserInfo)
		return user_info.ChannelId, true
	}
	return "", false
}

func (p *GWUsersMgr) UpdateUser(uid string) bool {
	p.Lock()
	defer p.Unlock()
	obj, ok := p.users.Get(uid)
	if !ok {
		return false
	}
	user := obj.(*UserInfo)
	user.LatestActiveTs = time.Now().Unix()
	p.users.Push(uid, user, true)
	return true
}

func (p *GWUsersMgr) GetOverdueUsers() []*UserInfo {
	var users []*UserInfo
	p.Lock()
	defer p.Unlock()
	now := time.Now().Unix()
	for {
		user := p.users.Top().(*UserInfo)
		if user == nil || user.LatestActiveTs + int64(p.timeout_sec) > now {
			break
		}
		users = append(users, user)
		p.users.Pop()
	}
	return users
}
