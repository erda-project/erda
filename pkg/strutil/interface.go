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

package strutil

import (
	"fmt"
	"strconv"
)

// String convert interface to string
func String(i interface{}) string {
	if i == nil {
		return ""
	}
	switch i.(type) {
	case int:
		return strconv.Itoa(i.(int))
	case int8:
		return strconv.FormatInt(int64(i.(int8)), 10)
	case int32:
		return strconv.FormatInt(int64(i.(int32)), 10)
	case int64:
		return strconv.FormatInt(int64(i.(int64)), 10)
	case uint:
		return strconv.FormatUint(uint64(i.(uint)), 10)
	case uint8:
		return strconv.FormatUint(uint64(i.(uint8)), 10)
	case uint32:
		return strconv.FormatUint(uint64(i.(uint32)), 10)
	case uint64:
		return strconv.FormatUint(uint64(i.(uint64)), 10)
	case float32:
		return strconv.FormatFloat(float64(i.(float32)), 'f', -1, 32)
	case float64:
		return strconv.FormatFloat(i.(float64), 'f', -1, 64)
	case []byte:
		return string(i.([]byte))
	case string:
		return i.(string)
	default:
		return fmt.Sprintf("%v", i)
	}
}
