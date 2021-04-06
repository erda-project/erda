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

// 原本的parser会把 global env 塞到各个 service 的 env 去
package diceyml

type EnvVisitor struct {
	DefaultVisitor
	globalEnv map[string]string
}

func NewEnvVisitor() DiceYmlVisitor {
	return &EnvVisitor{}
}

func (o *EnvVisitor) VisitObject(v DiceYmlVisitor, obj *Object) {
	override(obj.Envs, &o.globalEnv)
	obj.Services.Accept(v)
}

func (o *EnvVisitor) VisitService(v DiceYmlVisitor, obj *Service) {
	if obj.Envs == nil {
		obj.Envs = map[string]string{}
	}
	for k, v := range o.globalEnv {
		if _, ok := obj.Envs[k]; !ok {

			obj.Envs[k] = v
		}
	}
}

func ExpandGlobalEnv(obj *Object) {
	visitor := NewEnvVisitor()
	obj.Accept(visitor)
}
