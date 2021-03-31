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
