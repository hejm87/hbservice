package util

import (
	"log"
	"time"
	"errors"
	"context"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"github.com/coreos/etcd/clientv3"
)

const (
	ETCD_PUT		int32 = 0
	ETCD_DELETE		int32 = 1
)

type WatchResult struct {
	Type		int32
	Key			string
	Value		string
	LeaseId		int64
}

type Etcd struct {
	client		*clientv3.Client
}

type WatchCallback func([]WatchResult)

func NewEtcd(addrs []string, username, password string) *Etcd {
	var etcd_conf = clientv3.Config {
		Endpoints:		addrs,
		DialTimeout:	5 * time.Second,
		TLS:			nil,
	}
	if username != "" && password != "" {
		etcd_conf.Username = username
		etcd_conf.Password = password
	}

	etcd_client, err := clientv3.New(etcd_conf)
	if err != nil {
		log.Fatalf("clientv3.New error:%#v", err)
	}
	return &Etcd{client: etcd_client}
}

func (p *Etcd) Close() error {
	return p.client.Close()
}

func (p *Etcd) Put(key string, value string) error {
	_, err := p.client.KV.Put(context.TODO(), key, value)
	return err
}

// param:
// 		ttl: 过期时长（单位：秒）
// return:
//		1] lease_id: 租约id
//		2] err: 错误码	
func (p *Etcd) PutWithTimeout(key string, value string, ttl int64) (int64, error) {
	ctx := context.TODO()
	resp, err := p.client.Lease.Grant(ctx, ttl)
	if err != nil {
		return 0, err
	}
	if _, err := p.client.KV.Put(ctx, key, value, clientv3.WithLease(resp.ID)); err != nil {
		return 0, err
	}
	return int64(resp.ID), nil
}

func (p *Etcd) KeepAlive(lease_id int64) error {
	ctx := context.TODO()
	_, err := p.client.Lease.KeepAliveOnce(ctx, clientv3.LeaseID(lease_id))
	return err
}

func (p *Etcd) Get(key string) (string, error) {
	resp, err := p.client.KV.Get(context.TODO(), key)
	if err != nil {
		return "", err
	}
	for _, kv := range resp.Kvs {
		return string(kv.Value), nil
	}
	return "", nil
}

func (p *Etcd) GetWithPrefix(key_prefix string) (map[string]string, error) {
	kvs := make(map[string]string)
	resp, err := p.client.KV.Get(context.TODO(), key_prefix, clientv3.WithPrefix())
	if err != nil {
		return kvs, err
	}
	for _, kv := range resp.Kvs {
		kvs[string(kv.Key)] = string(kv.Value)
	}
	return kvs, nil
}

func (p *Etcd) Watch(ctx context.Context, key string, with_prefix bool, cb WatchCallback) error {
	var wch clientv3.WatchChan
	if with_prefix == false {
		wch = p.client.Watch(ctx, key)
	} else {
		wch = p.client.Watch(ctx, key, clientv3.WithPrefix())
	}
	for x := range wch {
		var resp []WatchResult
		if x.Err() != nil {
			log.Printf("ERROR|watch key:%s error:%#v", key, x.Err())
			return x.Err()
		}
		// etcd client close
		if len(x.Events) == 0 && x.Err() == nil {
			log.Printf("ERROR|watch key:%s etcd client close", key)
			return nil
		}
		for _, ev := range x.Events {
			var result WatchResult
			key := string(ev.Kv.Key)
			value := string(ev.Kv.Value)
			if ev.Type == mvccpb.PUT {
				result = WatchResult {ETCD_PUT, key, value, ev.Kv.Lease}
			} else {
				result = WatchResult {ETCD_DELETE, key, value, ev.Kv.Lease}
			}
			resp = append(resp, result)
		}
		cb(resp)
	}
	return nil
}