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
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/commonutil/thirdparty/gittarutil"
	"github.com/erda-project/erda/modules/pipeline/precheck"
	"github.com/erda-project/erda/modules/pipeline/precheck/checkers/actionchecker/release"
	"github.com/erda-project/erda/modules/pipeline/precheck/prechecktype"
	"github.com/erda-project/erda/modules/pipeline/services/apierrors"
	"github.com/erda-project/erda/modules/pipeline/spec"
	"github.com/erda-project/erda/pkg/discover"
	"github.com/erda-project/erda/pkg/parser/diceyml"
	"github.com/erda-project/erda/pkg/parser/pipelineyml"
)

func (s *PipelineSvc) PreCheck(p *spec.Pipeline) error {
	tasks, err := s.dbClient.ListPipelineTasksByPipelineID(p.ID)
	if err != nil {
		return apierrors.ErrPreCheckPipeline.InternalError(err)
	}

	// ItemsForCheck
	itemsForCheck := prechecktype.ItemsForCheck{
		PipelineYml:               p.PipelineYml,
		Files:                     make(map[string]string),
		ActionSpecs:               make(map[string]apistructs.ActionSpec),
		Labels:                    p.MergeLabels(),
		Envs:                      p.Snapshot.Envs,
		ClusterName:               p.ClusterName,
		GlobalSnippetConfigLabels: p.Labels,
	}

	// files
	if p.CommitDetail.RepoAbbr != "" {
		diceymlByte, err := gittarutil.NewRepo(discover.Gittar(), p.CommitDetail.RepoAbbr).FetchFile(p.GetCommitID(), "dice.yml")
		if err == nil {
			itemsForCheck.Files["dice.yml"] = string(diceymlByte)
		}
	}
	err = setItemForCheckRealDiceYml(p, &itemsForCheck)
	if err != nil {
		return err
	}

	// 从 extension marketplace 获取 action
	extSearchReq := make([]string, 0)
	actionTypeVerMap := make(map[string]struct{})
	for _, task := range tasks {
		typeVersion := task.Extra.Action.GetActionTypeVersion()
		if _, ok := actionTypeVerMap[typeVersion]; ok {
			continue
		}
		actionTypeVerMap[typeVersion] = struct{}{}
		extSearchReq = append(extSearchReq, typeVersion)
	}
	_, actionSpecs, err := s.extMarketSvc.SearchActions(extSearchReq)
	if err != nil {
		return apierrors.ErrPreCheckPipeline.InternalError(err)
	}
	for typeVersion, actionSpec := range actionSpecs {
		if actionSpec != nil {
			itemsForCheck.ActionSpecs[typeVersion] = *actionSpec
		}
	}

	if p.Extra.StorageConfig.EnableShareVolume() {
		for _, task := range tasks {
			typeVersion := task.Extra.Action.GetActionTypeVersion()
			value, exist := actionSpecs[typeVersion].Labels["new_workspace"]
			if exist && value == "true" {
				// action带有new_workspace标签,使用独立目录
				p.Extra.TaskWorkspaces = append(p.Extra.TaskWorkspaces, task.Name)
			}
		}
		err = s.dbClient.UpdatePipelineExtraByPipelineID(p.ID, &p.PipelineExtra)
		if err != nil {
			return apierrors.ErrPreCheckPipeline.InternalError(err)
		}
	}

	// secrets
	secrets, _, holdOnKeys, _, err := s.FetchSecrets(p)
	if err != nil {
		return apierrors.ErrPreCheckPipeline.InternalError(err)
	}
	platformSecrets, err := s.FetchPlatformSecrets(p, holdOnKeys)
	if err != nil {
		return apierrors.ErrPreCheckPipeline.InternalError(err)
	}
	itemsForCheck.Secrets = platformSecrets
	for k, v := range secrets {
		itemsForCheck.Secrets[k] = v
	}

	precheckCtx := prechecktype.InitContext()
	abort, showMessage := precheck.PreCheck(precheckCtx, []byte(p.PipelineYml), itemsForCheck)
	if len(showMessage.Stacks) > 0 {
		if err := s.dbClient.UpdatePipelineShowMessage(p.ID, showMessage); err != nil {
			return apierrors.ErrPreCheckPipeline.InternalError(err)
		}
	}
	if abort {
		return apierrors.ErrPreCheckPipeline.InvalidParameter("precheck failed")
	}

	analyzedCrossCluster, ok := prechecktype.GetContextResult(precheckCtx, prechecktype.CtxResultKeyCrossCluster).(bool)
	if ok {
		p.Snapshot.AnalyzedCrossCluster = &analyzedCrossCluster
		if err := s.dbClient.StoreAnalyzedCrossCluster(p.ID, analyzedCrossCluster); err != nil {
			return apierrors.ErrPreCheckPipeline.InternalError(err)
		}
	}

	return nil
}

