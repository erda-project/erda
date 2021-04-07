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

	"github.com/stretchr/testify/assert"
)

type testWalkType struct {
	a int `tagtagtag`
	b map[string]*bool
	c struct {
		d int
		f struct {
			g int
			h string
		}
	}
}

func TestBottomUpWalk(t *testing.T) {
	tp := reflect.TypeOf(testWalkType{})
	n := newNode(constructCtx{name: tp.Name()}, tp)
	BottomUpWalk(n, func(curr Node, children []Node) {
		fmt.Printf("%+v, %s\n", curr, curr.Name()) // debug print
		extra := curr.Extra()
		*extra = curr.Name()
		for _, c := range children {
			(*extra) = (*extra).(string) + (*c.Extra()).(string)
		}
	})
	assert.Equal(t, "testWalkTypeabcdfgh", (*n.Extra()).(string))
}
