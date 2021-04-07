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

type ComposeVisitor struct {
	DefaultVisitor
	envObj *Object
	env    string
}

func NewComposeVisitor(envObj *Object, env string) DiceYmlVisitor {
	return &ComposeVisitor{envObj: envObj, env: env}
}

func (o *ComposeVisitor) VisitEnvObjects(v DiceYmlVisitor, obj *EnvObjects) {

	if o.envObj == nil {
		return
	}

	if (*obj) == nil {
		obj_ := EnvObjects(map[string]*EnvObject{})
		*obj = obj_
	}

	if e, ok := (*obj)[o.env]; !ok || e == nil {
		(*obj)[o.env] = new(EnvObject)
	}

	overrideIfNotZero(o.envObj.Envs, &(*obj)[o.env].Envs)
	overrideIfNotZero(o.envObj.Services, &(*obj)[o.env].Services)
	overrideIfNotZero(o.envObj.AddOns, &(*obj)[o.env].AddOns)

}

func Compose(obj, envObj *Object, env string) {
	visitor := NewComposeVisitor(envObj, env)

	obj.Accept(visitor)

}
