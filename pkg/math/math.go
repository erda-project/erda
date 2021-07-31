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

package math

func AbsInt(x int) int {
	y := x >> 31
	return (x ^ y) - y
}

func AbsInt32(x int32) int32 {
	y := x >> 31
	return (x ^ y) - y
}

func AbsInt64(x int64) int64 {
	y := x >> 63
	return (x ^ y) - y
}
