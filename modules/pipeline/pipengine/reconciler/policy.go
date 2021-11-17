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

package reconciler

import (
	"fmt"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/dbclient"
	"github.com/erda-project/erda/modules/pipeline/spec"
	"github.com/erda-project/erda/pkg/parser/pipelineyml"
)

type PolicyHandlerOptions struct {
	dbClient *dbclient.Client
}

type PolicyType interface {
	ResetTask(*spec.PipelineTask, PolicyHandlerOptions) (*spec.PipelineTask, error)
}

var policyTypeAdaptor = map[apistructs.PolicyType]PolicyType{
	apistructs.TryLatestSuccessResultPolicyType: TryLastSuccessResult{},
	apistructs.NewRunPolicyType:                 NewRun{},
}

type NewRun struct{}

func (run NewRun) ResetTask(task *spec.PipelineTask, options PolicyHandlerOptions) (*spec.PipelineTask, error) {
	return task, nil
}

type TryLastSuccessResult struct{}

func (t TryLastSuccessResult) ResetTask(task *spec.PipelineTask, opt PolicyHandlerOptions) (*spec.PipelineTask, error) {
	if !task.IsSnippet {
		return task, nil
	}

	if task.Extra.Action.SnippetConfig == nil {
		return task, nil
	}

	ymlName, source := getPipelineSourceAndNameBySnippetConfig(task.Extra.Action.SnippetConfig)

	runSuccessPipeline, _, _, _, err := opt.dbClient.PageListPipelines(apistructs.PipelinePageListRequest{
		Sources:        []apistructs.PipelineSource{source},
		YmlNames:       []string{ymlName},
		Statuses:       []string{apistructs.PipelineStatusSuccess.String()},
		PageNum:        1,
		PageSize:       1,
		IncludeSnippet: true,
		DescCols:       []string{apistructs.PipelinePageListRequestIdColumn},
	})
	if err != nil {
		return task, err
	}
	if len(runSuccessPipeline) <= 0 {
		return task, nil
	}
	pipeline := runSuccessPipeline[0]

	beforeSuccessTask, err := opt.dbClient.GetPipelineTask(pipeline.ParentTaskID)
	if err != nil {
		return task, err
	}

	if beforeSuccessTask.ID <= 0 {
		return task, nil
	}

	task.Status = beforeSuccessTask.Status
	task.Result = beforeSuccessTask.Result
	task.IsSnippet = beforeSuccessTask.IsSnippet
	task.SnippetPipelineDetail = beforeSuccessTask.SnippetPipelineDetail
	task.TimeBegin = beforeSuccessTask.TimeBegin
	task.TimeEnd = beforeSuccessTask.TimeEnd
	task.CostTimeSec = beforeSuccessTask.CostTimeSec
	task.QueueTimeSec = beforeSuccessTask.QueueTimeSec
	task.SnippetPipelineID = beforeSuccessTask.SnippetPipelineID
	if task.Extra.Action.Policy != nil {
		task.Extra.CurrentPolicy = apistructs.Policy{
			Type: task.Extra.Action.Policy.Type,
		}
	}
	return task, nil
}

func getPipelineSourceAndNameBySnippetConfig(snippetConfig *pipelineyml.SnippetConfig) (ymlName string, sources apistructs.PipelineSource) {
	return snippetConfig.Name, apistructs.PipelineSource(snippetConfig.Source)
}

func (r *Reconciler) adaptPolicy(task *spec.PipelineTask) (result *spec.PipelineTask, err error) {
	if task == nil {
		return task, fmt.Errorf("task was empty")
	}
	if task.Extra.Action.Policy == nil {
		return task, nil
	}

	handler := policyTypeAdaptor[task.Extra.Action.Policy.Type]
	if handler == nil {
		return task, nil
	}

	opt := PolicyHandlerOptions{
		dbClient: r.dbClient,
	}

	return handler.ResetTask(task, opt)
}
