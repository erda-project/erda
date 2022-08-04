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

package rule

import (
	"context"
	"fmt"

	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/erda-project/erda-proto-go/dop/rule/pb"
	"github.com/erda-project/erda/internal/apps/dop/providers/rule/db"
	"github.com/erda-project/erda/pkg/common/apis"
	"github.com/erda-project/erda/pkg/common/errors"
	"github.com/erda-project/erda/pkg/strutil"
)

type ruleService struct {
	p *provider
}

func (s *ruleService) Fire(ctx context.Context, req *pb.FireRequest) (*pb.FireResponse, error) {
	fired, err := s.p.ruleExecutor.Fire(req)
	return &pb.FireResponse{
		Output: fired,
	}, err
}

func (s *ruleService) CreateRule(ctx context.Context, req *pb.CreateRuleRequest) (*pb.CreateRuleResponse, error) {
	userID := apis.GetUserID(ctx)
	if userID == "" {
		return nil, fmt.Errorf("not login")
	}
	if req.Code == nil || req.Params == nil {
		return nil, errors.NewInvalidParameterError("req", "code and params required")
	}

	r := &db.Rule{
		Scope:     req.Scope,
		ScopeID:   req.ScopeID,
		Params:    db.ActionParams(*req.Params),
		EventType: req.EventType,
		Enabled:   &req.Enabled,
		Updator:   userID,
		Name:      req.Name,
	}
	if req.Code != nil {
		r.Code = *req.Code
	}
	if err := s.p.db.CreateRule(r); err != nil {
		return nil, err
	}
	return &pb.CreateRuleResponse{}, nil
}

func (s *ruleService) GetRule(ctx context.Context, req *pb.GetRuleRequest) (*pb.GetRuleResponse, error) {
	r, err := s.p.db.GetRule(req.Id)
	if err != nil {
		return nil, err
	}
	return &pb.GetRuleResponse{
		Data:    ToPbRule(*r),
		UserIDs: []string{r.Updator},
	}, nil
}

func (s *ruleService) UpdateRule(ctx context.Context, req *pb.UpdateRuleRequest) (*pb.UpdateRuleResponse, error) {
	userID := apis.GetUserID(ctx)
	if userID == "" {
		return nil, fmt.Errorf("not login")
	}

	r := &db.Rule{
		ID:        req.Id,
		Scope:     req.Scope,
		ScopeID:   req.ScopeID,
		Code:      req.Code,
		EventType: req.EventType,
		Enabled:   req.Enabled,
		Name:      req.Name,
		Updator:   userID,
	}
	if req.Params != nil {
		r.Params = db.ActionParams(*req.Params)
	}
	if err := s.p.db.UpdateRule(r); err != nil {
		return nil, err
	}
	return &pb.UpdateRuleResponse{}, nil
}

func ToPbRule(r db.Rule) *pb.Rule {
	return &pb.Rule{
		Id:        r.ID,
		Name:      r.Name,
		Scope:     r.Scope,
		ScopeID:   r.ScopeID,
		Code:      r.Code,
		EventType: r.EventType,
		Params: &pb.ActionParams{
			Nodes: r.Params.Nodes,
		},
		CreatedAt: timestamppb.New(r.CreatedAt),
		UpdatedAt: timestamppb.New(r.UpdatedAt),
		Enabled:   *r.Enabled,
		Updator:   r.Updator,
	}
}

func (s *ruleService) ListRules(ctx context.Context, req *pb.ListRulesRequest) (*pb.ListRulesResponse, error) {
	userID := apis.GetUserID(ctx)
	if userID == "" {
		return nil, fmt.Errorf("not login")
	}

	if req.PageNo == 0 {
		req.PageNo = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 10
	}
	r, total, err := s.p.db.ListRules(req, false)
	if err != nil {
		return nil, err
	}
	res := make([]*pb.Rule, len(r))
	userIDs := make([]string, len(r))
	for i, obj := range r {
		res[i] = ToPbRule(obj)
		userIDs[i] = obj.Updator
	}

	return &pb.ListRulesResponse{
		Data: &pb.ListRulesResponseData{
			Total: total,
			List:  res,
		},
		UserIDs: strutil.DedupSlice(userIDs),
	}, nil
}

func (s *ruleService) DeleteRule(ctx context.Context, req *pb.DeleteRuleRequest) (*pb.DeleteRuleResponse, error) {
	userID := apis.GetUserID(ctx)
	if userID == "" {
		return nil, fmt.Errorf("not login")
	}

	err := s.p.db.DeleteRule(req.Id)
	return &pb.DeleteRuleResponse{}, err
}

func (s *ruleService) ListRuleExecHistory(ctx context.Context, req *pb.ListRuleExecHistoryRequest) (*pb.ListRuleExecHistoryResponse, error) {
	userID := apis.GetUserID(ctx)
	if userID == "" {
		return nil, fmt.Errorf("not login")
	}

	if req.PageNo == 0 {
		req.PageNo = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 10
	}
	r, total, err := s.p.db.ListRuleExecRecords(req)
	if err != nil {
		return nil, err
	}
	res := make([]*pb.RuleExecHistory, len(r))
	for i, obj := range r {
		res[i] = ToPbRuleExecHistory(obj)
	}
	return &pb.ListRuleExecHistoryResponse{
		Data: &pb.ListRuleExecHistoryResponseData{
			Total: total,
			List:  res,
		},
	}, nil
}

func ToPbRuleExecHistory(r db.RuleExecRecord) *pb.RuleExecHistory {
	env, _ := structpb.NewValue(map[string]interface{}(r.Env))
	return &pb.RuleExecHistory{
		Id:           r.ID,
		Scope:        r.Scope,
		ScopeID:      r.ScopeID,
		RuleID:       r.RuleID,
		Code:         r.Code,
		Env:          env,
		CreatedAt:    timestamppb.New(r.CreatedAt),
		Succeed:      *r.Succeed,
		ActionOutput: r.ActionOutput,
	}
}
