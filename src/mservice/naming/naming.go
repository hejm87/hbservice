package naming

import (
	"fmt"
	"log"
	"time"
	"sync"
	"errors"
	"strings"
	"strconv"
	"context"
	"path/filepath"
	"hbservice/src/util"
	"hbservice/src/mservice/util"
	"hbservice/src/mservice/define"
)

const (
	ServicePut		int = 1
	ServiceDel		int = 2
	
	kNameServiceDir			string = "NameService"

	kErrorKeyAlreadyWatch				string = "key already watch"
	kErrorServiceIdAlreadyRegister		string = "service_id already register"
	kErrorServiceIdNotRegister			string = "service_id not register"
)

type ServiceInfo struct {
	Name		string
	ServiceId	string		// {服务名}_{seq_id}
	Host		string
	Port		int
}

type ServiceChange struct {
	Service			ServiceInfo
	Op				int
}

// 服务发现数据格式:
// key:		{服务名}_{host}_{port}
// value:	{seq_id}

type Naming struct {
	watch_cancels		map[string]context.CancelFunc
	register_cancels	map[string]context.CancelFunc
	sync.Mutex
}

var (
	instance		*Naming
	once			sync.Once
)

func GetInstance() *Naming {
	once.Do(func() {
		instance = &Naming {
			watch_cancels:		make(map[string]context.CancelFunc),
			register_cancels:	make(map[string]context.CancelFunc),
		}
	})
	return instance
}

type ChangeCb func([]ServiceChange)

func (p *Naming) Find(name string) (infos []ServiceInfo, err error) {
	prefix := p.get_name_dir(name) + "/" + name
	result, err := mservice_util.GetEtcdInstance().GetWithPrefix(prefix)
	if err != nil {
		return infos, err
	}
	for k, v := range result {
		if info, ok := p.convert_str_to_service_info(k, v); ok {
			infos = append(infos, info)
		}
	}
	return infos, nil
}

func (p *Naming) Subscribe(name string, cb ChangeCb) error {
	p.Lock()
	defer p.Unlock()
	if _, ok := p.watch_cancels[name]; ok {
		return errors.New(kErrorKeyAlreadyWatch)
	}
	watch_f := func(results []util.WatchResult) {
		var changes []ServiceChange
		for _, w := range results {
			if service, ok := p.convert_str_to_service_info(w.Key, w.Value); ok {
				var op int
				if w.Type == util.ETCD_PUT {
					op = ServicePut
				} else {
					op = ServiceDel
				}
				change := ServiceChange {
					Service:	service,
					Op:			op,
				}
				changes = append(changes, change)
			}
		}
		cb(changes)
	}
	ctx, cancel := context.WithCancel(context.TODO())
	p.watch_cancels[name] = cancel
	go func() {
		prefix := fmt.Sprintf("%s/%s/%s", p.get_name_dir(name), name, name)
		mservice_util.GetEtcdInstance().Watch(ctx, prefix, true, watch_f)
	} ()
	return nil
}

func (p *Naming) Unsubscribe(name string) {
	p.Lock()
	defer p.Unlock()
	if cancel, ok := p.watch_cancels[name]; ok {
		cancel()
		delete(p.watch_cancels, name)
	}
}

func (p *Naming) Register(info ServiceInfo) error {
	p.Lock()
	if _, ok := p.register_cancels[info.ServiceId]; ok {
		return errors.New(kErrorServiceIdAlreadyRegister)
	}

	ctx, cancel := context.WithCancel(context.TODO())
	p.register_cancels[info.ServiceId] = cancel
	p.Unlock()

	cfg := util.GetConfigValue[mservice_define.MServiceConfig]().Service
	k, v := p.convert_service_info_to_str(info)

	exists_ttl := 3 * cfg.KeepAliveTtl
	lease_id, err := mservice_util.GetEtcdInstance().PutWithTimeout(k, v, int64(exists_ttl))
	if err != nil {
		return err
	}

	go func() {
		timer := time.NewTicker(time.Duration(cfg.KeepAliveTtl) * time.Second)
		for {
			select {
			case <-ctx.Done():
				break
			case <-timer.C:
				mservice_util.GetEtcdInstance().KeepAlive(lease_id)
			}
		}
	} ()
	return nil
}

func (p *Naming) Deregister(service_id string) error {
	p.Lock()
	defer p.Unlock()
	cancel, ok := p.register_cancels[service_id]
	if !ok {
		return errors.New(kErrorServiceIdNotRegister)
	}
	cancel()
	return nil
}

func (p *Naming) convert_str_to_service_info(key string, value string) (service ServiceInfo, ok bool) {
	_, name_id := filepath.Split(key)
	keys := strings.Split(name_id, "_")
	port, err := strconv.Atoi(keys[2])
	if len(keys) != 3 || err != nil {
		return service, false
	}
	service = ServiceInfo {
		Name:		keys[0],
		ServiceId:	value,
		Host:		keys[1],
		Port:		port,
	}
	return service, true
}

func (p *Naming) convert_service_info_to_str(info ServiceInfo) (k string, v string){
	k = fmt.Sprintf("%s/%s/%s_%s_%d", p.get_name_dir(info.Name), info.Name, info.Name, info.Host, info.Port)
	v = info.ServiceId
	return k, v
}

func (p *Naming) get_name_dir(name string) string {
	return kNameServiceDir + "/" + name
}
