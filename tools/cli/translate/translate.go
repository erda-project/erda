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

package translate

import (
	"strings"
)

var items = map[string]string{
	"session not exist":            "session expired, please login again",
	"lack of required auth header": "not login yet, please login first",
	//"failed to request http, status-code 404, content-type text/plain; charset=utf-8, raw body not found path": "http url interface not match, please check the version of CLI and dice openAPI server",
}

func Translate(err error) string {
	errStr := err.Error()
	for item, r := range items {
		if strings.Contains(errStr, item) {
			return r
		}
	}
	return errStr
}
