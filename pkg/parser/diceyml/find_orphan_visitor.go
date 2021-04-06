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

// 找出没有被其他service依赖的后端service， 前端service不考虑（也就是有expose的service）
package diceyml

type FindOrphanVisitor struct {
	DefaultVisitor
	orphan []string
}

func NewFindOrphanVisitor() DiceYmlVisitor {
	return &FindOrphanVisitor{}
}

func (o *FindOrphanVisitor) VisitServices(v DiceYmlVisitor, obj *Services) {
	orphan := []string{}
	depends := map[string]struct{}{}
	for _, srv := range *obj {
		for _, name := range srv.DependsOn {
			depends[name] = struct{}{}
		}
	}
	for name, srv := range *obj {
		_, ok := depends[name]
		if !ok && srv.Expose == nil {
			orphan = append(orphan, name)
		}
	}
	o.orphan = orphan
}

func FindOrphan(obj *Object) []string {
	visitor := NewFindOrphanVisitor()
	obj.Accept(visitor)
	return visitor.(*FindOrphanVisitor).orphan
}
