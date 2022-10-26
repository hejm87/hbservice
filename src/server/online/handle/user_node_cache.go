package online_handle

import (
	"errors"
	"hbservice/src/util"
	"hbservice/src/server/online/define"
)

type UserNodeInfo struct {
	UserNode		online_define.UserNode
	ExpireTs		int64
}

// lru特性的用户节点缓存
type UserNodeCache struct {
	zset				*util.SortedSet
	user_nodes			map[string]*UserNodeInfo
	max_size		int
}

func NewUserNodeCache(max_size int) *UserNodeCache {
	return &UserNodeCache {
		zset:		util.MakeSortedSet(),
		user_nodes:	make(map[string]*UserNodeInfo),
		max_size:	max_size,
	}
}

func (p *UserNodeCache) Set(uid string, user *UserNodeInfo) error {
	_, exists := p.user_nodes[uid]
	if !exists && p.zset.Len() >= int64(p.max_size) {
		if res := p.zset.PopMin(1); res != nil {
			delete(p.user_nodes, res[0].Member)
		} else {
			return errors.New(online_define.ERR_INNER_SERVER)
		}
	}
	if ok := p.zset.Add(uid, float64(user.ExpireTs)); !ok {
		return errors.New(online_define.ERR_INNER_SERVER)
	}
	p.user_nodes[uid] = user
	return nil
}

func (p *UserNodeCache) Get(uid string) (*UserNodeInfo, bool) {
	v, ok := p.user_nodes[uid]
	return v, ok
}

func (p *UserNodeCache) Del(uid string) bool {
	_, exists := p.user_nodes[uid]
	if !exists {
		return false
	}
	p.zset.Remove(uid)
	delete(p.user_nodes, uid)
	return true
}