// 用户可能在 release 中设置了 dice_development_yml,dice_test_yml,dice_staging_yml,dice_production_yml 等不同环境的 dice.yml, 但是对应的校验也要转化
func setItemForCheckRealDiceYml(p *spec.Pipeline, itemForCheck *prechecktype.ItemsForCheck) error {
	if p == nil {
		return nil
	}

	y, err := pipelineyml.New([]byte(p.PipelineYml))
	if err != nil {
		return err
	}
	var worn error
	// 遍历 pipeline yml 中的 action
	y.Spec().LoopStagesActions(func(stage int, action *pipelineyml.Action) {
		if action.Type != release.ActionType {
			return
		}
		// 拿到 release 中指名的 yml
		var realDiceYmlParam interface{}
		workspace := p.Extra.DiceWorkspace
		switch workspace {
		case "DEV":
			realDiceYmlParam = action.Params["dice_development_yml"]
		case "PROD":
			realDiceYmlParam = action.Params["dice_production_yml"]
		case "TEST":
			realDiceYmlParam = action.Params["dice_test_yml"]
		case "STAGING":
			realDiceYmlParam = action.Params["dice_staging_yml"]
		default:
			realDiceYmlParam = action.Params["dice_yml"]
		}

		if realDiceYmlParam == nil {
			return
		}

		if p.CommitDetail.RepoAbbr == "" {
			return
		}

		var realDiceYmlStr, ok = realDiceYmlParam.(string)
		if !ok {
			return
		}

		realDiceYmlSplit := strings.Split(realDiceYmlStr, "/")
		var length = len(realDiceYmlSplit)
		if length < 1 {
			return
		}

		realDiceYmlName := realDiceYmlSplit[length-1]
		diceYmlByte, err := gittarutil.NewRepo(discover.Gittar(), p.CommitDetail.RepoAbbr).FetchFile(p.GetCommitID(), realDiceYmlName)
		if err != nil {
			worn = err
			return
		}

		var check = true
		if action.Params["check_diceyml"] != nil {
			check, err = strconv.ParseBool(action.Params["check_diceyml"].(string))
			if err != nil {
				check = true
			}
		}

		// 将 dice.yml 和 其他环境的 yml 合并下
		yml, err := composeEnvYml(itemForCheck.Files["dice.yml"], check, string(diceYmlByte), workspace.String())
		if err != nil {
			logrus.Errorf("composeEnvYml dice.yml error: %v", err)
			worn = err
			return
		}

		itemForCheck.Files["dice.yml"] = yml
	})

	return worn
}

func composeEnvYml(diceYaml string, check bool, otherYaml string, workspace string) (string, error) {
	d, err := diceyml.New([]byte(diceYaml), check)
	if err != nil {
		return "", errors.Wrap(err, "new parser failed")
	}

	switch workspace {
	case string(apistructs.DevWorkspace):
		err = composeYaml(d, "development", otherYaml)
	case string(apistructs.TestWorkspace):
		err = composeYaml(d, "test", otherYaml)
	case string(apistructs.StagingWorkspace):
		err = composeYaml(d, "staging", otherYaml)
	case string(apistructs.ProdWorkspace):
		err = composeYaml(d, "production", otherYaml)
	}
	if err != nil {
		return "", errors.Wrap(err, "failed to compose diceyml")
	}

	return d.YAML()
}

func composeYaml(targetYml *diceyml.DiceYaml, env, envYmlFile string) error {
	envYml, err := diceyml.New([]byte(envYmlFile), false)
	if err != nil {
		return err
	}

	err = targetYml.Compose(env, envYml)
	if err != nil {
		return err
	}

	return nil
}
