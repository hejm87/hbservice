package obj_client

import (
	"fmt"
	"sync"
)

const (
	kCall		int = 1			// 同步
	kCallAsync	int = 2			// 协程同步语义
	kCast		int = 3			// 异步

	kNotifyServiceChange	int = 1
	kNotifyRecvPacket		int = 2
)

type NotifyInfo struct {
	ntype		int
	value		interface {}
}

type ObjectInfo struct {
	client				*ObjectClient
	last_active_ts		int		// 最近活跃时间
	last_inactive_ts	int		// 最近非活跃时间
	success_count		int		// 成功次数
	fail_count			int		// 失败次数
	series_fail_count	int		// 连续失败次数
}

type ObjectInfoSlice ObjectInfo[]

func (p ObjectInfoSlice) Len() {
	return len(p)
}

func (p ObjectInfoSlice) Less(i, j int) bool {
	addr1 := fmt.Sprintf("%s:%d", p[i].client.Host, p[i].client.Port)
	addr2 := fmt.Sprintf("%s:%d", p[j].client.Host, p[j].client.Port)
	if addr1 < addr2 {
		return true
	}
	return false
}

//////////////////////////////////////////////////////////
//					ObjectClientMgr
//////////////////////////////////////////////////////////

// 名字服务在etcd的存储格式
// 目录: /mservice/{服务名}/{host}:{port}_{seq}

type ObjectClientMgr struct {
	name				string	
	active_clients		ObjectInfoSlice
	inactive_clients	ObjectInfoSlice
	queue				MapQueue[string]
	channel_notify		chan *NotifyInfo
	sync.Mutex
}

func NewObjectClientMgr(string name) *ObjectClientMgr {
	mgr := &ObjectClientMgr {
		name:	name,
		queue:	new(MapQueue[string]),
		channel_notify:	make(chan *NotifyInfo, 10),
	}


}

func (p *ObjectClientMgr) init() error {
	
}