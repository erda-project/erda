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

package pipelineyml

import (
	"google.golang.org/protobuf/types/known/structpb"
	"gopkg.in/yaml.v3"

	"github.com/erda-project/erda-proto-go/core/pipeline/base/pb"
	"github.com/erda-project/erda/apistructs"
)

// ConvertGraphPipelineYmlContent: YAML(apistructs.PipelineYml) -> YAML(Spec)
func ConvertGraphPipelineYmlContent(data []byte) ([]byte, error) {
	var frontendYmlSpec apistructs.PipelineYml
	if err := yaml.Unmarshal(data, &frontendYmlSpec); err != nil {
		return nil, err
	}
	s := &Spec{}
	s.Version = frontendYmlSpec.Version
	s.Envs = frontendYmlSpec.Envs
	s.Cron = frontendYmlSpec.Cron
	if frontendYmlSpec.CronCompensator != nil {
		s.CronCompensator = &CronCompensator{
			Enable:               frontendYmlSpec.CronCompensator.Enable,
			LatestFirst:          frontendYmlSpec.CronCompensator.LatestFirst,
			StopIfLatterExecuted: frontendYmlSpec.CronCompensator.StopIfLatterExecuted,
		}
	}
	s.Stages = make([]*Stage, 0)
	for _, stage := range frontendYmlSpec.Stages {
		actions := make([]typedActionMap, 0)
		for _, frontendAction := range stage {

			maps := typedActionMap{
				ActionType(frontendAction.Type): &Action{
					Alias:       ActionAlias(frontendAction.Alias),
					Description: frontendAction.Description,
					Version:     frontendAction.Version,
					Params:      frontendAction.Params,
					Image:       frontendAction.Image,
					Commands:    frontendAction.Commands,
					Shell:       frontendAction.Shell,
					Timeout:     frontendAction.Timeout,
					Disable:     frontendAction.Disable,
					If:          frontendAction.If,
					Loop:        frontendAction.Loop,
					Type:        ActionType(frontendAction.Type),
					Namespaces:  frontendAction.Namespaces,
					Resources: Resources{
						CPU:     frontendAction.Resources.Cpu,
						Mem:     int(frontendAction.Resources.Mem),
						Disk:    int(frontendAction.Resources.Disk),
						Network: frontendAction.Resources.Network,
					},
				}}

			if frontendAction.SnippetConfig != nil {
				maps[ActionType(frontendAction.Type)].SnippetConfig = &SnippetConfig{
					Name:   frontendAction.SnippetConfig.Name,
					Source: frontendAction.SnippetConfig.Source,
					Labels: frontendAction.SnippetConfig.Labels,
				}
			}

			for _, cache := range frontendAction.Caches {
				maps[ActionType(frontendAction.Type)].Caches = append(maps[ActionType(frontendAction.Type)].Caches, ActionCache{
					Key:  cache.Key,
					Path: cache.Path,
				})
			}

			if frontendAction.Policy != nil {
				maps[ActionType(frontendAction.Type)].Policy = &Policy{
					Type: frontendAction.Policy.Type,
				}
			}

			actions = append(actions, maps)
		}
		s.Stages = append(s.Stages, &Stage{Actions: actions})
	}

	var pipelineParams []*PipelineParam
	for _, param := range frontendYmlSpec.Params {
		pipelineInput := toPipelineYamlParam(param)
		pipelineParams = append(pipelineParams, pipelineInput)
	}

	var pipelineOutputs []*PipelineOutput
	for _, output := range frontendYmlSpec.Outputs {
		pipelineOutput := toPipelineYamlOutput(output)
		pipelineOutputs = append(pipelineOutputs, pipelineOutput)
	}

	s.Params = pipelineParams
	s.Outputs = pipelineOutputs

	var lifecycle []*NetworkHookInfo
	for _, hookInfo := range frontendYmlSpec.Lifecycle {
		hookInfo := toPipelineYamlHookInfo(hookInfo)
		lifecycle = append(lifecycle, hookInfo)
	}
	s.Lifecycle = lifecycle

	return GenerateYml(s)
}

