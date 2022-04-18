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

package action

import (
	context "context"
	"fmt"
	"os"
	"strings"
	"time"

	"sigs.k8s.io/yaml"

	pb "github.com/erda-project/erda-proto-go/core/pipeline/action/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/providers/action/db"
	"github.com/erda-project/erda/modules/pipeline/services/apierrors"
	"github.com/erda-project/erda/pkg/common/apis"
	"github.com/erda-project/erda/pkg/crypto/uuid"
)

type actionService struct {
	p        *provider
	dbClient *db.Client
}

func (s *actionService) CheckInternalClient(ctx context.Context) error {
	if apis.GetInternalClient(ctx) == "" {
		return fmt.Errorf("not internal client req")
	}
	return nil
}

func (s *actionService) List(ctx context.Context, req *pb.PipelineActionListRequest) (*pb.PipelineActionListResponse, error) {
	if len(req.Locations) == 0 {
		return nil, apierrors.ErrListPipelineAction.InvalidParameter("locations was empty")
	}
	if req.ActionNameWithVersionQuery != nil {
		for _, query := range req.ActionNameWithVersionQuery {
			if query.Name == "" {
				return nil, apierrors.ErrListPipelineAction.InvalidParameter(fmt.Errorf("actionNameWithVersionQuery: name can not empty"))
			}
		}
	}

	actions, err := s.dbClient.ListPipelineAction(req)
	if err != nil {
		return nil, err
	}

	var data []*pb.Action
	for _, action := range actions {
		actionDto, err := action.Convert(req.YamlFormat)
		if err != nil {
			return nil, apierrors.ErrListPipelineAction.InternalError(err)
		}
		data = append(data, actionDto)
	}

	return &pb.PipelineActionListResponse{
		Data: data,
	}, nil
}

func (s *actionService) Save(ctx context.Context, req *pb.PipelineActionSaveRequest) (*pb.PipelineActionSaveResponse, error) {
	if err := s.CheckInternalClient(ctx); err != nil {
		return nil, apierrors.ErrListPipelineAction.InternalError(err)
	}

	if req.Spec == "" {
		return nil, apierrors.ErrSavePipelineAction.InvalidParameter("spec yml was empty")
	}
	if req.Dice == "" {
		return nil, apierrors.ErrSavePipelineAction.InvalidParameter("dice yml was empty")
	}
	if req.Location == "" {
		return nil, apierrors.ErrSavePipelineAction.InvalidParameter("location was empty")
	}
	if !strings.HasSuffix(req.Location, string(os.PathSeparator)) {
		return nil, apierrors.ErrSavePipelineAction.InvalidParameter(fmt.Errorf("location need %v suffix", string(os.PathSeparator)))
	}

	var spec apistructs.Spec
	err := yaml.Unmarshal([]byte(req.Spec), &spec)
	if err != nil {
		return nil, apierrors.ErrSavePipelineAction.InternalError(err)
	}

	if spec.Name == "" {
		return nil, apierrors.ErrSavePipelineAction.InvalidParameter("spec name was empty")
	}
	if spec.Version == "" {
		return nil, apierrors.ErrSavePipelineAction.InvalidParameter("spec version was empty")
	}

	actions, err := s.dbClient.ListPipelineAction(&pb.PipelineActionListRequest{
		Locations: []string{req.Location},
		ActionNameWithVersionQuery: []*pb.ActionNameWithVersionQuery{
			{
				Name:    spec.Name,
				Version: spec.Version,
			},
		},
	})
	if err != nil {
		return nil, apierrors.ErrSavePipelineAction.InternalError(err)
	}

	var saveAction *db.PipelineAction
	for _, action := range actions {
		if action.Location == req.Location {
			saveAction = &action
			break
		}
	}

	var insert = false
	if saveAction == nil {
		saveAction = &db.PipelineAction{}
		saveAction.ID = uuid.New()
		saveAction.TimeCreated = time.Now()
		insert = true
	}

	saveAction.VersionInfo = spec.Version
	saveAction.Name = spec.Name
	saveAction.Location = req.Location
	saveAction.Spec = req.Spec
	saveAction.Desc = spec.Desc
	saveAction.DisplayName = spec.DisplayName
	saveAction.IsDefault = spec.IsDefault
	saveAction.IsPublic = spec.Public
	saveAction.Dice = req.Dice
	saveAction.Readme = req.Readme
	saveAction.LogoUrl = spec.LogoUrl
	saveAction.Category = spec.Category
	saveAction.TimeUpdated = time.Now()

	if insert {
		err := s.dbClient.InsertPipelineAction(saveAction)
		if err != nil {
			return nil, apierrors.ErrSavePipelineAction.InternalError(err)
		}
	} else {
		err := s.dbClient.UpdatePipelineAction(saveAction.ID, saveAction)
		if err != nil {
			return nil, apierrors.ErrSavePipelineAction.InternalError(err)
		}
	}

	result, err := saveAction.Convert(false)
	if err != nil {
		return nil, apierrors.ErrSavePipelineAction.InternalError(err)
	}

	return &pb.PipelineActionSaveResponse{
		Action: result,
	}, nil
}

func (s *actionService) Delete(ctx context.Context, req *pb.PipelineActionDeleteRequest) (*pb.PipelineActionDeleteResponse, error) {
	if err := s.CheckInternalClient(ctx); err != nil {
		return nil, apierrors.ErrListPipelineAction.InternalError(err)
	}

	if req.Location == "" {
		return nil, apierrors.ErrDeletePipelineAction.InvalidParameter("location was empty")
	}
	if req.Name == "" {
		return nil, apierrors.ErrDeletePipelineAction.InvalidParameter("name was empty")
	}
	if req.Version == "" {
		return nil, apierrors.ErrDeletePipelineAction.InvalidParameter("version was empty")
	}

	actions, err := s.dbClient.ListPipelineAction(&pb.PipelineActionListRequest{
		Locations: []string{req.Location},
		ActionNameWithVersionQuery: []*pb.ActionNameWithVersionQuery{
			{
				Name:    req.Name,
				Version: req.Version,
			},
		},
	})
	if err != nil {
		return nil, apierrors.ErrSavePipelineAction.InternalError(err)
	}

	var deleteAction *db.PipelineAction
	for _, action := range actions {
		if action.Location == req.Location {
			deleteAction = &action
			break
		}
	}

	if deleteAction == nil {
		return nil, apierrors.ErrSavePipelineAction.InternalError(fmt.Errorf("not find action name %v version %v location %v", req.Name, req.Version, req.Location))
	}

	deleteAction.SoftDeletedAt = time.Now().UnixNano() / 1e6
	err = s.dbClient.DeletePipelineAction(deleteAction.ID, deleteAction)
	if err != nil {
		return nil, apierrors.ErrSavePipelineAction.InternalError(err)
	}

	return &pb.PipelineActionDeleteResponse{}, nil
}
