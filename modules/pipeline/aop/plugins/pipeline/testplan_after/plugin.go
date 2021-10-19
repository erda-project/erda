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

package testplan_after

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda-infra/base/servicehub"
	testplanpb "github.com/erda-project/erda-proto-go/core/dop/autotest/testplan/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/pipeline/aop"
	"github.com/erda-project/erda/modules/pipeline/aop/aoptypes"
	"github.com/erda-project/erda/modules/pipeline/spec"
)

// +provider
type provider struct {
	aoptypes.PipelineBaseTunePoint
	Bundle *bundle.Bundle
}

func (p *provider) Init(ctx servicehub.Context) error {
	p.Bundle = bundle.New(bundle.WithEventBox())
	err := aop.RegisterTunePoint(p)
	if err != nil {
		panic(err)
	}
	return nil
}

func init() {
	servicehub.Register(aop.NewProviderNameByPluginName(&provider{}), &servicehub.Spec{
		Creator: func() servicehub.Provider {
			return &provider{}
		},
	})
}

func (p *provider) Name() string { return "testplan-after" }

func (p *provider) Handle(ctx *aoptypes.TuneContext) error {
	// source = autotest
	if ctx.SDK.Pipeline.PipelineSource != apistructs.PipelineSourceAutoTest {
		return nil
	}

	var (
		sceneID          uint64
		sceneSetID       uint64
		testPlanID       uint64
		iterationID      uint64
		parentPipelineID uint64
		stepType         apistructs.StepAPIType
		err              error
		userID           = ctx.SDK.Pipeline.GetUserID()
	)

	// PipelineYmlName is autotest-plan-xxx
	pipelineNamePre := apistructs.PipelineSourceAutoTestPlan.String() + "-"
	if strings.HasPrefix(ctx.SDK.Pipeline.PipelineYmlName, pipelineNamePre) {
		stepType = apistructs.AutoTestPlan
		testPlanIDStr := strings.TrimPrefix(ctx.SDK.Pipeline.PipelineYmlName, pipelineNamePre)
		testPlanID, err = strconv.ParseUint(testPlanIDStr, 10, 64)
		if err != nil {
			return err
		}
	}
	labels := ctx.SDK.Pipeline.MergeLabels()
	switch labels[apistructs.LabelAutotestExecType] {
	case apistructs.SceneSetsAutotestExecType:
		stepType = apistructs.AutotestSceneSet
	case apistructs.SceneAutotestExecType:
		stepType = apistructs.StepTypeScene
	default:
		if stepType == "" {
			return nil
		}
	}
	if labels[apistructs.LabelSceneID] != "" {
		sceneID, err = strconv.ParseUint(labels[apistructs.LabelSceneID], 10, 64)
		if err != nil {
			return err
		}
	}
	if labels[apistructs.LabelSceneSetID] != "" {
		sceneSetID, err = strconv.ParseUint(labels[apistructs.LabelSceneSetID], 10, 64)
		if err != nil {
			return err
		}
	}
	if labels[apistructs.LabelIterationID] != "" {
		iterationID, err = strconv.ParseUint(labels[apistructs.LabelIterationID], 10, 64)
		if err != nil {
			return err
		}
	}
	if testPlanID == 0 && labels[apistructs.LabelTestPlanID] != "" {
		testPlanID, err = strconv.ParseUint(labels[apistructs.LabelTestPlanID], 10, 64)
		if err != nil {
			return err
		}
	}
	if ctx.SDK.Pipeline.ParentPipelineID != nil {
		parentPipelineID = *ctx.SDK.Pipeline.ParentPipelineID
	}

	statics, err := statistics(ctx, ctx.SDK.Pipeline.ID)
	if err != nil {
		return err
	}
	apiExecNum := statics.ApiExecNum
	apiSuccessNum := statics.ApiSuccessNum

	var req = testplanpb.Content{
		TestPlanID:      testPlanID,
		ExecuteTime:     ctx.SDK.Pipeline.TimeBegin.Format("2006-01-02 15:04:05"),
		PassRate:        statics.PassRate,
		ApiTotalNum:     int64(apiExecNum),
		ExecuteDuration: time.Unix(ctx.SDK.Pipeline.CostTimeSec, 0).In(time.UTC).Format("15:04:05"),
		ApiSuccessNum:   int64(apiSuccessNum),
		ApiExecNum:      int64(apiExecNum),
		PipelineYml:     ctx.SDK.Pipeline.PipelineYml,
		StepAPIType:     stepType.String(),
		Status:          ctx.SDK.Pipeline.Status.String(),
		SceneID:         sceneID,
		SceneSetID:      sceneSetID,
		ParentID:        parentPipelineID,
		CreatorID:       userID,
		IterationID:     iterationID,
		StepID:          0,
	}
	if err = p.sendMessage(req, ctx); err != nil {
		return err
	}
	if stepType == apistructs.StepTypeScene {
		if err = p.sendStepMessage(ctx, testPlanID, sceneID, sceneSetID, iterationID, parentPipelineID, userID); err != nil {
			return err
		}
	}

	return nil
}

