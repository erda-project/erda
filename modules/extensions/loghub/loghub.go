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

package loghub

import (
	"os"
)

// Init .
func Init() {
	setEnvDefault("LOGS_ES_URL", "ES_URL")
	setEnvDefault("LOGS_ES_SECURITY_ENABLE", "ES_SECURITY_ENABLE")
	setEnvDefault("LOGS_ES_SECURITY_USERNAME", "ES_SECURITY_USERNAME")
	setEnvDefault("LOGS_ES_SECURITY_PASSWORD", "ES_SECURITY_PASSWORD")

	_, ok := os.LookupEnv("CONFIG_NAME")
	if !ok {
		val, ok := os.LookupEnv("MONITOR_LOG_OUTPUT")
		if ok {
			os.Setenv("CONFIG_NAME", "conf/monitor/extensions/loghub/output/"+val)
		}
	}
}

func setEnvDefault(key string, from ...string) {
	_, ok := os.LookupEnv(key)
	if ok {
		return
	}
	for _, item := range from {
		val, ok := os.LookupEnv(item)
		if ok {
			os.Setenv(key, val)
			break
		}
	}
}
