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

	pb "github.com/erda-project/erda-proto-go/core/rule/pb"
	"github.com/erda-project/erda/internal/core/rule/dao"
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

func (s *ruleService) CreateRuleSet(ctx context.Context, req *pb.CreateRuleSetRequest) (*pb.CreateRuleSetResponse, error) {
	userID := apis.GetUserID(ctx)
	if userID == "" {
		return nil, fmt.Errorf("not login")
	}
	if req.Code == nil || req.Params == nil {
		return nil, errors.NewInvalidParameterError("req", "code and params required")
	}

	r := &dao.RuleSet{
		Scope:     req.Scope,
		ScopeID:   req.ScopeID,
		Params:    dao.ActionParams(*req.Params),
		EventType: req.EventType,
		Enabled:   req.Enabled,
		Updator:   userID,
		Name:      req.Name,
	}
	if req.Code != nil {
		r.Code = *req.Code
	}
	if err := s.p.db.CreateRuleSet(r); err != nil {
		return nil, err
	}
	return &pb.CreateRuleSetResponse{}, nil
}

func (s *ruleService) GetRuleSet(ctx context.Context, req *pb.GetRuleSetRequest) (*pb.GetRuleSetResponse, error) {
	r, err := s.p.db.GetRuleSet(req.Id)
	if err != nil {
		return nil, err
	}
	return &pb.GetRuleSetResponse{
		Data:    ToPbRuleSet(*r),
		UserIDs: []string{r.Updator},
	}, nil
}

func (s *ruleService) UpdateRuleSet(ctx context.Context, req *pb.UpdateRuleSetRequest) (*pb.UpdateRuleSetResponse, error) {
	userID := apis.GetUserID(ctx)
	if userID == "" {
		return nil, fmt.Errorf("not login")
	}

	r := &dao.RuleSet{
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
		r.Params = dao.ActionParams(*req.Params)
	}
	if err := s.p.db.UpdateRuleSet(r); err != nil {
		return nil, err
	}
	return &pb.UpdateRuleSetResponse{}, nil
}

func ToPbRuleSet(r dao.RuleSet) *pb.RuleSet {
	return &pb.RuleSet{
		Id:        r.ID,
		Name:      r.Name,
		Scope:     r.Scope,
		ScopeID:   r.ScopeID,
		Code:      r.Code,
		EventType: r.EventType,
		Params: &pb.ActionParams{
			DingTalk: r.Params.DingTalk,
			Snippet:  r.Params.Snippet,
		},
		CreatedAt: timestamppb.New(r.CreatedAt),
		UpdatedAt: timestamppb.New(r.UpdatedAt),
		Enabled:   r.Enabled,
		Updator:   r.Updator,
	}
}

func (s *ruleService) ListRuleSets(ctx context.Context, req *pb.ListRuleSetsRequest) (*pb.ListRuleSetsResponse, error) {
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
	r, total, err := s.p.db.ListRuleSets(req)
	if err != nil {
		return nil, err
	}
	res := make([]*pb.RuleSet, len(r))
	userIDs := make([]string, len(r))
	for i, obj := range r {
		res[i] = ToPbRuleSet(obj)
		userIDs[i] = obj.Updator
	}

	return &pb.ListRuleSetsResponse{
		Data: &pb.ListRuleSetsResponseData{
			Total: total,
			List:  res,
		},
		UserIDs: strutil.DedupSlice(userIDs),
	}, nil
}

func (s *ruleService) DeleteRuleSet(ctx context.Context, req *pb.DeleteRuleSetRequest) (*pb.DeleteRuleSetResponse, error) {
	userID := apis.GetUserID(ctx)
	if userID == "" {
		return nil, fmt.Errorf("not login")
	}

	err := s.p.db.DeleteRuleSet(req.Id)
	return &pb.DeleteRuleSetResponse{}, err
}

func (s *ruleService) ListRuleSetExecHistory(ctx context.Context, req *pb.ListRuleSetExecHistoryRequest) (*pb.ListRuleSetExecHistoryResponse, error) {
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
	r, total, err := s.p.db.ListRuleSetExecRecords(req)
	if err != nil {
		return nil, err
	}
	res := make([]*pb.RuleSetExecHistory, len(r))
	for i, obj := range r {
		res[i] = ToPbRuleSetExecHistory(obj)
	}
	return &pb.ListRuleSetExecHistoryResponse{
		Data: &pb.ListRuleSetExecHistoryResponseData{
			Total: total,
			List:  res,
		},
	}, nil
}

func ToPbRuleSetExecHistory(r dao.RuleSetExecRecord) *pb.RuleSetExecHistory {
	env, _ := structpb.NewValue(map[string]interface{}(r.Env))
	return &pb.RuleSetExecHistory{
		Id:           r.ID,
		Scope:        r.Scope,
		ScopeID:      r.ScopeID,
		RuleSetID:    r.RuleSetID,
		Code:         r.Code,
		Env:          env,
		CreatedAt:    timestamppb.New(r.CreatedAt),
		Succeed:      r.Succeed,
		ActionOutput: r.ActionOutput,
	}
}
