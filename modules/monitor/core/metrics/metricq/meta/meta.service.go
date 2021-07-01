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

package meta

import (
	"context"
	"encoding/json"

	"github.com/erda-project/erda-infra/modcom/api"
	transhttp "github.com/erda-project/erda-infra/pkg/transport/http"
	"github.com/erda-project/erda-infra/providers/i18n"
	"github.com/erda-project/erda-proto-go/core/monitor/metric/pb"
	"github.com/erda-project/erda/modules/monitor/core/metrics/metricq/metricmeta"
)

type metricMetaService struct {
	p *provider
}

func getLanguage(ctx context.Context) i18n.LanguageCodes {
	req := transhttp.ContextRequest(ctx)
	if req != nil {
		return api.Language(req)
	}
	return nil
}

// ListMetricNames GET /api/metric/names
func (m *metricMetaService) ListMetricNames(ctx context.Context, req *pb.ListMetricNamesRequest) (*pb.ListMetricNamesResponse, error) {
	names, err := m.p.Metricq.MetricNames(getLanguage(ctx), req.Scope, req.ScopeID)
	if err != nil {
		return nil, err
	}
	var data []*pb.NameDefine
	for _, v := range names {
		data = append(data, &pb.NameDefine{Key: v.Key, Name: v.Name})
	}
	return &pb.ListMetricNamesResponse{Data: data}, nil
}

// ListMetricMeta GET /api/metric/meta
func (m *metricMetaService) ListMetricMeta(ctx context.Context, req *pb.ListMetricMetaRequest) (*pb.ListMetricMetaResponse, error) {
	metrics, err := m.p.Metricq.MetricMeta(getLanguage(ctx), req.Scope, req.ScopeID, req.Metrics...)
	if err != nil {
		return nil, err
	}
	data := make([]*pb.MetricMeta, len(metrics))
	b, err := json.Marshal(metrics)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(b, &data)
	if err != nil {
		return nil, err
	}

	return &pb.ListMetricMetaResponse{Data: data}, nil
}

// ListMetricGroups GET /api/metric/groups
func (m *metricMetaService) ListMetricGroups(ctx context.Context, req *pb.ListMetricGroupsRequest) (*pb.ListMetricGroupsResponse, error) {
	groups, err := m.p.Metricq.MetricGroups(getLanguage(ctx), req.Scope, req.ScopeID, req.Mode)
	if err != nil {
		return nil, err
	}
	data := make([]*pb.Group, len(groups))
	b, err := json.Marshal(groups)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(b, &data)
	if err != nil {
		return nil, err
	}
	return &pb.ListMetricGroupsResponse{Data: data}, nil
}

// GetMetricGroup GET /api/metric/groups/{id}
func (m *metricMetaService) GetMetricGroup(ctx context.Context, req *pb.GetMetricGroupRequest) (*pb.GetMetricGroupResponse, error) {
	if len(req.Format) <= 0 {
		if req.Version == "v2" {
			req.Format = metricmeta.InfluxFormat
			req.AppendTags = true
		} else if len(req.Format) <= 0 && req.Mode != "analysis" {
			// However, alarm expressions do not support dot format, so metadata queries that are not in alarm mode are all in dot format.
			req.Format = metricmeta.DotFormat
		}
	}
	group, err := m.p.Metricq.MetricGroup(getLanguage(ctx), req.Scope, req.ScopeID, req.Id, req.Mode, req.Format, req.AppendTags)
	if err != nil {
		return nil, err
	}
	var data *pb.GetMetricGroupResponse
	b, err := json.Marshal(group)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(b, &data)
	if err != nil {
		return nil, err
	}

	return data, nil
}
