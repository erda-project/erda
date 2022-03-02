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
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	cmspb "github.com/erda-project/erda-proto-go/core/pipeline/cms/pb"
	dpb "github.com/erda-project/erda-proto-go/core/pipeline/definition/pb"
	spb "github.com/erda-project/erda-proto-go/core/pipeline/source/pb"
	guidepb "github.com/erda-project/erda-proto-go/dop/guide/pb"
	"github.com/erda-project/erda-proto-go/dop/projectpipeline/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/dop/providers/projectpipeline/deftype"
	"github.com/erda-project/erda/modules/dop/services/apierrors"
	"github.com/erda-project/erda/modules/dop/utils"
	def "github.com/erda-project/erda/modules/pipeline/providers/definition"
	"github.com/erda-project/erda/pkg/common/apis"
	"github.com/erda-project/erda/pkg/http/httpserver/errorresp"
	"github.com/erda-project/erda/pkg/limit_sync_group"
	"github.com/erda-project/erda/pkg/parser/pipelineyml"
	"github.com/erda-project/erda/pkg/time/mysql_time"
)

type CategoryType string

const (
	DefaultCategory  CategoryType = "default"
	StarCategory     CategoryType = "primary"
	DicePipelinePath string       = ".dice/pipelines"
	ErdaPipelinePath string       = ".erda/pipelines"

	CreateProjectPipelineNamePreCheckLocaleKey   string = "ProjectPipelineCreateNamePreCheckNotPass"
	CreateProjectPipelineSourcePreCheckLocaleKey string = "ProjectPipelineCreateSourcePreCheckNotPass"
)

func (c CategoryType) String() string {
	return string(c)
}

type PipelineType string

const (
	cicdPipelineType PipelineType = "cicd"
)

func (p PipelineType) String() string {
	return string(p)
}

func (s *ProjectPipelineService) ListPipelineYml(ctx context.Context, req *pb.ListAppPipelineYmlRequest) (*pb.ListAppPipelineYmlResponse, error) {

	app, err := s.bundle.GetApp(req.AppID)
	if err != nil {
		return nil, err
	}

	work := limit_sync_group.NewWorker(3)
	var list []*pb.PipelineYmlList
	var pathList = []string{"", DicePipelinePath, ErdaPipelinePath}
	for _, path := range pathList {
		work.AddFunc(func(locker *limit_sync_group.Locker, i ...interface{}) error {
			result, err := s.getPipelineYml(app, apis.GetUserID(ctx), req.Branch, i[0].(string))
			if err != nil {
				return nil
			}

			locker.Lock()
			defer locker.Unlock()
			list = append(list, result...)
			return nil
		}, path)
	}
	if err := work.Do().Error(); err != nil {
		return nil, err
	}

	return &pb.ListAppPipelineYmlResponse{
		Result: list,
	}, nil
}

func (s *ProjectPipelineService) getPipelineYml(app *apistructs.ApplicationDTO, userID string, branch string, findPath string) ([]*pb.PipelineYmlList, error) {
	var path string
	if findPath == "" {
		path = fmt.Sprintf("/wb/%v/%v/tree/%v", app.ProjectName, app.Name, branch)
	} else {
		path = fmt.Sprintf("/wb/%v/%v/tree/%v/%v", app.ProjectName, app.Name, branch, findPath)
	}

	diceEntrys, err := s.bundle.GetGittarTreeNode(path, strconv.Itoa(int(app.OrgID)), true, userID)
	if err != nil {
		return nil, err
	}

	var list []*pb.PipelineYmlList
	for _, entry := range diceEntrys {
		if !strings.HasSuffix(entry.Name, ".yml") {
			continue
		}
		if findPath == "" && entry.Name != apistructs.DefaultPipelineYmlName {
			continue
		}
		list = append(list, &pb.PipelineYmlList{
			YmlName: entry.Name,
			YmlPath: findPath,
		})
	}
	return list, nil
}

func (p *ProjectPipelineService) CreateSourcePreCheck(ctx context.Context, params *pb.CreateProjectPipelineSourcePreCheckRequest) (*pb.CreateProjectPipelineSourcePreCheckResponse, error) {
	if err := params.Validate(); err != nil {
		return nil, apierrors.ErrCreateProjectPipeline.InvalidParameter(err)
	}

	app, err := p.bundle.GetApp(params.AppID)
	if err != nil {
		return nil, err
	}

	resp, err := p.PipelineSource.List(ctx, &spb.PipelineSourceListRequest{
		SourceType: params.SourceType,
		Remote:     makeRemote(app),
		Ref:        params.Ref,
		Path:       params.Path,
		Name:       params.FileName,
	})
	if err != nil {
		return nil, apierrors.ErrCreateProjectPipeline.InvalidParameter(err)
	}
	if len(resp.Data) == 0 {
		return &pb.CreateProjectPipelineSourcePreCheckResponse{
			Pass: true,
		}, nil
	}

	definitionList, err := p.PipelineDefinition.List(ctx, &dpb.PipelineDefinitionListRequest{
		Location: makeLocation(&apistructs.ApplicationDTO{
			OrgName:     app.OrgName,
			ProjectName: app.ProjectName,
		}, cicdPipelineType),
		SourceIDList: []string{resp.Data[0].ID},
	})
	if err != nil {
		return nil, apierrors.ErrCreateProjectPipeline.InvalidParameter(err)
	}

	if len(definitionList.Data) == 0 {
		return &pb.CreateProjectPipelineSourcePreCheckResponse{
			Pass: true,
		}, nil
	}
	definitionName := definitionList.Data[0].Name

	return &pb.CreateProjectPipelineSourcePreCheckResponse{
		Pass:    false,
		Message: fmt.Sprintf(p.trans.Text(apis.Language(ctx), CreateProjectPipelineSourcePreCheckLocaleKey), definitionName),
	}, nil
}

