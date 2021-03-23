package loop

import (
	"math"
	"time"
)

// CalculateInterval 传入当前循环次数，计算间隔时间
func (l *Loop) CalculateInterval(loopedTimes uint64) time.Duration {
	// 间隔 等于 初始间隔 乘以 衰退比例的 loopedTimes 次方
	interval := time.Duration(float64(l.interval) * math.Pow(l.declineRatio, float64(loopedTimes+1)))
	// 间隔 不能大于衰退限制
	if interval > l.declineLimit {
		interval = l.declineLimit
	}
	// 间隔 不能小于0
	if interval < 0 {
		interval = time.Duration(math.Min(float64(l.declineLimit), float64(l.interval)))
	}
	return interval
}
