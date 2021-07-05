// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package resource

import (
	"context"
	"github.com/erda-project/erda-proto-go/msp/resource/pb"
	"github.com/erda-project/erda/modules/msp/resource/deploy/coordinator"
	"github.com/erda-project/erda/modules/msp/resource/deploy/handlers"
)

type resourceService struct {
	p           *provider
	coordinator coordinator.Interface
}

func (s *resourceService) CreateResource(ctx context.Context, req *pb.CreateResourceRequest) (*pb.CreateResourceResponse, error) {
	deployReq := handlers.ResourceDeployRequest{
		Az:          req.Az,
		Uuid:        req.Uuid,
		Plan:        req.Plan,
		Engine:      req.Engine,
		Callback:    req.Callback,
		Options:     req.Options,
		TenantGroup: req.Options["tenantGroup"],
	}

	result, err := s.coordinator.Deploy(deployReq)

	if err != nil {
		return nil, err
	}

	return &pb.CreateResourceResponse{
		Data: &pb.ResourceCreateResult{
			Id:        result.ID,
			Config:    result.Config,
			Status:    result.Status,
			Label:     map[string]string{},
			UpdateAt:  result.UpdatedTime.UTC().Format("2006-01-02T15:04:05Z"),
			CreatedAt: result.CreatedTime.UTC().Format("2006-01-02T15:04:05Z"),
		},
	}, nil
}

func (s *resourceService) DeleteResource(ctx context.Context, request *pb.DeleteResourceRequest) (*pb.DeleteResourceResponse, error) {
	err := s.coordinator.UnDeploy(request.Id)

	if err != nil {
		return nil, err
	}

	return &pb.DeleteResourceResponse{Data: true}, nil
}
