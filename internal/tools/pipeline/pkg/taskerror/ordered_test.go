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

package taskerror

import (
	"sort"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestOrderedResponses(t *testing.T) {
	now := time.Now()
	respA := &PipelineTaskErrResponse{
		Code: "codeA",
		Msg:  "msg of codeA",
		Ctx: PipelineTaskErrCtx{
			StartTime: time.Time{},
			EndTime:   now,
			Count:     0,
		},
	}
	respB := &PipelineTaskErrResponse{
		Code: "codeB",
		Msg:  "msg of codeB",
		Ctx: PipelineTaskErrCtx{
			StartTime: time.Time{},
			EndTime:   now.Add(-time.Second), // before codeA
			Count:     0,
		},
	}
	resps := OrderedResponses{respA, respB}
	sort.Sort(resps)

	assert.Equal(t, 2, len(resps))
	assert.Equal(t, "codeB", resps[0].Code)
	assert.Equal(t, "codeA", resps[1].Code)
}
