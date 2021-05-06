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

package marathon

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDeepcopyGroup(t *testing.T) {
	g := Group{
		Id: "id",
		Apps: []App{
			{
				Env: map[string]string{
					"env1": "v1",
					"env2": "v2",
				},
				Ports: []int{1, 2, 3, 4},
			},
		},
	}
	g2, err := deepcopyGroup(&g)
	assert.Nil(t, err)
	assert.Equal(t, g, g2)
	g2.Apps[0].Ports = append(g2.Apps[0].Ports, 5)
	assert.Equal(t, []int{1, 2, 3, 4}, g.Apps[0].Ports)
}

func TestIsSameConstraints(t *testing.T) {
	c1 := []Constraint{
		{"a1", "a2", "a2"},
		{"b1", "b2", "b3"},
	}
	c2 := []Constraint{
		{"b1", "b2", "b3"},
		{"a1", "a2", "a2"},
	}
	assert.True(t, isSameConstraints(c1, c2))
}

func TestClassifyApps(t *testing.T) {
	groups := [][]string{{"a", "b"}, {"c", "d", "e"}}
	g := &Group{
		Apps: []App{{Id: "x/a"}, {Id: "x/b"}, {Id: "x/c"}, {Id: "x/d"}},
	}
	apps, err := classifyApps(groups, g)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(apps))
	assert.Equal(t, 4, len(apps[0])+len(apps[1]))

}
