// Copyright (c) 2021 Terminus, Inc.

// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later (AGPL), as published by the Free Software Foundation.

// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.

// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package env

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// GetHostIndex .
func GetHostIndex() string {
	hostname, err := os.Hostname()
	fmt.Println("hostname", hostname)
	if err == nil {
		idx := strings.LastIndex(hostname, "-")
		if idx > 0 {
			_, err = strconv.Atoi(hostname[idx+1:])
			if err == nil {
				return hostname[idx+1:]
			}
		}
	}
	return ""
}

// Override .
func Override() {
	idx := GetHostIndex()
	if len(idx) <= 0 {
		return
	}
	prefix := "N" + idx + "_"
	for _, kv := range os.Environ() {
		parts := strings.SplitN(kv, "=", 2)
		if len(parts) < 2 {
			continue
		}
		key := parts[0]
		val := parts[1]
		if strings.HasPrefix(key, prefix) && len(val) > 0 {
			nkey := key[len(prefix):]
			if len(os.Getenv(nkey)) <= 0 {
				os.Setenv(nkey, val)
			}
		}
	}
	fmt.Println("Override envs \n", strings.Join(os.Environ(), "\n"))
}
