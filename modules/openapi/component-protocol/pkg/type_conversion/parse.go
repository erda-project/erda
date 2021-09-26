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

package type_conversion

import (
	"fmt"
	"strconv"
)

func InterfaceToUint64(value interface{}) (uint64, error) {
	if value == nil {
		return 0, fmt.Errorf("can not parse value")
	}

	switch value.(type) {
	case int:
		return uint64(value.(int)), nil
	case float64:
		return uint64(value.(float64)), nil
	case string:
		intValue, err := strconv.ParseInt(value.(string), 10, 64)
		if err != nil {
			return 0, err
		}
		return uint64(intValue), nil
	}
	return 0, fmt.Errorf("can not parse value")
}
