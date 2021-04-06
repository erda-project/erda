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

// 找出依赖却不存在的 service
package diceyml

import (
	"fmt"
)

type FindLackServiceVisitor struct {
	DefaultVisitor
	lack ValidateError
}

func NewFindLackServiceVisitor() DiceYmlVisitor {
	return &FindLackServiceVisitor{
		lack: ValidateError{},
	}
}

func (o *FindLackServiceVisitor) VisitServices(v DiceYmlVisitor, obj *Services) {
	for srvName, srv := range *obj {
		for _, name := range srv.DependsOn {
			if _, ok := (*obj)[name]; !ok {
				o.lack[yamlHeaderRegexWithUpperHeader([]string{"services", srvName}, "depends_on")] = fmt.Errorf("found lacked services: %v", name)
			}
		}
	}
}

func FindLackService(obj *Object) ValidateError {
	visitor := NewFindLackServiceVisitor()
	obj.Accept(visitor)
	return visitor.(*FindLackServiceVisitor).lack
}
