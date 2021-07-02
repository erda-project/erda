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
	"context"

	"github.com/erda-project/erda-proto-go/core/monitor/metric/pb"
	"github.com/erda-project/erda/modules/core/monitor/metric/query/metricmeta"
	"github.com/erda-project/erda/pkg/common/apis"
)

type metricMetaService struct {
	p    *provider
	meta *metricmeta.Manager
}

func (s *metricMetaService) ListMetricNames(ctx context.Context, req *pb.ListMetricNamesRequest) (*pb.ListMetricNamesResponse, error) {
	list, err := s.meta.MetricNames(apis.Language(ctx), req.Scope, req.ScopeID)
	if err != nil {
		return nil, err
	}
	return &pb.ListMetricNamesResponse{Data: list}, nil
}
func (s *metricMetaService) ListMetricMeta(ctx context.Context, req *pb.ListMetricMetaRequest) (*pb.ListMetricMetaResponse, error) {
	list, err := s.meta.MetricMeta(apis.Language(ctx), req.Scope, req.ScopeID, req.Metrics...)
	if err != nil {
		return nil, err
	}
	return &pb.ListMetricMetaResponse{Data: list}, nil
}
func (s *metricMetaService) ListMetricGroups(ctx context.Context, req *pb.ListMetricGroupsRequest) (*pb.ListMetricGroupsResponse, error) {
	list, err := s.meta.MetricGroups(apis.Language(ctx), req.Scope, req.ScopeID, req.Mode)
	if err != nil {
		return nil, err
	}
	return &pb.ListMetricGroupsResponse{Data: list}, nil
}
func (s *metricMetaService) GetMetricGroup(ctx context.Context, req *pb.GetMetricGroupRequest) (*pb.GetMetricGroupResponse, error) {
	if len(req.Format) <= 0 {
		if req.Version == "v2" {
			req.Format = metricmeta.InfluxFormat
			req.AppendTags = true
		} else if len(req.Format) <= 0 && req.Mode != "analysis" {
			// However, alarm expressions do not support dot format, so metadata queries that are not in alarm mode are all in dot format.
			req.Format = metricmeta.DotFormat
		}
	}
	list, err := s.meta.MetricGroup(apis.Language(ctx), req.Scope, req.ScopeID, req.Id, req.Mode, req.Format, req.AppendTags)
	if err != nil {
		return nil, err
	}
	return &pb.GetMetricGroupResponse{Data: list}, nil
}
