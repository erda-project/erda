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
	"encoding/json"
	"fmt"
	"github.com/erda-project/erda/pkg/http/httpserver/errorresp"
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

func (p *ProjectPipelineService) Run(ctx context.Context, params deftype.ProjectPipelineRun) (*deftype.ProjectPipelineRunResult, error) {
	if params.PipelineDefinitionID == "" {
		return nil, apierrors.ErrRunProjectPipeline.InvalidParameter(fmt.Errorf("pipelineDefinitionID：%s", params.PipelineDefinitionID))
	}

	definition, err := p.getPipelineDefinition(params.PipelineDefinitionID)
	if err != nil {
		return nil, apierrors.ErrRunProjectPipeline.InternalError(err)
	}

	source, err := p.getPipelineSource(definition.PipelineSourceId)
	if err != nil {
		return nil, apierrors.ErrRunProjectPipeline.InternalError(err)
	}

	var extraValue = apistructs.PipelineDefinitionExtraValue{}
	err = json.Unmarshal([]byte(definition.Extra.Extra), &extraValue)
	if err != nil {
		return nil, apierrors.ErrRunProjectPipeline.InternalError(fmt.Errorf("failed unmarshal pipeline extra error %v", err))
	}
	createV2 := extraValue.CreateRequest
	createV2.PipelineYml = source.PipelineYml
	createV2.AutoRunAtOnce = true
	createV2.DefinitionID = definition.ID

	value, err := p.bundle.CreatePipeline(createV2)
	if err != nil {
		return nil, apierrors.ErrRunProjectPipeline.InternalError(err)
	}
	return &deftype.ProjectPipelineRunResult{
		Pipeline: value,
	}, nil
}

func (p *ProjectPipelineService) StartCron(ctx context.Context, params deftype.ProjectPipelineStartCron) (*deftype.ProjectPipelineStartCronResult, error) {
	dto ,err := p.startOrEndCron(params.PipelineDefinitionID, true, apierrors.ErrStartCronProjectPipeline)
	if err != nil {
		return nil, err
	}

	return &deftype.ProjectPipelineStartCronResult{
		Cron: dto,
	}, nil
}

func (p *ProjectPipelineService) EndCron(ctx context.Context, params deftype.ProjectPipelineEndCron) (*deftype.ProjectPipelineEndCronResult, error) {

	dto ,err := p.startOrEndCron(params.PipelineDefinitionID, false, apierrors.ErrEndCronProjectPipeline)
	if err != nil {
		return nil, err
	}
	return &deftype.ProjectPipelineEndCronResult{
		Cron: dto,
	}, nil
}

func (p *ProjectPipelineService) ListExecHistory(ctx context.Context, params deftype.ProjectPipelineListExecHistory) (*deftype.ProjectPipelineListExecHistoryResult, error) {
	var pipelineDefinition = apistructs.PipelineDefinitionRequest{
		Name: params.Name,
		Creator: params.Executor,
	}
	if params.AppID != 0 {
		appDto, err := p.bundle.GetApp(params.AppID)
		if err != nil {
			return nil, apierrors.ErrListExecHistoryProjectPipeline.InternalError(err)
		}
		pipelineDefinition.PipelineSourceRequest = apistructs.PipelineSourceRequest{
			Remote: filepath.Join(appDto.OrgName, appDto.ProjectName, appDto.Name),
		}
	}
	jsonValue, err := json.Marshal(pipelineDefinition)
	if err != nil {
		return nil, apierrors.ErrListExecHistoryProjectPipeline.InternalError(err)
	}

	var pipelinePageListRequest = apistructs.PipelinePageListRequest{
		PageNum:  int(params.PageNo),
		PageSize: int(params.PageSize),
		AllSources: true,
		PipelineDefinitionRequestJSON: string(jsonValue),
	}
	if params.Status != "" {
		pipelinePageListRequest.Statuses = []string{params.Status}
	}

	data, err := p.bundle.PageListPipeline(pipelinePageListRequest)
	if err != nil {
		return nil, apierrors.ErrListExecHistoryProjectPipeline.InternalError(err)
	}
	return &deftype.ProjectPipelineListExecHistoryResult{
		Data: data,
	}, nil
}

func (p *ProjectPipelineService) BatchRun(ctx context.Context, params deftype.ProjectPipelineBatchRun) (*deftype.ProjectPipelineBatchRunResult, error) {
	panic("implement me")
}

