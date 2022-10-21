package online_handle

import (
	"sync"
	"errors"
	"hbservice/src/util"
	"hbservice/src/server/online/define"
)

const (
	eSetException			string = "UserNodeCache.Set exception"
)

type UserNodeInfo struct {
	UserNode		online_define.UserNode
	ExpireTs		int64
}

// lru特性的用户节点缓存
type UserNodeCache struct {
	zset				*util.SortedSet
	user_nodes			map[string]*online_define.UserNodeInfo
	max_size		int
}

func NewUserNodeCache(max_size) *UserNodeCache {
	return &UserNodeCache {
		zset:		util.MakeSortedSet(),
		user_nodes:	make(map[string]*online_define.UserNodeInfo),
		max_size:	max_size,
	}
}

func (p *UserNodeCache) Set(uid string, user *UserNodeInfo) error {
	_, exists := p.user_nodes[uid]
	if !exists && p.zset.Len() >= p.max_limit {
		if res := zset.PopMin(1); res != nil {
			delete(p.user_nodes, res[0].Uid)
		} else {
			return errors.New(eSetException)
		}
	}
	if ok := p.zset.Add(uid, user.ActiveTs); !ok {
		return errors.New(eSetException)
	}
	p.user_nodes[uid] = user
	return nil
}

func (p *UserNodeCache) Get(uid string) (*online_define.UserNodeInfo, bool) {
	return p.user_nodes[uid]
}
