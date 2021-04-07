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

package structparser

import (
	"fmt"
	"reflect"
	"testing"
	"time"
)

type testType struct {
	a *int `tagtagtag`
	b map[string]*bool
	c *struct {
		d **int
	}
}

func TestNewNode(t *testing.T) {
	tp := reflect.TypeOf(testType{})
	n := newNode(constructCtx{name: tp.Name()}, tp)
	fmt.Printf("%+v\n", n) // debug print
}

func TestCompress(t *testing.T) {
	tp := reflect.TypeOf(testType{})
	n := newNode(constructCtx{name: tp.Name()}, tp)
	fmt.Printf("%+v\n", n) // debug print
	nn := n.Compress()
	fmt.Printf("%+v\n", nn) // debug print
}

func TestTest(t *testing.T) {
	tp := reflect.TypeOf(time.Time{})
	n := newNode(constructCtx{}, tp)
	fmt.Printf("%+v\n", n.String()) // debug print
}