func (p *ProjectPipelineService) CreateNamePreCheck(ctx context.Context, req *pb.CreateProjectPipelineNamePreCheckRequest) (*pb.CreateProjectPipelineNamePreCheckResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, apierrors.ErrCreateProjectPipeline.InvalidParameter(err)
	}
	haveSameNameDefinition, err := p.checkDefinitionRemoteSameName(req.ProjectID, req.Name, apis.GetUserID(ctx))
	if err != nil {
		return nil, apierrors.ErrCreateProjectPipeline.InternalError(err)
	}
	if haveSameNameDefinition {
		return &pb.CreateProjectPipelineNamePreCheckResponse{
			Pass:    false,
			Message: p.trans.Text(apis.Language(ctx), CreateProjectPipelineNamePreCheckLocaleKey),
		}, nil
	}

	return &pb.CreateProjectPipelineNamePreCheckResponse{
		Pass: true,
	}, nil
}

func (p *ProjectPipelineService) Create(ctx context.Context, params *pb.CreateProjectPipelineRequest) (*pb.CreateProjectPipelineResponse, error) {
	if err := params.Validate(); err != nil {
		return nil, apierrors.ErrCreateProjectPipeline.InvalidParameter(err)
	}

	err := p.checkRolePermission(apistructs.IdentityInfo{
		UserID: apis.GetUserID(ctx),
	}, &apistructs.PipelineCreateRequestV2{
		Labels: map[string]string{
			apistructs.LabelAppID:  strconv.FormatUint(params.AppID, 10),
			apistructs.LabelBranch: params.Ref,
		},
	}, apierrors.ErrCreateProjectPipeline)
	if err != nil {
		return nil, err
	}

	nameCheckResult, err := p.CreateNamePreCheck(ctx, &pb.CreateProjectPipelineNamePreCheckRequest{
		ProjectID: params.ProjectID,
		Name:      params.Name,
	})
	if err != nil {
		return nil, err
	}
	if !nameCheckResult.Pass {
		return nil, apierrors.ErrCreateProjectPipeline.InternalError(fmt.Errorf(nameCheckResult.Message))
	}

	sourceCheckResult, err := p.CreateSourcePreCheck(ctx, &pb.CreateProjectPipelineSourcePreCheckRequest{
		SourceType: params.SourceType,
		Ref:        params.Ref,
		Path:       params.Path,
		FileName:   params.FileName,
		AppID:      params.AppID,
	})
	if err != nil {
		return nil, err
	}
	if !sourceCheckResult.Pass {
		return nil, apierrors.ErrCreateProjectPipeline.InternalError(fmt.Errorf(sourceCheckResult.Message))
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

	location, err := p.makeLocationByAppID(params.AppID)
	if err != nil {
		return nil, apierrors.ErrCreateProjectPipeline.InternalError(err)
	}

	definitionRsp, err := p.PipelineDefinition.Create(ctx, &dpb.PipelineDefinitionCreateRequest{
		Location:         location,
		Name:             params.Name,
		Creator:          apis.GetUserID(ctx),
		PipelineSourceID: sourceRsp.PipelineSource.ID,
		Category:         DefaultCategory.String(),
		Extra: &dpb.PipelineDefinitionExtra{
			Extra: p.pipelineSourceType.GetPipelineCreateRequestV2(),
		},
	})
	if err != nil {
		return nil, apierrors.ErrCreateProjectPipeline.InternalError(err)
	}

	go func() {
		_, err = p.GuideSvc.ProcessGuide(ctx, &guidepb.ProcessGuideRequest{AppID: params.AppID, Branch: params.Ref})
		if err != nil {
			p.logger.Errorf("failed to ProcessGuide, err: %v, appID: %d, branch: %s", err, params.AppID, params.Ref)
		}
	}()
	return &pb.CreateProjectPipelineResponse{ProjectPipeline: &pb.ProjectPipeline{
		ID:               definitionRsp.PipelineDefinition.ID,
		Name:             definitionRsp.PipelineDefinition.Name,
		Creator:          definitionRsp.PipelineDefinition.Creator,
		Category:         definitionRsp.PipelineDefinition.Category,
		TimeCreated:      definitionRsp.PipelineDefinition.TimeCreated,
		TimeUpdated:      definitionRsp.PipelineDefinition.TimeUpdated,
		SourceType:       sourceRsp.PipelineSource.SourceType,
		Remote:           sourceRsp.PipelineSource.Remote,
		Ref:              sourceRsp.PipelineSource.Ref,
		Path:             sourceRsp.PipelineSource.Path,
		FileName:         sourceRsp.PipelineSource.Name,
		PipelineSourceID: sourceRsp.PipelineSource.ID,
	}}, nil
}

func (p *ProjectPipelineService) checkDefinitionRemoteSameName(projectID uint64, name string, userID string) (bool, error) {
	projectDto, err := p.bundle.GetProject(projectID)
	if err != nil {
		return false, err
	}

	orgDto, err := p.bundle.GetOrg(projectDto.OrgID)
	if err != nil {
		return false, err
	}

	location := makeLocation(&apistructs.ApplicationDTO{
		OrgName:     orgDto.Name,
		ProjectName: projectDto.Name,
	}, cicdPipelineType)

	resp, err := p.PipelineDefinition.List(context.Background(), &dpb.PipelineDefinitionListRequest{
		Location: location,
		Name:     name,
		PageNo:   1,
		PageSize: 999,
		Creator:  []string{userID},
	})
	if err != nil {
		return false, err
	}
	// pipelineDefinition list method param name was like search, should be determine whether it is equal
	for _, definition := range resp.Data {
		if definition.Name == name {
			return true, nil
		}
	}
	return false, nil
}

func (p *ProjectPipelineService) getYmlFromGittar(app *apistructs.ApplicationDTO, ref, filePath, userID string) (string, error) {
	commit, err := p.bundle.GetGittarCommit(app.GitRepoAbbrev, ref, userID)
	if err != nil {
		return "", err
	}

	yml, err := p.bundle.GetGittarFile(app.GitRepo, commit.ID, filePath, "", "", userID)
	return yml, err
}

func (p *ProjectPipelineService) List(ctx context.Context, params deftype.ProjectPipelineList) ([]*dpb.PipelineDefinition, int64, error) {
	if err := params.Validate(); err != nil {
		return nil, 0, apierrors.ErrListProjectPipeline.InvalidParameter(err)
	}

	project, err := p.bundle.GetProject(params.ProjectID)
	if err != nil {
		return nil, 0, apierrors.ErrListProjectPipeline.InternalError(err)
	}

	org, err := p.bundle.GetOrg(project.OrgID)
	if err != nil {
		return nil, 0, apierrors.ErrListProjectPipeline.InternalError(err)
	}

	var apps []apistructs.ApplicationDTO
	if len(params.AppName) == 0 {
		appResp, err := p.bundle.GetMyAppsByProject(params.IdentityInfo.UserID, project.OrgID, project.ID, "")
		if err != nil {
			return nil, 0, err
		}
		apps = appResp.List
	}
	for _, v := range apps {
		params.AppName = append(params.AppName, v.Name)
	}
	// No application returned directly
	if len(params.AppName) == 0 {
		return []*dpb.PipelineDefinition{}, 0, nil
	}

	list, err := p.PipelineDefinition.List(ctx, &dpb.PipelineDefinitionListRequest{
		PageSize: int64(params.PageSize),
		PageNo:   int64(params.PageNo),
		Location: makeLocation(&apistructs.ApplicationDTO{
			OrgName:     org.Name,
			ProjectName: project.Name,
		}, cicdPipelineType),
		Name:     params.Name,
		Creator:  params.Creator,
		Executor: params.Executor,
		Category: params.Category,
		Ref:      params.Ref,
		Remote: func() []string {
			remotes := make([]string, 0, len(params.AppName))
			for _, v := range params.AppName {
				remotes = append(remotes, makeRemote(&apistructs.ApplicationDTO{
					OrgName:     org.Name,
					ProjectName: project.Name,
					Name:        v,
				}))
			}
			return remotes
		}(),
		TimeCreated: params.TimeCreated,
		TimeStarted: params.TimeStarted,
		Status:      params.Status,
		AscCols:     params.AscCols,
		DescCols:    params.DescCols,
	})
	if err != nil {
		return nil, 0, apierrors.ErrListProjectPipeline.InternalError(err)
	}

	return list.Data, list.Total, nil
}

func (p *ProjectPipelineService) Delete(ctx context.Context, params deftype.ProjectPipelineDelete) (*deftype.ProjectPipelineDeleteResult, error) {
	if err := params.Validate(); err != nil {
		return nil, apierrors.ErrDeleteProjectPipeline.InvalidParameter(err)
	}

	definition, source, err := p.getPipelineDefinitionAndSource(params.ID)
	if err != nil {
		return nil, apierrors.ErrDeleteProjectPipeline.InvalidParameter(err)
	}
	if definition.Creator != apis.GetUserID(ctx) {
		return nil, apierrors.ErrDeleteProjectPipeline.AccessDenied()
	}
	err = p.checkDataPermissionByProjectID(params.ProjectID, source)
	if err != nil {
		return nil, apierrors.ErrDeleteProjectPipeline.AccessDenied()
	}
	if apistructs.PipelineStatus(definition.Status).IsRunningStatus() {
		return nil, apierrors.ErrDeleteProjectPipeline.InternalError(fmt.Errorf("pipeline wass running status"))
	}

	extraValue, err := def.GetExtraValue(definition)
	if err != nil {
		return nil, apierrors.ErrDeleteProjectPipeline.InternalError(fmt.Errorf("failed unmarshal pipeline extra error %v", err))
	}
	crons, err := p.bundle.PageListPipelineCrons(apistructs.PipelineCronPagingRequest{
		Sources:  []apistructs.PipelineSource{extraValue.CreateRequest.PipelineSource},
		YmlNames: []string{extraValue.CreateRequest.PipelineYmlName},
		PageSize: 1,
		PageNo:   1,
	})
	if err != nil {
		return nil, apierrors.ErrDeleteProjectPipeline.InternalError(err)
	}
	if len(crons.Data) > 0 && crons.Data[0].Enable != nil && *crons.Data[0].Enable == true {
		return nil, apierrors.ErrDeleteProjectPipeline.InternalError(fmt.Errorf("pipeline cron is running status"))
	}

	_, err = p.PipelineDefinition.Delete(ctx, &dpb.PipelineDefinitionDeleteRequest{PipelineDefinitionID: params.ID})
	return nil, err
}

func (p *ProjectPipelineService) Update(ctx context.Context, params deftype.ProjectPipelineUpdate) (*deftype.ProjectPipelineUpdateResult, error) {
	if err := params.Validate(); err != nil {
		return nil, err
	}

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
		PipelineSourceID:     sourceRsp.PipelineSource.ID,
	})

	return nil, err
}

