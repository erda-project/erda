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

	"github.com/erda-project/erda-proto-go/core/pipeline/definition/pb"
	spb "github.com/erda-project/erda-proto-go/core/pipeline/source/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/dop/providers/project_pipeline/deftype"
)

func (p *ProjectPipelineSvc) Create(ctx context.Context, params deftype.ProjectPipelineCreate) (*deftype.ProjectPipelineCreateResult, error) {
	if err := params.Validate(); err != nil {
		return nil, err
	}

	app, err := p.bundle.GetApp(params.AppID)
	if err != nil {
		return nil, err
	}

	source, err := p.PipelineSource.Create(ctx, &spb.PipelineSourceCreateRequest{
		SourceType: params.SourceType.String(),
		Remote:     makeRemote(app),
		Ref:        params.Ref,
		Path:       params.Path,
		Name:       params.FileName,
	})
	if err != nil {
		return nil, err
	}

	var createReqV1 = apistructs.PipelineCreateRequest{
		PipelineYmlName:    params.FileName,
		AppID:              params.AppID,
		Branch:             params.Ref,
		PipelineYmlContent: "version: \"1.1\"\nstages: []",
		UserID:             params.IdentityInfo.UserID,
	}
	createReqV2, err := p.pipelineSvc.ConvertPipelineToV2(&createReqV1)
	if err != nil {
		return nil, err
	}
	b, err := json.Marshal(createReqV2)
	if err != nil {
		return nil, err
	}

	_, err = p.PipelineDefinition.Create(ctx, &pb.PipelineDefinitionCreateRequest{
		Name:             params.Name,
		Creator:          params.IdentityInfo.UserID,
		PipelineSourceId: source.PipelineSource.ID,
		Category:         "",
		Extra: &pb.PipelineDefinitionExtra{
			Extra: string(b),
		},
	})
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func makeRemote(app *apistructs.ApplicationDTO) string {
	return fmt.Sprintf("%d/%d/%d", app.OrgID, app.ProjectID, app.ID)
}

func (p *ProjectPipelineSvc) List(ctx context.Context, params deftype.ProjectPipelineList) (deftype.ProjectPipelineListResult, error) {
	panic("implement me")
}

func (p *ProjectPipelineSvc) Delete(ctx context.Context, params deftype.ProjectPipelineDelete) (deftype.ProjectPipelineDeleteResult, error) {
	panic("implement me")
}

func (p *ProjectPipelineSvc) Update(ctx context.Context, params deftype.ProjectPipelineUpdate) (deftype.ProjectPipelineUpdateResult, error) {
	panic("implement me")
}

func (p *ProjectPipelineSvc) Star(ctx context.Context, params deftype.ProjectPipelineStar) (deftype.ProjectPipelineStarResult, error) {
	panic("implement me")
}

func (p *ProjectPipelineSvc) UnStar(ctx context.Context, params deftype.ProjectPipelineUnStar) (deftype.ProjectPipelineUnStarResult, error) {
	panic("implement me")
}

func (p *ProjectPipelineSvc) Run(ctx context.Context, params deftype.ProjectPipelineRun) (deftype.ProjectPipelineRunResult, error) {
	panic("implement me")
}

func (p *ProjectPipelineSvc) FailRerun(ctx context.Context, params deftype.ProjectPipelineFailRerun) (deftype.ProjectPipelineFailRerunResult, error) {
	panic("implement me")
}

func (p *ProjectPipelineSvc) StartCron(ctx context.Context, params deftype.ProjectPipelineStartCron) (deftype.ProjectPipelineStartCronResult, error) {
	panic("implement me")
}

func (p *ProjectPipelineSvc) EndCron(ctx context.Context, params deftype.ProjectPipelineEndCron) (deftype.ProjectPipelineEndCronResult, error) {
	panic("implement me")
}

func (p *ProjectPipelineSvc) ListExecHistory(ctx context.Context, params deftype.ProjectPipelineListExecHistory) (deftype.ProjectPipelineListExecHistoryResult, error) {
	panic("implement me")
}
