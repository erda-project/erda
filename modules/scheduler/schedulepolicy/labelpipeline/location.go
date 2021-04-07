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
	"github.com/erda-project/erda/modules/scheduler/schedulepolicy/labelconfig"
)

// LocationLabelFilter LabelInfo.Selectors
func LocationLabelFilter(
	r *labelconfig.RawLabelRuleResult, r2 *labelconfig.RawLabelRuleResult2, li *labelconfig.LabelInfo) {
	if r.Location == nil {
		r.Location = make(map[string]interface{})
	}
	if r2.Location == nil {
		r2.Location = make(map[string]interface{})
	}
	for service, selectors := range li.Selectors {
		selector, ok := selectors["location"]
		if !ok {
			continue
		}
		r.Location[service] = selector
		r2.Location[service] = selector
	}
}