func (p *ProjectPipelineService) SetPrimary(ctx context.Context, params deftype.ProjectPipelineCategory) (*dpb.PipelineDefinitionUpdateResponse, error) {
	if err := params.Validate(); err != nil {
		return nil, apierrors.ErrSetPrimaryProjectPipeline.InvalidParameter(err)
	}

	_, source, err := p.getPipelineDefinitionAndSource(params.PipelineDefinitionID)
	if err != nil {
		return nil, apierrors.ErrSetPrimaryProjectPipeline.InvalidParameter(err)
	}
	err = p.checkDataPermissionByProjectID(params.ProjectID, source)
	if err != nil {
		return nil, apierrors.ErrSetPrimaryProjectPipeline.AccessDenied()
	}

	definition, err := p.PipelineDefinition.Update(ctx, &dpb.PipelineDefinitionUpdateRequest{
		PipelineDefinitionID: params.PipelineDefinitionID,
		Category:             StarCategory.String(),
	})
	if err != nil {
		return nil, apierrors.ErrSetPrimaryProjectPipeline.InternalError(err)
	}

	return definition, nil
}

func (p *ProjectPipelineService) UnSetPrimary(ctx context.Context, params deftype.ProjectPipelineCategory) (*dpb.PipelineDefinitionUpdateResponse, error) {
	if err := params.Validate(); err != nil {
		return nil, apierrors.ErrUnSetPrimaryProjectPipeline.InvalidParameter(err)
	}

	_, source, err := p.getPipelineDefinitionAndSource(params.PipelineDefinitionID)
	if err != nil {
		return nil, apierrors.ErrUnSetPrimaryProjectPipeline.InvalidParameter(err)
	}
	err = p.checkDataPermissionByProjectID(params.ProjectID, source)
	if err != nil {
		return nil, apierrors.ErrUnSetPrimaryProjectPipeline.AccessDenied()
	}

	definition, err := p.PipelineDefinition.Update(ctx, &dpb.PipelineDefinitionUpdateRequest{
		PipelineDefinitionID: params.PipelineDefinitionID,
		Category:             DefaultCategory.String(),
	})
	if err != nil {
		return nil, apierrors.ErrUnSetPrimaryProjectPipeline.InternalError(err)
	}

	return definition, nil
}

