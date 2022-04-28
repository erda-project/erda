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

package rbworkflow

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/erda-project/erda-proto-go/dop/rbworkflow/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/dop/providers/rbworkflow/db"
	"github.com/erda-project/erda/modules/dop/services/apierrors"
	"github.com/erda-project/erda/pkg/common/apis"
)

const resource = "rbWorkflow"

type Service interface {
	CreateRbWorkflow(context.Context, *pb.CreateRbWorkflowRequest) (*pb.CreateRbWorkflowResponse, error)
	UpdateRbWorkflow(context.Context, *pb.UpdateRbWorkflowRequest) (*pb.UpdateRbWorkflowResponse, error)
	DeleteRbWorkflow(context.Context, *pb.DeleteRbWorkflowRequest) (*pb.DeleteRbWorkflowResponse, error)
	ListRbWorkflows(context.Context, *pb.ListRbWorkflowRequest) (*pb.ListRbWorkflowResponse, error)
}

type ServiceImplement struct {
	db  *db.Client
	bdl *bundle.Bundle
}

func (s *ServiceImplement) CreateRbWorkflow(ctx context.Context, request *pb.CreateRbWorkflowRequest) (*pb.CreateRbWorkflowResponse, error) {
	if err := request.Validate(); err != nil {
		return nil, apierrors.ErrCreateRbWorkflow.InvalidParameter(err)
	}
	access, err := s.bdl.CheckPermission(&apistructs.PermissionCheckRequest{
		UserID:   apis.GetUserID(ctx),
		Scope:    apistructs.ProjectScope,
		ScopeID:  request.ProjectID,
		Resource: resource,
		Action:   apistructs.OperateAction,
	})
	if err != nil {
		return nil, apierrors.ErrCreateRbWorkflow.InternalError(err)
	}
	if !access.Access {
		return nil, apierrors.ErrCreateRbWorkflow.AccessDenied()
	}
	project, err := s.bdl.GetProject(request.ProjectID)
	if err != nil {
		return nil, apierrors.ErrCreateRbWorkflow.InternalError(err)
	}
	org, err := s.bdl.GetOrg(project.OrgID)
	if err != nil {
		return nil, apierrors.ErrCreateRbWorkflow.InternalError(err)
	}
	// Some check
	if err = s.checkInProjectWithCreate(request); err != nil {
		return nil, apierrors.ErrCreateRbWorkflow.InternalError(err)
	}

	wf := db.RBWorkflow{
		Scope: db.Scope{
			OrgID:       org.ID,
			OrgName:     org.Name,
			ProjectID:   project.ID,
			ProjectName: project.Name,
		},
		Operator: db.Operator{
			Creator: apis.GetUserID(ctx),
			Updater: apis.GetUserID(ctx),
		},
		Stage:       request.Stage,
		Sort:        request.Sort,
		Branch:      request.Branch,
		Artifact:    request.Artifact,
		Environment: request.Environment,
		SubFlows:    request.SubFlows,
	}
	err = s.db.CreateWf(&wf)
	if err != nil {
		return nil, apierrors.ErrCreateRbWorkflow.InternalError(err)
	}
	return &pb.CreateRbWorkflowResponse{Data: wf.Convert()}, nil
}

func (s *ServiceImplement) checkInProjectWithCreate(request *pb.CreateRbWorkflowRequest) error {
	// The same sort check
	exist, err := s.checkSameSortInProject(request.ProjectID, request.Sort, "")
	if err != nil {
		return err
	}
	if exist {
		return fmt.Errorf("the same sort is exist")
	}

	// The same stage check
	exist, err = s.checkSameStageInProject(request.ProjectID, request.Stage, "")
	if err != nil {
		return err
	}
	if exist {
		return fmt.Errorf("the same stage is exist")
	}

	return nil
}

func (s *ServiceImplement) checkSameSortInProject(projectID, sort uint64, id string) (bool, error) {
	wf, err := s.db.GetWfByProjectIDAndSort(projectID, sort)
	if err == nil && wf.ID.String != id {
		return true, nil
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return false, err
	}
	return false, nil
}

func (s *ServiceImplement) checkSameStageInProject(projectID uint64, stage, id string) (bool, error) {
	wf, err := s.db.GetWfByProjectIDAndStage(projectID, stage)
	if err == nil && wf.ID.String != id {
		return true, nil
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return false, err
	}
	return false, nil
}