func (p *ProjectPipelineService) Cancel(ctx context.Context, params deftype.ProjectPipelineCancel) (*deftype.ProjectPipelineCancelResult, error) {
	if params.PipelineDefinitionID == "" {
		return nil, apierrors.ErrCancelProjectPipeline.InvalidParameter(fmt.Errorf("pipelineDefinitionID：%s", params.PipelineDefinitionID))
	}

	definition, err := p.getPipelineDefinition(params.PipelineDefinitionID)
	if err != nil {
		return nil, apierrors.ErrCancelProjectPipeline.InternalError(err)
	}

	var extraValue = apistructs.PipelineDefinitionExtraValue{}
	err = json.Unmarshal([]byte(definition.Extra.Extra), &extraValue)
	if err != nil {
		return nil, apierrors.ErrCancelProjectPipeline.InternalError(fmt.Errorf("failed unmarshal pipeline extra error %v", err))
	}

	runningPipelineID, err := p.getRunningPipeline(extraValue.CreateRequest.PipelineSource.String(), extraValue.CreateRequest.PipelineYmlName)
	if err != nil{
		return nil, apierrors.ErrCancelProjectPipeline.InternalError(err)
	}
	if runningPipelineID == 0 {
		return nil, apierrors.ErrCancelProjectPipeline.InternalError(fmt.Errorf("not find running pipeline, can not cancel"))
	}

	var req apistructs.PipelineCancelRequest
	req.PipelineID = runningPipelineID
	req.IdentityInfo = params.IdentityInfo
	err = p.bundle.CancelPipeline(req)
	if err != nil {
		return nil, apierrors.ErrCancelProjectPipeline.InternalError(err)
	}

	return &deftype.ProjectPipelineCancelResult{}, nil
}

func (p *ProjectPipelineService) Rerun(ctx context.Context, params deftype.ProjectPipelineRerun) (*deftype.ProjectPipelineRerunResult, error) {
	dto, err := p.failRerunOrRerunPipeline(true, params.PipelineDefinitionID, params.IdentityInfo, apierrors.ErrRerunProjectPipeline)
	if err != nil {
		return nil, err
	}

	return &deftype.ProjectPipelineRerunResult{
		Pipeline: dto,
	}, nil
}

func (p *ProjectPipelineService) FailRerun(ctx context.Context, params deftype.ProjectPipelineFailRerun) (*deftype.ProjectPipelineFailRerunResult, error) {
	dto, err := p.failRerunOrRerunPipeline(false, params.PipelineDefinitionID, params.IdentityInfo, apierrors.ErrFailRerunProjectPipeline)
	if err != nil {
		return nil, err
	}
	return &deftype.ProjectPipelineFailRerunResult{
		Pipeline: dto,
	}, nil
}

func (p *ProjectPipelineService) failRerunOrRerunPipeline(rerun bool, pipelineDefinitionID string, identityInfo apistructs.IdentityInfo, apiError *errorresp.APIError) (*apistructs.PipelineDTO, error) {
	if pipelineDefinitionID == "" {
		return nil, apiError.InvalidParameter(fmt.Errorf("pipelineDefinitionID：%s", pipelineDefinitionID))
	}

	definition, err := p.getPipelineDefinition(pipelineDefinitionID)
	if err != nil {
		return nil, apiError.InternalError(err)
	}

	var extraValue = apistructs.PipelineDefinitionExtraValue{}
	err = json.Unmarshal([]byte(definition.Extra.Extra), &extraValue)
	if err != nil {
		return nil, apiError.InternalError(fmt.Errorf("failed unmarshal pipeline extra error %v", err))
	}

	runningPipelineID, err := p.getRunningPipeline(extraValue.CreateRequest.PipelineSource.String(), extraValue.CreateRequest.PipelineYmlName)
	if err != nil {
		return nil, apiError.InternalError(err)
	}
	if runningPipelineID > 0 {
		return nil, apiError.InternalError(fmt.Errorf("operation failed, pipeline %v was running status", runningPipelineID))
	}
	pipeline, err := p.getLatestPipeline(extraValue.CreateRequest.PipelineSource.String(), extraValue.CreateRequest.PipelineYmlName)
	if err != nil {
		return nil, apiError.InternalError(err)
	}
	if !pipeline.Status.IsFailedStatus() {
		return nil, apiError.InternalError(fmt.Errorf("operation failed, the latest pipeline is not in an error state"))
	}

	if rerun {
		var req apistructs.PipelineRerunRequest
		req.PipelineID = pipeline.ID
		req.AutoRunAtOnce = true
		req.IdentityInfo = identityInfo
		dto, err := p.bundle.RerunPipeline(req)
		if err != nil {
			return nil, apiError.InternalError(err)
		}
		return dto, nil
	}else{
		var req apistructs.PipelineRerunFailedRequest
		req.PipelineID = pipeline.ID
		req.AutoRunAtOnce = true
		req.IdentityInfo = identityInfo
		dto, err := p.bundle.RerunFailedPipeline(req)
		if err != nil {
			return nil, apiError.InternalError(err)
		}

		return dto, nil
	}
}

