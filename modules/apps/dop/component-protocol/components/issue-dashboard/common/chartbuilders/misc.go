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

package chartbuilders

type counterItem struct {
	Name  string
	Value int
}

type counterList []counterItem

func (l counterList) Less(i, j int) bool { return l[i].Value > l[j].Value }
func (l counterList) Swap(i, j int)      { l[i], l[j] = l[j], l[i] }
func (l counterList) Len() int           { return len(l) }

func FixEmptyWord(em string) string {
	if em == "" {
		return "无"
	}
	return em
}
