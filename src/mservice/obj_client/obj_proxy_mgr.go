package obj_client

import (
	"log"
	"sync"
	"hbservice/src/mservice/define"
)

var (
	instance	*ObjProxyMgr
	once		sync.Once
)

func GetInstance() *ObjProxyMgr {
	once.Do(func() {
		instance = &ObjProxyMgr {
			proxys:		make(map[string]*ObjectProxy),
		}
	})
	return instance
}

type ObjProxyMgr struct {
	proxys			map[string]*ObjectProxy		
	sync.Mutex
}

func (p *ObjProxyMgr) get_obj_proxy(name string) *ObjectProxy {
	p.Lock()	
	defer p.Unlock()
	if proxy, ok := p.proxys[name]; ok {
		return proxy
	}	
	proxy := NewObjectProxy(name, &mservice_define.MPacketHandle {}) 
	p.proxys[name] = proxy
	return proxy
}

func (p *ObjProxyMgr) Call(hash uint32, req *mservice_define.MServicePacket) (*mservice_define.MServicePacket, error) {
	return p.get_obj_proxy(req.Header.Service).Call(hash, req)
}

func (p *ObjProxyMgr) CallAsync(hash uint32, req *mservice_define.MServicePacket, cb RespPacketCb) {
	p.get_obj_proxy(req.Header.Service).CallAsync(hash, req, cb)
}

func (p *ObjProxyMgr) Cast(hash uint32, req *mservice_define.MServicePacket) {
	p.get_obj_proxy(req.Header.Service).Cast(hash, req)
}