// ConvertToGraphPipelineYml: YAML(Spec) -> pb.PipelineYml
func ConvertToGraphPipelineYml(data []byte) (*pb.PipelineYml, error) {

	pipelineYml, err := New(data, WithFlatParams(false))
	if err != nil {
		return nil, err
	}

	params := pipelineYml.Spec().Params
	var pipelineParams []*pb.PipelineParam
	if params != nil {
		for _, param := range params {
			pipelineInput, err := toApiParam(param)
			if err != nil {
				return nil, err
			}
			pipelineParams = append(pipelineParams, pipelineInput)
		}
	}

	outputs := pipelineYml.Spec().Outputs
	var pipelineOutputs []*pb.PipelineOutput
	if outputs != nil {
		for _, output := range outputs {
			pipelineOutput := toApiOutput(output)
			pipelineOutputs = append(pipelineOutputs, pipelineOutput)
		}
	}

	var on *pb.TriggerConfig
	if pipelineYml.Spec().On != nil {
		merge := pipelineYml.Spec().On.Merge
		push := pipelineYml.Spec().On.Push
		if merge != nil || push != nil {
			on = &pb.TriggerConfig{}
			if merge != nil {
				var branches []string
				if merge.Branches != nil {
					branches = merge.Branches
				}
				on.Merge = &pb.MergeTrigger{Branches: branches}
			}
			if push != nil {
				var branches, tags []string
				if push.Branches != nil {
					branches = push.Branches
				}
				if push.Tags != nil {
					tags = push.Tags
				}
				on.Push = &pb.PushTrigger{
					Branches: branches,
					Tags:     tags,
				}
			}
		}
	}

	triggers := pipelineYml.Spec().Triggers
	var pipelineTriggers []*pb.PipelineTrigger
	if triggers != nil {
		for _, trigger := range triggers {
			eventName := trigger.On
			filter := trigger.Filter
			if eventName != "" && filter != nil {
				pipelineTriggers = append(pipelineTriggers, &pb.PipelineTrigger{On: eventName, Filter: filter})
			}
		}
	}

	result := &pb.PipelineYml{
		Version:     pipelineYml.Spec().Version,
		Envs:        pipelineYml.Spec().Envs,
		Cron:        pipelineYml.Spec().Cron,
		NeedUpgrade: pipelineYml.needUpgrade,
		Params:      pipelineParams,
		Outputs:     pipelineOutputs,
		On:          on,
		Triggers:    pipelineYml.Spec().Triggers,
		CronCompensator: func() *pb.CronCompensator {
			if pipelineYml.Spec().CronCompensator == nil {
				return nil
			}
			return &pb.CronCompensator{
				Enable:               pipelineYml.Spec().CronCompensator.Enable,
				LatestFirst:          pipelineYml.Spec().CronCompensator.LatestFirst,
				StopIfLatterExecuted: pipelineYml.Spec().CronCompensator.StopIfLatterExecuted,
			}
		}(),
	}

	var lifecycle []*pb.NetworkHookInfo
	for _, hookInfo := range pipelineYml.Spec().Lifecycle {
		hook, err := hookInfo.Convert2PB()
		if err != nil {
			return nil, err
		}
		lifecycle = append(lifecycle, hook)
	}
	result.Lifecycle = lifecycle

	if result.NeedUpgrade {
		result.YmlContent = string(pipelineYml.upgradedYmlContent)
	} else {
		graphYmlContent, err := GenerateYml(pipelineYml.s)
		if err != nil {
			return nil, err
		}
		result.YmlContent = string(graphYmlContent)
	}

	stages := make([]interface{}, 0)
	for _, stage := range pipelineYml.Spec().Stages {
		stageActions := make([]interface{}, 0)
		for _, typedAction := range stage.Actions {
			for _, action := range typedAction {
				resultAction := &apistructs.PipelineYmlAction{}
				resultAction.Type = action.Type.String()
				resultAction.Alias = action.Alias.String()
				resultAction.Version = action.Version
				resultAction.Params = action.Params
				resultAction.Image = action.Image
				resultAction.Shell = action.Shell
				resultAction.Commands = action.Commands
				resultAction.Timeout = action.Timeout
				resultAction.Namespaces = action.Namespaces
				resultAction.If = action.If
				resultAction.Disable = action.Disable
				resultAction.Loop = action.Loop
				resultAction.Resources = apistructs.Resources{
					Cpu:     action.Resources.CPU,
					Mem:     float64(action.Resources.Mem),
					Disk:    float64(action.Resources.Disk),
					Network: action.Resources.Network,
				}

				caches := action.Caches
				if caches != nil {
					var resultActionCaches []apistructs.ActionCache
					for _, v := range caches {
						resultActionCaches = append(resultActionCaches, apistructs.ActionCache{
							Path: v.Path,
							Key:  v.Key,
						})
					}
					resultAction.Caches = resultActionCaches
				}

				if action.SnippetConfig != nil {
					resultAction.SnippetConfig = action.SnippetConfig.toPBSnippetConfig()
				}

				if action.Policy != nil {
					resultAction.Policy = &apistructs.Policy{
						Type: action.Policy.Type,
					}
				}
				structValue, err := resultAction.Convert2StructValue()
				if err != nil {
					return nil, err
				}

				stageActions = append(stageActions, structValue.AsInterface())
			}
		}
		stages = append(stages, stageActions)
	}
	stagesPB, err := structpb.NewList(stages)
	if err != nil {
		return nil, err
	}
	result.Stages = stagesPB

	result.CronCompensator = cronCompensatorReset(result.CronCompensator)
	return result, nil
}

