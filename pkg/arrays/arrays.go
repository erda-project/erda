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

package arrays

func Distinct(array []string) []string {
	data := make(map[string]interface{})
	for _, v := range array {
		data[v] = nil
	}
	var result []string
	for k := range data {
		result = append(result, k)
	}
	return result
}

func Concat(array []string, arrays ...[]string) []string {
	for _, arr := range arrays {
		array = append(array, arr...)
	}
	return array
}
