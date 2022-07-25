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
	"errors"

	"github.com/erda-project/erda-proto-go/dop/rule/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/internal/apps/dop/providers/rule/actions/api"
	"github.com/erda-project/erda/internal/apps/dop/providers/rule/db"
)

type Executor struct {
	RuleExecutor
	API api.Interface
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
	results := make([]bool, len(configs))
	actionOutputs := make([]string, len(configs))
	for i, v := range configs {
		res, err := e.Exec(v, ruleEnv.Env)
		if err != nil {
			return nil, err
		}
		if !res {
			results[i] = res
			continue
		}
		var output string
		actionRes, err := e.DingTalkAction(ruleEnv.Env, v.Action)
		if err != nil {
			output = err.Error()
		} else {
			output = actionRes
		}
		actionOutputs[i] = output
		results[i] = res
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

func (e Executor) DingTalkAction(content map[string]interface{}, params db.ActionParams) (string, error) {
	if len(params.Nodes) == 0 {
		return "", nil
	}
	// TODO: multiple nodes concurrency
	node := params.Nodes[0]
	d := node.DingTalk
	if d == nil {
		return "", nil
	}
	target := apistructs.Target{
		Receiver: d.Webhook,
		Secret:   d.Signature,
	}
	url, err := target.GetSignURL()
	if err != nil {
		return "", err
	}
	return e.API.Send(&api.API{
		URL:     url,
		Snippet: node.Snippet,
		TLARaw:  content,
	})
}
