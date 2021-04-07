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

package cpupolicy

import (
	"fmt"
	"math"
	"strconv"
)

// 应用公式 round(x+1.5**(-3.6*x)*1.6*x,1)
// 具体参考 http://git.terminus.io/dice/dice/blob/develop/scripts/cpu_policy/policy.org
func AdjustCPUSize(origin float64) float64 {
	value := origin + math.Pow(1.5, -3.0*origin)*0.9*origin
	value_, _ := strconv.ParseFloat(fmt.Sprintf("%.1f", value), 64)
	return value_
}

func CalcCPUSubscribeRatio(cpuSubscribeRatio float64, extra map[string]string) float64 {
	if ratio_, ok := extra["CPU_SUBSCRIBE_RATIO"]; ok && len(ratio_) > 0 {
		if ratio, err := strconv.ParseFloat(ratio_, 64); err == nil && ratio > 1.0 {
			return ratio
		}
	}
	if cpuSubscribeRatio > 1.0 {
		return cpuSubscribeRatio
	}
	return 1.0
}
