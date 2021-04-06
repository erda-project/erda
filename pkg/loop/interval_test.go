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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLoop_CalculateInterval(t *testing.T) {
	// 2, 4, 8, 10, 10
	l := New(WithMaxTimes(math.MaxUint64), WithDeclineLimit(10*time.Second), WithDeclineRatio(2))

	interval := l.CalculateInterval(0)
	assert.Equal(t, time.Second*2, interval)

	interval = l.CalculateInterval(1)
	assert.Equal(t, time.Second*4, interval)

	interval = l.CalculateInterval(2)
	assert.Equal(t, time.Second*8, interval)

	interval = l.CalculateInterval(3)
	assert.Equal(t, time.Second*10, interval)

	interval = l.CalculateInterval(4)
	assert.Equal(t, time.Second*10, interval)
}
