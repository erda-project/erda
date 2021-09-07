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

package common

import (
	"encoding/json"
	"fmt"
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

func GetPercent(a, b float64) string {
	return fmt.Sprintf("%.1f", a*100/b)
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

// SortByString sort by string value
func SortByString(data []interface{}, sortColumn string, asc bool) {
	sort.Slice(data, func(i, j int) bool {
		a := reflect.ValueOf(data[i])
		b := reflect.ValueOf(data[j])
		if asc {
			return a.FieldByName(sortColumn).String() > b.FieldByName(sortColumn).String()
		}
		return a.FieldByName(sortColumn).String() < b.FieldByName(sortColumn).String()

	})
}

// SortByNode sort by node struct
func SortByNode(data []interface{}, sortColumn string, asc bool) {
	sort.Slice(data, func(i, j int) bool {
		a := reflect.ValueOf(data[i])
		b := reflect.ValueOf(data[j])
		if asc {
			return a.FieldByName(sortColumn).FieldByName("Value").String() > b.FieldByName(sortColumn).FieldByName("Value").String()
		}
		return a.FieldByName(sortColumn).FieldByName("Value").String() < b.FieldByName(sortColumn).FieldByName("Value").String()
	})
}

// SortByDistribution sort by percent
func SortByDistribution(data []interface{}, sortColumn string, asc bool) {
	sort.Slice(data, func(i, j int) bool {
		a := reflect.ValueOf(data[i])
		b := reflect.ValueOf(data[j])
		if asc {
			return a.FieldByName(sortColumn).FieldByName("Value").FieldByName("Percent").Float() >
				b.FieldByName(sortColumn).FieldByName("Value").FieldByName("Percent").Float()
		}
		return a.FieldByName(sortColumn).FieldByName("Value").FieldByName("Percent").Float() <
			b.FieldByName(sortColumn).FieldByName("Value").FieldByName("Percent").Float()
	})
}
