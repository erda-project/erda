// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package pipelinesvc

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/mohae/deepcopy"
	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/commonutil/thirdparty/gittarutil"
	"github.com/erda-project/erda/modules/pipeline/conf"
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

type Graph struct {
	sLock              sync.Mutex
	s                  *PipelineSvc
	pipelines          []*GraphPipeline
	actionDiceYmlCache *passedDataWhenCreate
}

type GraphPipeline struct {
	p      *spec.Pipeline
	stages []*GraphStage
}

type GraphStage struct {
	stage *spec.PipelineStage
	tasks []*GraphTask
}

type GraphTask struct {
	task           *spec.PipelineTask
	pipeline       *spec.Pipeline
	parentPipeline *spec.Pipeline
}

func (graph *Graph) Create(p *spec.Pipeline) (err error) {

	if err := graph.parsePipelineAndTask(p); err != nil {
		return err
	}

	err = graph.s.dbClient.DB.Transaction(func(tx *gorm.DB) error {

		err = graph.batchCreatePipelinesAndTasks(tx)
		if err != nil {
			return fmt.Errorf("batchCreatePipelinesAndTasks error: %v", err)
		}

		err = graph.batchUpdatePipelines(tx)
		if err != nil {
			return fmt.Errorf("batchUpdatePipelines error: %v", err)
		}
		return nil
	})
	if err != nil {
		return err
	}

	//graph.batchGc()

	graph.batchSendEvent()

	return nil
}

func (graph *Graph) batchSendEvent() {
	for _, graphPipeline := range graph.pipelines {
		go func(p spec.Pipeline) {
			events.EmitPipelineInstanceEvent(&p, p.GetSubmitUserID())
		}(*graphPipeline.p)
	}
}

func (graph *Graph) batchGc() {
	//for index := range graph.pipelines {
	//	go func(index int) {
	//		p := graph.pipelines[index].p
	//		//_ = graph.s.PreCheck(p)
	//
	//		// put into db gc
	//		//p.EnsureGC()
	//		//graph.s.engine.WaitDBGC(p.ID, *p.Extra.GC.DatabaseGC.Analyzed.TTLSecond, *p.Extra.GC.DatabaseGC.Analyzed.NeedArchive)
	//	}(index)
	//}
}

// update some snippet info after insert
// pipelineBase or pipelineExtra need snippet pipeline ID or need parent pipeline ID
func (graph *Graph) batchUpdatePipelines(tx *gorm.DB) (err error) {

	var bases []*spec.PipelineBase
	for _, graphPipeline := range graph.pipelines {
		for _, stage := range graphPipeline.stages {
			for _, graphTask := range stage.tasks {

				if !graphTask.task.IsSnippet || graphTask.pipeline == nil {
					continue
				}

				graphTask.pipeline.ParentTaskID = &graphTask.task.ID
				graphTask.pipeline.ParentPipelineID = &graphTask.parentPipeline.ID
				bases = append(bases, &graphTask.pipeline.PipelineBase)
			}
		}
	}

	if err := graph.s.dbClient.BatchUpdatePipelineBaseParentID(bases, tx); err != nil {
		return err
	}

	var extras []*spec.PipelineExtra
	for _, graphPipeline := range graph.pipelines {
		for _, stage := range graphPipeline.stages {
			for _, graphTask := range stage.tasks {
				if !graphTask.task.IsSnippet || graphTask.pipeline == nil {
					continue
				}
				graphTask.pipeline.Extra.SnippetChain = append(graphTask.pipeline.Extra.SnippetChain, graphTask.parentPipeline.ID)
				extras = append(extras, &graphTask.pipeline.PipelineExtra)
			}
		}
	}

	if err := graph.s.dbClient.BatchUpdatePipelineExtra(extras, tx); err != nil {
		return err
	}

	return nil
}

