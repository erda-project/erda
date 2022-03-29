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

package extmarketsvc

import (
	"fmt"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/conf"
	"github.com/erda-project/erda/pkg/parser/diceyml"
	"github.com/erda-project/erda/pkg/parser/pipelineyml"
)

var (
	defaultVersion = "default"
)

type SearchOption struct {
	NeedRender   bool
	Placeholders map[string]string
}
type OpOption func(*SearchOption)

func SearchActionWithRender(placeholders map[string]string) OpOption {
	return func(so *SearchOption) {
		so.NeedRender = true
		so.Placeholders = placeholders
	}
}

func (s *ExtMarketSvc) constructAllActions() error {
	allExtensions, err := s.bdl.QueryExtensions(apistructs.ExtensionQueryRequest{
		All:  true,
		Type: "action",
	})
	if err != nil {
		return errors.Errorf("failed to query all extension: %v", err)
	}
	s.pools.Start()
	for i := range allExtensions {
		extension := allExtensions[i]
		s.pools.MustGo(func() {
			s.updateExtension(extension)
		})
	}
	s.pools.Stop()
	return nil
}

func (s *ExtMarketSvc) updateExtension(extension apistructs.Extension) {
	extensionVersions, err := s.bdl.QueryExtensionVersions(apistructs.ExtensionVersionQueryRequest{
		Name:               extension.Name,
		All:                "true",
		YamlFormat:         true,
		OrderByVersionDesc: true,
	})
	if err != nil {
		logrus.Errorf("failed to query extension version, name: %s, err: %v", extension.Name, err)
		return
	}
	s.Lock()
	defer s.Unlock()
	delete(s.defaultActions, extension.Name)
	for _, extensionVersion := range extensionVersions {
		s.actions[fmt.Sprintf("%s@%s", extension.Name, extensionVersion.Version)] = extensionVersion
		if extensionVersion.IsDefault {
			s.defaultActions[extension.Name] = extensionVersion
		}
	}
	// if not get the default version, set the first public version as default
	if _, ok := s.defaultActions[extension.Name]; !ok && len(extensionVersions) > 0 {
		for _, extensionVersion := range extensionVersions {
			if extensionVersion.Public {
				s.defaultActions[extension.Name] = extensionVersion
				break
			}
		}
	}
}

// getOrUpdateExtension get the fitted extension from the cache
// if not exist, try to update the cache by the given extension name
func (s *ExtMarketSvc) getOrUpdateExtension(nameVersion string) (action apistructs.ExtensionVersion, found bool) {
	splits := strings.SplitN(nameVersion, "@", 2)
	name := splits[0]
	version := ""
	if len(splits) > 1 {
		version = splits[1]
	}
	if version == "" {
		s.Lock()
		action, found = s.defaultActions[name]
		s.Unlock()
		if !found {
			newAction, err := s.bdl.GetExtensionVersion(apistructs.ExtensionVersionGetRequest{
				Name:       name,
				Version:    defaultVersion,
				YamlFormat: true,
			})
			if err != nil {
				found = false
				return
			}
			s.Lock()
			s.defaultActions[name] = *newAction
			s.Unlock()
			return *newAction, true
		}
		return
	}
	s.Lock()
	action, found = s.actions[nameVersion]
	s.Unlock()
	if !found {
		newAction, err := s.bdl.GetExtensionVersion(apistructs.ExtensionVersionGetRequest{
			Name:       name,
			Version:    version,
			YamlFormat: true,
		})
		if err != nil {
			found = false
			return
		}
		s.Lock()
		s.actions[nameVersion] = *newAction
		s.Unlock()
		return *newAction, true
	}
	return
}

func (s *ExtMarketSvc) continuousRefreshAction() {
	ticker := time.NewTicker(time.Minute * time.Duration(conf.ExtensionVersionRefreshIntervalMinute()))
	s.constructAllActions()
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if err := s.constructAllActions(); err != nil {
				logrus.Errorf("extension market failed to construct all actions: %v", err)
			}
		}
	}
}

// each search item: ActionType 或 ActionType@version
// output: map[item]Image
func (s *ExtMarketSvc) SearchActions(items []string, ops ...OpOption) (map[string]*diceyml.Job, map[string]*apistructs.ActionSpec, error) {
	so := SearchOption{
		NeedRender:   false,
		Placeholders: nil,
	}
	for _, op := range ops {
		op(&so)
	}

	actionDiceYmlJobMap := make(map[string]*diceyml.Job)
	actionSpecMap := make(map[string]*apistructs.ActionSpec)
	for _, nameVersion := range items {
		action, ok := s.getOrUpdateExtension(nameVersion)
		if !ok {
			return nil, nil, errors.Errorf("failed to find action: %s", nameVersion)
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
