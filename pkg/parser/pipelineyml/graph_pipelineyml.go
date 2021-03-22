package pipelineyml

import (
	"gopkg.in/yaml.v3"

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
					Timeout:     frontendAction.Timeout,
					If:          frontendAction.If,
					Loop:        frontendAction.Loop,
					Type:        ActionType(frontendAction.Type),
					Namespaces:  frontendAction.Namespaces,
					Resources: Resources{
						CPU:  frontendAction.Resources.Cpu,
						Mem:  int(frontendAction.Resources.Mem),
						Disk: int(frontendAction.Resources.Disk),
					},
				}}

			if frontendAction.SnippetConfig != nil {
				maps[ActionType(frontendAction.Type)].SnippetConfig = &SnippetConfig{
					Name:   frontendAction.SnippetConfig.Name,
					Source: frontendAction.SnippetConfig.Source,
					Labels: frontendAction.SnippetConfig.Labels,
				}
			}

			actions = append(actions, maps)
		}
		s.Stages = append(s.Stages, &Stage{Actions: actions})
	}

	params := frontendYmlSpec.Params
	var pipelineParams []*PipelineParam
	if params != nil {
		for _, param := range params {
			pipelineInput := toPipelineYamlParam(param)
			pipelineParams = append(pipelineParams, pipelineInput)
		}
	}
	outputs := frontendYmlSpec.Outputs
	var pipelineOutputs []*PipelineOutput
	if outputs != nil {
		for _, output := range outputs {
			pipelineOutput := toPipelineYamlOutput(output)
			pipelineOutputs = append(pipelineOutputs, pipelineOutput)
		}
	}
	s.Params = pipelineParams
	s.Outputs = pipelineOutputs

	return GenerateYml(s)
}

// ConvertToGraphPipelineYml: YAML(Spec) -> apistructs.PipelineYml
func ConvertToGraphPipelineYml(data []byte) (*apistructs.PipelineYml, error) {

	pipelineYml, err := New(data, WithFlatParams(false))
	if err != nil {
		return nil, err
	}

	params := pipelineYml.Spec().Params
	var pipelineParams []*apistructs.PipelineParam
	if params != nil {
		for _, param := range params {
			pipelineInput := toApiParam(param)
			pipelineParams = append(pipelineParams, pipelineInput)
		}
	}

	outputs := pipelineYml.Spec().Outputs
	var pipelineOutputs []*apistructs.PipelineOutput
	if outputs != nil {
		for _, output := range outputs {
			pipelineOutput := toApiOutput(output)
			pipelineOutputs = append(pipelineOutputs, pipelineOutput)
		}
	}

	var on *apistructs.TriggerConfig
	if pipelineYml.Spec().On != nil {
		on = &apistructs.TriggerConfig{}
		if pipelineYml.Spec().On.Merge != nil {
			var branches []string
			if pipelineYml.Spec().On.Merge.Branches != nil {
				branches = pipelineYml.Spec().On.Merge.Branches
			}
			on.Merge = &apistructs.MergeTrigger{Branches: branches}
		}
		if pipelineYml.Spec().On.Push != nil {
			var branches, tags []string
			if pipelineYml.Spec().On.Push.Branches != nil {
				branches = pipelineYml.Spec().On.Push.Branches
			}
			if pipelineYml.Spec().On.Push.Tags != nil {
				tags = pipelineYml.Spec().On.Push.Tags
			}
			on.Push = &apistructs.PushTrigger{
				Branches: branches,
				Tags:     tags,
			}
		}
	}
	result := &apistructs.PipelineYml{
		Version:     pipelineYml.Spec().Version,
		Envs:        pipelineYml.Spec().Envs,
		Cron:        pipelineYml.Spec().Cron,
		NeedUpgrade: pipelineYml.needUpgrade,
		Params:      pipelineParams,
		Outputs:     pipelineOutputs,
		On:          on,
	}

	if result.NeedUpgrade {
		result.YmlContent = string(pipelineYml.upgradedYmlContent)
	} else {
		graphYmlContent, err := GenerateYml(pipelineYml.s)
		if err != nil {
			return nil, err
		}
		result.YmlContent = string(graphYmlContent)
	}

	for _, stage := range pipelineYml.Spec().Stages {
		stageActions := make([]*apistructs.PipelineYmlAction, 0)
		for _, typedAction := range stage.Actions {
			for _, action := range typedAction {
				resultAction := &apistructs.PipelineYmlAction{}
				resultAction.Type = action.Type.String()
				resultAction.Alias = action.Alias.String()
				resultAction.Version = action.Version
				resultAction.Params = action.Params
				resultAction.Image = action.Image
				resultAction.Commands = action.Commands
				resultAction.Timeout = action.Timeout
				resultAction.Namespaces = action.Namespaces
				resultAction.If = action.If
				resultAction.Loop = action.Loop
				resultAction.Resources = apistructs.Resources{Cpu: action.Resources.CPU, Mem: float64(action.Resources.Mem), Disk: float64(action.Resources.Disk)}

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
					resultAction.SnippetConfig = action.SnippetConfig.toApiSnippetConfig()
				}

				stageActions = append(stageActions, resultAction)
			}
		}
		result.Stages = append(result.Stages, stageActions)
	}
	return result, nil
}

func toApiParam(pipelineInput *PipelineParam) (params *apistructs.PipelineParam) {
	return &apistructs.PipelineParam{
		Name:     pipelineInput.Name,
		Required: pipelineInput.Required,
		Default:  pipelineInput.Default,
		Desc:     pipelineInput.Desc,
		Type:     pipelineInput.Type,
	}
}

func toApiOutput(pipelineOutput *PipelineOutput) (outputs *apistructs.PipelineOutput) {
	return &apistructs.PipelineOutput{
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