func (graph *Graph) batchCreatePipelinesAndTasks(tx *gorm.DB) (err error) {

	// batchCreate all pipeline
	var pipelines []*spec.Pipeline
	for _, v := range graph.pipelines {
		pipelines = append(pipelines, v.p)
	}
	if err := graph.s.dbClient.BatchCreatePipelines(pipelines, tx); err != nil {
		return err
	}

	// all stage and task add pipelineID and namespace
	for _, graphPipeline := range graph.pipelines {
		pipelineID := graphPipeline.p.PipelineID
		for _, stage := range graphPipeline.stages {
			stage.stage.PipelineID = pipelineID
			for _, graphTask := range stage.tasks {
				graphTask.task.PipelineID = pipelineID
				graphTask.task.Extra.Namespace = graphPipeline.p.Extra.Namespace
				if graphTask.task.IsSnippet && graphTask.pipeline != nil {
					graphTask.task.SnippetPipelineID = &graphTask.pipeline.PipelineID
				}
			}
		}
	}

	// batchCreate all stages
	var stages []*spec.PipelineStage
	for _, graphPipeline := range graph.pipelines {
		for _, graphStage := range graphPipeline.stages {
			stages = append(stages, graphStage.stage)
		}
	}
	if err := graph.s.dbClient.BatchCreatePipelineStages(stages, tx); err != nil {
		return err
	}

	// all task add stageID
	for _, graphPipeline := range graph.pipelines {
		for _, stage := range graphPipeline.stages {
			for _, graphTask := range stage.tasks {
				graphTask.task.StageID = stage.stage.ID
			}
		}
	}

	// batchCreate all task
	var tasks []*spec.PipelineTask
	for _, graphPipeline := range graph.pipelines {
		for _, stage := range graphPipeline.stages {
			for _, graphTask := range stage.tasks {
				tasks = append(tasks, graphTask.task)
			}
		}
	}
	if err := graph.s.dbClient.BatchCreatePipelineTasks(tasks, tx); err != nil {
		return err
	}

	return nil
}

func (graph *Graph) initPipelineAndSnippet(p *spec.Pipeline, pipelineYml *pipelineyml.PipelineYml) ([]*GraphTask, *GraphPipeline, error) {
	// init graphPipeline
	var graphPipeline = GraphPipeline{p: p}
	var snippetGraphTasks []*GraphTask
	lastSuccessTaskMap, _, err := graph.s.dbClient.ParseRerunFailedDetail(p.Extra.RerunFailedDetail)
	if err != nil {
		return nil, nil, apierrors.ErrCreatePipelineGraph.InternalError(err)
	}
	for si, stage := range pipelineYml.Spec().Stages {

		// init stage
		var graphStage GraphStage
		ps := &spec.PipelineStage{
			Name:        "",
			Status:      apistructs.PipelineStatusAnalyzed,
			CostTimeSec: -1,
			Extra:       spec.PipelineStageExtra{StageOrder: si},
		}
		graphStage.stage = ps

		// init tasks
		var graphTasks []*GraphTask
		for _, typedAction := range stage.Actions {
			for actionType, action := range typedAction {
				var pt *spec.PipelineTask
				var task GraphTask
				lastSuccessTask, ok := lastSuccessTaskMap[string(action.Alias)]
				if ok {
					pt = lastSuccessTask
					pt.ID = 0
				} else {
					switch actionType {
					case apistructs.ActionTypeSnippet:
						pt, err = graph.s.makeSnippetPipelineTask(p, ps, action)
						if err != nil {
							return nil, nil, apierrors.ErrCreatePipelineTask.InternalError(err)
						}
					default:
						pt, err = graph.s.makeNormalPipelineTask(p, ps, action, graph.actionDiceYmlCache)
						if err != nil {
							return nil, nil, apierrors.ErrCreatePipelineTask.InternalError(err)
						}
					}
				}
				task.task = pt
				if actionType == apistructs.ActionTypeSnippet {
					snippetGraphTasks = append(snippetGraphTasks, &task)
				}
				graphTasks = append(graphTasks, &task)
			}
		}

		graphStage.tasks = graphTasks
		graphPipeline.stages = append(graphPipeline.stages, &graphStage)
	}
	graph.sLock.Lock()
	graph.pipelines = append(graph.pipelines, &graphPipeline)
	graph.sLock.Unlock()

	return snippetGraphTasks, &graphPipeline, nil
}

