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
