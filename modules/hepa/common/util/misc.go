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
