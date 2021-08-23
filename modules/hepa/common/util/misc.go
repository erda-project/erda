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

package util

import (
	"runtime/debug"
	"sort"

	log "github.com/sirupsen/logrus"
)

func UniqStringSlice(slice []string) []string {
	sort.Strings(slice)
	i := 0
	for j := 1; j < len(slice); j++ {
		if slice[i] == slice[j] {
			continue
		} else {
			i++
			slice[i] = slice[j]
		}
	}
	size := i + 1
	if size > len(slice) {
		return slice
	} else {
		return slice[:size]
	}
}

func DoRecover() {
	if r := recover(); r != nil {
		log.Errorf("recovered from: %+v ", r)
		debug.PrintStack()
	}
}
