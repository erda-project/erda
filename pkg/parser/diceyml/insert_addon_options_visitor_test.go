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