func (graph *Graph) searchSnippetPipelineYml(snippetGraphTasks []*GraphTask) (map[string]string, error) {
	// search snippetConfig pipeline yml
	var sourceSnippetConfigs []apistructs.SnippetConfig
	for _, snippetTask := range snippetGraphTasks {
		yamlSnippetConfig := snippetTask.task.Extra.Action.SnippetConfig
		snippetConfig := apistructs.SnippetConfig{
			Source: yamlSnippetConfig.Source,
			Name:   yamlSnippetConfig.Name,
			Labels: yamlSnippetConfig.Labels,
		}
		sourceSnippetConfigs = append(sourceSnippetConfigs, snippetConfig)
	}
	sourceSnippetConfigYmls, err := graph.s.HandleQueryPipelineYamlBySnippetConfigs(sourceSnippetConfigs)
	if err != nil {
		return nil, apierrors.ErrQuerySnippetYaml.InternalError(err)
	}
	return sourceSnippetConfigYmls, nil
}

func (graph *Graph) parsePipelineAndTask(p *spec.Pipeline) (err error) {

	var (
		snippetGraphTasks        []*GraphTask
		graphPipeline            *GraphPipeline
		sourceSnippetConfigYamls map[string]string
	)

	pipelineYml, err := pipelineyml.New(
		[]byte(p.PipelineYml),
	)
	if err != nil {
		return apierrors.ErrParsePipelineYml.InternalError(err)
	}

	err = graph.actionDiceYmlCache.putPassedDataByPipelineYml(pipelineYml)
	if err != nil {
		return err
	}

	snippetGraphTasks, graphPipeline, err = graph.initPipelineAndSnippet(p, pipelineYml)
	if err != nil {
		return err
	}

	sourceSnippetConfigYamls, err = graph.searchSnippetPipelineYml(snippetGraphTasks)
	if err != nil {
		return apierrors.ErrQuerySnippetYaml.InternalError(err)
	}

	// create snippet pipeline
	var sErrs []error
	var sLock sync.Mutex
	var wg sync.WaitGroup
	for i := range snippetGraphTasks {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			snippet := snippetGraphTasks[index].task
			snippetConfig := apistructs.SnippetConfig{
				Source: snippet.Extra.Action.SnippetConfig.Source,
				Name:   snippet.Extra.Action.SnippetConfig.Name,
				Labels: snippet.Extra.Action.SnippetConfig.Labels,
			}
			snippetPipeline, err := graph.s.makeSnippetPipeline4Create(p, snippet, sourceSnippetConfigYamls[snippetConfig.ToString()])
			if err != nil {
				sLock.Lock()
				sErrs = append(sErrs, err)
				sLock.Unlock()
				return
			}
			if err := graph.parsePipelineAndTask(snippetPipeline); err != nil {
				sLock.Lock()
				sErrs = append(sErrs, err)
				sLock.Unlock()
				return
			}
			snippetGraphTasks[index].pipeline = snippetPipeline
			snippetGraphTasks[index].parentPipeline = p
			snippet.Extra.AppliedResources = snippetPipeline.Snapshot.AppliedResources
		}(i)
	}
	wg.Wait()
	if len(sErrs) > 0 {
		var errMsgs []string
		for _, err := range sErrs {
			errMsgs = append(errMsgs, err.Error())
		}
		return apierrors.ErrCreatePipelineGraph.InternalError(fmt.Errorf(strutil.Join(errMsgs, "; ")))
	}

	// calculation task resources
	var allStagedTasks [][]*spec.PipelineTask
	for _, stage := range graphPipeline.stages {
		var tasks []*spec.PipelineTask
		for _, graphTask := range stage.tasks {
			tasks = append(tasks, graphTask.task)
		}
		allStagedTasks = append(allStagedTasks, tasks)
	}
	pipelineAppliedResources := graph.s.calculatePipelineResources(allStagedTasks)
	p.Snapshot.AppliedResources = pipelineAppliedResources

	return nil
}

// createPipelineGraph recursively create pipeline graph.
// passedData stores data passed recursively.
func (s *PipelineSvc) createPipelineGraph(p *spec.Pipeline, passedDataOpt ...passedDataWhenCreate) (err error) {
	var graph Graph
	graph.s = s
	graph.actionDiceYmlCache = &passedDataWhenCreate{}
	graph.actionDiceYmlCache.initData(s.extMarketSvc)
	return graph.Create(p)
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
