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

package pipeline

import (
	"context"
	"strconv"
	"time"

	"github.com/pkg/errors"

	basepb "github.com/erda-project/erda-proto-go/core/pipeline/base/pb"
	"github.com/erda-project/erda-proto-go/core/pipeline/pipeline/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/internal/tools/pipeline/commonutil/thirdparty/gittarutil"
	"github.com/erda-project/erda/internal/tools/pipeline/conf"
	"github.com/erda-project/erda/internal/tools/pipeline/providers/cms"
	"github.com/erda-project/erda/internal/tools/pipeline/services/apierrors"
	"github.com/erda-project/erda/internal/tools/pipeline/spec"
	"github.com/erda-project/erda/pkg/common/apis"
	"github.com/erda-project/erda/pkg/discover"
	"github.com/erda-project/erda/pkg/parser/pipelineyml"
)

func (s *pipelineService) PipelineCreate(ctx context.Context, req *pb.PipelineCreateRequest) (*pb.PipelineCreateResponse, error) {
	identityInfo := apis.GetIdentityInfo(ctx)
	if req.UserID == "" && identityInfo != nil {
		req.UserID = identityInfo.UserID
	}

	p, err := s.makePipelineFromRequest(req)
	if err != nil {
		return nil, err
	}

	var stages []spec.PipelineStage
	if stages, err = s.CreatePipelineGraph(p); err != nil {
		return nil, err
	}

	// PreCheck
	_ = s.PreCheck(p, stages, p.GetUserID(), req.AutoRun)

	// 是否自动执行
	if req.AutoRun {
		if p, err = s.run.RunOnePipeline(ctx, &pb.PipelineRunRequest{
			PipelineID: p.ID,
			UserID:     req.UserID,
		}); err != nil {
			return nil, err
		}
	}

	return &pb.PipelineCreateResponse{
		Data: s.ConvertPipeline(p),
	}, nil
}

