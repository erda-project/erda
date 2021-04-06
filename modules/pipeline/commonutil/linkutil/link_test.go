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

package linkutil

import (
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/pipeline/spec"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetPipelineLink(t *testing.T) {
	fakeBdl := bundle.New()

	type invalidCase struct {
		p    spec.Pipeline
		desc string
	}

	invalidCases := []invalidCase{
		{
			p: spec.Pipeline{
				Labels: map[string]string{
					apistructs.LabelOrgID: "a",
				},
			},
			desc: "invalid orgID",
		},
		{
			p: spec.Pipeline{
				Labels: map[string]string{
					apistructs.LabelOrgID:     "1",
					apistructs.LabelProjectID: "a",
				},
			},
			desc: "invalid projectID",
		},
		{
			p: spec.Pipeline{
				Labels: map[string]string{
					apistructs.LabelOrgID:     "1",
					apistructs.LabelProjectID: "1",
					apistructs.LabelAppID:     "a",
				},
			},
			desc: "invalid appID",
		},
	}

	for _, c := range invalidCases {
		valid, link := GetPipelineLink(fakeBdl, c.p)
		assert.False(t, valid)
		assert.Empty(t, link)
	}
}
