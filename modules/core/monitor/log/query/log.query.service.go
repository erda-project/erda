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
		return nil, err
	}

	logs, err := p.getLogItems(r)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	return &pb.GetLogByOrganizationResponse{Lines: logs}, nil
}
