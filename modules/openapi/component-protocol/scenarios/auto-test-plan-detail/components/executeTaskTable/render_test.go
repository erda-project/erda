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

package executeTaskTable

import (
	"context"
	"testing"

	"github.com/alecthomas/assert"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
)

func Test_handlerListOperation(t *testing.T) {
	ctx := protocol.ContextBundle{
		InParams: map[string]interface{}{},
	}
	ctx1 := context.WithValue(context.Background(), protocol.GlobalInnerKeyCtxBundle.String(), ctx)
	a := &ExecuteTaskTable{}
	a.State.PipelineDetail = nil
	err := a.Render(ctx1, &apistructs.Component{}, apistructs.ComponentProtocolScenario{},
		apistructs.ComponentEvent{
			Operation:     apistructs.ExecuteChangePageNoOperationKey,
			OperationData: nil,
		}, nil)
	assert.NoError(t, err)
	a.State.PipelineDetail = &apistructs.PipelineDetailDTO{
		PipelineDTO: apistructs.PipelineDTO{
			ID: 0,
		},
	}
	assert.NoError(t, err)
}

func TestGetCostTime(t *testing.T) {
	tt := []struct {
		task apistructs.PipelineTaskDTO
		want string
	}{
		{
			apistructs.PipelineTaskDTO{
				Status: apistructs.PipelineStatusRunning,
			},
			"-",
		},
		{
			apistructs.PipelineTaskDTO{
				Status:      apistructs.PipelineStatusSuccess,
				IsSnippet:   false,
				CostTimeSec: 59,
			},
			"00:00:59",
		},
		{
			apistructs.PipelineTaskDTO{
				Status:      apistructs.PipelineStatusSuccess,
				IsSnippet:   false,
				CostTimeSec: 3600,
			},
			"01:00:00",
		},
		{
			apistructs.PipelineTaskDTO{
				Status:      apistructs.PipelineStatusSuccess,
				IsSnippet:   true,
				CostTimeSec: 59*60 + 59,
			},
			"00:59:59",
		},
		{
			apistructs.PipelineTaskDTO{
				Status:      apistructs.PipelineStatusSuccess,
				IsSnippet:   true,
				CostTimeSec: -1,
			},
			"-",
		},
	}
	r := ExecuteTaskTable{
		CtxBdl: protocol.ContextBundle{
			Bdl: bundle.New(),
		},
	}
	for _, v := range tt {
		assert.Equal(t, v.want, r.getCostTime(v.task))
	}
}
