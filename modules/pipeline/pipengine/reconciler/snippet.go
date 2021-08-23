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
	"regexp"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/commonutil/statusutil"
	"github.com/erda-project/erda/modules/pipeline/spec"
	"github.com/erda-project/erda/pkg/strutil"
)

// fulfillParentSnippetTask 填充 parent snippet task 信息
func (r *Reconciler) fulfillParentSnippetTask(p *spec.Pipeline) error {
	if !p.IsSnippet || p.ParentTaskID == nil {
		return nil
	}
	pWithTasks, err := r.dbClient.GetPipelineWithTasks(p.ID)
	if err != nil {
		return err
	}
	p = pWithTasks.Pipeline
	tasks := pWithTasks.Tasks
	// outputs
	outputValues, err := r.calculateAndUpdatePipelineOutputValues(p, tasks)
	if err != nil {
		return fmt.Errorf("failed to calculate pipeline outputs, pipelineID: %d, err: %v", p.ID, err)
	}
	// update parent task
	if err := r.handleParentSnippetTaskOutputs(p, outputValues); err != nil {
		return fmt.Errorf("failed to handler parent snippet task outputs, pipelineID: %d, err: %v", p.ID, err)
	}
	// 根据 tasks 拿到 snippet task 的状态
	calcStatus := statusutil.CalculatePipelineStatusV2(tasks)
	logrus.Infof("snippt pipeline %d calculated pipeline status: %s", p.ID, calcStatus)
	// 更新 snippet task 状态
	if err := r.dbClient.UpdatePipelineTaskStatus(*p.ParentTaskID, calcStatus); err != nil {
		return err
	}
	return nil
}

// handleParentSnippetTaskOutputs 处理 parentSnippetTask 的 outputs
// 1. task 的 snippetPipelineDetail.Outputs 用于记录
// 2. task 的 result.metadata 用于作为普通任务值引用
func (r *Reconciler) handleParentSnippetTaskOutputs(snippetPipeline *spec.Pipeline, outputValues []apistructs.PipelineOutputWithValue) error {
	parentTaskID := *snippetPipeline.ParentTaskID

	// update snippetPipelineDetail.Outputs, not overwrite
	parentTask, err := r.dbClient.GetPipelineTask(parentTaskID)
	if err != nil {
		return err
	}
	snippetDetail := parentTask.SnippetPipelineDetail
	if snippetDetail == nil {
		snippetDetail = &apistructs.PipelineTaskSnippetDetail{
			Outputs:                  nil,
			DirectSnippetTasksNum:    -1,
			RecursiveSnippetTasksNum: -1,
		}
	}
	snippetDetail.Outputs = outputValues
	if err := r.dbClient.UpdatePipelineTaskSnippetDetail(parentTaskID, *snippetDetail); err != nil {
		return fmt.Errorf("failed to update pipeline task snippet detail, err: %v", err)
	}

	// update result.metadata for value-context reference
	for _, outputValue := range snippetPipeline.Snapshot.OutputValues {
		parentTask.Result.Metadata = append(parentTask.Result.Metadata, apistructs.MetadataField{
			Name:  outputValue.Name,
			Value: strutil.String(outputValue.Value),
		})
	}
	if err := r.dbClient.UpdatePipelineTaskResult(parentTaskID, parentTask.Result); err != nil {
		return err
	}

	return nil
}

// calculatePipelineOutputs 计算 pipeline
func (r *Reconciler) calculateAndUpdatePipelineOutputValues(p *spec.Pipeline, tasks []*spec.PipelineTask) ([]apistructs.PipelineOutputWithValue, error) {
	// 所有任务的输出
	allTaskOutputs := make(map[string]map[string]interface{})
	for _, task := range tasks {
		for _, meta := range task.Result.Metadata {
			if allTaskOutputs[task.Name] == nil {
				allTaskOutputs[task.Name] = make(map[string]interface{})
			}
			allTaskOutputs[task.Name][meta.Name] = meta.Value
		}
	}

	// 根据定义塞入流水线级别的输出
	var outputValues []apistructs.PipelineOutputWithValue
	for _, define := range p.Extra.DefinedOutputs {
		// handle ref v1
		reffedTask, reffedKey, err := parsePipelineOutputRef(define.Ref)
		if err == nil {
			reffedValue := allTaskOutputs[reffedTask][reffedKey]
			outputWithValue := apistructs.PipelineOutputWithValue{PipelineOutput: define, Value: reffedValue}
			outputValues = append(outputValues, outputWithValue)
		}

		// handle ref v2
		reffedTask, reffedKey, err = parsePipelineOutputRefV2(define.Ref)
		if err == nil {
			reffedValue := allTaskOutputs[reffedTask][reffedKey]
			outputWithValue := apistructs.PipelineOutputWithValue{PipelineOutput: define, Value: reffedValue}
			outputValues = append(outputValues, outputWithValue)
		}
	}

	// update pipeline outputs
	p.Snapshot.OutputValues = outputValues
	if err := r.dbClient.UpdatePipelineExtraSnapshot(p.ID, p.Snapshot); err != nil {
		logrus.Errorf("failed to update pipeline outputValues, err: %v", err)
	}

	return outputValues, nil
}

