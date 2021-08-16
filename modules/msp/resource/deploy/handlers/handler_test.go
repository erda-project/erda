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

package handlers

import (
	"encoding/json"
	"testing"

	"github.com/erda-project/erda/modules/msp/resource/utils"
)

type Obj struct {
	Name string `json:"name,omitempty"`
	Age  string `json:"age,omitempty"`
}

func TestJsonStringMarshal(t *testing.T) {
	var a = "1111"
	var b string
	utils.JsonConvertObjToType(a, &b)
	if a != b {
		t.Errorf("expect: %s, actual: %s", a, b)
	}
}

func TestJsonObjMarshal(t *testing.T) {
	obj := Obj{Name: "xiao ming", Age: "18"}
	bytes, _ := json.Marshal(obj)
	objStr := string(bytes)

	var unObj1 Obj
	utils.JsonConvertObjToType(obj, &unObj1)

	if obj.Name != unObj1.Name || obj.Age != unObj1.Age {
		t.Error("unmarshal from obj should same")
	}

	var unObj2 Obj
	utils.JsonConvertObjToType(objStr, &unObj2)

	if obj.Name != unObj2.Name || obj.Age != unObj2.Age {
		t.Error("unmarshal from str should same")
	}

	obj.Name = "xiao hong"
	utils.JsonConvertObjToType(obj, &unObj1)

	if obj.Name != unObj1.Name || obj.Age != unObj1.Age {
		t.Error("unmarshal from str should same")
	}

	var str string
	err := utils.JsonConvertObjToType(obj, &str)
	if err == nil {
		t.Error("obj should fail re-marshal to string")
	}

	var mapObj map[string]string
	utils.JsonConvertObjToType(objStr, &mapObj)

}

func TestMapJsonMarshal(t *testing.T) {
	var mapObj map[string]string
	err := utils.JsonConvertObjToType("", &mapObj)
	if err == nil {
		t.Error("convert empty string to map should fail")
	}
	if mapObj != nil {
		t.Error("when convert fail, map value should keep same")
	}

	mapIni := map[string]string{}
	err = utils.JsonConvertObjToType("", &mapIni)
	if err == nil {
		t.Error("convert empty string to map should fail")
	}
	if mapIni == nil {
		t.Error("when convert fail, map value should keep same")
	}
}

func TestMapAppend(t *testing.T) {
	map1 := map[string]string{"1": "1"}
	map2 := map[string]string{"2": "2"}
	utils.AppendMap(map1, map2)
	if len(map1) != 2 {
		t.Error("append should be in-place")
	}
}

func TestNewDefaultHandler(t *testing.T) {
	instance := NewDefaultHandler(nil, nil)
	if instance.Bdl == nil {
		t.Errorf("DefaultHandler.Bdl should not nil")
	}
}
