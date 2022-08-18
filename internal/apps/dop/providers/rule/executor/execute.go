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
	"github.com/erda-project/erda/internal/apps/dop/providers/rule/actions/dingtalkworknotice"
	"github.com/erda-project/erda/internal/apps/dop/providers/rule/actions/pipeline"
	"github.com/erda-project/erda/pkg/strutil"
)

type Executor struct {
	RuleExecutor
	API      api.Interface
	Pipeline pipeline.Interface
	DingTalk dingtalkworknotice.Interface
}

func (e *Executor) Fire(req *pb.FireRequest) ([]bool, error) {
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
	actors := make([]string, len(configs))
	for i, v := range configs {
		res, err := e.Exec(v, ruleEnv.Env)
		if err != nil {
			return nil, err
		}
		if !res {
			results[i] = res
			continue
		}
		actionOutputs[i] = e.Do(ruleEnv.Env, v)
		results[i] = res
		actors[i] = v.Actor
	}

	err = e.AddExecutionRecords(&RecordConfig{
		Scope:         req.Scope,
		ScopeID:       req.ScopeID,
		Results:       results,
		RuleConfigs:   configs,
		Env:           ruleEnv.Env,
		ActionOutputs: actionOutputs,
		Actors:        actors,
	})
	if err != nil {
		return nil, err
	}

	return results, nil
}

func (e *Executor) Do(content map[string]interface{}, config *RuleConfig) string {
	params := config.Action
	if len(params.Nodes) == 0 {
		return "no valid action nodes"
	}
	// TODO: multiple nodes concurrency
	results := make([]string, 0, len(params.Nodes))
	for _, node := range params.Nodes {
		switch node.Type {
		case "api":
			res, err := e.API.Send(&api.API{
				Snippet: node.Snippet,
				TLARaw:  content,
				Actor:   config.Actor,
			})
			results = addToResults(res, err, results)
		case "dingtalk":
			d := node.DingTalk
			target := apistructs.Target{
				Receiver: d.Webhook,
				Secret:   d.Signature,
			}
			url, err := target.GetSignURL()
			if err != nil {
				results = addToResults(url, err, results)
				continue
			}
			res, err := e.API.Send(&api.API{
				URL:     url,
				Snippet: node.Snippet,
				TLARaw:  content,
				Actor:   config.Actor,
			})
			results = addToResults(res, err, results)
		case "pipeline":
			res, err := e.Pipeline.CreatePipeline(content)
			results = addToResults(res, err, results)
		case "dingtalkworknotice":
			res, err := e.DingTalk.Send(&dingtalkworknotice.JsonnetParam{
				Snippet: node.Snippet,
				TLARaw:  content,
			})
			results = addToResults(res, err, results)
		default:
			results = addToResults("", errors.New("invalid action type"), results)
		}
	}
	return strutil.Join(results, ";")
}

func addToResults(output string, err error, results []string) []string {
	if err != nil {
		results = append(results, err.Error())
	} else {
		results = append(results, output)
	}
	return results
}
