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

	"github.com/antonmedv/expr"

	rulepb "github.com/erda-project/erda-proto-go/dop/rule/pb"
	"github.com/erda-project/erda/internal/apps/dop/providers/rule/db"
)

type RuleExecutor interface {
	Exec(r *RuleConfig, env map[string]interface{}) (bool, error)
	BuildRuleEnv(req *rulepb.FireRequest) (*RuleEnv, error)
	AddExecutionRecords(c *RecordConfig) error
}

type ExprExecutor struct {
	DB *db.DBClient
}

type RuleEnv struct {
	Configs []*RuleConfig
	Env     map[string]interface{}
}

type RuleConfig struct {
	RuleID string
	Code   string
	Action db.ActionParams
	Actor  string
}

func (e *ExprExecutor) Exec(r *RuleConfig, env map[string]interface{}) (bool, error) {
	program, err := expr.Compile(r.Code, expr.Env(env))
	if err != nil {
		return false, err
	}

	output, err := expr.Run(program, env)
	if err != nil {
		return false, err
	}
	res, ok := output.(bool)
	if !ok {
		return false, errors.New("invalid expr")
	}
	return res, nil
}

func (e *ExprExecutor) BuildRuleEnv(req *rulepb.FireRequest) (*RuleEnv, error) {
	if req.Scope == "" || req.ScopeID == "" || req.EventType == "" {
		return nil, errors.New("invalid request, missing scope info")
	}

	enabled := true
	rules, _, err := e.DB.ListRules(&rulepb.ListRulesRequest{
		Scope:     req.Scope,
		ScopeID:   req.ScopeID,
		EventType: req.EventType,
		Enabled:   &enabled,
		PageNo:    1,
		PageSize:  10,
	}, true)
	if err != nil {
		return nil, err
	}

	ruleConfigs := make([]*RuleConfig, 0, len(rules))
	for _, r := range rules {
		ruleConfigs = append(ruleConfigs, &RuleConfig{
			RuleID: r.ID,
			Code:   r.Code,
			Action: r.Params,
			Actor:  r.Actor,
		})
	}

	env := make(map[string]interface{})
	for k, v := range req.Env {
		env[k] = v.AsInterface()
	}

	return &RuleEnv{
		Configs: ruleConfigs,
		Env:     env,
	}, nil
}

type RecordConfig struct {
	Scope         string
	ScopeID       string
	Env           map[string]interface{}
	Results       []bool
	RuleConfigs   []*RuleConfig
	ActionOutputs []string
	Actors        []string
}

func (e *ExprExecutor) AddExecutionRecords(r *RecordConfig) error {
	records := make([]db.RuleExecRecord, 0, len(r.RuleConfigs))
	for i, c := range r.RuleConfigs {
		records = append(records, db.RuleExecRecord{
			Scope:        r.Scope,
			ScopeID:      r.ScopeID,
			RuleID:       c.RuleID,
			Code:         c.Code,
			Env:          r.Env,
			Succeed:      &r.Results[i],
			ActionOutput: r.ActionOutputs[i],
			Actor:        r.Actors[i],
		})
	}
	return e.DB.BatchCreateRuleExecRecords(records)
}
