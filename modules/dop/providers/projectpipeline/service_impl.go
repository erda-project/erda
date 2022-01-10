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

package projectpipeline

import (
	"context"
	"path/filepath"

	dpb "github.com/erda-project/erda-proto-go/core/pipeline/definition/pb"
	spb "github.com/erda-project/erda-proto-go/core/pipeline/source/pb"
	"github.com/erda-project/erda-proto-go/dop/projectpipeline/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/dop/providers/projectpipeline/deftype"
	"github.com/erda-project/erda/modules/dop/services/apierrors"
	"github.com/erda-project/erda/pkg/common/apis"
)

func (p *ProjectPipelineService) Create(ctx context.Context, params *pb.CreateProjectPipelineRequest) (*pb.CreateProjectPipelineResponse, error) {
	if err := params.Validate(); err != nil {
		return nil, apierrors.ErrCreateProjectPipeline.InvalidParameter(err)
	}
	if err := p.checkCreatePermission(ctx, params); err != nil {
		return nil, apierrors.ErrCreateProjectPipeline.AccessDenied()
	}

	p.pipelineSourceType = NewProjectSourceType(params.SourceType)
	sourceReq, err := p.pipelineSourceType.GenerateReq(ctx, p, params)
	if err != nil {
		return nil, apierrors.ErrCreateProjectPipeline.InternalError(err)
	}

	sourceRsp, err := p.PipelineSource.Create(ctx, sourceReq)
	if err != nil {
		return nil, apierrors.ErrCreateProjectPipeline.InternalError(err)
	}

	definitionRsp, err := p.PipelineDefinition.Create(ctx, &dpb.PipelineDefinitionCreateRequest{
		Name:             params.Name,
		Creator:          apis.GetUserID(ctx),
		PipelineSourceId: sourceRsp.PipelineSource.ID,
		Category:         "",
		Extra: &dpb.PipelineDefinitionExtra{
			Extra: p.pipelineSourceType.GetPipelineCreateRequestV2(),
		},
	})
	if err != nil {
		return nil, apierrors.ErrCreateProjectPipeline.InternalError(err)
	}
	return &pb.CreateProjectPipelineResponse{ID: definitionRsp.PipelineDefinition.ID}, nil
}

func (p *ProjectPipelineService) checkCreatePermission(ctx context.Context, params *pb.CreateProjectPipelineRequest) error {
	if !apis.IsInternalClient(ctx) {
		if params.SourceType == deftype.ErdaProjectPipelineType.String() {
			app, err := p.bundle.GetApp(params.AppID)
			if err != nil {
				return err
			}
			req := apistructs.PermissionCheckRequest{
				UserID:   apis.GetUserID(ctx),
				Scope:    apistructs.AppScope,
				ScopeID:  app.ID,
				Resource: apistructs.AppResource,
				Action:   apistructs.GetAction,
			}
			if access, err := p.bundle.CheckPermission(&req); err != nil || !access.Access {
				return err
			}
		}

		req := apistructs.PermissionCheckRequest{
			UserID:   apis.GetUserID(ctx),
			Scope:    apistructs.ProjectScope,
			ScopeID:  params.ProjectID,
			Resource: apistructs.ProjectPipelineResource,
			Action:   apistructs.CreateAction,
		}
		if access, err := p.bundle.CheckPermission(&req); err != nil || !access.Access {
			return err
		}
	}
	return nil
}

func (p *ProjectPipelineService) getYmlFromGittar(app *apistructs.ApplicationDTO, ref, filePath, userID string) (string, error) {
	commit, err := p.bundle.GetGittarCommit(app.GitRepoAbbrev, ref, userID)
	if err != nil {
		return "", err
	}

	yml, err := p.bundle.GetGittarFile(app.GitRepo, commit.ID, filePath, "", "", userID)
	return yml, err
}

