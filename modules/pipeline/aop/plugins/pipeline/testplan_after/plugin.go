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

type ApiReportMeta struct {
	ApiTotalNum   int `json:"apiTotalNum"`
	ApiSuccessNum int `json:"apiSuccessNum"`
}

// +provider
type provider struct {
	aoptypes.PipelineBaseTunePoint
	Bundle *bundle.Bundle
}

func (p *provider) Name() string { return "testplan-after" }

func (p *provider) Handle(ctx *aoptypes.TuneContext) error {
	// source = autotest
	if ctx.SDK.Pipeline.PipelineSource != apistructs.PipelineSourceAutoTest || ctx.SDK.Pipeline.IsSnippet {
		return nil
	}

	// PipelineYmlName is autotest-plan-xxx
	pipelineNamePre := apistructs.PipelineSourceAutoTestPlan.String() + "-"
	if !strings.HasPrefix(ctx.SDK.Pipeline.PipelineYmlName, pipelineNamePre) {
		return nil
	}
	testPlanIDStr := strings.TrimPrefix(ctx.SDK.Pipeline.PipelineYmlName, pipelineNamePre)
	testPlanID, err := strconv.ParseUint(testPlanIDStr, 10, 64)
	if err != nil {
		return err
	}

	var allTasks []*spec.PipelineTask
	// get tasks from ctx
	tasks, ok := ctx.TryGet(aoptypes.CtxKeyTasks)
	if ok {
		if _tasks, ok := tasks.([]*spec.PipelineTask); ok {
			allTasks = _tasks
		}
	} else {
		result, err := ctx.SDK.DBClient.GetPipelineWithTasks(ctx.SDK.Pipeline.ID)
		if err != nil {
			return err
		}
		allTasks = result.Tasks
	}
	// 过滤出 api_test task 以及 snippetTask
	apiTestTasks, snippetTaskPipelineIDs := filterPipelineTask(allTasks)

	apiTotalNum := 0
	apiSuccessNum := 0
	// snippetTask get execute detail from snippetPipeline api-test
	snippetReports, err := ctx.SDK.DBClient.BatchListPipelineReportsByPipelineID(
		snippetTaskPipelineIDs,
		[]string{string(apistructs.PipelineReportTypeAPITest)},
	)
	if err != nil {
		return err
	}

	for _, apiTestTask := range apiTestTasks {
		apiTotalNum++
		if apiTestTask.Status.IsSuccessStatus() {
			apiSuccessNum++
		}
	}
	for pipelineID, reports := range snippetReports {
		for _, report := range reports {
			meta, err := convertReport(pipelineID, report)
			if err != nil {
				continue
			}
			apiTotalNum += meta.ApiTotalNum
			apiSuccessNum += meta.ApiSuccessNum
		}
	}

	var req = testplanpb.Content{
		TestPlanID:     testPlanID,
		ExecuteTime:    ctx.SDK.Pipeline.TimeBegin.Format("2006-01-02 15:04:05"),
		ApiTotalNum:    int64(apiTotalNum),
		ExecuteMinutes: ctx.SDK.Pipeline.TimeEnd.Sub(*ctx.SDK.Pipeline.TimeBegin).Minutes(),
	}

	if apiTotalNum == 0 {
		req.PassRate = 0
	} else {
		req.PassRate = float64(apiSuccessNum) / float64(apiTotalNum) * 100
	}

	ev2 := &apistructs.EventCreateRequest{
		EventHeader: apistructs.EventHeader{
			Event:         bundle.AutoTestPlanExecuteEvent,
			Action:        bundle.UpdateAction,
			OrgID:         "-1",
			ProjectID:     "-1",
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

func filterPipelineTask(allTasks []*spec.PipelineTask) ([]*spec.PipelineTask, []uint64) {
	var apiTestTasks []*spec.PipelineTask
	var snippetTaskPipelineIDs []uint64
	for _, task := range allTasks {
		if task.Type == apistructs.ActionTypeAPITest && task.Extra.Action.Version == "2.0" {
			apiTestTasks = append(apiTestTasks, task)
			continue
		}
		if task.Type == apistructs.ActionTypeSnippet {
			snippetTaskPipelineIDs = append(snippetTaskPipelineIDs, *task.SnippetPipelineID)
			continue
		}
	}
	return apiTestTasks, snippetTaskPipelineIDs
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
