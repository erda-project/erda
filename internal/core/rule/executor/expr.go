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

	rulepb "github.com/erda-project/erda-proto-go/core/rule/pb"
	"github.com/erda-project/erda/internal/core/rule/dao"
)

type RuleSetExecutor interface {
	Exec(r *RuleConfig, env map[string]interface{}) (bool, error)
	BuildRuleEnv(req *rulepb.FireRequest) (*RuleEnv, error)
	AddExecutionRecords(c *RecordConfig) error
}

type ExprExecutor struct {
	DB *dao.DBClient
}

type RuleEnv struct {
	Configs []*RuleConfig
	Env     map[string]interface{}
}

type RuleConfig struct {
	RuleSetID string
	Code      string
	Action    dao.ActionParams
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
	ruleSets, _, err := e.DB.ListRuleSets(&rulepb.ListRuleSetsRequest{
		Scope:     req.Scope,
		ScopeID:   req.ScopeID,
		EventType: req.EventType,
		Enabled:   &enabled,
		PageNo:    1,
		PageSize:  10,
	})
	if err != nil {
		return nil, err
	}

	ruleConfigs := make([]*RuleConfig, 0, len(ruleSets))
	for _, r := range ruleSets {
		ruleConfigs = append(ruleConfigs, &RuleConfig{
			RuleSetID: r.ID,
			Code:      r.Code,
			Action:    r.Params,
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
}

func (e *ExprExecutor) AddExecutionRecords(r *RecordConfig) error {
	records := make([]dao.RuleSetExecRecord, 0, len(r.RuleConfigs))
	for i, c := range r.RuleConfigs {
		records = append(records, dao.RuleSetExecRecord{
			Scope:        r.Scope,
			ScopeID:      r.ScopeID,
			RuleSetID:    c.RuleSetID,
			Code:         c.Code,
			Env:          r.Env,
			Succeed:      r.Results[i],
			ActionOutput: r.ActionOutputs[i],
		})
	}
	return e.DB.BatchCreateRuleSetExecRecords(records)
}
