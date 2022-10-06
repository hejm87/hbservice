package config_test

import (
	"os"
	"time"
	"reflect"
	"testing"
	"io/ioutil"
	"encoding/json"
	"hbservice/src/util"
)

type EtcdConfigData struct {
	A	int
	B	float32
	C	[]string
}

type FileConfigData struct {
	A	int
	B	float32
	C	[]string
}

func TestEtcdConfig(t *testing.T) {
	etcd, err := util.NewEtcd([]string{"127.0.0.1:2379"}, "", "")
	if err != nil {
		t.Fatalf("etcd new error:%#v", err)
	}

	key := "etcd_config"
	data := EtcdConfigData {
		A:	1,
		B:	1.23,
		C:	[]string{"abc", "123"},
	}
	if err := put_etcd_data(t, etcd, key, data); err != nil {
		t.Fatalf("put_etcd_data error:%#v", err)
	}

	if err := util.SetConfigByEtcdLoader[EtcdConfigData](etcd, key); err != nil {
		t.Fatalf("SetConfigByEtcdLoader error:%#v", err)
	}

	v := util.GetConfigValue[EtcdConfigData]()
	if v.A != 1 {
		t.Fatal("EtcdConfigData.A is not equal")
	}
	if v.B != 1.23 {
		t.Fatal("EtcdConfigData.B is not equal")
	}
	if reflect.DeepEqual(v.C, []string{"abc", "123"}) == false {
		t.Fatal("EtcdConfigData.C is not equal")
	}

	// 修改配置值，验证是否有更新
	data.A = 2
	data.B = 2.34
	data.C = []string{"2", "2.34"}
	if err := put_etcd_data(t, etcd, key, data); err != nil {
		t.Fatalf("put_etcd_data error:%#v", err)
	}
	time.Sleep(time.Second)
	v = util.GetConfigValue[EtcdConfigData]()
	if v.A != 2 {
		t.Fatal("EtcdConfigData.A is not equal")
	}
	if v.B != 2.34 {
		t.Fatal("EtcdConfigData.B is not equal")
	}
	if reflect.DeepEqual(v.C, []string{"2", "2.34"}) == false {
		t.Fatal("EtcdConfigData.C is not equal")
	}
}

func TestFileConfig(t *testing.T) {
	data := FileConfigData {
		A:	1,
		B:	1.23,
		C:	[]string{"abc", "123"},
	}

	path := "./file_config.json"
	if err := put_file_data(t, path, data); err != nil {
		t.Fatalf("put_etcd_data error:%#v", err)
	}

	if err := util.SetConfigByFileLoader[FileConfigData](path); err != nil {
		t.Fatalf("SetConfigByEtcdLoader error:%#v", err)
	}

	v := util.GetConfigValue[FileConfigData]()
	if v.A != 1 {
		t.Fatal("FileConfigData.A is not equal")
	}
	if v.B != 1.23 {
		t.Fatal("FileConfigData.B is not equal")
	}
	if reflect.DeepEqual(v.C, []string{"abc", "123"}) == false {
		t.Fatal("FileConfigData.C is not equal")
	}

	// 修改配置值，验证是否有更新
	time.Sleep(time.Second)
	data.A = 2
	data.B = 2.34
	data.C = []string{"2", "2.34"}
	if err := put_file_data(t, path, data); err != nil {
		t.Fatalf("put_etcd_data error:%#v", err)
	}
	time.Sleep(time.Second)
	v = util.GetConfigValue[FileConfigData]()
	if v.A != 2 {
		t.Fatal("FileConfigData.A is not equal")
	}
	if v.B != 2.34 {
		t.Fatal("FileConfigData.B is not equal")
	}
	if reflect.DeepEqual(v.C, []string{"2", "2.34"}) == false {
		t.Fatal("FileConfigData.C is not equal")
	}
}

func put_etcd_data(t *testing.T, etcd *util.Etcd, key string, data EtcdConfigData) error {
	var err error
	var jsonBytes []byte
	if jsonBytes, err = json.Marshal(&data); err != nil {
		return err
	}
	if err = etcd.Put(key, string(jsonBytes)); err != nil {
		return err
	}
	return nil
}

func put_file_data(t *testing.T, path string, data FileConfigData) error {
	var err error
	var jsonBytes []byte
	if jsonBytes, err = json.Marshal(&data); err != nil {
		return err
	}
	if err = ioutil.WriteFile(path, jsonBytes, os.ModePerm); err != nil {
		return err
	}
	return nil
}