func cronCompensatorReset(cronCompensator *pb.CronCompensator) *pb.CronCompensator {
	if cronCompensator != nil {
		if cronCompensator.Enable == DefaultCronCompensator.Enable &&
			cronCompensator.LatestFirst == DefaultCronCompensator.LatestFirst &&
			cronCompensator.StopIfLatterExecuted == DefaultCronCompensator.StopIfLatterExecuted {
			return nil
		}
	}
	return cronCompensator
}

func toApiParam(pipelineInput *PipelineParam) (params *pb.PipelineParam, err error) {
	paramDefault, err := structpb.NewValue(pipelineInput.Default)
	if err != nil {
		return nil, err
	}
	return &pb.PipelineParam{
		Name:     pipelineInput.Name,
		Required: pipelineInput.Required,
		Default:  paramDefault,
		Desc:     pipelineInput.Desc,
		Type:     pipelineInput.Type,
	}, nil
}

func toApiOutput(pipelineOutput *PipelineOutput) (outputs *pb.PipelineOutput) {
	return &pb.PipelineOutput{
		Desc: pipelineOutput.Desc,
		Name: pipelineOutput.Name,
		Ref:  pipelineOutput.Ref,
	}
}

func toPipelineYamlParam(params *apistructs.PipelineParam) (pipelineInput *PipelineParam) {
	return &PipelineParam{
		Name:     params.Name,
		Required: params.Required,
		Default:  params.Default,
		Desc:     params.Desc,
		Type:     params.Type,
	}
}

func toPipelineYamlOutput(outputs *apistructs.PipelineOutput) (pipelineOutput *PipelineOutput) {
	return &PipelineOutput{
		Desc: outputs.Desc,
		Name: outputs.Name,
		Ref:  outputs.Ref,
	}
}

func toPipelineYamlHookInfo(hookInfo *apistructs.NetworkHookInfo) (pipelineHookInfo *NetworkHookInfo) {
	return &NetworkHookInfo{
		Hook:   hookInfo.Hook,
		Client: hookInfo.Client,
		Labels: hookInfo.Labels,
	}
}
