//package map_queue_test
package hbservice_test

import (
	"testing"
	"hbservice/src/util"
)

func TestMapQueue(t *testing.T) {
	var dt util.MapQueue[string]
	dt.Init(3)
	dt.Set("haobi1", 1, false)
	dt.Set("haobi2", 2, false)
	dt.Set("haobi3", 3, false)
	if ok := dt.Set("haobi4", 4, false); !ok {
		t.Fatalf("dt.Set not limit size")
	}
	if ok := dt.Set("haobi5", 5, true); !ok {
		t.Fatalf("dt.Set not eliminate")
	}

	if dt.GetCount() != 3 {
		t.Fatalf("dt.GetCount() != 3")
	}

	if v, ok := dt.GetAndDelete("haobi3"); ok {
		if v.(int) != 3 {
			t.Fatalf("dt.GetAndDelete value not match")
		}
	} else {
		t.Fatalf("dt.GetAndDelete error")
	}

	check_exists(t, &dt, "haobi2")
	check_exists(t, &dt, "haobi5")

	check_get_kv(t, &dt, "haobi2", 2)
	check_get_kv(t, &dt, "haobi5", 5)

	check_top_kv(t, &dt, "haobi2", 2)
	check_top_kv(t, &dt, "haobi5", 5)
}

func TestMapQueueDeleteHead(t *testing.T) {
	var dt util.MapQueue[string]
	dt.Init(3)
	dt.Set("haobi1", 1, false)
	dt.Set("haobi2", 2, false)
	dt.Set("haobi3", 3, false)
	if dt.GetCount() != 3 {
		t.Fatalf("dt.GetCount() != 3")
	}
	dt.Delete("haobi1")
	k, v := dt.Front()
	if k != "haobi2" || v.(int) != 2 {
		t.Fatalf("DeleteHead not match")
	}
	check_top_kv(t, &dt, "haobi2", 2)
	check_top_kv(t, &dt, "haobi3", 3)
}

func TestMapQueueDeleteMiddle(t *testing.T) {
	var dt util.MapQueue[string]
	dt.Init(3)
	dt.Set("haobi1", 1, false)
	dt.Set("haobi2", 2, false)
	dt.Set("haobi3", 3, false)
	if dt.GetCount() != 3 {
		t.Fatalf("dt.GetCount() != 3")
	}
	dt.Delete("haobi2")
	if dt.GetCount() != 2 {
		t.Fatalf("dt.GetCount() != 2")
	}
	k, v := dt.Front()
	if k != "haobi1" || v.(int) != 1 {
		t.Fatalf("DeleteHead not match")
	}
	check_top_kv(t, &dt, "haobi1", 1)
	check_top_kv(t, &dt, "haobi3", 3)
}

func TestMapQueueDeleteTail(t *testing.T) {
	var dt util.MapQueue[string]
	dt.Init(3)
	dt.Set("haobi1", 1, false)
	dt.Set("haobi2", 2, false)
	dt.Set("haobi3", 3, false)
	if dt.GetCount() != 3 {
		t.Fatalf("dt.GetCount() != 3")
	}
	dt.Delete("haobi3")
	if dt.GetCount() != 2 {
		t.Fatalf("dt.GetCount() != 2")
	}
	k, v := dt.Front()
	if k != "haobi1" || v.(int) != 1 {
		t.Fatalf("DeleteHead not match")
	}
	check_top_kv(t, &dt, "haobi1", 1)
	check_top_kv(t, &dt, "haobi2", 2)
}

func TestMapQueueSearchAndRemove(t *testing.T) {
	var dt util.MapQueue[int]
	dt.Init(5)
	dt.Set(1, 1, false)
	dt.Set(2, 2, false)
	dt.Set(3, 3, false)

	check_map_value := func(dt map[int]interface{}, key int) {
		if _, ok := dt[key]; !ok {
			t.Fatalf("check_map_value, key:%d fail", key)
		}
	}

	fsearch := func(k int, v interface{}) (bool, bool) {
		if k <= 2 {
			return false, true
		}
		return false, false
	}
	result1 := dt.Search(fsearch)
	if len(result1) != 2 {
		t.Fatalf("result1.size != 2")
	}
	check_map_value(result1, 1)
	check_map_value(result1, 2)

	fremove := func(k int, v interface{}) (bool , bool) {
		if k <= 2 {
			return false, true
		}
		return false, false
	}
	result2 := dt.Remove(fremove)
	check_map_value(result2, 1)
	check_map_value(result2, 2)

	fcheck_top_kv := func(dt *util.MapQueue[int], key int, value int) {
		if k, v := dt.Front(); k != key || v.(int) != value {
			t.Fatalf("fcheck_top_kv|k:%d, key:%d not match", k, key)
		}
		dt.PopFront()
	}
	fcheck_top_kv(&dt, 3, 3)
}

func TestMapQueueGet(t *testing.T) {
	var dt util.MapQueue[string]
	dt.Init(5)
	dt.Set("haobi1", 1, false)
	dt.Set("haobi2", 2, false)
	dt.Set("haobi3", 3, false)

	check_get(t, &dt, false, "haobi1", 1)
	check_back(t, &dt, "haobi3", 3)

	check_get(t, &dt, true, "haobi1", 1)
	check_back(t, &dt, "haobi1", 1)

	check_get(t, &dt, true, "haobi3", 3)
	check_back(t, &dt, "haobi3", 3)
}

func check_get(t *testing.T, dt *util.MapQueue[string], move_to_tail bool, k string, v int) {
	if get_v, ok := dt.Get(k, move_to_tail); ok {
		if get_v.(int) != v {
			t.Fatalf("dt.Get, key:%s value not match", k)
		}
	} else {
		t.Fatalf("dt.Get, key:%s error", k)
	}
}

func check_back(t *testing.T, dt *util.MapQueue[string], k string, v int) {
	get_k, get_v := dt.Back()
	if k != get_k || v != get_v.(int) {
		t.Fatalf("dt.Back, not match")
	}
}

func check_exists(t *testing.T, dt *util.MapQueue[string], k string) {
	if dt.Exists(k) == false {
		t.Fatalf("check_exists|k:%s not exists", k)
	}
}

func check_get_kv(t *testing.T, dt *util.MapQueue[string], k string, dv int) {
	 v, ok := dt.Get(k, false)
	 if !ok {
		t.Fatalf("check_get_kv|k:%s not exist", k)
	 }
	 if v != dv {
		t.Fatalf("check_get_kv|v:%d value not equal", dv)
	 }
}

func check_top_kv(t *testing.T, dt *util.MapQueue[string], dk string, dv int) {
	if k, v := dt.Front(); k != dk || v.(int) != dv {
		t.Fatalf("check_top_kv|k:%s, dk:%s not match", k, dk)
	}
	dt.PopFront()
}
