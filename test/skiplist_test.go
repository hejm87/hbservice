package hbservice_test

import (
	"testing"
	"hbservice/src/util"
)

func TestMapSkiplist(t *testing.T) {
	zset := util.MakeSortedSet()
	zset.Add("u1000", 10)
	zset.Add("u1001", 11)
	zset.Add("u1002", 11)
	zset.Add("u1003", 13)
	zset.Add("u1004", 14)
	zset.Add("u1005", 15)
	zset.Add("u1006", 16)

	type Element struct {
		Member string
		Score  float64
	}

	if zset.Len() != 7 {
		t.Fatalf("zset.Len() != 7")
	}
	check_list_get(t, zset, "u1000", 10)
	check_list_get(t, zset, "u1002", 11)
	check_list_get(t, zset, "u1004", 14)

	if zset.GetRank("u1005", true) != 1 {
		t.Fatalf("zset.GetRank u1005 not match")
	}
	if zset.GetRank("u1004", true) != 2 {
		t.Fatalf("zset.GetRank u1004 not match")
	}

	// 查区间元素
	var elems []*util.Element
	// 1. 某值到无穷大
	min := &util.ScoreBorder {Value: 11}
	max := &util.ScoreBorder {Inf: 1}
	elems = zset.RangeByScore(min, max, 0, -1, true)
	if len(elems) != 6 {
		t.Fatalf("RangeByScore.size:%d != 6, range:[12 ~ +inf)", len(elems))
	}
	if elems[0].Member != "u1006" || elems[0].Score != 16.0 {
		t.Fatalf("elem[0] not match")
	}
	if elems[1].Member != "u1005" || elems[1].Score != 15.0 {
		t.Fatalf("elem[1] not match")
	}
	if elems[2].Member != "u1004" || elems[2].Score != 14.0 {
		t.Fatalf("elem[2] not match")
	}
	if elems[3].Member != "u1003" || elems[3].Score != 13.0 {
		t.Fatalf("elem[3] not match")
	}
	if elems[4].Member != "u1002" || elems[4].Score != 11.0 {
		t.Fatalf("elem[4] not match")
	}
	if elems[5].Member != "u1001" || elems[5].Score != 11.0 {
		t.Fatalf("elem[5] not match")
	}

	// 2. 无穷小到某值
	min = &util.ScoreBorder {Inf: -1}
	max = &util.ScoreBorder {Value: 13}
	elems = zset.RangeByScore(min, max, 0, 2, false)
	if len(elems) != 2 {
		t.Fatalf("RangeByScore.size:%d != 2, range:[-inf, 13], limit:2", len(elems))
	}
	if elems[0].Member != "u1000" || elems[0].Score != 10.0 {
		t.Fatalf("elem[0] not match")
	}
	if elems[1].Member != "u1001" || elems[1].Score != 11.0 {
		t.Fatalf("elem[1] not match")
	}

	// 3. 值区间
	min = &util.ScoreBorder {Value: 13}
	max = &util.ScoreBorder {Value: 15}
	elems = zset.RangeByScore(min, max, 0, -1, false)
	if len(elems) != 3 {
		t.Fatalf("RangeByScore.size:%d != 3, range:[13, 15]", len(elems))
	}
	if elems[0].Member != "u1003" || elems[0].Score != 13.0 {
		t.Fatalf("elem[0] not match")
	}
	if elems[1].Member != "u1004" || elems[1].Score != 14.0 {
		t.Fatalf("elem[1] not match")
	}
	if elems[2].Member != "u1005" || elems[2].Score != 15.0 {
		t.Fatalf("elem[2] not match")
	}
	
}

func check_list_get(t *testing.T, zset *util.SortedSet, k string, v float64) {
	e, ok := zset.Get(k)
	if !ok {
		t.Fatalf("zset.Get key:%s not found", k)
	}
	if e.Member != k || e.Score != v {
		t.Fatalf("zset.Get not match, result:{%s, %f}, check:{%s, %f}", e.Member, e.Score, k, v)
	}
}

//type Element int
//func (e Element) ExtractKey() float64 {
//	return float64(e)
//}
//func (e Element) String() string {
//	return fmt.Sprintf("%003d", e)
//}
//
//func TestSkipList(t *testing.T) {
//	list := skiplist.New()
//
//    // Insert some elements
//	list.Insert(Element(10))
//	list.Insert(Element(11))
//	list.Insert(Element(11))
//	list.Insert(Element(12))
//	list.Insert(Element(12))
//	list.Insert(Element(13))
//
//	fmt.Printf("count:%d\n", list.GetNodeCount())
//
//    // Find an element
//    if e, ok := list.Find(Element(11)); ok {
//		for {
//			fmt.Printf("elem:%s\n", e.GetValue().String())
//			if e == list.GetLargestNode() {
//				break
//			} else {
//				e = list.Next(e)
//			}
//		}
//    }
//
//    // Delete all elements
//    for i := 0; i < 100; i++ {
//        list.Delete(Element(i))
//    }
//}
