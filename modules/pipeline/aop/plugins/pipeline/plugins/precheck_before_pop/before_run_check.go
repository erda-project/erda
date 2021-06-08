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

package precheck_before_pop

import (
	"fmt"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/pipeline/dbclient"
	"github.com/erda-project/erda/pkg/parser/pipelineyml"
	"github.com/erda-project/erda/pkg/pipeline_network_hook_client"
)

const CheckResultSuccess = "success"
const CheckResultFailed = "failed"
const HookType = "before-run-check"

type HttpBeforeCheckRun struct {
	PipelineID uint64
	Bdl        *bundle.Bundle
	DBClient   *dbclient.Client
}

type RetryOption struct {
	IntervalSecond      uint64 `json:"intervalSecond"`
	IntervalMillisecond uint64 `json:"intervalMillisecond"`
}

type CheckRunResult struct {
	CheckResult string      `json:"checkResult"`
	RetryOption RetryOption `json:"retryOption"`
	Message     string      `json:"message"`
}

type CheckRunResultRequest struct {
	Hook                string                 `json:"hook"`
	Labels              map[string]interface{} `json:"labels"`
	Source              string                 `json:"source"`
	PipelineYamlName    string                 `json:"pipelineYamlName"`
	PipelineYamlContent string                 `json:"pipelineYamlContent"`
}

type CheckRunResultResponse struct {
	apistructs.Header
	CheckRunResult CheckRunResult `json:"data"`
}

type HookBeforeCheckRun interface {
	CheckRun() (*CheckRunResult, error)
	GetPipelineWithTasks()
}

// filter out the corresponding type of hook
func (beforeCheckRun *HttpBeforeCheckRun) matchHookType(lifecycle []*pipelineyml.NetworkHookInfo) []*pipelineyml.NetworkHookInfo {
	var result []*pipelineyml.NetworkHookInfo
	for _, v := range lifecycle {
		if v.Hook == HookType {
			result = append(result, v)
		}
	}
	return result
}

func (beforeCheckRun HttpBeforeCheckRun) CheckRun() (result *CheckRunResult, err error) {
	if beforeCheckRun.PipelineID <= 0 {
		return nil, fmt.Errorf("pipelineID can not empty")
	}

	pipelineWithTasks, err := beforeCheckRun.DBClient.GetPipelineWithTasks(beforeCheckRun.PipelineID)
	if err != nil {
		return nil, err
	}

	yml, err := pipelineyml.New(
		[]byte(pipelineWithTasks.Pipeline.PipelineYml),
	)
	if err != nil {
		return nil, err
	}

	// if the network hook is not specified
	//the network hook will be passed by success
	matchInfo := beforeCheckRun.matchHookType(yml.Spec().Lifecycle)
	if matchInfo == nil {
		return &CheckRunResult{
			CheckResult: CheckResultSuccess,
		}, nil
	}
	pipeline := pipelineWithTasks.Pipeline

	for _, info := range matchInfo {
		var checkRunResultRequest CheckRunResultRequest

		checkRunResultRequest.Hook = info.Hook
		checkRunResultRequest.PipelineYamlContent = pipeline.PipelineYml
		checkRunResultRequest.Source = pipeline.PipelineSource.String()
		checkRunResultRequest.PipelineYamlName = pipeline.PipelineYmlName

		checkRunResultRequest.Labels = map[string]interface{}{}
		checkRunResultRequest.Labels["pipelineID"] = pipeline.ID
		checkRunResultRequest.Labels["pipelineLabels"] = info.Labels

		var response CheckRunResultResponse
		err := pipeline_network_hook_client.PostLifecycleHookHttpClient(info.Client, checkRunResultRequest, &response)
		if err != nil {
			return nil, err
		}

		if !response.Success {
			return nil, fmt.Errorf("response is empty or response not success")
		}

		// return directly if there is an failed
		if response.CheckRunResult.CheckResult == CheckResultFailed {
			return &response.CheckRunResult, nil
		}
	}

	return &CheckRunResult{
		CheckResult: CheckResultSuccess,
	}, nil
}