func (s *pipelineService) makePipelineFromRequest(req *pb.PipelineCreateRequest) (*spec.Pipeline, error) {
	p := &spec.Pipeline{
		PipelineExtra: spec.PipelineExtra{
			NormalLabels: make(map[string]string),
			Extra: spec.PipelineExtraInfo{
				// --- empty user ---
				SubmitUser: &basepb.PipelineUser{},
				RunUser:    &basepb.PipelineUser{},
				CancelUser: &basepb.PipelineUser{},
			},
		},
		Labels: make(map[string]string),
	}

	// --- app info ---
	app, err := s.appSvc.GetWorkspaceApp(req.AppID, req.Branch)
	if err != nil {
		return nil, apierrors.ErrGetApp.InternalError(err)
	}
	p.Labels[apistructs.LabelOrgID] = strconv.FormatUint(app.OrgID, 10)
	p.NormalLabels[apistructs.LabelOrgName] = app.OrgName
	p.Labels[apistructs.LabelProjectID] = strconv.FormatUint(app.ProjectID, 10)
	p.NormalLabels[apistructs.LabelProjectName] = app.ProjectName
	p.Labels[apistructs.LabelAppID] = strconv.FormatUint(app.ID, 10)
	p.NormalLabels[apistructs.LabelAppName] = app.Name
	p.ClusterName = app.ClusterName
	p.Extra.DiceWorkspace = app.Workspace

	// --- repo info ---
	repo := gittarutil.NewRepo(discover.Gittar(), app.GitRepoAbbrev)
	commit, err := repo.GetCommit(req.Branch, req.UserID)
	if err != nil {
		return nil, apierrors.ErrGetGittarRepo.InternalError(err)
	}
	p.Labels[apistructs.LabelBranch] = req.Branch
	p.CommitDetail = apistructs.CommitDetail{
		CommitID: commit.ID,
		Repo:     app.GitRepo,
		RepoAbbr: app.GitRepoAbbrev,
		Author:   commit.Committer.Name,
		Email:    commit.Committer.Email,
		Time: func() *time.Time {
			commitTime, err := time.Parse("2006-01-02T15:04:05-07:00", commit.Committer.When)
			if err != nil {
				return nil
			}
			return &commitTime
		}(),
		Comment: commit.CommitMessage,
	}

	// --- yaml info ---
	if req.Source == "" {
		return nil, apierrors.ErrCreatePipeline.MissingParameter("source")
	}
	if !apistructs.PipelineSource(req.Source).Valid() {
		return nil, apierrors.ErrCreatePipeline.InvalidParameter(errors.Errorf("source: %s", req.Source))
	}
	p.PipelineSource = apistructs.PipelineSource(req.Source)

	if req.PipelineYmlName == "" {
		req.PipelineYmlName = apistructs.DefaultPipelineYmlName
	}

	// PipelineYmlNameV1 用于从 gittar 中获取 pipeline.yml 内容
	p.Extra.PipelineYmlNameV1 = req.PipelineYmlName
	p.PipelineYmlName = p.GenerateV1UniquePipelineYmlName(p.Extra.PipelineYmlNameV1)

	if req.PipelineYmlSource == "" {
		return nil, apierrors.ErrCreatePipeline.MissingParameter("pipelineYmlSource")
	}
	if !apistructs.PipelineYmlSource(req.PipelineYmlSource).Valid() {
		return nil, apierrors.ErrCreatePipeline.InvalidParameter(errors.Errorf("pipelineYmlSource: %s", req.PipelineYmlSource))
	}
	p.Extra.PipelineYmlSource = apistructs.PipelineYmlSource(req.PipelineYmlSource)
	switch apistructs.PipelineYmlSource(req.PipelineYmlSource) {
	case apistructs.PipelineYmlSourceGittar:
		// get yaml
		f, err := repo.FetchFile(req.Branch, p.Extra.PipelineYmlNameV1, req.UserID)
		if err != nil {
			return nil, apierrors.ErrGetGittarRepoFile.InternalError(err)
		}
		p.PipelineYml = string(f)
	case apistructs.PipelineYmlSourceContent:
		if req.PipelineYmlContent == "" {
			return nil, apierrors.ErrCreatePipeline.MissingParameter("pipelineYmlContent (pipelineYmlSource=content)")
		}
		p.PipelineYml = req.PipelineYmlContent
	}

	// --- run info ---
	p.Type = apistructs.PipelineTypeNormal
	p.TriggerMode = apistructs.PipelineTriggerModeManual
	if req.IsCronTriggered {
		p.TriggerMode = apistructs.PipelineTriggerModeCron
	}
	p.Status = apistructs.PipelineStatusAnalyzed

	// set storageConfig
	p.Extra.StorageConfig.EnableNFS = true
	if conf.DisablePipelineVolume() {
		p.Extra.StorageConfig.EnableNFS = false
		p.Extra.StorageConfig.EnableLocal = false
	}

	// --- extra ---
	p.Extra.ConfigManageNamespaceOfSecretsDefault = cms.MakeAppDefaultSecretNamespace(strconv.FormatUint(req.AppID, 10))
	ns, err := cms.MakeAppBranchPrefixSecretNamespace(strconv.FormatUint(req.AppID, 10), req.Branch)
	if err != nil {
		return nil, apierrors.ErrMakeConfigNamespace.InvalidParameter(err)
	}
	p.Extra.ConfigManageNamespaceOfSecrets = ns
	if req.UserID != "" {
		p.Extra.SubmitUser = s.user.TryGetUser(context.Background(), req.UserID)
	}
	p.Extra.IsAutoRun = req.AutoRun
	p.Extra.CallbackURLs = req.CallbackURLs

	// --- cron ---
	pipelineYml, err := pipelineyml.New([]byte(p.PipelineYml))
	if err != nil {
		return nil, apierrors.ErrParsePipelineYml.InternalError(err)
	}
	p.Extra.CronExpr = pipelineYml.Spec().Cron
	if err := s.UpdatePipelineCron(p, nil, nil, pipelineYml.Spec().CronCompensator); err != nil {
		return nil, apierrors.ErrCreatePipeline.InternalError(err)
	}

	version, err := pipelineyml.GetVersion([]byte(p.PipelineYml))
	if err != nil {
		return nil, apierrors.ErrParsePipelineYml.InvalidParameter("version")
	}
	p.Extra.Version = version

	p.CostTimeSec = -1
	p.Progress = -1

	return p, nil
}
