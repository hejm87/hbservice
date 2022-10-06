package util

import (
	"sync"
	"errors"
	"reflect"
	"context"
	"io/ioutil"
	"encoding/json"
	"github.com/fsnotify/fsnotify"
)

////////////////////////////////////////////////////
//					模块需求
////////////////////////////////////////////////////
// 1. 配置内容为json格式
// 2. 支持配置自动更新以及手动更新

var (
	configs		map[string] interface{}
	mutex 		sync.Mutex
)

func SetConfig(loader ConfigLoader) error {
	key := loader.Key()
	mutex.Lock()
	if configs == nil {
		configs = make(map[string] interface{})
	}
	if _, ok := configs[key]; ok {
		return errors.New("exists config")
	}
	var err error
	var result interface{}
	if result, err = loader.Get(); err != nil {
		return err
	}
	configs[key] = result
	mutex.Unlock()
	go func(loader ConfigLoader) {
		loader.Watch()
	} (loader)
	return nil
}

func GetConfigValue[T any]() *T {
	mutex.Lock()
	defer mutex.Unlock()
	key := get_string[T]()
	if obj, ok := configs[key]; ok {
		return obj.(*T)
	}
	return nil
}

func ConfigUpdateCallback(key string, config interface{}) {
	mutex.Lock()
	defer mutex.Unlock()
	configs[key] = config
}

func get_string[T any]() string {
	var obj T
	return reflect.TypeOf(obj).String()
}

////////////////////////////////////////////////////
//					ConfigLoader
////////////////////////////////////////////////////
type ConfigUpdateCb func(key string, config interface{})

type ConfigLoader interface {
	Key() string
	Get() (interface{}, error)
	Watch() error
}

// ConfigFileLoader
type ConfigFileLoader[T any] struct {
	config_path		string		// 配置文件路径
	update_cb		ConfigUpdateCb
}

func SetConfigByFileLoader[T any](config_path string) error {
	loader := &ConfigFileLoader[T] {
		config_path:	config_path,
		update_cb:		ConfigUpdateCallback,
	}
	return SetConfig(loader)
}

func (p *ConfigFileLoader[T]) Key() string {
	return get_string[T]()
}

func (p *ConfigFileLoader[T]) Get() (interface{}, error) {
	var err error
	var content []byte
	if content, err = ioutil.ReadFile(p.config_path); err != nil {
		return nil, err
	}
	var obj T
	if err = json.Unmarshal(content, &obj); err != nil {
		return nil, err
	}
	return &obj, nil
}

func (p *ConfigFileLoader[T]) Watch() error {
	var err error
	var watcher *fsnotify.Watcher
	if watcher, err = fsnotify.NewWatcher(); err != nil {
		return err
	}
	defer watcher.Close()
	if err = watcher.Add(p.config_path); err != nil {
		return err
	}

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				break
			}
			if event.Op & fsnotify.Write == fsnotify.Write {
				config, err := p.Get()
				if err != nil {
					continue
				}
				p.update_cb(get_string[T](), config)
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				break
			}
			return err
		}
	}
	return nil
}

// ConfigEtcdLoader
type ConfigEtcdLoader[T any] struct {
	etcd			*Etcd
	config_key		string
	update_cb		ConfigUpdateCb
}

func SetConfigByEtcdLoader[T any](etcd *Etcd, config_key string) error {
	loader := &ConfigEtcdLoader[T] {
		etcd:		etcd,
		config_key:	config_key,
		update_cb:	ConfigUpdateCallback,
	}
	return SetConfig(loader)
}

func (p *ConfigEtcdLoader[T]) Key() string {
	return get_string[T]()
}

func (p *ConfigEtcdLoader[T]) Get() (interface{}, error) {
	var err error
	var value string
	if value, err = p.etcd.Get(p.config_key); err != nil {
		return nil, err
	}
	var obj T
	if err = json.Unmarshal([]byte(value), &obj); err != nil {
		return nil, err
	}
	return &obj, nil
}

func (p *ConfigEtcdLoader[T]) Watch() error {
	return p.etcd.Watch(context.TODO(), p.config_key, false, p.watch_cb)
}

func (p *ConfigEtcdLoader[T]) watch_cb(results []WatchResult) {
	for _, x := range results {
		if x.Type == ETCD_PUT {
			var obj T
			if err := json.Unmarshal([]byte(x.Value), &obj); err == nil {
				p.update_cb(get_string[T](), &obj)
			}
		}
	}
}
