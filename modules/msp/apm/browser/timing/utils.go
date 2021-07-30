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

package timing

import (
	"strconv"
)

func parseInt64(value string, def int64) int64 {
	if num, err := strconv.ParseInt(value, 10, 64); err == nil {
		return num
	}
	return def
}

func parseInt64WithRadix(value string, def int64, radix int) int64 {
	if num, err := strconv.ParseInt(value, radix, 64); err == nil {
		return num
	}
	return def
}
