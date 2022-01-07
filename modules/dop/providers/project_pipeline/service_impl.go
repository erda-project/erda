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

package project_pipeline

import (
	"context"
	"encoding/json"
	"fmt"

	definitionpb "github.com/erda-project/erda-proto-go/core/pipeline/definition/pb"
	sourcepb "github.com/erda-project/erda-proto-go/core/pipeline/source/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/dop/providers/project_pipeline/deftype"
	"github.com/erda-project/erda/modules/dop/services/apierrors"
)

func (p *provider) Create(ctx context.Context, params deftype.ProjectPipelineCreate) (deftype.ProjectPipelineCreateResult, error) {
	panic("implement me")
}

func (p *provider) List(ctx context.Context, params deftype.ProjectPipelineList) (deftype.ProjectPipelineListResult, error) {
	panic("implement me")
}

func (p *provider) Delete(ctx context.Context, params deftype.ProjectPipelineDelete) (deftype.ProjectPipelineDeleteResult, error) {
	panic("implement me")
}

func (p *provider) Update(ctx context.Context, params deftype.ProjectPipelineUpdate) (deftype.ProjectPipelineUpdateResult, error) {
	panic("implement me")
}

func (p *provider) Star(ctx context.Context, params deftype.ProjectPipelineStar) (deftype.ProjectPipelineStarResult, error) {
	panic("implement me")
}

func (p *provider) UnStar(ctx context.Context, params deftype.ProjectPipelineUnStar) (deftype.ProjectPipelineUnStarResult, error) {
	panic("implement me")
}

func (p *provider) Run(ctx context.Context, params deftype.ProjectPipelineRun) (*deftype.ProjectPipelineRunResult, error) {
	if params.PipelineDefinitionID == "" {
		return nil, apierrors.ErrRunProjectPipeline.InvalidParameter(fmt.Errorf("pipelineDefinitionID：%s", params.PipelineDefinitionID))
	}
	var getReq definitionpb.PipelineDefinitionGetRequest
	getReq.PipelineDefinitionID = params.PipelineDefinitionID

	resp, err := p.PipelineDefinition.Get(context.Background(), &getReq)
	if err != nil {
		return nil, err
	}
	if resp == nil || resp.PipelineDefinition == nil || resp.PipelineDefinition.Extra == nil {
		return nil, apierrors.ErrRunProjectPipeline.InternalError(fmt.Errorf("not find pipeline"))
	}
	var extraValue = apistructs.PipelineDefinitionExtraValue{}
	err = json.Unmarshal([]byte(resp.PipelineDefinition.Extra.Extra), &extraValue)
	if err != nil {
		return nil, apierrors.ErrRunProjectPipeline.InternalError(fmt.Errorf("failed unmarshal pipeline extra error %v", err))
	}

	var sourceGetReq = sourcepb.PipelineSourceGetRequest{}
	sourceGetReq.PipelineSourceID = resp.PipelineDefinition.PipelineSourceId
	source, err := p.PipelineSource.Get(context.Background(), &sourceGetReq)
	if err != nil {
		return nil, err
	}

	if source == nil || source.PipelineSource == nil || source.PipelineSource.PipelineYml == "" {
		return nil, apierrors.ErrRunProjectPipeline.InternalError(fmt.Errorf("failed to run pipeline, yml was empty"))
	}

	createV2 := extraValue.CreateRequest
	createV2.PipelineYml = source.PipelineSource.PipelineYml
	createV2.AutoRunAtOnce = true

	value, err := p.bundle.CreatePipeline(createV2)
	if err != nil {
		return nil, apierrors.ErrRunProjectPipeline.InternalError(err)
	}
	return &deftype.ProjectPipelineRunResult{
		Pipeline: value,
	}, nil
}

func (p *provider) FailRerun(ctx context.Context, params deftype.ProjectPipelineFailRerun) (deftype.ProjectPipelineFailRerunResult, error) {
	return deftype.ProjectPipelineFailRerunResult{}, nil
}

func (p *provider) StartCron(ctx context.Context, params deftype.ProjectPipelineStartCron) (deftype.ProjectPipelineStartCronResult, error) {
	return deftype.ProjectPipelineStartCronResult{}, nil
}

func (p *provider) EndCron(ctx context.Context, params deftype.ProjectPipelineEndCron) (deftype.ProjectPipelineEndCronResult, error) {
	return deftype.ProjectPipelineEndCronResult{}, nil
}

func (p *provider) ListExecHistory(ctx context.Context, params deftype.ProjectPipelineListExecHistory) (deftype.ProjectPipelineListExecHistoryResult, error) {
	return deftype.ProjectPipelineListExecHistoryResult{}, nil
}
