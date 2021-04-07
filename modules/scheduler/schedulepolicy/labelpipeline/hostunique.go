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
	"encoding/json"

	"github.com/erda-project/erda/modules/scheduler/schedulepolicy/labelconfig"

	"github.com/sirupsen/logrus"
)

// HostUniqueLabelFilter 处理 Pass1ScheduleInfo.label 中的 HOST_UNIQUE
func HostUniqueLabelFilter(
	r *labelconfig.RawLabelRuleResult, r2 *labelconfig.RawLabelRuleResult2, li *labelconfig.LabelInfo) {
	hostUniqueStr, ok := li.Label[labelconfig.HOST_UNIQUE]
	if !ok {
		return
	}
	var hostUniqueGroup [][]string
	if err := json.Unmarshal([]byte(hostUniqueStr), &hostUniqueGroup); err != nil {
		logrus.Errorf("bad input label: %v, err: %v", labelconfig.HOST_UNIQUE, err)
		return
	}
	r.HostUnique = true
	r.HostUniqueInfo = hostUniqueGroup

	r2.HasHostUnique = true
	r2.HostUnique = hostUniqueGroup
}
