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

package diceyml

import (
	"reflect"
)

type FindCycleVisitor struct {
	DefaultVisitor
	result bool
	chain  []string
}

func NewFindCycleVisitor() DiceYmlVisitor {
	return &FindCycleVisitor{}
}

func (o *FindCycleVisitor) VisitServices(v DiceYmlVisitor, obj *Services) {
	all := make([]string, 0, 10)
	nodes := make(map[string]map[string]struct{})
	for name, srv := range *obj {
		depends := make(map[string]struct{})
		for _, dependName := range srv.DependsOn {
			depends[dependName] = struct{}{}
		}
		nodes[name] = depends
		all = append(all, name)
	}
	for {
		removed := false

		for _, name := range all {
			depends, ok := nodes[name]
			if ok && len(depends) == 0 {
				delete(nodes, name)
				for _, depends := range nodes {
					delete(depends, name)
				}
				removed = true
			}
		}
		if len(nodes) == 0 {
			o.result = false
			return
		}
		if !removed {
			o.result = true
			o.chain = extractCycle(nodes)

			return
		}
	}
	o.result = false
	return
}

func extractCycle(nodes map[string]map[string]struct{}) []string {
	chain := make([]string, 0)
	name := pickOne(nodes)
	start := name
	for {
		chain = append(chain, name)
		deps := nodes[name]
		name = pickOne(deps)
		if name == start {
			chain = append(chain, name)
			return chain
		}
	}
}

func pickOne(m interface{}) string {
	keys := reflect.ValueOf(m).MapKeys()
	return keys[0].Interface().(string)
}

// if 'has' == false, 'chain' is meaningless
func FindCycle(obj *Object) (has bool, chain []string) {
	visitor := NewFindCycleVisitor()
	obj.Accept(visitor)
	has = visitor.(*FindCycleVisitor).result
	chain = visitor.(*FindCycleVisitor).chain
	return
}
