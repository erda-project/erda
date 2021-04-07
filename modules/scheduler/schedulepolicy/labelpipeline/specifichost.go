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

package labelpipeline

import (
	"strings"

	"github.com/erda-project/erda/modules/scheduler/schedulepolicy/labelconfig"
)

func SpecificHostLabelFilter(
	r *labelconfig.RawLabelRuleResult, r2 *labelconfig.RawLabelRuleResult2, li *labelconfig.LabelInfo) {
	v, ok := li.Label[labelconfig.SPECIFIC_HOSTS]
	if !ok {
		return
	}
	result := []string{}
	hosts := strings.Split(v, ",")
	for _, host := range hosts {
		trimmedHost := strings.TrimSpace(host)
		if trimmedHost != "" {
			result = append(result, trimmedHost)
		}
	}
	r.SpecificHost = result
	r2.SpecificHost = result
}
