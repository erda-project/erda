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

// 该 visitor 处理 diceyml 中一些字段的默认值
package diceyml

type DefaultValueVisitor struct {
	DefaultVisitor
}

func NewDefaultValueVisitor() DiceYmlVisitor {
	return &DefaultValueVisitor{}
}

func (o *DefaultValueVisitor) VisitResources(v DiceYmlVisitor, obj *Resources) {
	if obj.Network == nil {
		obj.Network = map[string]string{}
	}
	if _, ok := obj.Network["mode"]; !ok {
		obj.Network["mode"] = "container"
	}
}

func SetDefaultValue(obj *Object) {
	visitor := NewDefaultValueVisitor()
	obj.Accept(visitor)
}
