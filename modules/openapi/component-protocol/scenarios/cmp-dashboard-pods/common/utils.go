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

package common

import (
	"encoding/json"
	"reflect"
	"sort"
)

// Transfer transfer a to b with json, kind of b must be pointer
func Transfer(a, b interface{}) error {
	if reflect.ValueOf(b).Kind() != reflect.Ptr {
		return PtrRequiredErr
	}
	if a == nil {
		return NothingToBeDoneErr
	}
	aBytes, err := json.Marshal(a)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(aBytes, b); err != nil {
		return err
	}
	return nil
}

// ConvertToMap transfer any struct to map
func ConvertToMap(obj interface{}) (map[string]interface{}, error) {
	out := make(map[string]interface{})
	t := reflect.TypeOf(obj)
	v := reflect.ValueOf(obj)

	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return nil, TypeNotAvailableErr
	}
	var (
		fieldData interface{}
		err       error
	)
	for i := 0; i < t.NumField(); i++ {
		field := v.Field(i)
		switch field.Kind() {
		case reflect.Struct:
			fallthrough
		case reflect.Ptr:
			if fieldData, err = ConvertToMap(field.Interface()); err != nil {
				return nil, err
			}
		default:
			fieldData = field.Interface()
		}
		out[t.Field(i).Name] = fieldData
	}
	return out, nil
}

func GetPercent(a, b float64) int {
	return int(a * 100 / b)
}
func GetInt64Len(a int64) int {
	length := 0
	for a > 0 {
		length++
		a /= 10
	}
	return length
}

/* ResetNumberBase
* e.g. : 20 100 to 2 10 , 0.1 1000 to 1 10000
 */
func ResetNumberBase(a, b float64) (float64, float64) {
	if a <= 0 || b <= 0 {
		return a, b
	}
	for a < 1 || b < 1 {
		a *= 10
		b *= 10
	}
	for a >= 10 || b >= 10 {
		a /= 10
		b /= 10
	}
	return a, b
}

// GetNodeStatus return node status from steve
func GetNodeStatus(status int) []SteveStatusEnum {
	return nodeStatusMap[status]
}

func GetPodStatus(status SteveStatusEnum) []SteveStatusEnum {
	return podStatusMap[status]
}

// SortByString sort by string value
func SortByString(data []interface{}, sortColumn string, order OrderEnum) {
	sort.Slice(data, func(i, j int) bool {
		a := reflect.ValueOf(data[i])
		b := reflect.ValueOf(data[j])
		if order == Asc {
			return a.FieldByName(sortColumn).String() > b.FieldByName(sortColumn).String()
		}else{
			return a.FieldByName(sortColumn).String() < b.FieldByName(sortColumn).String()
		}
	})
}

// SortByNode sort by node struct
func SortByNode(data []interface{}, sortColumn string,  order OrderEnum) {
	sort.Slice(data, func(i, j int) bool {
		a := reflect.ValueOf(data[i])
		b := reflect.ValueOf(data[j])
		if order ==  Asc {
			return a.FieldByName(sortColumn).FieldByName("Value").String() > b.FieldByName(sortColumn).FieldByName("Value").String()
		}else{
			return a.FieldByName(sortColumn).FieldByName("Value").String() < b.FieldByName(sortColumn).FieldByName("Value").String()
		}
	})
}

// SortByDistribution sort by percent
func SortByDistribution(data []interface{}, sortColumn string,  order OrderEnum) {
	sort.Slice(data, func(i, j int) bool {
		a := reflect.ValueOf(data[i])
		b := reflect.ValueOf(data[j])
		if order == Asc {
			return a.FieldByName(sortColumn).FieldByName("Value").FieldByName("Percent").Float() >
				b.FieldByName(sortColumn).FieldByName("Value").FieldByName("Percent").Float()
		}
		return a.FieldByName(sortColumn).FieldByName("Value").FieldByName("Percent").Float() <
			b.FieldByName(sortColumn).FieldByName("Value").FieldByName("Percent").Float()
	})
}
