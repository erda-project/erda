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

type CompatibleExposeVisitor struct {
	DefaultVisitor
}

func NewCompatibleExposeVisitor() DiceYmlVisitor {
	return &CompatibleExposeVisitor{}
}

func (o *CompatibleExposeVisitor) VisitServices(v DiceYmlVisitor, obj *Services) {
	for _, v := range *obj {
		if len(v.Expose) > 0 && len(v.Ports) > 0 {
			v.Ports[0].Expose = true
		}
	}
}

func CompatibleExpose(obj *Object) {
	visitor := NewCompatibleExposeVisitor()
	obj.Accept(visitor)
}
