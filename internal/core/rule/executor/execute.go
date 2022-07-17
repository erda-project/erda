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

package executor

import (
	"encoding/json"
	"errors"

	"github.com/erda-project/erda-proto-go/core/rule/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/internal/core/rule/actions"
	"github.com/erda-project/erda/internal/core/rule/dao"
	"github.com/erda-project/erda/internal/core/rule/jsonnet"
)

type Executor struct {
	RuleSetExecutor
	TemplateParser jsonnet.Engine
}

func (e Executor) Fire(req *pb.FireRequest) ([]bool, error) {
	if req == nil {
		return nil, errors.New("empty req")
	}
	ruleEnv, err := e.BuildRuleEnv(req)
	if err != nil {
		return nil, err
	}

	configs := ruleEnv.Configs
	results := make([]bool, 0, len(configs))
	actionOutputs := make([]string, 0, len(configs))
	for _, v := range configs {
		res, err := e.Exec(v, ruleEnv.Env)
		if err != nil {
			return nil, err
		}
		if !res {
			results = append(results, res)
			continue
		}
		var output string
		actionRes, err := e.Action(req.EventType, ruleEnv.Env, v.Action)
		if err != nil {
			output = err.Error()
		} else {
			output = actionRes
		}
		actionOutputs = append(actionOutputs, output)
		results = append(results, res)
	}

	err = e.AddExecutionRecords(&RecordConfig{
		Scope:         req.Scope,
		ScopeID:       req.ScopeID,
		Results:       results,
		RuleConfigs:   configs,
		Env:           ruleEnv.Env,
		ActionOutputs: actionOutputs,
	})
	if err != nil {
		return nil, err
	}

	return results, nil
}

// Outgoing API:  dingtalk config is available now
func (e Executor) Action(eventType string, content map[string]interface{}, params dao.ActionParams) (string, error) {
	if d := params.DingTalk; d != nil {
		target := apistructs.Target{
			Receiver: d.Webhook,
			Secret:   d.Signature,
		}
		url, err := target.GetSignURL()
		if err != nil {
			return "", err
		}
		bytes, err := json.Marshal(content)
		if err != nil {
			return "", err
		}
		configs := make([]jsonnet.TLACodeConfig, 0)
		configs = append(configs, jsonnet.TLACodeConfig{
			Key:   "ctx",
			Value: string(bytes),
		})
		b, err := e.TemplateParser.EvaluateBySnippet(params.Snippet, configs)
		if err != nil {
			return "", err
		}
		action := actions.APIConfig{
			URL:  url,
			Body: b,
		}
		return action.Send()
	}
	return "", nil
}
