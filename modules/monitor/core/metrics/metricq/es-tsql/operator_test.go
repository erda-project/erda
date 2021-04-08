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

package tsql

import (
	"reflect"
	"testing"
	"time"
)

type testItem struct {
	a, b    interface{}
	op      Operator
	reverse bool
	r       interface{}
	err     error
}

var now = time.Now()

var testList = []*testItem{
	&testItem{
		a:       1,
		op:      ADD,
		b:       2,
		reverse: true,
		r:       int64(3),
	},
	&testItem{
		a:       float32(1),
		op:      ADD,
		b:       2,
		reverse: true,
		r:       float64(3),
	},
	&testItem{
		a:       uint32(1),
		op:      ADD,
		b:       2,
		reverse: true,
		r:       uint64(3),
	},
	&testItem{
		a:       true,
		op:      ADD,
		b:       uint64(2),
		reverse: true,
		r:       uint64(3),
	},
	&testItem{
		a:       time.Second,
		op:      ADD,
		b:       int64(time.Second),
		reverse: true,
		r:       time.Duration(2 * time.Second),
	},
	&testItem{
		a:       now,
		op:      ADD,
		b:       time.Hour,
		reverse: true,
		r:       now.Add(time.Hour),
	},
	&testItem{
		a:       now,
		op:      ADD,
		b:       int64(time.Hour),
		reverse: true,
		r:       now.Add(time.Hour),
	},
	&testItem{
		a:       int64(time.Hour),
		op:      ADD,
		b:       now,
		reverse: true,
		r:       now.Add(time.Hour),
	},
	&testItem{
		a:  now.Add(time.Hour),
		op: SUB,
		b:  now,
		r:  time.Hour,
	},
	&testItem{
		a:       2,
		op:      MUL,
		b:       time.Hour,
		reverse: true,
		r:       2 * time.Hour,
	},
	&testItem{
		a:       "2",
		op:      MUL,
		b:       time.Hour,
		reverse: true,
		r:       2 * time.Hour,
	},
	&testItem{
		a:       "2",
		op:      MUL,
		b:       2,
		reverse: true,
		r:       float64(4),
	},
}

func TestParseTree_operateValues(t *testing.T) {
	for _, item := range testList {
		if !checkTestItem(t, item.a, item.op, item.b, item.r, item.err) {
			return
		}
		if item.reverse {
			if !checkTestItem(t, item.b, item.op, item.a, item.r, item.err) {
				return
			}
		}
	}
}

func checkTestItem(t *testing.T, a interface{}, op Operator, b interface{}, ret interface{}, e error) bool {
	r, err := OperateValues(a, op, b)
	if err != nil {
		if e == nil || err.Error() != e.Error() {
			t.Fatalf("%v %s %v -> unexpected error: %s", a, op, b, err)
		}
		return false
	}
	if reflect.TypeOf(r) != reflect.TypeOf(ret) {
		t.Fatalf("%v %s %v -> unexpected type: %s", a, op, b, reflect.TypeOf(r))
		return false
	}
	if r != ret {
		t.Fatalf("%v %s %v -> unexpected value: %s", a, op, b, r)
		return false
	}
	return true
}
