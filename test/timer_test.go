//package timer_test
package hbservice_test

import (
	"time"
	"testing"
	"hbservice/src/util"
)

var g_test_delay_counter = 0
var g_test_period_counter = 0

func TestDelayTimer(t *testing.T) {
	tm := new(util.TimerMgr)	
	var value interface {} = &g_test_delay_counter
	if err := tm.SetTimerDelay("delay_timer", value, 2, delay_timer_cb); err != nil {
		t.Fatalf("SetTimerDelay error:%#v", err)
	}
	time.Sleep(time.Duration(3) * time.Second)
	if g_test_delay_counter != 1 {
		t.Fatalf("g_test_delay_counter != 1")
	}
}

func TestPeriodTimer(t *testing.T) {
	tm := new(util.TimerMgr)	
	var value interface {} = &g_test_period_counter
	if err := tm.SetTimerPeriod("period_timer", value, 0, 1, period_timer_cb); err != nil {
		t.Fatalf("SetTimerPeriod error:%#v", err)
	}
	time.Sleep(time.Duration(5) * time.Second)
	tm.CancelTimer("period_timer")
	if g_test_period_counter != 5 {
		t.Fatalf("g_test_period_counter != 5")
	}
}

func delay_timer_cb(id string, value interface {}) {
	p := value.(*int)
	*p = 1
}

func period_timer_cb(id string, value interface {}) {
	p := value.(*int)
	*p = *p + 1
}