package structparser

// BottomUpWalk 自底向上遍历 `node`
func BottomUpWalk(node Node, f func(curr Node, children []Node)) {
	switch v := node.(type) {
	case *BoolNode:
		f(node, nil)
	case *IntNode:
		f(node, nil)
	case *Int64Node:
		f(node, nil)
	case *FloatNode:
		f(node, nil)
	case *StringNode:
		f(node, nil)
	case *SliceNode:
		BottomUpWalk(v.value, f)
		f(v, []Node{v.value})
	case *MapNode:
		BottomUpWalk(v.key, f)
		BottomUpWalk(v.value, f)
		f(v, []Node{v.key, v.value})
	case *PtrNode:
		BottomUpWalk(v.inner, f)
		f(v, []Node{v.inner})
	case *StructNode:
		for _, field := range v.fields {
			BottomUpWalk(field, f)
		}
		f(v, v.fields)
	default:
		panic("unreachable")
	}
}