// parsePipelineOutputRef 解析 pipeline 的 output ref 表达式
func parsePipelineOutputRef(ref string) (string, string, error) {
	ref = strings.TrimSpace(ref)
	ref = strings.TrimPrefix(ref, "${")
	ref = strings.TrimSuffix(ref, "}")
	ss := strings.SplitN(ref, ":", 3)
	if len(ss) < 3 {
		return "", "", fmt.Errorf("invalid ref: %s", ref)
	}
	if ss[1] != "OUTPUT" {
		return "", "", fmt.Errorf("ref is not output, ref: %s", ref)
	}
	return ss[0], ss[2], nil
}

// parsePipelineOutputRefV2 解析 pipeline 的 output ref 表达式
// ${{ outputs.xxx.key }}
func parsePipelineOutputRefV2(ref string) (string, string, error) {
	ref = strings.TrimSpace(ref)
	ref = strings.TrimPrefix(ref, "${{ outputs.")
	ref = strings.TrimSuffix(ref, " }}")
	ss := strings.SplitN(ref, ".", 2)
	if len(ss) < 2 {
		return "", "", fmt.Errorf("invalid ref: %s", ref)
	}
	return ss[0], ss[1], nil
}

// copyParentPipelineRunInfo 从父流水线拷贝执行信息
func (r *Reconciler) copyParentPipelineRunInfo(snippetPipeline *spec.Pipeline) error {
	// 从根流水线拷贝执行信息到嵌套流水线
	rootPipelineID := snippetPipeline.Extra.SnippetChain[0]
	rootPipeline, err := r.dbClient.GetPipeline(rootPipelineID)
	if err != nil {
		return err
	}
	snippetPipeline.Snapshot.PlatformSecrets = rootPipeline.Snapshot.PlatformSecrets
	snippetPipeline.Snapshot.Secrets = rootPipeline.Snapshot.Secrets
	snippetPipeline.Snapshot.Envs = rootPipeline.Snapshot.Envs

	// 处理 runParams，嵌套流水线的 runParams 即为 parentSnippetTask 的 params，已经在创建时存入，其中占位符需要被替换
	// 占位符包括：previousTaskOutputs, pipelineRunParams
	// 获取 parentPipeline 所有前置 task 的 outputs
	parentPipelinePreviousOutputs, err := r.dbClient.GetPipelineOutputs(*snippetPipeline.ParentPipelineID)
	if err != nil {
		return err
	}
	// 获取 parent pipeline runParams
	parentPipeline, err := r.dbClient.GetPipeline(*snippetPipeline.ParentPipelineID)
	if err != nil {
		return err
	}
	parentPipelineRunParams := parentPipeline.Snapshot.RunPipelineParams
	parentPipelineRunParamMap := make(map[string]interface{})
	for _, rp := range parentPipelineRunParams {
		value := rp.TrueValue
		if value == nil {
			value = rp.PipelineRunParam.Value
		}
		parentPipelineRunParamMap[rp.Name] = value
	}
	// 遍历处理 snippetPipeline 的 runParams
	for i := range snippetPipeline.Snapshot.RunPipelineParams {
		runParam := snippetPipeline.Snapshot.RunPipelineParams[i]
		if runParam.TrueValue != nil {
			continue
		}
		var reffedValue interface{}

		// outputs 替换
		findOutputRef := false
		ref := strutil.String(runParam.PipelineRunParam.Value)
		// ${alias:OUTPUT:key}
		ref = strutil.ReplaceAllStringSubmatchFunc(regexp.MustCompile(`\${([^:]+):OUTPUT:([^:]+)}`), ref, func(subs []string) string {
			alias := subs[1]
			key := subs[2]
			if metas, ok := parentPipelinePreviousOutputs[alias]; ok {
				if value, ok := metas[key]; ok {
					findOutputRef = true
					return value
				}
			}
			return subs[0]
		})
		// ${{ outputs.alias.key }}
		ref = strutil.ReplaceAllStringSubmatchFunc(regexp.MustCompile(`\${{[ ]{1}outputs.([^{}\s.]+).(.+)[ ]{1}}}`), ref, func(subs []string) string {
			alias := subs[1]
			key := subs[2]
			if metas, ok := parentPipelinePreviousOutputs[alias]; ok {
				if value, ok := metas[key]; ok {
					findOutputRef = true
					return value
				}
			}
			return subs[0]
		})
		if findOutputRef {
			reffedValue = ref
		}

		// params 整体替换
		ss := regexp.MustCompile(`^\${params.(.+)}$`).FindStringSubmatch(strutil.String(runParam.Value))
		if len(ss) == 2 {
			key := ss[1]
			reffedValue = parentPipelineRunParamMap[key]
		}
		ss = regexp.MustCompile(`^\${{ params.(.+) }}$`).FindStringSubmatch(strutil.String(runParam.Value))
		if len(ss) == 2 {
			key := ss[1]
			reffedValue = parentPipelineRunParamMap[key]
		}

		snippetPipeline.Snapshot.RunPipelineParams[i].TrueValue = reffedValue
	}
	if err := r.dbClient.UpdatePipelineExtraSnapshot(snippetPipeline.ID, snippetPipeline.Snapshot); err != nil {
		return err
	}
	return nil
}
