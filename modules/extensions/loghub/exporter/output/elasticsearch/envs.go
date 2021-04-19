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

package elasticsearch

import (
	"os"
	"strings"
)

func setEnvDefault(key, val string) {
	if len(val) > 0 {
		_, ok := os.LookupEnv(key)
		if ok {
			return
		}
		os.Setenv(key, val)
	}
}

func init() {
	env := os.Getenv("MONITOR_LOG_OUTPUT_CONFIG")
	if len(env) > 0 {
		lines := strings.Split(env, " ")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if len(line) <= 0 {
				continue
			}
			kv := strings.SplitN(line, "=", 2)
			if len(kv) == 1 {
				if strings.HasPrefix(kv[0], "http://") || strings.HasPrefix(kv[0], "https://") {
					setEnvDefault("ES_URLS", kv[0])
				}
			} else if len(kv) == 2 {
				setEnvDefault(strings.ToUpper(kv[0]), kv[1])

			}
		}
	}
}
