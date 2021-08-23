// Copyright (c) 2021 Terminus, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