func (s *ServiceImplement) UpdateRbWorkflow(ctx context.Context, request *pb.UpdateRbWorkflowRequest) (*pb.UpdateRbWorkflowResponse, error) {
	if err := request.Validate(); err != nil {
		return nil, apierrors.ErrUpdateRbWorkflow.InvalidParameter(err)
	}

	wf, err := s.db.GetWf(request.ID)
	if err != nil {
		return nil, apierrors.ErrUpdateRbWorkflow.InternalError(err)
	}

	access, err := s.bdl.CheckPermission(&apistructs.PermissionCheckRequest{
		UserID:   apis.GetUserID(ctx),
		Scope:    apistructs.ProjectScope,
		ScopeID:  wf.ProjectID,
		Resource: resource,
		Action:   apistructs.OperateAction,
	})
	if err != nil {
		return nil, apierrors.ErrUpdateRbWorkflow.InternalError(err)
	}
	if !access.Access {
		return nil, apierrors.ErrUpdateRbWorkflow.AccessDenied()
	}

	if err = s.checkInProjectWithUpdate(request, wf.ProjectID); err != nil {
		return nil, apierrors.ErrUpdateRbWorkflow.InternalError(err)
	}

	wf.Stage = request.Stage
	wf.Sort = request.Sort
	wf.Branch = request.Branch
	wf.Artifact = request.Artifact
	wf.Environment = request.Environment
	wf.SubFlows = request.SubFlows
	wf.Operator.Updater = apis.GetUserID(ctx)
	if err = s.db.UpdateWf(wf); err != nil {
		return nil, apierrors.ErrUpdateRbWorkflow.InternalError(err)
	}

	return &pb.UpdateRbWorkflowResponse{Data: wf.Convert()}, nil
}

func (s *ServiceImplement) checkInProjectWithUpdate(request *pb.UpdateRbWorkflowRequest, projectID uint64) error {
	// The same sort check
	exist, err := s.checkSameSortInProject(projectID, request.Sort, request.ID)
	if err != nil {
		return err
	}
	if exist {
		return fmt.Errorf("the same sort is exist")
	}

	// The same stage check
	exist, err = s.checkSameStageInProject(projectID, request.Stage, request.ID)
	if err != nil {
		return err
	}
	if exist {
		return fmt.Errorf("the same stage is exist")
	}

	return nil
}

func (s *ServiceImplement) DeleteRbWorkflow(ctx context.Context, request *pb.DeleteRbWorkflowRequest) (*pb.DeleteRbWorkflowResponse, error) {
	wf, err := s.db.GetWf(request.ID)
	if err != nil {
		return nil, apierrors.ErrDeleteRbWorkflow.InternalError(err)
	}

	access, err := s.bdl.CheckPermission(&apistructs.PermissionCheckRequest{
		UserID:   apis.GetUserID(ctx),
		Scope:    apistructs.ProjectScope,
		ScopeID:  wf.ProjectID,
		Resource: resource,
		Action:   apistructs.OperateAction,
	})
	if err != nil {
		return nil, apierrors.ErrDeleteRbWorkflow.InternalError(err)
	}
	if !access.Access {
		return nil, apierrors.ErrDeleteRbWorkflow.AccessDenied()
	}
	if err = s.db.DeleteWf(wf); err != nil {
		return nil, apierrors.ErrDeleteRbWorkflow.InternalError(err)
	}
	return &pb.DeleteRbWorkflowResponse{}, nil
}

func (s *ServiceImplement) ListRbWorkflows(ctx context.Context, request *pb.ListRbWorkflowRequest) (*pb.ListRbWorkflowResponse, error) {
	_, err := s.bdl.GetProject(request.ProjectID)
	if err != nil {
		return nil, apierrors.ErrListRbWorkflow.InternalError(err)
	}
	wfs, err := s.db.ListWfByProjectID(request.ProjectID)
	if err != nil {
		return nil, apierrors.ErrListRbWorkflow.InternalError(err)
	}

	data := make([]*pb.RbWorkflow, 0, len(wfs))
	for i := range wfs {
		data = append(data, wfs[i].Convert())
	}
	return &pb.ListRbWorkflowResponse{Data: data}, nil
}
