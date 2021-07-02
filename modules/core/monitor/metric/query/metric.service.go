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
	"fmt"
	"reflect"

	"github.com/erda-project/erda-infra/modcom/api"
	"github.com/erda-project/erda-proto-go/core/monitor/metric/pb"
	"github.com/erda-project/erda/modules/core/monitor/metric/query/es-tsql/formats/influxdb"
	"github.com/erda-project/erda/modules/core/monitor/metric/query/query"
	"github.com/erda-project/erda/pkg/common/errors"
)

type metricService struct {
	p     *provider
	query query.Queryer
}

func (s *metricService) QueryWithInfluxFormat(ctx context.Context, req *pb.QueryWithInfluxFormatRequest) (*pb.QueryWithInfluxFormatResponse, error) {
	if len(req.Statement) <= 0 {
		return &pb.QueryWithInfluxFormatResponse{}, nil
	}
	rs, data, err := s.query.QueryWithFormat("influxql", req.Statement, "influxdb", nil, convertParams(req.Params), convertFilters(req.Filters), convertOptions(req.Start, req.End, req.Options))
	if err != nil {
		return nil, errors.NewServiceInvokingError("metric.query", err)
	}
	if rs.Details != nil {
		fmt.Println(rs.Details)
	}
	if data == nil {
		return &pb.QueryWithInfluxFormatResponse{}, nil
	}
	rawResp, ok := data.(*api.RawResponse)
	if !ok {
		return nil, errors.NewInternalServerError(fmt.Errorf("%s not *api.RawResponse", reflect.TypeOf(data)))
	}
	resp, ok := rawResp.Body().(*influxdb.Response)
	if !ok {
		return nil, errors.NewInternalServerError(fmt.Errorf("%s is not *influxdb.Response", reflect.TypeOf(rawResp.Body())))
	}
	if resp.Error != nil {
		return nil, errors.NewInternalServerError(resp.Error)
	}
	return &pb.QueryWithInfluxFormatResponse{Results: resp.Results}, nil
}

func (s *metricService) SearchWithInfluxFormat(ctx context.Context, req *pb.QueryWithInfluxFormatRequest) (*pb.QueryWithInfluxFormatResponse, error) {
	return s.QueryWithInfluxFormat(ctx, req)
}

func (s *metricService) QueryWithTableFormat(ctx context.Context, req *pb.QueryWithTableFormatRequest) (*pb.QueryWithTableFormatResponse, error) {
	if len(req.Statement) <= 0 {
		return &pb.QueryWithTableFormatResponse{Data: &pb.TableResult{}}, nil
	}
	opts := convertOptions(req.Start, req.End, req.Options)
	opts.Set("type", "_")
	rs, data, err := s.query.QueryWithFormat("influxql", req.Statement, "chartv2", nil, convertParams(req.Params), convertFilters(req.Filters), opts)
	if err != nil {
		return nil, errors.NewServiceInvokingError("metric.query", err)
	}
	if rs.Details != nil {
		fmt.Println(rs.Details)
	}
	if data == nil {
		return &pb.QueryWithTableFormatResponse{Data: &pb.TableResult{}}, nil
	}
	result, ok := data.(*pb.TableResult)
	if !ok {
		return nil, errors.NewInternalServerError(fmt.Errorf("%s is not *pb.TableResult", reflect.TypeOf(data)))
	}
	return &pb.QueryWithTableFormatResponse{Data: result}, nil
}
func (s *metricService) SearchWithTableFormat(ctx context.Context, req *pb.QueryWithTableFormatRequest) (*pb.QueryWithTableFormatResponse, error) {
	return s.QueryWithTableFormat(ctx, req)
}
