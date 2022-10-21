package hbservice_test

import (
	"time"
	"testing"
	"hbservice/src/util"
)

func TestMergeSearch(t *testing.T) {
	f := func(key string) (string, error) {
		time.Sleep(time.Duration(2) * time.Second)
		return key, nil
	}

	obj := util.NewMergeSearch[string, string](f)

	key := "haobi"
	for x := 0; x < 10; x += 1 {
		go func(x int) {
			time.Sleep(time.Duration(1) * time.Second)
			if value, err := obj.Call(key); err == nil {
				if value != "haobi" {
					t.Fatal("corouting value not match")
				}
			} else {
				t.Fatal("corouting Call error")
			}
		} (x)
	}
	if value, err := obj.Call(key); err == nil {
		if value != "haobi" {
			t.Fatal("main value not match")
		}
	} else {
		t.Fatal("main Call error")
	}
	time.Sleep(time.Duration(1) * time.Second)
}