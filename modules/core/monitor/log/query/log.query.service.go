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

package query

import (
	context "context"

	pb "github.com/erda-project/erda-proto-go/core/monitor/log/query/pb"
	"github.com/erda-project/erda/pkg/common/errors"
)

type logQueryService struct {
	p *provider
}

func (s *logQueryService) GetLog(ctx context.Context, req *pb.GetLogRequest) (*pb.GetLogResponse, error) {
	r, err := convertToRequestCtx(req)
	if err != nil {
		return nil, err
	}
	err = normalizeRequest(r)
	if err != nil {
		return nil, errors.NewInvalidParameterError("", err.Error())
	}
	p := s.p

	logs, err := p.getLogItems(r)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	return &pb.GetLogResponse{Lines: logs}, nil
}

func (s *logQueryService) GetLogByRuntime(ctx context.Context, req *pb.GetLogByRuntimeRequest) (*pb.GetLogByRuntimeResponse, error) {
	r, err := convertToRequestCtx(req)
	if err != nil {
		return nil, err
	}
	err = normalizeRequest(r)
	if err != nil {
		return nil, errors.NewInvalidParameterError("", err.Error())
	}
	p := s.p

	result, err := p.checkLogMeta(r.Source, r.ID, "dice_application_id", r.ApplicationID)
	if err != nil {
		return nil, err
	} else if !result {
		return &pb.GetLogByRuntimeResponse{}, nil
	}

	logs, err := p.getLogItems(r)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	return &pb.GetLogByRuntimeResponse{Lines: logs}, nil
}

func (s *logQueryService) GetLogByOrganization(ctx context.Context, req *pb.GetLogByOrganizationRequest) (*pb.GetLogByOrganizationResponse, error) {
	r, err := convertToRequestCtx(req)
	if err != nil {
		return nil, err
	}
	err = normalizeRequest(r)
	if err != nil {
		return nil, errors.NewInvalidParameterError("", err.Error())
	}
	p := s.p

	result, err := p.checkLogMeta(r.Source, r.ID, "dice_cluster_name", r.ClusterName)
	if err != nil {
		return nil, err
	} else if !result {
		return &pb.GetLogByOrganizationResponse{}, nil
	}

	logs, err := p.getLogItems(r)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	return &pb.GetLogByOrganizationResponse{Lines: logs}, nil
}
