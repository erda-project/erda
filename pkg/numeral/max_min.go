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

package numeral

// MinFloat64 return the minimal num of nums.
//
// 1, 2, 3, 4, 5, 0 => 0
// 1, 2, 3, 4, 5, 0 (ignoreZero) => 1
func MinFloat64(nums []float64, ignoreZeroOpt ...bool) float64 {
	ignoreZero := false
	if len(ignoreZeroOpt) > 0 {
		ignoreZero = ignoreZeroOpt[0]
	}
	var r float64 = 0
	if len(nums) == 0 {
		return r
	}
	// set r first
	for _, num := range nums {
		if num == 0 && ignoreZero {
			continue
		}
		r = num
		break
	}
	// compare
	for _, num := range nums {
		if num == 0 && ignoreZero {
			continue
		}
		if num < r {
			r = num
			continue
		}
	}
	return r
}

// MaxFloat64 return the maximum num of nums.
//
// 1, 2, 3, 4, 5, 0 => 5
func MaxFloat64(nums []float64) float64 {
	var r float64 = 0
	if len(nums) == 0 {
		return r
	}
	r = nums[0]
	for _, num := range nums {
		if num > r {
			r = num
			continue
		}
	}
	return r
}