func (p *provider) sendStepMessage(ctx *aoptypes.TuneContext, testPlanID, sceneID, sceneSetID, iterationID, parentPipelineID uint64, userID string) error {
	result, err := ctx.SDK.DBClient.GetPipelineWithTasks(ctx.SDK.Pipeline.PipelineID)
	if err != nil {
		return err
	}
	allTasks := result.Tasks
	for _, task := range allTasks {
		if task.Type == apistructs.ActionTypeAPITest &&
			task.Extra.Action.Version == "2.0" {
			var (
				passRate      float64
				apiSuccessNum int64
			)
			if task.Status.IsSuccessStatus() {
				passRate = 100.00
				apiSuccessNum = 1
			}
			stepID, _ := strconv.ParseUint(task.Name, 10, 64)

			err = p.sendMessage(testplanpb.Content{
				TestPlanID:      testPlanID,
				ExecuteTime:     task.TimeBegin.Format("2006-01-02 15:04:05"),
				PassRate:        passRate,
				ApiTotalNum:     1,
				ExecuteDuration: time.Unix(task.CostTimeSec, 0).In(time.UTC).Format("15:04:05"),
				ApiSuccessNum:   apiSuccessNum,
				ApiExecNum:      1,
				PipelineYml:     "",
				StepAPIType:     apistructs.StepTypeAPI.String(),
				Status:          task.Status.String(),
				SceneID:         sceneID,
				SceneSetID:      sceneSetID,
				ParentID:        parentPipelineID,
				CreatorID:       userID,
				StepID:          stepID,
				IterationID:     iterationID,
			}, ctx)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

type ApiNumStatistics struct {
	ApiExecNum    int
	ApiSuccessNum int
	PassRate      float64
}

func statistics(ctx *aoptypes.TuneContext, pipelineID uint64) (*ApiNumStatistics, error) {
	var allTasks []*spec.PipelineTask
	// get tasks from ctx
	tasks, ok := ctx.TryGet(aoptypes.CtxKeyTasks)
	if ok {
		if _tasks, ok := tasks.([]*spec.PipelineTask); ok {
			allTasks = _tasks
		}
	} else {
		result, err := ctx.SDK.DBClient.GetPipelineWithTasks(pipelineID)
		if err != nil {
			return nil, err
		}
		allTasks = result.Tasks
	}
	// 过滤出 api_test task 以及 snippetTask
	apiTestTasks, snippetTaskPipelineIDs := filterPipelineTask(allTasks)

	apiExecNum := 0
	apiSuccessNum := 0
	// snippetTask get execute detail from snippetPipeline api-test
	snippetReports, err := ctx.SDK.DBClient.BatchListPipelineReportsByPipelineID(
		snippetTaskPipelineIDs,
		[]string{string(apistructs.PipelineReportTypeAPITest)},
	)
	if err != nil {
		return nil, err
	}

	for _, apiTestTask := range apiTestTasks {
		apiExecNum++
		if apiTestTask.Status.IsSuccessStatus() {
			apiSuccessNum++
		}
	}
	for id, reports := range snippetReports {
		for _, report := range reports {
			meta, err := convertReport(id, report)
			if err != nil {
				continue
			}
			apiExecNum += meta.ApiTotalNum
			apiSuccessNum += meta.ApiSuccessNum
		}
	}

	var passRate float64
	if apiExecNum == 0 {
		passRate = 0
	} else {
		passRate = float64(apiSuccessNum) / float64(apiExecNum) * 100
	}
	return &ApiNumStatistics{
		ApiExecNum:    apiExecNum,
		ApiSuccessNum: apiSuccessNum,
		PassRate:      passRate,
	}, nil

}

func filterPipelineTask(allTasks []*spec.PipelineTask) ([]*spec.PipelineTask, []uint64) {
	var apiTestTasks []*spec.PipelineTask
	var snippetTaskPipelineIDs []uint64
	for _, task := range allTasks {
		if task.Type == apistructs.ActionTypeAPITest && task.Extra.Action.Version == "2.0" {
			apiTestTasks = append(apiTestTasks, task)
			continue
		}
		if task.Type == apistructs.ActionTypeSnippet {
			if task.SnippetPipelineID != nil {
				snippetTaskPipelineIDs = append(snippetTaskPipelineIDs, *task.SnippetPipelineID)
			}
			continue
		}
	}
	return apiTestTasks, snippetTaskPipelineIDs
}

type ApiReportMeta struct {
	ApiTotalNum   int `json:"apiTotalNum"`
	ApiSuccessNum int `json:"apiSuccessNum"`
}

func convertReport(pipelineID uint64, report spec.PipelineReport) (ApiReportMeta, error) {
	b, err := json.Marshal(report.Meta)
	if err != nil {
		logrus.Warnf("failed to marshal api-test report, snippet pipelineID: %d, reportID: %d, err: %v",
			pipelineID, report.ID, err)
		return ApiReportMeta{}, err
	}
	var meta ApiReportMeta
	if err := json.Unmarshal(b, &meta); err != nil {
		logrus.Warnf("failed to unmarshal api-test report to meta, snippet pipelineID: %d, reportID: %d, err: %v",
			pipelineID, report.ID, err)
		return ApiReportMeta{}, err
	}
	return meta, nil
}

func (p *provider) sendMessage(req testplanpb.Content, ctx *aoptypes.TuneContext) error {
	ev2 := &apistructs.EventCreateRequest{
		EventHeader: apistructs.EventHeader{
			Event:         bundle.AutoTestPlanExecuteEvent,
			Action:        bundle.UpdateAction,
			OrgID:         ctx.SDK.Pipeline.Labels[apistructs.LabelOrgID],
			ProjectID:     ctx.SDK.Pipeline.Labels[apistructs.LabelProjectID],
			ApplicationID: "-1",
			TimeStamp:     time.Now().Format("2006-01-02 15:04:05"),
		},
		Sender:  bundle.SenderDOP,
		Content: req,
	}
	// create event
	if err := p.Bundle.CreateEvent(ev2); err != nil {
		logrus.Warnf("failed to send autoTestPlan update event, (%v)", err)
		return err
	}
	return nil
}
