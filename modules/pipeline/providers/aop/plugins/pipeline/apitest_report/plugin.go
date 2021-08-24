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

package apitest_report

import (
	"encoding/json"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/providers/aop/aoptypes"
	"github.com/erda-project/erda/modules/pipeline/providers/aop/plugins_manage"
	"github.com/erda-project/erda/modules/pipeline/spec"
)

const (
	actionTypeAPITest = "api-test"
	version2          = "2.0"
)

type Plugin struct {
	aoptypes.PipelineBaseTunePoint
}

func (p *Plugin) Name() string {
	return "apitest_report"
}

type ApiReportMeta struct {
	ApiTotalNum   int `json:"apiTotalNum"`
	ApiSuccessNum int `json:"apiSuccessNum"`
	ApiFailedNum  int `json:"apiFailedNum"`
	ApiNotExecNum int `json:"apiNotExecNum"`
}

func (p *Plugin) Handle(ctx *aoptypes.TuneContext) error {
	// source = autotest
	if ctx.SDK.Pipeline.PipelineSource != apistructs.PipelineSourceAutoTest {
		return nil
	}
	var allTasks []*spec.PipelineTask
	// 尝试从上下文中获取，减少不必要的网络、数据库请求
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
	var apiTestTasks []*spec.PipelineTask
	var snippetTaskPipelineIDs []uint64
	for _, task := range allTasks {
		if task.Type == actionTypeAPITest && task.Extra.Action.Version == version2 {
			apiTestTasks = append(apiTestTasks, task)
			continue
		}
		if task.Type == apistructs.ActionTypeSnippet {
			snippetTaskPipelineIDs = append(snippetTaskPipelineIDs, *task.SnippetPipelineID)
			continue
		}
	}

	apiTotalNum := 0
	apiSuccessNum := 0
	apiFailedNum := 0

	// snippetTask 从对应的 snippetPipeline api-test 报告里获取接口执行情况
	snippetReports, err := ctx.SDK.DBClient.BatchListPipelineReportsByPipelineID(
		snippetTaskPipelineIDs,
		[]string{string(apistructs.PipelineReportTypeAPITest)},
	)
	if err != nil {
		return err
	}

	for _, apiTestTask := range apiTestTasks {
		// 总数
		apiTotalNum++
		// 执行成功
		if apiTestTask.Status.IsSuccessStatus() {
			apiSuccessNum++
		}
		// 执行失败
		if apiTestTask.Status.IsFailedStatus() {
			apiFailedNum++
		}
	}
	for pipelineID, reports := range snippetReports {
		for _, report := range reports {
			b, err := json.Marshal(report.Meta)
			if err != nil {
				logrus.Warnf("failed to marshal api-test report, snippet pipelineID: %d, reportID: %d, err: %v",
					pipelineID, report.ID, err)
				continue
			}
			var meta ApiReportMeta
			if err := json.Unmarshal(b, &meta); err != nil {
				logrus.Warnf("failed to unmarshal api-test report to meta, snippet pipelineID: %d, reportID: %d, err: %v",
					pipelineID, report.ID, err)
				continue
			}
			// 总数
			apiTotalNum += meta.ApiTotalNum
			apiSuccessNum += meta.ApiSuccessNum
			apiFailedNum += meta.ApiFailedNum
		}
	}

	// 处理 notExecNum
	apiNotExecNum := apiTotalNum - apiSuccessNum - apiFailedNum

	// 构造 reportMeta
	reportMeta := ApiReportMeta{
		ApiTotalNum:   apiTotalNum,
		ApiSuccessNum: apiSuccessNum,
		ApiFailedNum:  apiFailedNum,
		ApiNotExecNum: apiNotExecNum,
	}
	var reqMeta map[string]interface{}
	b, _ := json.Marshal(reportMeta)
	_ = json.Unmarshal(b, &reqMeta)

	// result 信息
	_, err = ctx.SDK.Report.Create(apistructs.PipelineReportCreateRequest{
		PipelineID: ctx.SDK.Pipeline.ID,
		Type:       actionTypeAPITest,
		Meta:       reqMeta,
	})
	if err != nil {
		return err
	}
	return nil
}

func New() *Plugin {
	var p Plugin
	return &p
}

type config struct {
	TuneType    aoptypes.TuneType      `file:"tune_type"`
	TuneTrigger []aoptypes.TuneTrigger `file:"tune_trigger" `
}

// +provider
type provider struct {
	Cfg *config
}

func (p *provider) Init(ctx servicehub.Context) error {
	for _, tuneTrigger := range p.Cfg.TuneTrigger {
		err := plugins_manage.RegisterTunePointToTuneGroup(p.Cfg.TuneType, tuneTrigger, New())
		if err != nil {
			panic(err)
		}
	}
	return nil
}

func init() {
	servicehub.Register("erda.core.pipeline.aop.plugins.pipeline.apitest-report", &servicehub.Spec{
		ConfigFunc: func() interface{} {
			return &config{}
		},
		Creator: func() servicehub.Provider {
			return &provider{}
		},
	})
}
