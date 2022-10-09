package util

import (
	"time"
	"sync"
	"errors"
)

const (
	TIMER_DELAY			int = 1
	TIMER_PERIOD		int = 2
)

var (
	instance	*TimerMgr
	once		sync.Once
)

func GetTimerMgrInstance() *TimerMgr {
	once.Do(func() {
		instance = new(TimerMgr)
	})
	return instance
}

type TimerInfo struct {
	id				string			// 定时器id
	ntype			int				// 定时器类型：TIMER_DELAY, TIMER_PERIOD
	timer			interface {}	// 定时器对象
	value			interface {}	// 定时器传参
}

type TimerMgr struct {
	timers			map[string]*TimerInfo
	sync.Mutex
}

type TimerCallback func(id string, value interface {})

func (p *TimerMgr) SetTimerDelay(id string, value interface {}, delay int, cb TimerCallback) error {
	timer := time.NewTimer(time.Duration(delay) * time.Second)
	ti := &TimerInfo {
		id:			id,
		ntype:		TIMER_DELAY,
		timer:		timer,
		value:		value,
	}
	if err := p.add_timer(ti); err != nil {
		return err
	}
	go func() {
		if delay > 0 {
			time.Sleep(time.Duration(delay) * time.Second)
		}
		<-timer.C
		cb(ti.id, ti.value)
	} ()
	return nil
}

func (p *TimerMgr) SetTimerPeriod(id string, value interface {}, delay int, period int, cb TimerCallback) error {
	timer := time.NewTicker(time.Duration(period) * time.Second)
	ti := &TimerInfo {
		id:			id,
		ntype:		TIMER_PERIOD,
		timer:		timer,
		value:		value,
	}
	if err := p.add_timer(ti); err != nil {
		return err
	}
	go func() {
		if delay > 0 {
			time.Sleep(time.Duration(delay) * time.Second)
		}
		for {
			select {
			case <-timer.C:
				cb(ti.id, ti.value)
			}
		}
	} ()
	return nil
}

func (p *TimerMgr) IsExistsTimer(id string) bool {
	p.Lock()
	defer p.Unlock()
	_, ok := p.timers[id]
	if ok {
		return true
	}
	return false
}

func (p *TimerMgr) CancelTimer(id string) error {
	var ok bool
	var ti *TimerInfo = nil
	p.Lock()
	ti, ok = p.timers[id]
	if !ok {
		return errors.New("no exists timer")
	}
	delete(p.timers, id)
	p.Unlock()
	if ti.ntype == TIMER_DELAY {
		ti.timer.(*time.Timer).Stop()
	} else {
		ti.timer.(*time.Ticker).Stop()
	}
	return nil
}

func (p *TimerMgr) add_timer(t *TimerInfo) error {
	p.Lock()
	defer p.Unlock()
	_, ok := p.timers[t.id]
	if ok {
		return errors.New("exists timer")
	}
	return nil
}