func (p *ProjectPipelineService) List(ctx context.Context, params deftype.ProjectPipelineList) ([]*dpb.PipelineDefinition, error) {
	if err := params.Validate(); err != nil {
		return nil, apierrors.ErrListProjectPipeline.InvalidParameter(err)
	}
	if err := p.checkListPermission(ctx, params); err != nil {
		return nil, apierrors.ErrListProjectPipeline.AccessDenied()
	}

	project, err := p.bundle.GetProject(params.ProjectID)
	if err != nil {
		return nil, apierrors.ErrListProjectPipeline.InternalError(err)
	}

	apps, err := p.bundle.GetAppsByProject(params.ProjectID, project.OrgID, params.IdentityInfo.UserID)
	if err != nil {
		return nil, apierrors.ErrListProjectPipeline.InternalError(err)
	}

	list, err := p.PipelineDefinition.List(ctx, &dpb.PipelineDefinitionListRequest{
		PageSize: int64(params.PageSize),
		PageNo:   int64(params.PageNo),
		Creator:  params.Creator,
		Executor: params.Executor,
		Category: params.Category,
		Ref:      params.Ref,
		Name:     params.Name,
		Remote: func() []string {
			remotes := make([]string, 0, len(apps.List))
			for _, v := range apps.List {
				remotes = append(remotes, makeRemote(&v))
			}
			return remotes
		}(),
		TimeCreated: params.TimeCreated,
		TimeStarted: params.TimeStarted,
		Status:      params.Status,
	})
	if err != nil {
		return nil, apierrors.ErrListProjectPipeline.InternalError(err)
	}

	return list.Data, nil
}

func (p *ProjectPipelineService) checkListPermission(ctx context.Context, params deftype.ProjectPipelineList) error {
	if !apis.IsInternalClient(ctx) {
		req := apistructs.PermissionCheckRequest{
			UserID:   params.IdentityInfo.UserID,
			Scope:    apistructs.ProjectScope,
			ScopeID:  params.ProjectID,
			Resource: apistructs.ProjectPipelineResource,
			Action:   apistructs.GetAction,
		}
		if access, err := p.bundle.CheckPermission(&req); err != nil || !access.Access {
			return err
		}
	}
	return nil
}

func (p *ProjectPipelineService) Delete(ctx context.Context, params deftype.ProjectPipelineDelete) (*deftype.ProjectPipelineDeleteResult, error) {
	if err := params.Validate(); err != nil {
		return nil, apierrors.ErrDeleteProjectPipeline.InvalidParameter(err)
	}
	// TODO check permission

	_, err := p.PipelineDefinition.Delete(ctx, &dpb.PipelineDefinitionDeleteRequest{PipelineDefinitionID: params.ID})
	return nil, err
}

func (p *ProjectPipelineService) Update(ctx context.Context, params deftype.ProjectPipelineUpdate) (*deftype.ProjectPipelineUpdateResult, error) {
	if err := params.Validate(); err != nil {
		return nil, err
	}

	// TODO check permission

	app, err := p.bundle.GetApp(params.AppID)
	if err != nil {
		return nil, err
	}

	yml, err := p.getYmlFromGittar(app, params.Ref, filepath.Join(params.Path, params.FileName), params.IdentityInfo.UserID)
	if err != nil {
		return nil, err
	}

	sourceRsp, err := p.PipelineSource.Create(ctx, &spb.PipelineSourceCreateRequest{
		SourceType:  params.SourceType.String(),
		Remote:      makeRemote(app),
		Ref:         params.Ref,
		Path:        params.Path,
		Name:        params.FileName,
		PipelineYml: yml,
	})
	if err != nil {
		return nil, err
	}
	_, err = p.PipelineDefinition.Update(ctx, &dpb.PipelineDefinitionUpdateRequest{
		PipelineDefinitionID: params.ID,
		PipelineSourceId:     sourceRsp.PipelineSource.ID,
	})

	return nil, err
}

func (p *ProjectPipelineService) Star(ctx context.Context, params deftype.ProjectPipelineStar) (deftype.ProjectPipelineStarResult, error) {
	panic("implement me")
}

func (p *ProjectPipelineService) UnStar(ctx context.Context, params deftype.ProjectPipelineUnStar) (deftype.ProjectPipelineUnStarResult, error) {
	panic("implement me")
}

func (p *ProjectPipelineService) Run(ctx context.Context, params deftype.ProjectPipelineRun) (deftype.ProjectPipelineRunResult, error) {
	panic("implement me")
}

func (p *ProjectPipelineService) FailRerun(ctx context.Context, params deftype.ProjectPipelineFailRerun) (deftype.ProjectPipelineFailRerunResult, error) {
	panic("implement me")
}

func (p *ProjectPipelineService) StartCron(ctx context.Context, params deftype.ProjectPipelineStartCron) (deftype.ProjectPipelineStartCronResult, error) {
	panic("implement me")
}

func (p *ProjectPipelineService) EndCron(ctx context.Context, params deftype.ProjectPipelineEndCron) (deftype.ProjectPipelineEndCronResult, error) {
	panic("implement me")
}

func (p *ProjectPipelineService) ListExecHistory(ctx context.Context, params deftype.ProjectPipelineListExecHistory) (deftype.ProjectPipelineListExecHistoryResult, error) {
	panic("implement me")
}
