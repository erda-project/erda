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
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda/modules/scheduler/schedulepolicy/labelconfig"
	"github.com/erda-project/erda/pkg/parser/diceyml"
)

func TestLocationLabelFilter(t *testing.T) {
	r := labelconfig.RawLabelRuleResult{}
	r2 := labelconfig.RawLabelRuleResult2{}
	LocationLabelFilter(&r, &r2, &labelconfig.LabelInfo{
		Selectors: map[string]diceyml.Selectors{
			"servicename": {"location": diceyml.Selector{Values: []string{"xxx", "yyy"}}},
		},
	})
	assert.Equal(t, map[string]interface{}{"servicename": diceyml.Selector{Values: []string{"xxx", "yyy"}}}, r.Location)
}
