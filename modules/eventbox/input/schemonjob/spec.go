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

package schemonjob

type StatusForEventbox struct {
	// runtime namespace
	Namespace string `json:"namespace"`
	// runtime name
	Name string `json:"name"`
	// the "nofity" get the status and post it to this url
	Addr []string `json:"addrs"`
	// runtime status
	Status string `json:"status,omitempty"`
	// 扩展字段，比如存储runtime下每个服务的名字及状态
	More map[string]string `json:"more,omitempty"`
}

// true if same
func Diff(j1, j2 *StatusForEventbox) bool {
	if j1.Status != j2.Status || !mapDiff(j1.More, j2.More) || !sliceDiff(j1.Addr, j2.Addr) {
		return false
	}
	return true
}

// true if same
func mapDiff(m1, m2 map[string]string) bool {
	for k1, v1 := range m1 {
		v2, ok := m2[k1]
		if !ok {
			return false
		}
		if v1 != v2 {
			return false
		}
	}
	return true
}

// true if same
func sliceDiff(s1, s2 []string) bool {
	for _, v1 := range s1 {
		flag := false
		for _, v2 := range s2 {
			if v1 == v2 {
				flag = true
				break
			}
		}
		if !flag {
			return false
		}
	}
	return true
}
