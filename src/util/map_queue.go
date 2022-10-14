package util

import (
	"fmt"
	"reflect"
)

type any = interface {}

type Node struct {
	prev		*Node
	next		*Node
	value		any
}

type MapQueue[K comparable] struct {
	dict		map[K] *Node
	head		*Node
	tail		*Node
	size		int
	max_size	int
}

func (p *MapQueue[K]) Init(max_size int) {
	p.dict = make(map[K] *Node)
	p.max_size = max_size
}

func (p *MapQueue[K]) GetCount() int {
	return p.size
}

func (p *MapQueue[K]) Exists(key K) bool {
	if _, ok := p.dict[key]; !ok {
		return false
	}
	return true
}

func (p *MapQueue[K]) Front() (K, any) {
	if p.size == 0 {
		var temp K
		panic(fmt.Sprintf("MapQueue[%#v] is empty", reflect.TypeOf(temp).String()))
	}
	result := (p.head.value).(struct {k K; v any})
	return result.k, result.v
}

func (p *MapQueue[K]) Back() (K, any) {
	if p.size == 0 {
		var temp K
		panic(fmt.Sprintf("MapQueue[%#v] is empty", reflect.TypeOf(temp).String()))
	}
	result := (p.tail.value).(struct {k K; v any})
	return result.k, result.v
}

func (p *MapQueue[K]) PopFront() {
	if p.head == nil {
		return
	}
	result := (p.head.value).(struct{k K; v any})
	p.head = p.head.next
	delete(p.dict, result.k)
	p.size--
}

func (p *MapQueue[K]) Set(key K, value any, auto_eliminate bool) bool {
	if _, ok := p.dict[key]; ok {
		p.Delete(key)
	}
	if p.size >= p.max_size {
		if auto_eliminate == false {
			return false
		}
		p.PopFront()
	}
	node := &Node {value: struct {k K; v any} {key, value}}
	node.prev = p.tail
	if p.size == 0 {
		p.head = node
	} else {
		p.tail.next = node
	}
	p.tail = node
	p.dict[key] = node
	p.size++
	return true
}

func (p *MapQueue[K]) Get(key K, move_to_tail bool) (any, bool) {
	node, ok := p.dict[key]
	if !ok {
		return nil, false
	}
	if p.size > 0 && move_to_tail {
		p.detach_to_queue(node)
		p.tail.next = node
		node.prev = p.tail
		node.next = nil
		p.tail = node
	}
	result := (node.value).(struct {k K; v any})
	return result.v, true
}

func (p *MapQueue[K]) GetAndDelete(key K) (any, bool) {
	node, ok := p.dict[key]
	if !ok {
		return nil, false
	}
	p.Delete(key)
	result := (node.value).(struct {k K; v any})
	return result.v, true
}

func (p *MapQueue[K]) Delete(key K) bool {
	node, ok := p.dict[key]
	if !ok {
		return false
	}
	if node.prev == nil {
		p.head = node.next
		p.head.prev = nil
	} else if node.next == nil {
		prev := node.prev
		prev.next = nil
		p.tail = prev
	} else {
		prev := node.prev
		next := node.next
		prev.next = next
		next.prev = prev
	}
	delete(p.dict, key)
	p.size--
	return true
}

func (p *MapQueue[K]) Keys() []K {
	arr := make([]K, 10)
	for k, _ := range p.dict {
		arr = append(arr, k)
	}
	return arr
}

func (p *MapQueue[K]) Values() []interface{} {
	arr := make([]interface{}, 10)
	for _, v := range p.dict {
		arr = append(arr, v)
	}
	return arr
}

// 回调函数返回值说明：
// [1] 查询是否停止
// [2] 查询是否命中
func (p *MapQueue[K]) Search(f func(K, interface{}) (bool, bool)) map[K]interface{} {
	return p.search_and_remove(f, false)
}

func (p *MapQueue[K]) Remove(f func(K, interface{}) (bool, bool)) map[K]interface{} {
	return p.search_and_remove(f, true)
}

func (p *MapQueue[K]) search_and_remove(f func(K, interface{}) (bool, bool), remove bool) map[K]interface{} {
	result := make(map[K]interface{})
	if p.size == 0 {
		return result
	}
	node := p.head
	for {
		if node == nil {
			break
		}
		cvalue := (node.value).(struct {k K; v any})
		stop, ok := f(cvalue.k, cvalue.v)
		if ok {
			result[cvalue.k] = cvalue.v
		}
		if stop {
			break
		}
		if ok && remove {
			p.Delete(cvalue.k)
		}
		node = node.next
	}
	return result
}

func (p *MapQueue[K]) detach_to_queue(node *Node) {
	if node.prev == nil {
		p.head = node.next
		p.head.prev = nil
	} else if node.next == nil {
		prev := node.prev
		prev.next = nil
		p.tail = prev
	} else {
		prev := node.prev
		next := node.next
		prev.next = next
		next.prev = prev
	}
}
