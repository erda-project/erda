// Copyright (c) 2021 Terminus, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package addon

import (
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/erda-project/erda/apistructs"
)

func TestUnixTimeFormat(t *testing.T) {
	fmt.Println(time.Now().Format("2006-01-02")) //2018-7-15 15:23:00
}

func TestOne(t *testing.T) {
	a := 2.0000
	b := 1.60000
	c := 0.30000
	fmt.Println(Smaller(a-b, c))
}

func Smaller(a, b float64) bool {
	return math.Max(a, b) == b && math.Abs(a-b) > 0.00001
}

func TestSplitN(t *testing.T) {
	image := "addon-registry.default.svc.cluster.local:5000/terminus-customer-engagement/acl-addon-demo:migration-1584515323233379060"
	ss := strings.SplitN(image, "/", 2)
	if len(ss) == 2 {
		repo := strings.Split(ss[1], ":")[0]
		var repoTag string
		if strings.Contains(ss[1], ":") {
			repoTag = strings.Split(ss[1], ":")[1]
		} else {
			repoTag = "latest"
		}
		fmt.Printf("repo: %s, repoTag: %s", repo, repoTag)
	}

}

func TestSubTime(t *testing.T) {
	t1 := "2019-12-01 18:00:30"
	timeTemplate1 := "2006-01-02 15:04:05" //常规类型
	stamp, _ := time.ParseInLocation(timeTemplate1, t1, time.Local)

	now := time.Now()
	subM := now.Sub(stamp)
	fmt.Println(subM.Minutes(), "分钟")
}

func TestTransPlan(t *testing.T) {
	var addons = make([]apistructs.AddonCreateItem, 2)
	addons[0] = apistructs.AddonCreateItem{
		Name: "1111",
		Type: "mysql",
		Plan: "small",
	}
	addons[1] = apistructs.AddonCreateItem{
		Name: "222",
		Type: "redis",
		Plan: "small",
	}
	fmt.Printf("-------%+v", addons)
	fmt.Println("")
	addonsss := transPlan(&addons)
	fmt.Printf("=======%+v", addonsss)
}

func transPlan(addons *[]apistructs.AddonCreateItem) *[]apistructs.AddonCreateItem {
	if len(*addons) == 0 {
		return nil
	}
	var addon_result = make([]apistructs.AddonCreateItem, 0, len(*addons))
	for _, v := range *addons {
		addon_item := apistructs.AddonCreateItem{
			Name: v.Name,
			Type: v.Type,
			Plan: v.Plan,
		}
		switch v.Plan {
		case "large", apistructs.AddonUltimate:
			addon_item.Plan = apistructs.AddonUltimate
		case "medium", apistructs.AddonProfessional:
			addon_item.Plan = apistructs.AddonProfessional
		case "small", apistructs.AddonBasic:
			addon_item.Plan = apistructs.AddonBasic
		default:
			addon_item.Plan = apistructs.AddonBasic
		}
		addon_result = append(addon_result, addon_item)
	}
	return &addon_result
}

func TestOptionsMap(t *testing.T) {
	fmt.Println(float64(895))
	fmt.Println(strconv.ParseFloat(fmt.Sprintf("%.2f", float64(0.3)), 64))
}

func TestConfigMap(t *testing.T) {
	var configMap = map[string]interface{}{}
	configMap["MYSQL_HOST"] = "123456"
	configMap["MYSQL_PORT"] = 3306.00
	fmt.Println(reflect.TypeOf(configMap["MYSQL_PORT"]).String())
}

func TestBuild(t *testing.T) {
	//fmt.Print(fmt.Sprintf("%.f", 2048*0.7))
	aa := StructToMap(apistructs.AddonCreateOptions{
		ClusterName: "terminus-dev",
		OrgName:     "terminus",
	}, 0, "json")

	fmt.Printf("sssssss----%+v", &aa)

}

func StructToMap(data interface{}, depth int, tag ...string) map[string]interface{} {
	m := make(map[string]interface{})
	values := reflect.ValueOf(data)
	types := reflect.TypeOf(data)
	for types.Kind() == reflect.Ptr {
		values = values.Elem()
		types = types.Elem()
	}
	num := types.NumField()
	depth = depth - 1
	if len(tag) <= 0 || tag[0] == "" {
		if depth == -1 {
			for i := 0; i < num; i++ {
				t := types.Field(i)
				v := values.Field(i)
				if v.CanInterface() {
					m[t.Name] = v.Interface()
				}
			}
		} else {
			for i := 0; i < num; i++ {
				t := types.Field(i)
				v := values.Field(i)
				v_struct := v
				v_struct_ptr := v
				for v_struct.Kind() == reflect.Ptr {
					v_struct_ptr = v_struct
					v_struct = v_struct.Elem()
				}
				if v_struct.Kind() == reflect.Struct && v_struct_ptr.CanInterface() {
					m[t.Name] = StructToMap(v_struct_ptr.Interface(), depth, tag[0])
				} else {
					if v.CanInterface() {
						m[t.Name] = v.Interface()
					}
				}
			}
		}
	} else {
		tagName := tag[0]
		if depth == -1 {
			for i := 0; i < num; i++ {
				t := types.Field(i)
				v := values.Field(i)
				tagVal := t.Tag.Get(tagName)
				if v.CanInterface() && tagVal != "" && tagVal != "-" {
					m[tagVal] = v.Interface()
				}
			}
		} else {
			for i := 0; i < num; i++ {
				t := types.Field(i)
				v := values.Field(i)
				tagVal := t.Tag.Get(tagName)
				if tagVal != "" && tagVal != "-" {
					v_struct := v
					v_struct_ptr := v
					for v_struct.Kind() == reflect.Ptr {
						v_struct_ptr = v_struct
						v_struct = v_struct.Elem()
					}
					if v_struct.Kind() == reflect.Struct && v_struct_ptr.CanInterface() {
						m[tagVal] = StructToMap(v_struct_ptr.Interface(), depth, tag[0])
						continue
					}
					if v.CanInterface() {
						m[tagVal] = v.Interface()
					}
				}
			}
		}
	}
	return m
}
