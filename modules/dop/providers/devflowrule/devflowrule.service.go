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

package devflowrule

import (
	"context"
	"encoding/json"

	"github.com/erda-project/erda-proto-go/dop/devflowrule/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/dop/providers/devflowrule/db"
	"github.com/erda-project/erda/modules/dop/services/apierrors"
	"github.com/erda-project/erda/pkg/common/apis"
)

const resource = "branch_rule"

type Service interface {
	CreateDevFlowRule(context.Context, *pb.CreateDevFlowRuleRequest) (*pb.CreateDevFlowRuleResponse, error)
	DeleteDevFlowRule(context.Context, *pb.DeleteDevFlowRuleRequest) (*pb.DeleteDevFlowRuleResponse, error)
	UpdateDevFlowRule(context.Context, *pb.UpdateDevFlowRuleRequest) (*pb.UpdateDevFlowRuleResponse, error)
	GetDevFlowRulesByProjectID(context.Context, *pb.GetDevFlowRuleRequest) (*pb.GetDevFlowRuleResponse, error)
}

type ServiceImplement struct {
	db  *db.Client
	bdl *bundle.Bundle
}

func (s *ServiceImplement) CreateDevFlowRule(ctx context.Context, request *pb.CreateDevFlowRuleRequest) (*pb.CreateDevFlowRuleResponse, error) {
	project, err := s.bdl.GetProject(request.ProjectID)
	if err != nil {
		return nil, apierrors.ErrCreateDevFlowRule.InternalError(err)
	}
	org, err := s.bdl.GetOrg(project.OrgID)
	if err != nil {
		return nil, apierrors.ErrCreateDevFlowRule.InternalError(err)
	}

	flows := s.InitFlows()
	b, err := json.Marshal(&flows)
	if err != nil {
		return nil, apierrors.ErrCreateDevFlowRule.InternalError(err)
	}
	devFlow := db.DevFlowRule{
		Scope: db.Scope{
			OrgID:       org.ID,
			OrgName:     org.Name,
			ProjectID:   project.ID,
			ProjectName: project.Name,
		},
		Operator: db.Operator{
			Creator: request.UserID,
			Updater: request.UserID,
		},
		Flows: b,
	}
	err = s.db.CreateDevFlowRule(&devFlow)
	if err != nil {
		return nil, apierrors.ErrCreateDevFlowRule.InternalError(err)
	}
	return &pb.CreateDevFlowRuleResponse{Data: devFlow.Convert()}, nil
}

func (s *ServiceImplement) InitFlows() db.Flows {
	return db.Flows{
		{
			Name:             "DEV",
			FlowType:         "two_branch",
			TargetBranch:     "develop",
			ChangeFromBranch: "",
			ChangeBranch:     "feature/*,bugfix/*",
			EnableAutoMerge:  true,
			AutoMergeBranch:  "dev",
			Artifact:         "alpha",
			Environment:      "DEV",
			StartWorkflowHints: []db.StartWorkflowHint{
				{
					Place:            "TASK",
					ChangeBranchRule: "feature/*",
				},
				{
					Place:            "BUG",
					ChangeBranchRule: "bugfix/*",
				},
			},
		},
		{
			Name:               "TEST",
			FlowType:           "single_branch",
			TargetBranch:       "develop",
			ChangeFromBranch:   "",
			ChangeBranch:       "",
			EnableAutoMerge:    false,
			AutoMergeBranch:    "",
			Artifact:           "beta",
			Environment:        "TEST",
			StartWorkflowHints: nil,
		},
		{
			Name:               "STAGING",
			FlowType:           "three_branch",
			TargetBranch:       "master",
			ChangeFromBranch:   "develop",
			ChangeBranch:       "release/*",
			EnableAutoMerge:    false,
			AutoMergeBranch:    "",
			Artifact:           "rc",
			Environment:        "STAGING",
			StartWorkflowHints: nil,
		},
		{
			Name:               "PROD",
			FlowType:           "single_branch",
			TargetBranch:       "master",
			ChangeFromBranch:   "",
			ChangeBranch:       "",
			EnableAutoMerge:    false,
			AutoMergeBranch:    "",
			Artifact:           "stable",
			Environment:        "PROD",
			StartWorkflowHints: nil,
		},
	}

}

func (s *ServiceImplement) UpdateDevFlowRule(ctx context.Context, request *pb.UpdateDevFlowRuleRequest) (*pb.UpdateDevFlowRuleResponse, error) {
	if err := request.Validate(); err != nil {
		return nil, apierrors.ErrUpdateDevFlowRule.InvalidParameter(err)
	}

	devFlow, err := s.db.GetDevFlowRule(request.ID)
	if err != nil {
		return nil, apierrors.ErrUpdateDevFlowRule.InternalError(err)
	}

	access, err := s.bdl.CheckPermission(&apistructs.PermissionCheckRequest{
		UserID:   apis.GetUserID(ctx),
		Scope:    apistructs.ProjectScope,
		ScopeID:  devFlow.ProjectID,
		Resource: resource,
		Action:   apistructs.OperateAction,
	})
	if err != nil {
		return nil, apierrors.ErrUpdateDevFlowRule.InternalError(err)
	}
	if !access.Access {
		return nil, apierrors.ErrUpdateDevFlowRule.AccessDenied()
	}

	devFlow.Flows = db.JSON(request.Flows)
	devFlow.Operator.Updater = apis.GetUserID(ctx)
	if err = s.db.UpdateDevFlowRule(devFlow); err != nil {
		return nil, apierrors.ErrUpdateDevFlowRule.InternalError(err)
	}

	return &pb.UpdateDevFlowRuleResponse{Data: devFlow.Convert()}, nil
}

func (s *ServiceImplement) GetDevFlowRulesByProjectID(ctx context.Context, request *pb.GetDevFlowRuleRequest) (*pb.GetDevFlowRuleResponse, error) {
	access, err := s.bdl.CheckPermission(&apistructs.PermissionCheckRequest{
		UserID:   apis.GetUserID(ctx),
		Scope:    apistructs.ProjectScope,
		ScopeID:  request.ProjectID,
		Resource: resource,
		Action:   apistructs.OperateAction,
	})
	if err != nil {
		return nil, apierrors.ErrGetDevFlowRule.InternalError(err)
	}
	if !access.Access {
		return nil, apierrors.ErrGetDevFlowRule.AccessDenied()
	}

	wfs, err := s.db.GetDevFlowRuleByProjectID(request.ProjectID)
	if err != nil {
		return nil, apierrors.ErrGetDevFlowRule.InternalError(err)
	}

	return &pb.GetDevFlowRuleResponse{Data: wfs.Convert()}, nil
}

func (s *ServiceImplement) DeleteDevFlowRule(ctx context.Context, request *pb.DeleteDevFlowRuleRequest) (*pb.DeleteDevFlowRuleResponse, error) {
	if err := s.db.DeleteDevFlowRuleByProjectID(request.ProjectID); err != nil {
		return nil, apierrors.ErrDeleteDevFlowRule.InternalError(err)
	}
	return &pb.DeleteDevFlowRuleResponse{}, nil
}
