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
	"strings"
)

type InsertAddonOptionsVisitor struct {
	DefaultVisitor
	env           EnvType
	addonPlan     string
	options       map[string]string
	fromEnvObject bool
}

func NewInsertAddonOptionsVisitor(env EnvType, addonPlan string, options map[string]string) DiceYmlVisitor {
	return &InsertAddonOptionsVisitor{addonPlan: addonPlan, options: options, env: env}
}

func (o *InsertAddonOptionsVisitor) VisitAddOns(v DiceYmlVisitor, obj *AddOns) {
	if o.env != BaseEnv && !o.fromEnvObject {
		return
	}
	for _, v_ := range *obj {
		v_.Accept(v)
	}
}

func (o *InsertAddonOptionsVisitor) VisitAddOn(v DiceYmlVisitor, obj *AddOn) {
	splitted := strings.Split(obj.Plan, ":")

	if len(splitted) < 1 {
		return
	}

	if strings.TrimSpace(splitted[0]) == o.addonPlan {
		for k, v := range o.options {
			obj.Options[k] = v
		}
	}
}

func (o *InsertAddonOptionsVisitor) VisitEnvObjects(v DiceYmlVisitor, obj *EnvObjects) {
	if o.env == BaseEnv {
		return
	}
	for name, v_ := range *obj {
		if name == o.env.String() {
			v_.Accept(v)
		}
	}
}

func (o *InsertAddonOptionsVisitor) VisitEnvObject(v DiceYmlVisitor, obj *EnvObject) {
	o.fromEnvObject = true
	obj.AddOns.Accept(v)
	o.fromEnvObject = false
}

func InsertAddonOptions(obj *Object, env EnvType, addonPlan string, options map[string]string) {
	visitor := NewInsertAddonOptionsVisitor(env, addonPlan, options)
	obj.Accept(visitor)
}
