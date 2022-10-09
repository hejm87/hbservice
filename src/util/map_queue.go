package util

import (
	"errors"
)

type any = interface {}

type Node struct {
	prev	*Node
	next	*Node
	value	any
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

func (p *MapQueue[K]) Top() any {
	if p.head == nil {
		return nil
	}
	result := (p.head.value).(struct{k K; v any})
	return result.v
}

func (p *MapQueue[K]) Pop() {
	if p.head == nil {
		return
	}
	result := (p.head.value).(struct{k K; v any})
	p.head = p.head.next
	delete(p.dict, result.k)
	p.size--
}

func (p *MapQueue[K]) Push(key K, value any, replace bool) error {
	add := 0
	if replace == false {
		add = 1
	}
	if p.size + add > p.max_size {
		return errors.New("exceed max size")
	}
	if _, ok := p.dict[key]; ok {
		if replace == true {
			p.Delete(key)
		} else {
			return errors.New("key already exists")
		}
	}
	node := &Node {value: struct {k K; v any} {key, value}}
	node.next = p.head
	p.head = node
	p.dict[key] = node
	p.size++
	return nil
}

func (p *MapQueue[K]) Get(key K) (any, bool) {
	node, ok := p.dict[key]
	if !ok {
		return nil, false
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
	if node := p.remove_to_queue(key); node != nil {
		p.size--
		delete(p.dict, key)
		return true
	}
	return false
}

func (p *MapQueue[K]) Refresh(key K) bool {
	if node := p.remove_to_queue(key); node != nil {
		node.next = p.head
		p.head = node
		p.dict[key] = node
		return true
	}
	return false
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

func (p *MapQueue[K]) remove_to_queue(key K) *Node {
	if node, ok := p.dict[key]; ok {
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
		return node
	}
	return nil
}