func (p *ProjectPipelineService) startOrEndCron(pipelineDefinitionID string, enable bool, apiError *errorresp.APIError) (*apistructs.PipelineCronDTO, error) {
	if pipelineDefinitionID == "" {
		return nil, apiError.InvalidParameter(fmt.Errorf("pipelineDefinitionID：%s", pipelineDefinitionID))
	}

	definition, err := p.getPipelineDefinition(pipelineDefinitionID)
	if err != nil {
		return nil, apiError.InternalError(err)
	}

	var extraValue = apistructs.PipelineDefinitionExtraValue{}
	err = json.Unmarshal([]byte(definition.Extra.Extra), &extraValue)
	if err != nil {
		return nil, apiError.InternalError(fmt.Errorf("failed unmarshal pipeline extra error %v", err))
	}

	var req apistructs.PipelineCronPagingRequest
	req.PageNo = 1
	req.PageSize = 1
	req.Sources = []apistructs.PipelineSource{extraValue.CreateRequest.PipelineSource}
	req.YmlNames = []string{extraValue.CreateRequest.PipelineYmlName}
	cron, err := p.bundle.PageListPipelineCrons(req)
	if err != nil {
		return nil, apiError.InternalError(err)
	}
	if len(cron.Data) == 0 {
		return nil, apiError.InternalError(fmt.Errorf("not find cron"))
	}

	var dto *apistructs.PipelineCronDTO
	if enable {
		dto, err = p.bundle.StartPipelineCron(cron.Data[0].ID)
		if err != nil {
			return nil, apiError.InternalError(err)
		}
	}else{
		dto, err = p.bundle.StopPipelineCron(cron.Data[0].ID)
		if err != nil {
			return nil, apiError.InternalError(err)
		}
	}
	return dto, nil
}

func getRunningStatus() []string {
	var runningStatus []string
	for _, status := range apistructs.ReconcilerRunningStatuses() {
		runningStatus = append(runningStatus, status.String())
	}
	return runningStatus
}

func (p *ProjectPipelineService) getRunningPipeline(pipelineSource string, pipelineYmlName string) (pipelineID uint64, err error) {
	var pipelinePageListRequest = apistructs.PipelinePageListRequest{
		PageNum:  1,
		PageSize: 1,
		Sources: []apistructs.PipelineSource{
			apistructs.PipelineSource(pipelineSource),
		},
		YmlNames: []string{
			pipelineYmlName,
		},
		Statuses: getRunningStatus(),
		DescCols: []string{
			"id",
		},
	}
	data, err := p.bundle.PageListPipeline(pipelinePageListRequest)
	if err != nil {
		return 0, apierrors.ErrCancelProjectPipeline.InternalError(err)
	}
	if len(data.Pipelines) == 0 {
		return 0, nil
	}
	return data.Pipelines[0].ID, nil
}

func (p *ProjectPipelineService) getLatestPipeline(pipelineSource string, pipelineYmlName string) (pipeline *apistructs.PagePipeline, err error) {
	var pipelinePageListRequest = apistructs.PipelinePageListRequest{
		PageNum:  1,
		PageSize: 1,
		Sources: []apistructs.PipelineSource{
			apistructs.PipelineSource(pipelineSource),
		},
		YmlNames: []string{
			pipelineYmlName,
		},
		NotStatuses: getRunningStatus(),
		DescCols: []string{
			"id",
		},
	}
	data, err := p.bundle.PageListPipeline(pipelinePageListRequest)
	if err != nil {
		return nil, apierrors.ErrCancelProjectPipeline.InternalError(err)
	}
	if len(data.Pipelines) == 0 {
		return nil, nil
	}
	return &data.Pipelines[0], nil
}

func (p *ProjectPipelineService) getPipelineDefinition(pipelineDefinitionID string) (PipelineDefinition *dpb.PipelineDefinition, err error) {
	var getReq dpb.PipelineDefinitionGetRequest
	getReq.PipelineDefinitionID = pipelineDefinitionID
	resp, err := p.PipelineDefinition.Get(context.Background(), &getReq)
	if err != nil {
		return nil, err
	}
	if resp == nil || resp.PipelineDefinition == nil || resp.PipelineDefinition.Extra == nil {
		return nil, fmt.Errorf("not find pipeline definition")
	}
	return resp.PipelineDefinition, nil
}

func (p *ProjectPipelineService) getPipelineSource(sourceID string) (pipelineSource *spb.PipelineSource, err error) {
	var sourceGetReq = spb.PipelineSourceGetRequest{}
	sourceGetReq.PipelineSourceID = sourceID
	source, err := p.PipelineSource.Get(context.Background(), &sourceGetReq)
	if err != nil {
		return nil, err
	}
	if source == nil || source.PipelineSource == nil || source.PipelineSource.PipelineYml == "" {
		return nil, fmt.Errorf("source %v pipeline yml was empty", sourceID)
	}
	return source.PipelineSource, nil
}
