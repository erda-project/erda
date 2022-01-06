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

	"github.com/erda-project/erda/modules/dop/providers/project_pipeline/deftype"
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

func (p *provider) Run(ctx context.Context, params deftype.ProjectPipelineRun) (deftype.ProjectPipelineRunResult, error) {
	panic("implement me")
}

func (p *provider) FailRerun(ctx context.Context, params deftype.ProjectPipelineFailRerun) (deftype.ProjectPipelineFailRerunResult, error) {
	panic("implement me")
}

func (p *provider) StartCron(ctx context.Context, params deftype.ProjectPipelineStartCron) (deftype.ProjectPipelineStartCronResult, error) {
	panic("implement me")
}

func (p *provider) EndCron(ctx context.Context, params deftype.ProjectPipelineEndCron) (deftype.ProjectPipelineEndCronResult, error) {
	panic("implement me")
}

func (p *provider) ListExecHistory(ctx context.Context, params deftype.ProjectPipelineListExecHistory) (deftype.ProjectPipelineListExecHistoryResult, error) {
	panic("implement me")
}
