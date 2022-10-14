package naming

import (
	"time"
	"sync"
	"strings"
	"strconv"
	"context"
	"hbservice/src/util"
)

const (
	kServicePut		int = 1
	kServiceDel		int = 2
	
	kNameServiceDir			string = "NameService"

	kErrorKeyAlreadyWatch				string = "key already watch"
	kErrorServiceIdAlreadyRegister		string = "service_id already register"
	kErrorServiceIdNotRegister			string = "service_id not register"
)

type ServiceInfo {
	Name		string
	ServiceId	string		// {服务名}_{seq_id}
	Host		string
	Port		int
}

type ServiceChange {
	Service			ServiceInfo
	Op				int
}

// 服务发现数据格式:
// key: {服务名}_{seq_id}
// value: {host}:{port}

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
		instance = new(Naming)
	})
	return instance
}

type ChangeCb func([]ServiceChange)

func (p *Naming) Find(name string) (infos []ServiceInfo, err error) {
	prefix := p.get_prefix()
	result, err := mservice_util.GetEtcdInstance().GetWithPrefix(prefix)
	if err != nil {
		return services, err
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
	watch_f := func(results []WatchResult) {
		var changes []ServiceChange
		for w := range results {
			if service, ok := p.get_service_info(w.Key, w.Value); ok {
				var op int
				if w.Type == ETCD_PUT {
					op = kServicePut
				} else {
					op = kServiceDel
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
	ctx, cancel := context.WithCancel()
	p.watch_cancels[name] = cancel
	go func() {
		mservice_util.GetEtcdInstance().Watch(ctx, name, true, watch_f)
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
	defer p.Unlock()
	if _, ok := p.register_cancels[info.ServiceId]; ok {
		return errors.New(kErrorServiceIdAlreadyRegister)
	}

	ctx, cancel := context.WithCancel(context.TODO())
	p.register_cancels[info.ServiceId] = cancel

	cfg := util.GetConfigValue[mservice_define.MServiceConfig].Service
	k, v := p.convert_service_info_to_str(info)
	if lease_id, err := mservice_util.GetEtcdInstance().PutWithTimeout(k, v, cfg.KeepAliveTtl); err != nil {
		return err
	}

	go func() {
		timer := time.NewTricker(time.Duration(cfg.KeepAlive) * time.Second)
		for {
			select {
			case <-ctx.Done():
				break
			case <-timer.C:
				mservice_util.GetEtcdInstance().KeepAlive(lease_id)
			}
		}
		mservice_util.GetEtcdInstance().KeepAlive(lease_id)
	} ()
	return nil
}

func (p *Naming) Deregister(service_id string) error {
	p.Lock()
	defer p.Unlock()
	if cancel, ok := p.register_cancels[service_id]; !ok {
		return errors.New(kErrorServiceIdNotRegister)
	}
	cancel()
	return nil
}

func (p *Naming) convert_str_to_service_info(key string, value string) (service ServiceInfo, ok bool) {
	keys := strings.Split(key, "_")
	values := strings.Split(value, ":")
	if len(keys) != 2 || len(values) != 2 {
		return service, false
	}
	result := ServiceInfo {
		Name:	keys[0],
		SeqId:	key,
		Host:	values[0],
		Port:	strconv.Atoi(values[1]),
	}
	return result, true
}

func (p *Naming) convert_service_info_to_str(info ServiceInfo) (k string, v string){
	k = info.ServiceId
	v = fmt.Sprintf("%s:%d", info.Host, info.Port)
	return k, v
}

func (p *Naming) get_prefix(name) string {
	return kNameServiceDir + "/" + name + "/" + name
}
