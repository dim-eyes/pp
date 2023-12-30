package timer

import (
	"pp/service/conn"
	"sync"
	"time"
)

var (
	timerMgr *TickTimerMgr
	onceMgr  sync.Once
)

func GetTickTimerMgr() *TickTimerMgr {
	onceMgr.Do(func() {
		if timerMgr == nil {
			timerMgr = &TickTimerMgr{}
			now := time.Now()
			timerMgr.timestamp = now.Unix() - int64(now.Second()) - int64(60*now.Minute())
		}
	})

	return timerMgr
}

// TickTimerMgr 系统消息管理
type TickTimerMgr struct {
	timestamp int64
}

// Timer 启动计时器
func (t *TickTimerMgr) Timer() {
	interval := time.Millisecond * 10
	timer := time.NewTimer(interval)
	var count int64
	for {
		select {
		case <-timer.C:
			count++
			if count%100 == 0 {
				t.Timer1s()
			}

			if count%3000 == 0 {
				t.Timer30s()
			}
			if count%6000 == 0 {
				t.Timer1min()
			}

			if count%30000 == 0 {
				t.Timer5min()
			}
			if count%360000 == 0 {
				t.Timer1hour()
			}

			timer.Reset(interval)
		}
	}
}

// Timer1s 1秒执行一次
func (t *TickTimerMgr) Timer1s() {
	go conn.GetGateClientMgr().Timer1s()
	now := time.Now()
	if now.Unix()-t.timestamp >= 3600 {
		t.everyHourTimer()
		nowTime := time.Now()
		t.timestamp = nowTime.Unix() - int64(nowTime.Second()) - int64(60*nowTime.Minute())
	}
}

// Timer1min 1分钟执行一次
func (t *TickTimerMgr) Timer1min() {
}

// Timer5min 5分钟执行一次
func (t *TickTimerMgr) Timer5min() {
}

// Timer30s 30秒执行一次
func (t *TickTimerMgr) Timer30s() {
}

// Timer1hour 1小时执行一次
func (t *TickTimerMgr) Timer1hour() {
}

// 整点调用
func (t *TickTimerMgr) everyHourTimer() {
	// go rankService.NewFortuneRank().Timer()
}
