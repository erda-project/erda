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
	"reflect"
	"regexp"

	yaml "gopkg.in/yaml.v2"
)

func assignWithoutEmpty(p interface{}, src interface{}) {
	value := reflect.ValueOf(p).Elem()
	if !value.CanSet() {
		return
	}

	if v, ok := src.(string); ok && v != "" {
		value.Set(reflect.ValueOf(v))
	}
	if v, ok := src.([]string); ok && v != nil && len(v) != 0 {
		value.Set(reflect.ValueOf(v))
	}
	if v, ok := src.(int); ok && v != 0 {
		value.Set(reflect.ValueOf(v))
	}
	if v, ok := src.([]int); ok && v != nil && len(v) != 0 {
		value.Set(reflect.ValueOf(v))
	}
	if v, ok := src.(int64); ok && v != 0 {
		value.Set(reflect.ValueOf(v))
	}
	if v, ok := src.([]int64); ok && v != nil && len(v) != 0 {
		value.Set(reflect.ValueOf(v))
	}
	if v, ok := src.(float64); ok && v != 0 {
		value.Set(reflect.ValueOf(v))
	}
	if v, ok := src.([]float64); ok && v != nil && len(v) != 0 {
		value.Set(reflect.ValueOf(v))
	}
}

func isZero(v interface{}) bool {
	switch reflect.TypeOf(v).Kind() {
	case reflect.Int:
		return v.(int) == 0
	case reflect.Int64:
		return v.(int64) == 0
	case reflect.Float64:
		return v.(float64) == 0
	case reflect.String:
		return v.(string) == ""
	case reflect.Map:
		return reflect.ValueOf(v).Len() == 0
	case reflect.Slice:
		return reflect.ValueOf(v).Len() == 0
	default:
		panic("not support kind: " + reflect.TypeOf(v).Kind().String())
	}
}

func overrideIfNotZero(src, dst interface{}) {
	if isZero(src) {
		return
	}
	override(&src, dst)
}

func override(src, dst interface{}) {
	r, err := yaml.Marshal(src)
	if err != nil {
		panic("deepcopy marshal")
	}
	if err := yaml.Unmarshal(r, dst); err != nil {
		panic("deepcopy unmarshal")
	}
}

func CopyObj(src *Object) *Object {
	dst := new(Object)
	override(src, dst)
	return dst
}

func yamlHeaderRegex(head string) *regexp.Regexp {
	return regexp.MustCompile("(?m)^[[:blank:]]*" + head + "[[:blank:]]*:")
}

func yamlHeaderRegexWithUpperHeader(upper []string, head string) *regexp.Regexp {
	r := "(?sm)"
	for _, uphead := range upper {
		r += "(?:^[[:blank:]]*?" + uphead + "[[:blank:]]*?:.*?)"
	}
	r += "(^[[:blank:]]*?" + head + "[[:blank:]]*?:)"
	return regexp.MustCompile(r)
}

func ComposeIntPortsFromServicePorts(servicePorts []ServicePort) []int {
	var ports []int
	for _, port := range servicePorts {
		ports = append(ports, port.Port)
	}
	return ports
}
