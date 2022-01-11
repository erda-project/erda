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

package pipelineTabs

import (
	"context"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/component-protocol/cpregister/base"
	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	dpb "github.com/erda-project/erda-proto-go/core/pipeline/definition/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/dop/component-protocol/components/project-pipeline/common"
	"github.com/erda-project/erda/modules/dop/providers/projectpipeline/deftype"
)

func init() {
	base.InitProviderWithCreator(common.ScenarioKey, "pipelineTabs", func() servicehub.Provider {
		return &Tab{}
	})
}

func (t *Tab) Render(ctx context.Context, c *cptype.Component, scenario cptype.Scenario, event cptype.ComponentEvent, gs *cptype.GlobalStateData) error {
	if err := t.InitFromProtocol(ctx, c, gs); err != nil {
		return err
	}

	t.SetType()
	t.SetState(func() string {
		if event.Operation == cptype.InitializeOperation ||
			event.Operation == cptype.RenderingOperation ||
			t.State.Value == "" {
			return defaultState.String()
		}
		return t.State.Value
	}())
	t.SetOperations(t.State.Value)

	list, total, err := t.ProjectPipelineSvc.List(ctx, deftype.ProjectPipelineList{
		ProjectID:    t.InParams.ProjectID,
		IdentityInfo: apistructs.IdentityInfo{UserID: t.sdk.Identity.UserID},
	})
	if err != nil {
		return err
	}
	t.SetData(ctx, Num{
		MinePipelineNum: func() uint64 {
			return uint64(len(pipelineFilterIn(list, func(pipeline *dpb.PipelineDefinition) bool {
				return pipeline.Creator == t.sdk.Identity.UserID
			})))
		}(),
		PrimaryPipelineNum: func() uint64 {
			return uint64(len(pipelineFilterIn(list, func(pipeline *dpb.PipelineDefinition) bool {
				return pipeline.Category == "primary"
			})))
		}(),
		AllPipelineNum: uint64(total),
	})
	t.gsHelper.SetGlobalPipelineTab(t.State.Value)
	return t.SetToProtocolComponent(c)
}

func pipelineFilterIn(pipelines []*dpb.PipelineDefinition, fn func(pipeline *dpb.PipelineDefinition) bool) []*dpb.PipelineDefinition {
	newPipelines := make([]*dpb.PipelineDefinition, 0)
	for _, v := range pipelines {
		if fn(v) {
			newPipelines = append(newPipelines, v)
		}
	}
	return newPipelines
}