func (p *ProjectPipelineService) Run(ctx context.Context, params deftype.ProjectPipelineRun) (*deftype.ProjectPipelineRunResult, error) {
	if params.PipelineDefinitionID == "" {
		return nil, apierrors.ErrRunProjectPipeline.InvalidParameter(fmt.Errorf("pipelineDefinitionID：%s", params.PipelineDefinitionID))
	}

	definition, source, err := p.getPipelineDefinitionAndSource(params.PipelineDefinitionID)
	if err != nil {
		return nil, apierrors.ErrRunProjectPipeline.InternalError(err)
	}
	err = p.checkDataPermissionByProjectID(params.ProjectID, source)
	if err != nil {
		return nil, apierrors.ErrRunProjectPipeline.AccessDenied()
	}

	value, err := p.autoRunPipeline(params.IdentityInfo, definition, source)
	if err != nil {
		return nil, err
	}

	return &deftype.ProjectPipelineRunResult{
		Pipeline: value,
	}, nil
}

func (p *ProjectPipelineService) StartCron(ctx context.Context, params deftype.ProjectPipelineStartCron) (*deftype.ProjectPipelineStartCronResult, error) {
	dto, err := p.startOrEndCron(params.IdentityInfo, params.PipelineDefinitionID, params.ProjectID, true, apierrors.ErrStartCronProjectPipeline)
	if err != nil {
		return nil, err
	}

	return &deftype.ProjectPipelineStartCronResult{
		Cron: dto,
	}, nil
}

func (p *ProjectPipelineService) EndCron(ctx context.Context, params deftype.ProjectPipelineEndCron) (*deftype.ProjectPipelineEndCronResult, error) {

	dto, err := p.startOrEndCron(params.IdentityInfo, params.PipelineDefinitionID, params.ProjectID, false, apierrors.ErrEndCronProjectPipeline)
	if err != nil {
		return nil, err
	}
	return &deftype.ProjectPipelineEndCronResult{
		Cron: dto,
	}, nil
}

