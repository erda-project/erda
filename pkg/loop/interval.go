// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

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
