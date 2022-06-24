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

package resource

import (
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/internal/tools/pipeline/conf"
	"github.com/erda-project/erda/internal/tools/pipeline/pkg/action_info"
	"github.com/erda-project/erda/internal/tools/pipeline/spec"
	"github.com/erda-project/erda/pkg/parser/diceyml"
	"github.com/erda-project/erda/pkg/parser/pipelineyml"
)

type Interface interface {
	// CalculatePipelineResources calculate pipeline resources according to all tasks grouped by stages.
	CalculatePipelineResources(pipelineYml *pipelineyml.PipelineYml, p *spec.Pipeline) (*apistructs.PipelineAppliedResources, error)
	CalculateNormalTaskResources(action *pipelineyml.Action, actionDefine *diceyml.Job) apistructs.PipelineAppliedResources
}

func (s *provider) CalculatePipelineResources(pipelineYml *pipelineyml.PipelineYml, p *spec.Pipeline) (*apistructs.PipelineAppliedResources, error) {
	if pipelineYml.Spec() == nil || len(pipelineYml.Spec().Stages) <= 0 {
		return nil, nil
	}

	// load pipelineYml all action define and spec
	var passedDataWhenCreate action_info.PassedDataWhenCreate
	passedDataWhenCreate.InitData(s.ActionMgr)
	if err := passedDataWhenCreate.PutPassedDataByPipelineYml(pipelineYml, p); err != nil {
		return nil, err
	}

	// summarize the resources required for all actions
	var stagesPipelineAppliedResources = make([][]*apistructs.PipelineAppliedResources, len(pipelineYml.Spec().Stages))
	pipelineYml.Spec().LoopStagesActions(func(stage int, action *pipelineyml.Action) {
		if !action.Type.IsSnippet() {
			resources := s.CalculateNormalTaskResources(action, passedDataWhenCreate.GetActionJobDefine(s.ActionMgr.MakeActionTypeVersion(action)))
			stagesPipelineAppliedResources[stage] = append(stagesPipelineAppliedResources[stage], &resources)
		}
	})

	// merge result
	pipelineResource := apistructs.PipelineAppliedResources{
		Limits:   calculatePipelineLimitResource(stagesPipelineAppliedResources),
		Requests: calculatePipelineRequestResource(stagesPipelineAppliedResources),
	}

	return &pipelineResource, nil
}

func (s *provider) CalculateNormalTaskResources(action *pipelineyml.Action, actionDefine *diceyml.Job) apistructs.PipelineAppliedResources {
	defaultRes := apistructs.PipelineAppliedResource{CPU: conf.TaskDefaultCPU(), MemoryMB: conf.TaskDefaultMEM()}
	overSoldRes := apistructs.PipelineOverSoldResource{CPURate: conf.TaskDefaultCPUOverSoldRate(), MaxCPU: conf.TaskMaxAllowedOverSoldCPU()}
	return apistructs.PipelineAppliedResources{
		Limits:   calculateOversoldTaskLimitResource(calculateNormalTaskLimitResource(action, actionDefine, defaultRes), overSoldRes),
		Requests: calculateNormalTaskRequestResource(action, actionDefine, defaultRes),
	}
}