func (p *ProjectPipelineService) ListExecHistory(ctx context.Context, params deftype.ProjectPipelineListExecHistory) (*deftype.ProjectPipelineListExecHistoryResult, error) {
	var pipelineDefinition = apistructs.PipelineDefinitionRequest{}
	pipelineDefinition.Name = params.Name
	pipelineDefinition.Creators = params.Executors

	if params.ProjectID == 0 {
		return nil, apierrors.ErrListExecHistoryProjectPipeline.InternalError(fmt.Errorf("projectID can not empty"))
	}
	project, err := p.bundle.GetProject(params.ProjectID)
	if err != nil {
		return nil, apierrors.ErrListExecHistoryProjectPipeline.InternalError(err)
	}
	appData, err := p.bundle.GetMyAppsByProject(params.IdentityInfo.UserID, project.OrgID, project.ID, "")
	if err != nil {
		return nil, apierrors.ErrListExecHistoryProjectPipeline.InternalError(err)
	}
	for _, app := range appData.List {
		// search all project app
		if len(params.AppIDList) <= 0 {
			pipelineDefinition.SourceRemotes = append(pipelineDefinition.SourceRemotes, makeRemote(&app))
			continue
		}
		// search user choose project app
		for _, appID := range params.AppIDList {
			if appID != app.ID {
				continue
			}
			pipelineDefinition.SourceRemotes = append(pipelineDefinition.SourceRemotes, makeRemote(&app))
		}
	}
	// No application returned directly
	if len(pipelineDefinition.SourceRemotes) == 0 {
		return &deftype.ProjectPipelineListExecHistoryResult{
			Data: &apistructs.PipelinePageListData{
				Total:           0,
				CurrentPageSize: int64(params.PageSize),
			},
		}, nil
	}

	jsonValue, err := json.Marshal(pipelineDefinition)
	if err != nil {
		return nil, apierrors.ErrListExecHistoryProjectPipeline.InternalError(err)
	}

	var pipelinePageListRequest = apistructs.PipelinePageListRequest{
		PageNum:                             int(params.PageNo),
		PageSize:                            int(params.PageSize),
		Statuses:                            params.Statuses,
		AllSources:                          true,
		StartTimeBeginTimestamp:             params.StartTimeBegin.Unix(),
		EndTimeBeginTimestamp:               params.StartTimeEnd.Unix(),
		PipelineDefinitionRequestJSONBase64: base64.StdEncoding.EncodeToString(jsonValue),
		DescCols:                            params.DescCols,
		AscCols:                             params.AscCols,
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
	definitionMap, err := p.batchGetPipelineDefinition(params.PipelineDefinitionIDs, params.ProjectID)
	if err != nil {
		return nil, apierrors.ErrBatchRunProjectPipeline.InternalError(err)
	}

	var pipelineSourceIDArray []string
	for _, v := range definitionMap {
		if v.PipelineSourceID == "" {
			return nil, apierrors.ErrBatchRunProjectPipeline.InternalError(fmt.Errorf("definition %v pipeline source was empty", v.ID))
		}
		pipelineSourceIDArray = append(pipelineSourceIDArray, v.PipelineSourceID)
	}

	sourceMap, err := p.batchGetPipelineSources(pipelineSourceIDArray)
	if err != nil {
		return nil, apierrors.ErrRunProjectPipeline.InternalError(err)
	}

	project, err := p.bundle.GetProject(params.ProjectID)
	if err != nil {
		return nil, apierrors.ErrRunProjectPipeline.InternalError(err)
	}
	org, err := p.bundle.GetOrg(project.OrgID)
	if err != nil {
		return nil, apierrors.ErrRunProjectPipeline.InternalError(err)
	}
	for _, source := range sourceMap {
		err := p.checkDataPermission(project, org, source)
		if err != nil {
			return nil, err
		}
	}

	work := limit_sync_group.NewWorker(5)
	var result = map[string]*apistructs.PipelineDTO{}

	for _, v := range definitionMap {
		work.AddFunc(func(locker *limit_sync_group.Locker, i ...interface{}) error {
			var definitionID = i[0].(string)
			var sourceID = i[1].(string)

			value, err := p.autoRunPipeline(params.IdentityInfo, definitionMap[definitionID], sourceMap[sourceID])
			if err != nil {
				return err
			}

			locker.Lock()
			result[definitionID] = value
			locker.Unlock()
			return nil
		}, v.ID, v.PipelineSourceID)
	}
	if err := work.Do().Error(); err != nil {
		return nil, err
	}

	return &deftype.ProjectPipelineBatchRunResult{
		PipelineMap: result,
	}, nil
}

func (p *ProjectPipelineService) Cancel(ctx context.Context, params deftype.ProjectPipelineCancel) (*deftype.ProjectPipelineCancelResult, error) {
	if params.PipelineDefinitionID == "" {
		return nil, apierrors.ErrCancelProjectPipeline.InvalidParameter(fmt.Errorf("pipelineDefinitionID：%s", params.PipelineDefinitionID))
	}

	definition, source, err := p.getPipelineDefinitionAndSource(params.PipelineDefinitionID)
	if err != nil {
		return nil, apierrors.ErrCancelProjectPipeline.InternalError(err)
	}
	err = p.checkDataPermissionByProjectID(params.ProjectID, source)
	if err != nil {
		return nil, apierrors.ErrCancelProjectPipeline.AccessDenied()
	}

	extraValue, err := def.GetExtraValue(definition)
	if err != nil {
		return nil, apierrors.ErrCancelProjectPipeline.InternalError(fmt.Errorf("failed unmarshal pipeline extra error %v", err))
	}

	if err := p.checkRolePermission(params.IdentityInfo, extraValue.CreateRequest, apierrors.ErrCancelProjectPipeline); err != nil {
		return nil, err
	}

	pipelineInfo, err := p.bundle.GetPipeline(uint64(definition.PipelineID))
	if err != nil {
		return nil, apierrors.ErrCancelProjectPipeline.InternalError(err)
	}

	if pipelineInfo.Status.IsRunningStatus() {
		var req apistructs.PipelineCancelRequest
		req.PipelineID = uint64(definition.PipelineID)
		req.IdentityInfo = params.IdentityInfo
		err = p.bundle.CancelPipeline(req)
		if err != nil {
			return nil, apierrors.ErrCancelProjectPipeline.InternalError(err)
		}

		_, err = p.PipelineDefinition.Update(context.Background(), &dpb.PipelineDefinitionUpdateRequest{PipelineDefinitionID: definition.ID, Status: string(apistructs.PipelineStatusStopByUser), PipelineID: definition.PipelineID})
		if err != nil {
			return nil, apierrors.ErrCancelProjectPipeline.InternalError(err)
		}

		return &deftype.ProjectPipelineCancelResult{}, nil
	}

	_, err = p.PipelineDefinition.Update(context.Background(), &dpb.PipelineDefinitionUpdateRequest{PipelineDefinitionID: definition.ID, Status: string(pipelineInfo.Status), PipelineID: definition.PipelineID})
	if err != nil {
		return nil, apierrors.ErrCancelProjectPipeline.InternalError(err)
	}

	return &deftype.ProjectPipelineCancelResult{}, nil
}

func (p *ProjectPipelineService) Rerun(ctx context.Context, params deftype.ProjectPipelineRerun) (*deftype.ProjectPipelineRerunResult, error) {
	dto, err := p.failRerunOrRerunPipeline(true, params.PipelineDefinitionID, params.ProjectID, params.IdentityInfo, apierrors.ErrRerunProjectPipeline)
	if err != nil {
		return nil, err
	}

	return &deftype.ProjectPipelineRerunResult{
		Pipeline: dto,
	}, nil
}

func (p *ProjectPipelineService) FailRerun(ctx context.Context, params deftype.ProjectPipelineFailRerun) (*deftype.ProjectPipelineFailRerunResult, error) {
	dto, err := p.failRerunOrRerunPipeline(false, params.PipelineDefinitionID, params.ProjectID, params.IdentityInfo, apierrors.ErrFailRerunProjectPipeline)
	if err != nil {
		return nil, err
	}
	return &deftype.ProjectPipelineFailRerunResult{
		Pipeline: dto,
	}, nil
}

func (p *ProjectPipelineService) failRerunOrRerunPipeline(rerun bool, pipelineDefinitionID string, projectID uint64, identityInfo apistructs.IdentityInfo, apiError *errorresp.APIError) (*apistructs.PipelineDTO, error) {
	if pipelineDefinitionID == "" {
		return nil, apiError.InvalidParameter(fmt.Errorf("pipelineDefinitionID：%s", pipelineDefinitionID))
	}

	definition, source, err := p.getPipelineDefinitionAndSource(pipelineDefinitionID)
	if err != nil {
		return nil, apiError.InternalError(err)
	}
	err = p.checkDataPermissionByProjectID(projectID, source)
	if err != nil {
		return nil, apiError.AccessDenied()
	}

	extraValue, err := def.GetExtraValue(definition)
	if err != nil {
		return nil, apiError.InternalError(fmt.Errorf("failed unmarshal pipeline extra error %v", err))
	}

	if err := p.checkRolePermission(identityInfo, extraValue.CreateRequest, apiError); err != nil {
		return nil, err
	}

	var dto *apistructs.PipelineDTO
	if rerun {
		var req apistructs.PipelineRerunRequest
		req.PipelineID = uint64(definition.PipelineID)
		req.AutoRunAtOnce = true
		req.IdentityInfo = identityInfo
		dto, err = p.bundle.RerunPipeline(req)
	} else {
		var req apistructs.PipelineRerunFailedRequest
		req.PipelineID = uint64(definition.PipelineID)
		req.AutoRunAtOnce = true
		req.IdentityInfo = identityInfo
		dto, err = p.bundle.RerunFailedPipeline(req)
	}
	if err != nil {
		return nil, apiError.InternalError(err)
	}

	definitionUpdateReq := &dpb.PipelineDefinitionUpdateRequest{
		PipelineDefinitionID: definition.ID,
		Status:               string(apistructs.StatusRunning),
		Executor:             identityInfo.UserID,
		EndedAt:              timestamppb.New(*mysql_time.GetMysqlDefaultTime()),
		PipelineID:           int64(dto.ID)}
	if rerun {
		definitionUpdateReq.ExecutedActionNum = -1
		definitionUpdateReq.StartedAt = timestamppb.New(time.Now())
	}

	_, err = p.PipelineDefinition.Update(context.Background(), definitionUpdateReq)
	if err != nil {
		return nil, apierrors.ErrRunProjectPipeline.InternalError(err)
	}
	return dto, nil
}

func (p *ProjectPipelineService) startOrEndCron(identityInfo apistructs.IdentityInfo, pipelineDefinitionID string, projectID uint64, enable bool, apiError *errorresp.APIError) (*apistructs.PipelineCronDTO, error) {
	if pipelineDefinitionID == "" {
		return nil, apiError.InvalidParameter(fmt.Errorf("pipelineDefinitionID：%s", pipelineDefinitionID))
	}

	definition, source, err := p.getPipelineDefinitionAndSource(pipelineDefinitionID)
	if err != nil {
		return nil, apiError.InternalError(err)
	}
	err = p.checkDataPermissionByProjectID(projectID, source)
	if err != nil {
		return nil, apiError.AccessDenied()
	}

	extraValue, err := def.GetExtraValue(definition)
	if err != nil {
		return nil, apiError.InternalError(fmt.Errorf("failed unmarshal pipeline extra error %v", err))
	}

	if err := p.checkRolePermission(identityInfo, extraValue.CreateRequest, apiError); err != nil {
		return nil, err
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

	if cron.Data[0].PipelineDefinitionID == "" {
		err := p.bundle.UpdatePipelineCron(apistructs.PipelineCronUpdateRequest{
			ID:                     cron.Data[0].ID,
			PipelineYml:            cron.Data[0].PipelineYml,
			PipelineDefinitionID:   pipelineDefinitionID,
			CronExpr:               cron.Data[0].CronExpr,
			ConfigManageNamespaces: cron.Data[0].ConfigManageNamespaces,
		})
		if err != nil {
			return nil, apiError.InternalError(err)
		}
	}

	var dto *apistructs.PipelineCronDTO
	if enable {
		orgStr := extraValue.CreateRequest.Labels[apistructs.LabelOrgID]
		orgID, err := strconv.ParseUint(orgStr, 10, 64)
		if err != nil {
			return nil, apiError.InternalError(fmt.Errorf("not found orgID"))
		}
		// update CmsNsConfigs
		if err = p.UpdateCmsNsConfigs(identityInfo.UserID, orgID); err != nil {
			return nil, apiError.InternalError(err)
		}

		dto, err = p.bundle.StartPipelineCron(cron.Data[0].ID)
		if err != nil {
			return nil, apiError.InternalError(err)
		}
	} else {
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

func (p *ProjectPipelineService) getPipelineDefinitionAndSource(pipelineDefinitionID string) (PipelineDefinition *dpb.PipelineDefinition, pipelineSource *spb.PipelineSource, err error) {
	definition, err := p.getPipelineDefinition(pipelineDefinitionID)
	if err != nil {
		return nil, nil, err
	}
	source, err := p.getPipelineSource(definition.PipelineSourceID)
	if err != nil {
		return nil, nil, err
	}
	return definition, source, nil
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

func (p *ProjectPipelineService) batchGetPipelineDefinition(pipelineDefinitionIDArray []string, projectID uint64) (map[string]*dpb.PipelineDefinition, error) {
	location, err := p.makeLocationByProjectID(projectID)
	if err != nil {
		return nil, err
	}
	resp, err := p.PipelineDefinition.List(context.Background(), &dpb.PipelineDefinitionListRequest{
		PageNo:   1,
		PageSize: int64(len(pipelineDefinitionIDArray)),
		Location: location,
		IdList:   pipelineDefinitionIDArray,
	})
	if err != nil {
		return nil, err
	}
	var pipelineDefinitionMap = map[string]*dpb.PipelineDefinition{}
	for _, v := range resp.Data {
		pipelineDefinitionMap[v.ID] = v
	}
	return pipelineDefinitionMap, nil
}

func (p *ProjectPipelineService) batchGetPipelineSources(pipelineSourceIDArray []string) (map[string]*spb.PipelineSource, error) {
	var pipelineSourceListRequest spb.PipelineSourceListRequest
	pipelineSourceListRequest.IdList = pipelineSourceIDArray
	resp, err := p.PipelineSource.List(context.Background(), &pipelineSourceListRequest)
	if err != nil {
		return nil, err
	}
	var pipelineSourceMap = map[string]*spb.PipelineSource{}
	for _, v := range resp.Data {
		pipelineSourceMap[v.ID] = v
	}
	return pipelineSourceMap, nil
}

func (p *ProjectPipelineService) autoRunPipeline(identityInfo apistructs.IdentityInfo, definition *dpb.PipelineDefinition, source *spb.PipelineSource) (*apistructs.PipelineDTO, error) {
	extraValue, err := def.GetExtraValue(definition)
	if err != nil {
		return nil, apierrors.ErrRunProjectPipeline.InternalError(fmt.Errorf("failed unmarshal pipeline extra error %v", err))
	}
	createV2 := extraValue.CreateRequest

	if err := p.checkRolePermission(identityInfo, extraValue.CreateRequest, apierrors.ErrRunProjectPipeline); err != nil {
		return nil, err
	}

	appIDString := createV2.Labels[apistructs.LabelAppID]
	appID, err := strconv.ParseUint(appIDString, 10, 64)
	if err != nil {
		return nil, err
	}

	// update user gittar token
	var worker = limit_sync_group.NewWorker(3)
	worker.AddFunc(func(locker *limit_sync_group.Locker, i ...interface{}) error {
		orgStr := createV2.Labels[apistructs.LabelOrgID]
		orgID, err := strconv.ParseUint(orgStr, 10, 64)
		if err != nil {
			return apierrors.ErrRunProjectPipeline.InternalError(fmt.Errorf("not find orgID"))
		}
		// update CmsNsConfigs
		if err = p.UpdateCmsNsConfigs(identityInfo.UserID, orgID); err != nil {
			return apierrors.ErrRunProjectPipeline.InternalError(err)
		}
		return nil
	})

	worker.AddFunc(func(locker *limit_sync_group.Locker, i ...interface{}) error {
		createV2, err = p.pipelineSvc.ConvertPipelineToV2(&apistructs.PipelineCreateRequest{
			PipelineYmlName:    filepath.Join(source.Path, source.Name),
			AppID:              appID,
			Branch:             createV2.Labels[apistructs.LabelBranch],
			PipelineYmlContent: source.PipelineYml,
			UserID:             identityInfo.UserID,
		})
		if err != nil {
			return err
		}
		return nil
	})
	if worker.Do().Error() != nil {
		return nil, worker.Error()
	}

	createV2.AutoRunAtOnce = true
	createV2.DefinitionID = definition.ID
	createV2.UserID = identityInfo.UserID

	value, err := p.bundle.CreatePipeline(createV2)
	if err != nil {
		return nil, apierrors.ErrRunProjectPipeline.InternalError(err)
	}
	totalActionNum, err := countActionNumByPipelineYml(source.PipelineYml)
	if err != nil {
		return nil, apierrors.ErrRunProjectPipeline.InternalError(err)
	}
	_, err = p.PipelineDefinition.Update(context.Background(), &dpb.PipelineDefinitionUpdateRequest{
		PipelineDefinitionID: definition.ID,
		Status:               string(apistructs.StatusRunning),
		Executor:             identityInfo.UserID,
		StartedAt:            timestamppb.New(time.Now()),
		EndedAt:              timestamppb.New(*mysql_time.GetMysqlDefaultTime()),
		TotalActionNum:       totalActionNum,
		ExecutedActionNum:    -1,
		CostTime:             -1,
		PipelineID:           int64(value.ID)})
	if err != nil {
		return nil, apierrors.ErrRunProjectPipeline.InternalError(err)
	}
	return value, nil
}

func countActionNumByPipelineYml(pipelineYmlStr string) (int64, error) {
	pipelineYml, err := pipelineyml.New([]byte(pipelineYmlStr))
	if err != nil {
		return 0, apierrors.ErrCreateProjectPipeline.InternalError(err)
	}

	var totalActionNum int64
	pipelineYml.Spec().LoopStagesActions(func(stage int, action *pipelineyml.Action) {
		totalActionNum++
	})
	return totalActionNum, nil
}

func (p *ProjectPipelineService) ListApp(ctx context.Context, params *pb.ListAppRequest) (*pb.ListAppResponse, error) {
	if err := params.Validate(); err != nil {
		return nil, apierrors.ErrListAppProjectPipeline.InvalidParameter(err)
	}

	project, err := p.bundle.GetProject(params.ProjectID)
	if err != nil {
		return nil, apierrors.ErrListAppProjectPipeline.InternalError(err)
	}

	org, err := p.bundle.GetOrg(project.OrgID)
	if err != nil {
		return nil, apierrors.ErrListAppProjectPipeline.InternalError(err)
	}

	appResp, err := p.bundle.GetMyAppsByProject(apis.GetUserID(ctx), project.OrgID, project.ID, params.Name)
	if err != nil {
		return nil, apierrors.ErrListAppProjectPipeline.InternalError(err)
	}

	appNames := make([]string, 0, len(appResp.List))
	for _, v := range appResp.List {
		appNames = append(appNames, v.Name)
	}

	statics, err := p.PipelineDefinition.StaticsGroupByRemote(ctx, &dpb.PipelineDefinitionStaticsRequest{
		Location: makeLocation(&apistructs.ApplicationDTO{
			OrgName:     org.Name,
			ProjectName: project.Name,
		}, cicdPipelineType),
	})
	if err != nil {
		return nil, apierrors.ErrListAppProjectPipeline.InternalError(err)
	}

	appNamePipelineNumMap := make(map[string]*pipelineNum)
	for _, v := range appNames {
		appNamePipelineNumMap[v] = &pipelineNum{
			RunningNum: 0,
			FailedNum:  0,
			TotalNum:   0,
		}
	}

	for _, v := range statics.GetPipelineDefinitionStatistics() {
		appName := getAppNameByRemote(v.Remote)
		if v2, ok := appNamePipelineNumMap[appName]; ok {
			v2.FailedNum = int(v.FailedNum)
			v2.RunningNum = int(v.RunningNum)
			v2.TotalNum = int(v.TotalNum)
		}
	}

	apps := make([]*pb.Application, 0, len(appResp.List))
	for _, v := range appResp.List {
		apps = append(apps, &pb.Application{
			ID:             v.ID,
			Name:           v.Name,
			DisplayName:    v.DisplayName,
			Mode:           v.Mode,
			Desc:           v.Desc,
			Logo:           v.Logo,
			IsPublic:       v.IsPublic,
			Creator:        v.Creator,
			GitRepo:        v.GitRepo,
			OrgID:          v.OrgID,
			OrgDisplayName: v.OrgDisplayName,
			ProjectID:      v.ProjectID,
			ProjectName:    v.ProjectName,
			IsExternalRepo: v.IsExternalRepo,
			CreatedAt:      timestamppb.New(v.CreatedAt),
			UpdatedAt:      timestamppb.New(v.UpdatedAt),
			RunningNum:     uint64(appNamePipelineNumMap[v.Name].RunningNum),
			FailedNum:      uint64(appNamePipelineNumMap[v.Name].FailedNum),
			TotalNum:       uint64(appNamePipelineNumMap[v.Name].TotalNum),
		})
	}
	sort.Slice(apps, func(i, j int) bool {
		return apps[i].TotalNum > apps[j].TotalNum
	})
	return &pb.ListAppResponse{
		Data: apps,
	}, nil
}

type pipelineNum struct {
	RunningNum int `json:"runningNum"`
	FailedNum  int `json:"failedNum"`
	TotalNum   int `json:"totalNum"`
}

func (p *ProjectPipelineService) checkRolePermission(identityInfo apistructs.IdentityInfo, createRequest *apistructs.PipelineCreateRequestV2, apiError *errorresp.APIError) error {
	appIDString := createRequest.Labels[apistructs.LabelAppID]
	appID, err := strconv.ParseInt(appIDString, 10, 64)
	if err != nil {
		return apiError.InternalError(fmt.Errorf("definition extras not find appID, err: %v", err.Error()))
	}
	if err := p.Permission.CheckRuntimeBranch(identityInfo, uint64(appID), createRequest.Labels[apistructs.LabelBranch], apistructs.OperateAction); err != nil {
		return apiError.AccessDenied()
	}
	return nil
}

func (p *ProjectPipelineService) checkDataPermission(project *apistructs.ProjectDTO, org *apistructs.OrgDTO, source *spb.PipelineSource) error {
	if !strings.HasPrefix(source.Remote, filepath.Join(org.Name, project.Name)) {
		return fmt.Errorf("no permission")
	}
	return nil
}

func (p *ProjectPipelineService) checkDataPermissionByProjectID(projectID uint64, source *spb.PipelineSource) error {
	project, err := p.bundle.GetProject(projectID)
	if err != nil {
		return err
	}

	org, err := p.bundle.GetOrg(project.OrgID)
	if err != nil {
		return err
	}

	return p.checkDataPermission(project, org, source)
}

// UpdateCmsNsConfigs update CmsNsConfigs
func (e *ProjectPipelineService) UpdateCmsNsConfigs(userID string, orgID uint64) error {
	members, err := e.bundle.GetMemberByUserAndScope(apistructs.OrgScope, userID, orgID)
	if err != nil {
		return err
	}

	if len(members) <= 0 {
		return errors.New("the member is not exist")
	}

	_, err = e.PipelineCms.UpdateCmsNsConfigs(apis.WithInternalClientContext(context.Background(), "dop"),
		&cmspb.CmsNsConfigsUpdateRequest{
			Ns:             utils.MakeUserOrgPipelineCmsNs(userID, orgID),
			PipelineSource: apistructs.PipelineSourceDice.String(),
			KVs: map[string]*cmspb.PipelineCmsConfigValue{
				utils.MakeOrgGittarUsernamePipelineCmsNsConfig(): {Value: "git", EncryptInDB: true},
				utils.MakeOrgGittarTokenPipelineCmsNsConfig():    {Value: members[0].Token, EncryptInDB: true}},
		})

	return err
}

func makeLocation(app *apistructs.ApplicationDTO, t PipelineType) string {
	return filepath.Join(t.String(), app.OrgName, app.ProjectName)
}

func (p *ProjectPipelineService) makeLocationByProjectID(projectID uint64) (string, error) {
	projectDto, err := p.bundle.GetProject(projectID)
	if err != nil {
		return "", err
	}
	orgDto, err := p.bundle.GetOrg(projectDto.OrgID)
	if err != nil {
		return "", err
	}

	return makeLocation(&apistructs.ApplicationDTO{
		OrgName:     orgDto.Name,
		ProjectName: projectDto.Name,
	}, cicdPipelineType), nil
}

func (p *ProjectPipelineService) makeLocationByAppID(appID uint64) (string, error) {
	app, err := p.bundle.GetApp(appID)
	if err != nil {
		return "", err
	}
	return makeLocation(&apistructs.ApplicationDTO{
		OrgName:     app.OrgName,
		ProjectName: app.ProjectName,
	}, cicdPipelineType), nil
}

func getAppNameByRemote(remote string) string {
	splits := strings.Split(remote, string(filepath.Separator))
	if len(splits) != 3 {
		return ""
	}
	return splits[2]
}
