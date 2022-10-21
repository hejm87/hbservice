package util

import (
	"sync"
)

type Result struct {
	err			error
	value		interface {}
}

type Results []chan Result

type MergeSearch[K comparable, V any] struct {
	caches			map[K] int
	waits			map[K] Results
	fun				func(K) (V, error)
	sync.Mutex
}

func NewMergeSearch[K comparable, V any](f func(K) (V, error)) *MergeSearch[K, V] {
	obj := &MergeSearch[K, V] {
		caches:		make(map[K] int),
		waits:		make(map[K] Results),
		fun:		f,	
	}
	return obj
}

func (p *MergeSearch[K, V]) Call(key K) (V, error) {
	var err error
	var value V
	p.Lock()
	if _, ok := p.caches[key]; !ok {
		p.caches[key] = 1
		p.Unlock()
		value, err = p.fun(key)

		var waits Results
		p.Lock()
		waits = p.waits[key]
		delete(p.caches, key)
		delete(p.waits, key)
		p.Unlock()

		if waits != nil {
			result := Result {
				err:	err,
				value:	value,
			}
			for _, w := range waits {
				w <-result
			}
		}
	} else {
		ch := make(chan Result)
		waits, ok := p.waits[key]
		if ok {
			p.waits[key] = append(waits, ch)
		} else {
			p.waits[key] = append(make(Results, 0), ch)
		}

		p.Unlock()
		resp := <-ch

		err = resp.err
		value = (resp.value).(V)
	}
	return value, err
}
