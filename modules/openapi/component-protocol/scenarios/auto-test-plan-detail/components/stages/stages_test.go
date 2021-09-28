// Copyright (c) 2021 Terminus, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package stages

import (
	"testing"

	"github.com/erda-project/erda/apistructs"
	"github.com/stretchr/testify/assert"
)

func TestFindFirstLastStepInGroup(t *testing.T) {
	tt := []struct {
		steps     []*apistructs.TestPlanV2Step
		wantFirst uint64
		wantLast  uint64
	}{
		{
			steps: []*apistructs.TestPlanV2Step{
				{ID: 2, PreID: 1},
				{ID: 3, PreID: 2},
				{ID: 4, PreID: 3},
				{ID: 5, PreID: 4},
			},
			wantFirst: 2,
			wantLast:  5,
		},
		{
			steps: []*apistructs.TestPlanV2Step{
				{ID: 4, PreID: 0},
				{ID: 2, PreID: 4},
				{ID: 3, PreID: 5},
				{ID: 5, PreID: 2},
			},
			wantFirst: 4,
			wantLast:  3,
		},
		{
			steps: []*apistructs.TestPlanV2Step{
				{ID: 1, PreID: 0},
			},
			wantFirst: 1,
			wantLast:  1,
		},
	}
	for _, v := range tt {
		firstStep, lastStep := findFirstLastStepInGroup(v.steps)
		firstStepID, lastStepID := firstStep.ID, lastStep.ID
		assert.Equal(t, v.wantFirst, firstStepID)
		assert.Equal(t, v.wantLast, lastStepID)
	}
}
