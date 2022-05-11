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

package devworkflow

import (
	"context"
	"encoding/json"

	"github.com/erda-project/erda-proto-go/dop/devworkflow/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/dop/providers/devworkflow/db"
	"github.com/erda-project/erda/modules/dop/services/apierrors"
	"github.com/erda-project/erda/pkg/common/apis"
)

const resource = "devWorkflow"

type Service interface {
	CreateDevWorkflow(context.Context, *pb.CreateDevWorkflowRequest) (*pb.CreateDevWorkflowResponse, error)
	DeleteDevWorkflow(context.Context, *pb.DeleteDevWorkflowRequest) (*pb.DeleteDevWorkflowResponse, error)
	UpdateDevWorkflow(context.Context, *pb.UpdateDevWorkflowRequest) (*pb.UpdateDevWorkflowResponse, error)
	GetDevWorkflowsByProjectID(context.Context, *pb.GetDevWorkflowRequest) (*pb.GetDevWorkflowResponse, error)
}

type ServiceImplement struct {
	db  *db.Client
	bdl *bundle.Bundle
}

func (s *ServiceImplement) CreateDevWorkflow(ctx context.Context, request *pb.CreateDevWorkflowRequest) (*pb.CreateDevWorkflowResponse, error) {
	project, err := s.bdl.GetProject(request.ProjectID)
	if err != nil {
		return nil, apierrors.ErrCreateDevWorkflow.InternalError(err)
	}
	org, err := s.bdl.GetOrg(project.OrgID)
	if err != nil {
		return nil, apierrors.ErrCreateDevWorkflow.InternalError(err)
	}

	flows := s.InitWorkFlows()
	b, err := json.Marshal(&flows)
	if err != nil {
		return nil, apierrors.ErrCreateDevWorkflow.InternalError(err)
	}
	wf := db.DevWorkflow{
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
		WorkFlows: b,
	}
	err = s.db.CreateWf(&wf)
	if err != nil {
		return nil, apierrors.ErrCreateDevWorkflow.InternalError(err)
	}
	return &pb.CreateDevWorkflowResponse{Data: wf.Convert()}, nil
}

func (s *ServiceImplement) InitWorkFlows() db.WorkFlows {
	return db.WorkFlows{
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

func (s *ServiceImplement) UpdateDevWorkflow(ctx context.Context, request *pb.UpdateDevWorkflowRequest) (*pb.UpdateDevWorkflowResponse, error) {
	if err := request.Validate(); err != nil {
		return nil, apierrors.ErrUpdateDevWorkflow.InvalidParameter(err)
	}

	wf, err := s.db.GetWf(request.ID)
	if err != nil {
		return nil, apierrors.ErrUpdateDevWorkflow.InternalError(err)
	}

	access, err := s.bdl.CheckPermission(&apistructs.PermissionCheckRequest{
		UserID:   apis.GetUserID(ctx),
		Scope:    apistructs.ProjectScope,
		ScopeID:  wf.ProjectID,
		Resource: resource,
		Action:   apistructs.OperateAction,
	})
	if err != nil {
		return nil, apierrors.ErrUpdateDevWorkflow.InternalError(err)
	}
	if !access.Access {
		return nil, apierrors.ErrUpdateDevWorkflow.AccessDenied()
	}

	wf.WorkFlows = db.JSON(request.WorkFlows)
	wf.Operator.Updater = apis.GetUserID(ctx)
	if err = s.db.UpdateWf(wf); err != nil {
		return nil, apierrors.ErrUpdateDevWorkflow.InternalError(err)
	}

	return &pb.UpdateDevWorkflowResponse{Data: wf.Convert()}, nil
}

func (s *ServiceImplement) GetDevWorkflowsByProjectID(ctx context.Context, request *pb.GetDevWorkflowRequest) (*pb.GetDevWorkflowResponse, error) {
	access, err := s.bdl.CheckPermission(&apistructs.PermissionCheckRequest{
		UserID:   apis.GetUserID(ctx),
		Scope:    apistructs.ProjectScope,
		ScopeID:  request.ProjectID,
		Resource: resource,
		Action:   apistructs.ListAction,
	})
	if err != nil {
		return nil, apierrors.ErrGetDevWorkflow.InternalError(err)
	}
	if !access.Access {
		return nil, apierrors.ErrGetDevWorkflow.AccessDenied()
	}

	wfs, err := s.db.GetWfByProjectID(request.ProjectID)
	if err != nil {
		return nil, apierrors.ErrGetDevWorkflow.InternalError(err)
	}

	return &pb.GetDevWorkflowResponse{Data: wfs.Convert()}, nil
}

func (s *ServiceImplement) DeleteDevWorkflow(ctx context.Context, request *pb.DeleteDevWorkflowRequest) (*pb.DeleteDevWorkflowResponse, error) {
	if err := s.db.DeleteWfByProjectID(request.ProjectID); err != nil {
		return nil, apierrors.ErrDeleteDevWorkflow.InternalError(err)
	}
	return &pb.DeleteDevWorkflowResponse{}, nil
}
