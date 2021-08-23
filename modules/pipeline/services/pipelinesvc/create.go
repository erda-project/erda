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

package pipelinesvc

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/mohae/deepcopy"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/commonutil/thirdparty/gittarutil"
	"github.com/erda-project/erda/modules/pipeline/conf"
	"github.com/erda-project/erda/modules/pipeline/dbclient"
	"github.com/erda-project/erda/modules/pipeline/events"
	"github.com/erda-project/erda/modules/pipeline/providers/cms"
	"github.com/erda-project/erda/modules/pipeline/services/apierrors"
	"github.com/erda-project/erda/modules/pipeline/spec"
	"github.com/erda-project/erda/pkg/discover"
	"github.com/erda-project/erda/pkg/parser/pipelineyml"
	"github.com/erda-project/erda/pkg/strutil"
)

// Deprecated
func (s *PipelineSvc) Create(req *apistructs.PipelineCreateRequest) (*spec.Pipeline, error) {
	p, err := s.makePipelineFromRequest(req)
	if err != nil {
		return nil, err
	}
	if err := s.createPipelineGraph(p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *PipelineSvc) makePipelineFromRequest(req *apistructs.PipelineCreateRequest) (*spec.Pipeline, error) {
	p := &spec.Pipeline{
		PipelineExtra: spec.PipelineExtra{
			NormalLabels: make(map[string]string),
			Extra: spec.PipelineExtraInfo{
				// --- empty user ---
				SubmitUser: &apistructs.PipelineUser{},
				RunUser:    &apistructs.PipelineUser{},
				CancelUser: &apistructs.PipelineUser{},
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
	commit, err := repo.GetCommit(req.Branch)
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
	if !req.Source.Valid() {
		return nil, apierrors.ErrCreatePipeline.InvalidParameter(errors.Errorf("source: %s", req.Source))
	}
	p.PipelineSource = req.Source

	if req.PipelineYmlName == "" {
		req.PipelineYmlName = apistructs.DefaultPipelineYmlName
	}

	// PipelineYmlNameV1 用于从 gittar 中获取 pipeline.yml 内容
	p.Extra.PipelineYmlNameV1 = req.PipelineYmlName
	p.PipelineYmlName = p.GenerateV1UniquePipelineYmlName(p.Extra.PipelineYmlNameV1)

	if req.PipelineYmlSource == "" {
		return nil, apierrors.ErrCreatePipeline.MissingParameter("pipelineYmlSource")
	}
	if !req.PipelineYmlSource.Valid() {
		return nil, apierrors.ErrCreatePipeline.InvalidParameter(errors.Errorf("pipelineYmlSource: %s", req.PipelineYmlSource))
	}
	p.Extra.PipelineYmlSource = req.PipelineYmlSource
	switch req.PipelineYmlSource {
	case apistructs.PipelineYmlSourceGittar:
		// get yaml
		f, err := repo.FetchFile(req.Branch, p.Extra.PipelineYmlNameV1)
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
		p.Extra.SubmitUser = s.tryGetUser(req.UserID)
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

// createPipelineGraph recursively create pipeline graph.
// passedData stores data passed recursively.
func (s *PipelineSvc) createPipelineGraph(p *spec.Pipeline, passedDataOpt ...passedDataWhenCreate) (err error) {
	var passedData passedDataWhenCreate
	if len(passedDataOpt) > 0 {
		passedData = passedDataOpt[0]
	}
	passedData.initData(s.extMarketSvc)

	// tx
	txSession := s.dbClient.NewSession()
	defer txSession.Close()
	if err := txSession.Begin(); err != nil {
		return apierrors.ErrCreatePipelineGraph.InternalError(err)
	}
	defer func() {
		if err != nil {
			rbErr := txSession.Rollback()
			if rbErr != nil {
				logrus.Errorf("[alert] failed to rollback when createPipelineGraph failed, pipeline: %+v, rollbackErr: %v",
					p, rbErr)
			}
			return
		}
		// metrics.PipelineCounterTotalAdd(*p, 1)
	}()

	// 创建 pipeline
	if err := s.dbClient.CreatePipeline(p, dbclient.WithTxSession(txSession.Session)); err != nil {
		return apierrors.ErrCreatePipeline.InternalError(err)
	}

	// 创建 stage -> task
	pipelineYml, err := pipelineyml.New(
		[]byte(p.PipelineYml),
		//pipelineyml.WithRunParams(p.Snapshot.RunPipelineParams), // runParams 在执行时才渲染，提前渲染在嵌套流水线中会导致渲染为 outputs 占位符，会引起歧义
		//pipelineyml.WithRenderSnippet(p.Labels, caches),
	)
	if err != nil {
		return apierrors.ErrParsePipelineYml.InternalError(err)
	}

	// search and cache action define and spec
	if err := passedData.putPassedDataByPipelineYml(pipelineYml); err != nil {
		return err
	}

	lastSuccessTaskMap, _, err := s.dbClient.ParseRerunFailedDetail(p.Extra.RerunFailedDetail)
	if err != nil {
		return apierrors.ErrCreatePipelineGraph.InternalError(err)
	}

	var snippetTasks []*spec.PipelineTask
	var allStagedTasks [][]*spec.PipelineTask
	for si, stage := range pipelineYml.Spec().Stages {
		var stagedTasks []*spec.PipelineTask
		ps := &spec.PipelineStage{
			PipelineID:  p.ID,
			Name:        "",
			Status:      apistructs.PipelineStatusAnalyzed,
			CostTimeSec: -1,
			Extra:       spec.PipelineStageExtra{StageOrder: si},
		}
		if err := s.dbClient.CreatePipelineStage(ps, dbclient.WithTxSession(txSession.Session)); err != nil {
			return apierrors.ErrCreatePipelineGraph.InternalError(err)
		}

		// create tasks
		for _, typedAction := range stage.Actions {
			for actionType, action := range typedAction {
				var pt *spec.PipelineTask
				lastSuccessTask, ok := lastSuccessTaskMap[string(action.Alias)]
				if ok {
					pt = lastSuccessTask
					pt.ID = 0
					pt.PipelineID = p.ID
					pt.StageID = ps.ID
				} else {
					switch actionType {
					case apistructs.ActionTypeSnippet: // 生成嵌套流水线任务
						pt, err = s.makeSnippetPipelineTask(p, ps, action)
						if err != nil {
							return apierrors.ErrCreatePipelineTask.InternalError(err)
						}
						snippetTasks = append(snippetTasks, pt)
					default: // 生成普通任务
						pt, err = s.makeNormalPipelineTask(p, ps, action, passedData)
						if err != nil {
							return apierrors.ErrCreatePipelineTask.InternalError(err)
						}
					}
				}
				// 创建当前节点
				if err := s.dbClient.CreatePipelineTask(pt, dbclient.WithTxSession(txSession.Session)); err != nil {
					logrus.Errorf("[alert] failed to create pipeline task when create pipeline graph: %v", err)
					return apierrors.ErrCreatePipelineTask.InternalError(err)
				}
				stagedTasks = append(stagedTasks, pt)
			}
		}
		allStagedTasks = append(allStagedTasks, stagedTasks)
	}

	// commit transaction
	if err := txSession.Commit(); err != nil {
		logrus.Errorf("[alert] failed to commit when createPipelineGraph success, pipeline: %+v, commitErr: %v",
			p, err)
		return apierrors.ErrCreatePipelineGraph.InternalError(err)
	}

	_ = s.PreCheck(p)

	// put into db gc
	p.EnsureGC()
	//s.engine.WaitDBGC(p.ID, *p.Extra.GC.DatabaseGC.Analyzed.TTLSecond, *p.Extra.GC.DatabaseGC.Analyzed.NeedArchive)

	// events
	events.EmitPipelineInstanceEvent(p, p.GetSubmitUserID())

	// 统一处理嵌套流水线
	// 批量查询 snippet yaml
	var sourceSnippetConfigs []apistructs.SnippetConfig
	for _, snippetTask := range snippetTasks {
		yamlSnippetConfig := snippetTask.Extra.Action.SnippetConfig
		snippetConfig := apistructs.SnippetConfig{
			Source: yamlSnippetConfig.Source,
			Name:   yamlSnippetConfig.Name,
			Labels: yamlSnippetConfig.Labels,
		}
		sourceSnippetConfigs = append(sourceSnippetConfigs, snippetConfig)
	}
	sourceSnippetConfigYamls, err := s.HandleQueryPipelineYamlBySnippetConfigs(sourceSnippetConfigs)
	if err != nil {
		return apierrors.ErrQuerySnippetYaml.InternalError(err)
	}

	// 创建嵌套流水线
	var sErrs []error
	var sLock sync.Mutex
	var wg sync.WaitGroup
	for i := range snippetTasks {
		snippet := snippetTasks[i]
		snippetConfig := apistructs.SnippetConfig{
			Source: snippet.Extra.Action.SnippetConfig.Source,
			Name:   snippet.Extra.Action.SnippetConfig.Name,
			Labels: snippet.Extra.Action.SnippetConfig.Labels,
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			// snippetTask 转换为 pipeline 结构体
			snippetPipeline, err := s.makeSnippetPipeline4Create(p, snippet, sourceSnippetConfigYamls[snippetConfig.ToString()])
			if err != nil {
				sLock.Lock()
				sErrs = append(sErrs, err)
				sLock.Unlock()
				return
			}
			// 创建嵌套流水线
			if err := s.createPipelineGraph(snippetPipeline, passedData); err != nil {
				sLock.Lock()
				sErrs = append(sErrs, err)
				sLock.Unlock()
				return
			}
			// 创建好的流水线数据塞回 snippetTask
			snippet.SnippetPipelineID = &snippetPipeline.ID
			snippet.Extra.AppliedResources = snippetPipeline.Snapshot.AppliedResources
			if err := s.dbClient.UpdatePipelineTask(snippet.ID, snippet); err != nil {
				sLock.Lock()
				sErrs = append(sErrs, err)
				sLock.Unlock()
				return
			}
		}()
	}
	wg.Wait()
	if len(sErrs) > 0 {
		var errMsgs []string
		for _, err := range sErrs {
			errMsgs = append(errMsgs, err.Error())
		}
		return apierrors.ErrCreatePipelineGraph.InternalError(fmt.Errorf(strutil.Join(errMsgs, "; ")))
	}

	// calculate pipeline applied resource after all snippetTask created
	pipelineAppliedResources := s.calculatePipelineResources(allStagedTasks)
	p.Snapshot.AppliedResources = pipelineAppliedResources
	if err := s.dbClient.UpdatePipelineExtraSnapshot(p.ID, p.Snapshot); err != nil {
		return apierrors.ErrCreatePipelineGraph.InternalError(
			fmt.Errorf("failed to update pipeline snapshot for applied resources, err: %v", err))
	}

	return nil
}

func getString(v interface{}) string {
	if v == nil {
		return ""
	}
	return fmt.Sprintf("%v", v)
}

func (s *PipelineSvc) makePipelineFromCopy(o *spec.Pipeline) (p *spec.Pipeline, err error) {
	r := deepcopy.Copy(o)
	p, ok := r.(*spec.Pipeline)
	if !ok {
		return nil, errors.New("failed to copy pipeline")
	}

	now := time.Now()

	// 初始化一些字段
	p.ID = 0
	p.Status = apistructs.PipelineStatusAnalyzed
	p.PipelineExtra.PipelineID = 0
	p.Snapshot = spec.Snapshot{}
	p.Snapshot.Envs = o.Snapshot.Envs
	p.Snapshot.RunPipelineParams = o.Snapshot.RunPipelineParams
	p.Extra.Namespace = o.Extra.Namespace
	p.Extra.SubmitUser = &apistructs.PipelineUser{}
	p.Extra.RunUser = &apistructs.PipelineUser{}
	p.Extra.CancelUser = &apistructs.PipelineUser{}
	p.Extra.ShowMessage = nil
	p.Extra.CopyFromPipelineID = &o.ID
	p.Extra.RerunFailedDetail = nil
	p.Extra.CronTriggerTime = nil
	p.Extra.CompleteReconcilerGC = false
	p.TriggerMode = apistructs.PipelineTriggerModeManual // 手动触发
	p.TimeCreated = &now
	p.TimeUpdated = &now
	p.TimeBegin = nil
	p.TimeEnd = nil
	p.CostTimeSec = -1

	return p, nil
}
