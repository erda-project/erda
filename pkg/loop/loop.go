// Package loop 提供了一个循环执行某个方法的功能，可以自定义执行时间等。
// 用法：
// l := loop.New()
// l.Do(方法)
package loop

import (
	"math"
	"time"
)

// Loop 定义循环执行任务的对象.
type Loop struct {
	maxTimes      uint64
	declineRatio  float64
	declineLimit  time.Duration
	interval      time.Duration
	lastSleepTime time.Duration
}

// Option 定义 Loop 对象的参数配置选项.
type Option func(*Loop)

// New 创建一个新的 Loop 实例.
func New(options ...Option) *Loop {
	loop := &Loop{
		interval:     time.Second,
		maxTimes:     math.MaxUint64,
		declineRatio: 1,
		declineLimit: 0,
	}

	for _, op := range options {
		op(loop)
	}

	loop.lastSleepTime = loop.interval

	return loop
}

// Do 循环的执行一个方法
// 这个被循环执行的方法有两个返回值，第一个返回值控制是否需要退出循环，第二为任务执行的错误信息.
func (l *Loop) Do(f func() (bool, error)) error {
	var i uint64
	for i = 0; i < l.maxTimes; i++ {
		abort, err := f()
		if abort {
			return err
		}
		if err != nil {
			// 暂停上次睡眠的时间乘以衰退比
			l.lastSleepTime = time.Duration(float64(l.lastSleepTime) * l.declineRatio)
			if l.declineLimit > 0 && l.lastSleepTime > l.declineLimit {
				l.lastSleepTime = l.declineLimit
			}
			time.Sleep(l.lastSleepTime)
			continue
		}

		// 成功执行 reset 暂停时间
		l.lastSleepTime = l.interval
		time.Sleep(l.lastSleepTime)
	}
	return nil
}

// WithMaxTimes 设置循环的最大次数.
func WithMaxTimes(n uint64) Option {
	return func(l *Loop) {
		l.maxTimes = n
	}
}

// WithDeclineRatio 设置衰退延迟的比例，默认是 1.
func WithDeclineRatio(n float64) Option {
	return func(l *Loop) {
		if n < 1 {
			return
		}
		l.declineRatio = n
	}
}

// WithDeclineLimit 设置衰退延迟的最大值，默认不限制最大值.
func WithDeclineLimit(t time.Duration) Option {
	return func(l *Loop) {
		if t < 0 {
			return
		}
		l.declineLimit = t
	}
}

// WithInterval 设置每次循环的间隔时间.
func WithInterval(t time.Duration) Option {
	return func(l *Loop) {
		if t < time.Millisecond {
			return
		}
		l.interval = t
	}
}
