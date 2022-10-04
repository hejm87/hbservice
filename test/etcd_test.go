package etcd_test

import (
	"fmt"
	"time"
	"context"
	"testing"
	"hbservice/src/util"
)

func TestFlow(t *testing.T) {
	etcd, err := util.NewEtcd([]string{"127.0.0.1:2379"}, "", "")
	if err != nil {
		t.Error("etcd new error")
	}

	// 测试正常流程
	if err = etcd.Put("key1", "value1"); err != nil {
		t.Error("etcd put error")
	}
	if value, err := etcd.Get("key1"); err == nil {
		if value != "value1" {
			t.Error("etcd get not equal")
		}
	} else {
		t.Error("etcd get error")
	}

	// 测试key过期流程
	if _, err := etcd.PutWithTimeout("key2", "value2", 1); err != nil {
		t.Error("etcd put with timeout error")
	}
	time.Sleep(3 * time.Second)
	if value, err := etcd.Get("key2"); err == nil {
		if value == "value2" {
			t.Error("etcd get value not equal")
		}
	} else {
		t.Error("etcd get error")
	}

	if err := etcd.Close(); err != nil {
		t.Error("etcd close error")
	}
}

func TestKeepAlive(t *testing.T) {
	etcd, err := util.NewEtcd([]string{"127.0.0.1:2379"}, "", "")
	if err != nil {
		t.Error("etcd new error")
	}
	if lease_id, err := etcd.PutWithTimeout("key2", "value2", 2); err == nil {
		time.Sleep(1 * time.Second)
		if err := etcd.KeepAlive(lease_id); err != nil {
			t.Error(fmt.Sprintf("etcd keep alive lease_id:%d, error:%#v", lease_id, err))
		}
		time.Sleep(2 * time.Second)
		if value, err := etcd.Get("key2"); err != nil {
			t.Error("etcd get error")
		} else {
			if value != "value2" {
				t.Error("etcd get value not exist")
			}
		}
		time.Sleep(1 * time.Second)
		if value, err := etcd.Get("key2"); err != nil {
			t.Error("etcd get error")
		} else {
			if value == "value2" {
				t.Error("etcd get unexpected value")
			}
		}
	} else {
		t.Error("etcd put with timeout error")
	}
}

func TestWatch(t *testing.T) {
	etcd, err := util.NewEtcd([]string{"127.0.0.1:2379"}, "", "")
	if err != nil {
		t.Error("etcd new error")
	}
	if err := etcd.Put("key_watch", "value_watch"); err != nil {
		t.Error("etcd put error")
	}

	go func(t *testing.T, etcd *util.Etcd) {
		i := 0
		for {
			time.Sleep(time.Second)
			etcd.Put("key_watch", fmt.Sprintf("value_change%d", i))
			i++
			t.Logf("etcd put2 succeed ...")
		}
	} (t, etcd)

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(3 * time.Second)
		cancel()
	} ()

	cb := func (results []util.WatchResult) {
		for _, x := range results {
			t.Logf("watch.Result:%#v", x)
		}
	}
	if err := etcd.Watch(ctx, "key_watch", false, cb); err != nil {
		t.Errorf("etcd watch error:%#v", err)
	}
}