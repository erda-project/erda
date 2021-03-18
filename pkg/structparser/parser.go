package structparser

import (
	"reflect"
)

// Parse `structure` into structparser.Node
func Parse(structure interface{}) Node {
	return newNode(constructCtx{}, reflect.TypeOf(structure))
}
