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

// 检查 service name 合法性
package diceyml

import (
	"fmt"
	"regexp"
)

var serviceNameRegex = regexp.MustCompile("^((([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\\-]*[a-zA-Z0-9])\\.)*([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\\-]*[a-zA-Z0-9]))$")

// var serviceNameMaxLen = 14 // service name 最长长度

var envNameRegex = regexp.MustCompile("^([A-Za-z_][A-Za-z0-9_]*)$")

type ServiceNameVisitor struct {
	DefaultVisitor
	collectErrors ValidateError
}

func NewServiceNameVisitor() DiceYmlVisitor {
	return &ServiceNameVisitor{
		collectErrors: ValidateError{},
	}
}

func (o *ServiceNameVisitor) VisitService(v DiceYmlVisitor, obj *Service) {
	if !serviceNameRegex.MatchString(o.currentService) {
		o.collectErrors[yamlHeaderRegexWithUpperHeader([]string{"services"}, o.currentService)] = fmt.Errorf(
			"service name[%s] not match regex:\n %s, \n e.g: aa.bb-cc.dd-ee", o.currentService, serviceNameRegex)
	}

	for k := range obj.Envs {
		if !envNameRegex.MatchString(k) {
			o.collectErrors[yamlHeaderRegexWithUpperHeader([]string{"services", o.currentService, "envs"}, k)] = fmt.Errorf("env name [%s] not match regex:\n %s", k, envNameRegex)
		}
	}
}

func (o *ServiceNameVisitor) VisitObject(v DiceYmlVisitor, obj *Object) {
	for k := range obj.Envs {
		if !envNameRegex.MatchString(k) {
			o.collectErrors[yamlHeaderRegexWithUpperHeader([]string{"envs"}, k)] = fmt.Errorf("env name [%s] not match regex:\n %s", k, envNameRegex)
		}
	}
}

func ServiceNameCheck(obj *Object) ValidateError {
	visitor := NewServiceNameVisitor()
	obj.Accept(visitor)
	return visitor.(*ServiceNameVisitor).collectErrors
}
