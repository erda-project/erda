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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInsertAddonOptions(t *testing.T) {
	obj := Object{
		AddOns: AddOns{"xxx": &AddOn{Plan: "mysql:small", Options: map[string]string{"op1": "op1v"}}, "yyy": &AddOn{Plan: "mysql:large", Options: map[string]string{"op2": "op2v"}}},
	}
	InsertAddonOptions(&obj, BaseEnv, "mysql", map[string]string{"op3": "op3v"})
	assert.True(t, obj.AddOns["xxx"].Options["op3"] == "op3v")
	assert.True(t, obj.AddOns["yyy"].Options["op3"] == "op3v")

}

func TestInsertAddonOptionsEnv(t *testing.T) {
	obj := Object{
		AddOns:       AddOns{"xxx": &AddOn{Plan: "mysql:small", Options: map[string]string{"op1": "op1v"}}, "yyy": &AddOn{Plan: "mysql:large", Options: map[string]string{"op2": "op2v"}}},
		Environments: EnvObjects{"test": &EnvObject{AddOns: AddOns{"xxx": &AddOn{Plan: "mysql:small", Options: map[string]string{"op1": "op1v"}}, "yyy": &AddOn{Plan: "mysql:large", Options: map[string]string{"op2": "op2v"}}}}},
	}
	InsertAddonOptions(&obj, TestEnv, "mysql", map[string]string{"op3": "op3v"})

	assert.True(t, obj.Environments["test"].AddOns["xxx"].Options["op3"] == "op3v")
	assert.True(t, obj.Environments["test"].AddOns["yyy"].Options["op3"] == "op3v")
	assert.True(t, obj.AddOns["yyy"].Options["op3"] == "")
	assert.True(t, obj.AddOns["xxx"].Options["op3"] == "")

}

func TestInsertAddonOptionsWrongEnv(t *testing.T) {
	obj := Object{
		AddOns: AddOns{"xxx": &AddOn{Plan: "mysql:small", Options: map[string]string{"op1": "op1v"}}, "yyy": &AddOn{Plan: "mysql:large", Options: map[string]string{"op2": "op2v"}}},
	}
	InsertAddonOptions(&obj, TestEnv, "mysql", map[string]string{"op3": "op3v"})
	assert.False(t, obj.AddOns["xxx"].Options["op3"] == "op3v")
	assert.False(t, obj.AddOns["yyy"].Options["op3"] == "op3v")

}
