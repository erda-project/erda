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
	"fmt"
)

type CheckEnvSizeVisitor struct {
	DefaultVisitor
	collectErrors ValidateError
}

func NewCheckEnvSizeVisitor() DiceYmlVisitor {
	return &CheckEnvSizeVisitor{
		collectErrors: ValidateError{},
	}
}

func (o *CheckEnvSizeVisitor) VisitObject(v DiceYmlVisitor, obj *Object) {
	for k, v := range obj.Envs {
		if toolong(v) {
			o.collectErrors[yamlHeaderRegexWithUpperHeader([]string{"envs"}, k)] = fmt.Errorf("global env too long(more than 20000 chars, ~20kb): [%s]", k)
		}
	}
}

func (o *CheckEnvSizeVisitor) VisitService(v DiceYmlVisitor, obj *Service) {
	for k, v := range obj.Envs {
		if toolong(v) {
			o.collectErrors[yamlHeaderRegexWithUpperHeader([]string{o.currentService, "envs"}, k)] = fmt.Errorf("env too long(more than 20000 chars, ~20kb): %s", k)
		}
	}
}

func toolong(s string) bool {
	return len(s) > 20000 // ~20kb
}

func CheckEnvSize(obj *Object) ValidateError {
	visitor := NewCheckEnvSizeVisitor()
	obj.Accept(visitor)
	return visitor.(*CheckEnvSizeVisitor).collectErrors
}
