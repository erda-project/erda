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
