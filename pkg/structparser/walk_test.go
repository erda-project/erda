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
