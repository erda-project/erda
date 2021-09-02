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

// @Title  this file is used to query the action definition
// @Description  query action definition and spec
package action_info

import (
	"errors"
	"fmt"
	"sync"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/pipeline/services/apierrors"
	"github.com/erda-project/erda/modules/pipeline/services/extmarketsvc"
	"github.com/erda-project/erda/pkg/parser/diceyml"
	"github.com/erda-project/erda/pkg/parser/pipelineyml"
	"github.com/erda-project/erda/pkg/strutil"
)

// passedDataWhenCreate stores data passed recursively when create graph.
type PassedDataWhenCreate struct {
	bdl              *bundle.Bundle
	actionJobDefines *sync.Map
	actionJobSpecs   *sync.Map
}

func (that *PassedDataWhenCreate) GetActionJobDefine(actionTypeVersion string) *diceyml.Job {
	if that == nil {
		return nil
	}
	if that.actionJobDefines == nil {
		return nil
	}

	if value, ok := that.actionJobDefines.Load(actionTypeVersion); ok {
		if job, ok := value.(*diceyml.Job); ok {
			return job
		}
	}
	return nil
}

func (that *PassedDataWhenCreate) GetActionJobSpecs(actionTypeVersion string) *apistructs.ActionSpec {

	if that == nil {
		return nil
	}
	if that.actionJobDefines == nil {
		return nil
	}

	if value, ok := that.actionJobSpecs.Load(actionTypeVersion); ok {
		if spec, ok := value.(*apistructs.ActionSpec); ok {
			return spec
		}
	}
	return nil
}

func (that *PassedDataWhenCreate) InitData(bdl *bundle.Bundle) {
	if that == nil {
		return
	}

	if that.actionJobDefines == nil {
		that.actionJobDefines = &sync.Map{}
	}
	if that.actionJobSpecs == nil {
		that.actionJobSpecs = &sync.Map{}
	}
	that.bdl = bdl
}

func (that *PassedDataWhenCreate) PutPassedDataByPipelineYml(pipelineYml *pipelineyml.PipelineYml) error {
	if that == nil {
		return nil
	}
	// batch search extensions
	var extItems []string
	for _, stage := range pipelineYml.Spec().Stages {
		for _, typedAction := range stage.Actions {
			for _, action := range typedAction {
				if action.Type.IsSnippet() {
					continue
				}
				extItem := extmarketsvc.MakeActionTypeVersion(action)
				// extension already searched, skip
				if _, ok := that.actionJobDefines.Load(extItem); ok {
					continue
				}
				extItems = append(extItems, extmarketsvc.MakeActionTypeVersion(action))
			}
		}
	}

	extItems = strutil.DedupSlice(extItems, true)
	actionJobDefines, actionJobSpecs, err := searchActions(that.bdl, extItems)
	if err != nil {
		return apierrors.ErrCreatePipelineGraph.InternalError(err)
	}

	for extItem, actionJobDefine := range actionJobDefines {
		that.actionJobDefines.Store(extItem, actionJobDefine)
	}
	for extItem, actionJobSpec := range actionJobSpecs {
		that.actionJobSpecs.Store(extItem, actionJobSpec)
	}
	return nil
}

type SearchOption struct {
	NeedRender   bool
	Placeholders map[string]string
}

func searchActions(bdl *bundle.Bundle, items []string) (map[string]*diceyml.Job, map[string]*apistructs.ActionSpec, error) {
	req := apistructs.ExtensionSearchRequest{Extensions: items, YamlFormat: true}
	actions, err := bdl.SearchExtensions(req)
	if err != nil {
		return nil, nil, err
	}

	so := SearchOption{
		NeedRender:   false,
		Placeholders: nil,
	}

	actionDiceYmlJobMap := make(map[string]*diceyml.Job)
	for nameVersion, action := range actions {
		if action.NotExist() {
			errMsg := fmt.Sprintf("action %q not exist in Extension Market", nameVersion)
			logrus.Errorf("[alert] %s", errMsg)
			return nil, nil, errors.New(errMsg)
		}

		diceYmlStr, ok := action.Dice.(string)
		if !ok {
			errMsg := fmt.Sprintf("failed to search action from extension market, action: %s, err: %s", nameVersion, "action's dice.yml is not string")
			logrus.Errorf("[alert] %s, action's dice.yml: %#v", errMsg, action.Dice)
			return nil, nil, errors.New(errMsg)
		}
		if so.NeedRender && len(so.Placeholders) > 0 {
			rendered, err := pipelineyml.RenderSecrets([]byte(diceYmlStr), so.Placeholders)
			if err != nil {
				errMsg := fmt.Sprintf("failed to render action's dice.yml, action: %s, err: %v", nameVersion, err)
				logrus.Errorf("[alert] %s, action's dice.yml: %#v", errMsg, action.Dice)
				return nil, nil, errors.New(errMsg)
			}
			diceYmlStr = string(rendered)
		}
		diceYml, err := diceyml.New([]byte(diceYmlStr), false)
		if err != nil {
			errMsg := fmt.Sprintf("failed to parse action's dice.yml, action: %s, err: %v", nameVersion, err)
			logrus.Errorf("[alert] %s, action's dice.yml: %#v", errMsg, action.Dice)
			return nil, nil, errors.New(errMsg)
		}
		for _, job := range diceYml.Obj().Jobs {
			actionDiceYmlJobMap[nameVersion] = job
			break
		}
	}
	actionSpecMap := make(map[string]*apistructs.ActionSpec)
	for nameVersion, action := range actions {
		actionSpecMap[nameVersion] = nil
		specYmlStr, ok := action.Spec.(string)
		if !ok {
			errMsg := fmt.Sprintf("failed to search action from extension market, action: %s, err: %s", nameVersion, "action's spec.yml is not string")
			logrus.Errorf("[alert] %s, action's spec.yml: %#v", errMsg, action.Spec)
			return nil, nil, errors.New(errMsg)
		}
		var actionSpec apistructs.ActionSpec
		if err := yaml.Unmarshal([]byte(specYmlStr), &actionSpec); err != nil {
			errMsg := fmt.Sprintf("failed to parse action's spec.yml, action: %s, err: %v", nameVersion, err)
			logrus.Errorf("[alert] %s, action's spec.yml: %#v", errMsg, action.Spec)
			return nil, nil, errors.New(errMsg)
		}
		actionSpecMap[nameVersion] = &actionSpec
	}

	return actionDiceYmlJobMap, actionSpecMap, nil
}
