package obj_client

import (
	"sync"
	"hbservice/src/net/net_core"
	"hbservice/src/mservice/define"
)

var (
	instance	*ObjClientMgr
	once		sync.Once
)

func GetInstance() *ObjClientMgr {
	once.Do(func() {
		instance = new(ObjClientMgr)
	})
	return instance
}

type ObjClientMgr struct {
	objs		map[string]*ObjectProxy
	handle		net_core.PacketHandle
	sync.Mutex
}

func (p *ObjClientMgr) Call(hash uint32, req *mservice_define.MServicePacket) (*mservice_define.MServicePacket, error) {
	proxy, err := p.get_proxy(req.Header.Service)
	if err != nil {
		return nil, err
	}
	return proxy.Call(hash, req)
}

func (p *ObjClientMgr) AsyncCall(hash uint32, req *mservice_define.MServicePacket, cb RespPacketCb) error {
	proxy, err := p.get_proxy(req.Header.Service)
	if err != nil {
		return err
	}
	proxy.AsyncCall(hash, req, cb)
	return nil
}

func (p *ObjClientMgr) get_proxy(name string) (*ObjectProxy, error) {
	var ok bool
	var proxy *ObjectProxy = nil
	p.Lock()
	defer p.Unlock()
	if proxy, ok = p.objs[name]; !ok {
		proxy = NewObjectProxy(name, p.handle)
		if err := proxy.Start(); err != nil {
			return nil, err
		}
		p.objs[name] = proxy
	}
	return proxy, nil
}