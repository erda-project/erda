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

package actionmgr

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"

	"github.com/erda-project/erda-proto-go/core/extension/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/limit_sync_group"
	"github.com/erda-project/erda/pkg/parser/diceyml"
	"github.com/erda-project/erda/pkg/parser/pipelineyml"
)

// SearchActions .
// each search item: ActionType or ActionType@version
// output: map[item]Image
func (s *provider) SearchActions(items []string, ops ...OpOption) (map[string]*diceyml.Job, map[string]*apistructs.ActionSpec, error) {
	so := SearchOption{
		NeedRender:   false,
		Placeholders: nil,
	}
	for _, op := range ops {
		op(&so)
	}

	// search from dicehub
	pipelineActionMap := s.searchExtensionVersions(items)

	actionDiceYmlJobMap := make(map[string]*diceyml.Job)
	actionSpecMap := make(map[string]*apistructs.ActionSpec)
	for _, nameVersion := range items {
		action, ok := pipelineActionMap[nameVersion]
		if !ok {
			return nil, nil, errors.Errorf("failed to find action: %s", nameVersion)
		}

		diceYmlStr := action.Dice.GetStringValue()
		if so.NeedRender && len(so.Placeholders) > 0 {
			rendered, err := pipelineyml.RenderSecrets([]byte(diceYmlStr), so.Placeholders)
			if err != nil {
				errMsg := fmt.Sprintf("failed to render action's dice.yml, action: %s, err: %v", nameVersion, err)
				logrus.Errorf("[alert] %s, action's dice.yml: %#v", errMsg, action.Dice)
				return nil, nil, errors.New(errMsg)
			}
			diceYmlStr = string(rendered)
		}

		diceYmlOps := make([]diceyml.Options, 0)
		if so.ClusterInfo != nil {
			diceYmlOps = append(diceYmlOps, diceyml.WithPlatformInfo(so.ClusterInfo))
		}
		diceYml, err := diceyml.New([]byte(diceYmlStr), false, diceYmlOps...)
		if err != nil {
			errMsg := fmt.Sprintf("failed to parse action's dice.yml, action: %s, err: %v", nameVersion, err)
			logrus.Errorf("[alert] %s, action's dice.yml: %#v", errMsg, action.Dice)
			return nil, nil, errors.New(errMsg)
		}
		for _, job := range diceYml.Obj().Jobs {
			actionDiceYmlJobMap[nameVersion] = job
			break
		}
		actionSpecMap[nameVersion] = nil
		specYmlStr := action.Spec.GetStringValue()
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

func (s *provider) searchExtensionVersions(items []string) map[string]*pb.ExtensionVersion {
	notFindActionMap := make(map[string]*pb.ExtensionVersion)

	if s.EdgeRegister.IsEdge() {
		return notFindActionMap
	}

	worker := limit_sync_group.NewWorker(5)
	for _, nameVersion := range items {
		worker.AddFunc(func(locker *limit_sync_group.Locker, i ...interface{}) error {
			nameVersion := i[0].(string)
			action, ok := s.getOrUpdateExtensionFromCache(nameVersion)

			locker.Lock()
			defer locker.Unlock()

			if ok {
				notFindActionMap[nameVersion] = action
			}
			return nil
		}, nameVersion)
	}
	worker.Do()

	return notFindActionMap
}

// MakeActionTypeVersion return ext item.
// Example: git, git@1.0, git@1.1
func (s *provider) MakeActionTypeVersion(action *pipelineyml.Action) string {
	r := action.Type.String()
	if action.Version != "" {
		r = r + "@" + action.Version
	}
	return r
}

func (s *provider) MakeActionLocationsBySource(source apistructs.PipelineSource) []string {
	var locations []string
	switch source {
	case apistructs.PipelineSourceCDPDev, apistructs.PipelineSourceCDPTest, apistructs.PipelineSourceCDPStaging, apistructs.PipelineSourceCDPProd, apistructs.PipelineSourceBigData:
		locations = append(locations, apistructs.PipelineTypeFDP.String()+"/")
	case apistructs.PipelineSourceDice, apistructs.PipelineSourceProject, apistructs.PipelineSourceProjectLocal, apistructs.PipelineSourceOps, apistructs.PipelineSourceQA:
		locations = append(locations, apistructs.PipelineTypeCICD.String()+"/")
	}

	locations = append(locations, apistructs.PipelineTypeDefault.String()+"/")
	return locations
}
