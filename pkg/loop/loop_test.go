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

//import (
//	"errors"
//	"fmt"
//	"testing"
//	"time"
//
//	"github.com/stretchr/testify/assert"
//)
//
//func TestLoopMaxTimes(t *testing.T) {
//	count := 0
//	l := New(WithMaxTimes(20))
//	l.Do(func() (bool, error) {
//		count += 1
//		return false, nil
//	})
//	assert.Equal(t, 20, count)
//}
//
//func TestLoopInterval(t *testing.T) {
//	l := New(WithMaxTimes(10), WithInterval(time.Second*2))
//
//	start := time.Now()
//	l.Do(func() (bool, error) {
//		return false, nil
//	})
//	duration := time.Now().Sub(start)
//	if duration < 20*time.Second {
//		t.Error("interval is not good working")
//	}
//}
//
//func TestLoopRaiot(t *testing.T) {
//	l := New(WithDeclineRatio(1.5), WithDeclineLimit(time.Second*10))
//	begin := time.Now()
//	l.Do(func() (bool, error) {
//		executeTime := time.Now()
//		fmt.Println("interval:", executeTime.Sub(begin))
//		begin = executeTime
//		return false, errors.New("effect")
//	})
//}
