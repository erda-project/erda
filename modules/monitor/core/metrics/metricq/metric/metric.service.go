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

package metric

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"strconv"

	"google.golang.org/protobuf/types/known/anypb"

	"github.com/erda-project/erda-infra/modcom/api"
	"github.com/erda-project/erda-infra/pkg/protobuf/goany"
	transhttp "github.com/erda-project/erda-infra/pkg/transport/http"
	"github.com/erda-project/erda-infra/providers/httpserver"
	"github.com/erda-project/erda-infra/providers/i18n"
	"github.com/erda-project/erda-proto-go/core/monitor/metric/pb"
)

func getLanguage(ctx context.Context) i18n.LanguageCodes {
	req := transhttp.ContextRequest(ctx)
	if req != nil {
		return api.Language(req)
	}
	return nil
}

// QueryWithInfluxFormat POST query
func (m *metricService) QueryWithInfluxFormat(ctx context.Context, req *pb.QueryWithInfluxFormatRequest) (*pb.QueryWithInfluxFormatResponse, error) {
	return m.metricsInfluxQuery(ctx, req)
}

// SearchWithInfluxFormat GET query
func (m *metricService) SearchWithInfluxFormat(ctx context.Context, req *pb.QueryWithInfluxFormatRequest) (*pb.QueryWithInfluxFormatResponse, error) {
	return m.metricsInfluxQuery(ctx, req)
}

// QueryWithTableFormat POST api/query
func (m *metricService) QueryWithTableFormat(ctx context.Context, req *pb.QueryWithTableFormatRequest) (*pb.QueryWithTableFormatResponse, error) {
	return m.metricsTableQuery(ctx, req)
}

// SearchWithTableFormat GET api/query
func (m *metricService) SearchWithTableFormat(ctx context.Context, req *pb.QueryWithTableFormatRequest) (*pb.QueryWithTableFormatResponse, error) {
	return m.metricsTableQuery(ctx, req)
}

func (m *metricService) metricsInfluxQuery(ctx context.Context, req *pb.QueryWithInfluxFormatRequest) (*pb.QueryWithInfluxFormatResponse, error) {
	ql, q, format := "influxql", req.Statement, "influxdb"
	if len(q) == 0 {
		q = req.Options["body"]
	}
	params := make(map[string][]string)
	for k, v := range req.Options {
		params[k] = append(params[k], v)
	}
	delete(params, "ql")
	delete(params, "q")
	delete(params, "format")
	_, result, err := m.metricq.QueryWithFormat(ql, q, format, getLanguage(ctx), nil, params)
	if err != nil {
		return nil, err
	}
	var data pb.QueryWithInfluxFormatResponse

	if response, ok := result.(httpserver.Response); ok {
		b, err := ioutil.ReadAll(response.ReadCloser(nil))
		if err != nil {
			return nil, err
		}
		var influxqlRespone InfluxqlRespone
		err = json.Unmarshal(b, &influxqlRespone)
		if err != nil {
			return nil, err
		}
		var results []*pb.Result
		for _, r := range influxqlRespone.Result {
			var result pb.Result
			result.StatementId = strconv.Itoa(r.StatementId)
			var arrSeries []*pb.Serie
			for _, s := range r.Series {
				var series pb.Serie
				series.Name = s.Name
				series.Columns = s.Columns
				var arrValues []*pb.Row
				for _, v := range s.Values {
					var arrAny []*anypb.Any
					for _, k := range v {
						any, err := goany.Marshal(k)
						if err != nil {
							return nil, err
						}
						arrAny = append(arrAny, any)
					}

					arrValues = append(arrValues, &pb.Row{Values: arrAny})
				}
				series.Rows = arrValues
				arrSeries = append(arrSeries, &series)
			}
			result.Series = arrSeries
			results = append(results, &result)
		}
		data = pb.QueryWithInfluxFormatResponse{Results: results}
	}
	return &data, nil
}

func (m *metricService) metricsTableQuery(ctx context.Context, req *pb.QueryWithTableFormatRequest) (*pb.QueryWithTableFormatResponse, error) {
	ql, q, format := "influxql", req.Options["q"], req.Options["format"]
	if len(q) == 0 {
		q = req.Options["body"]
	}
	if len(format) == 0 {
		format = "chartv2"
	}
	params := make(map[string][]string)
	for k, v := range req.Options {
		params[k] = append(params[k], v)
	}
	delete(params, "ql")
	delete(params, "q")
	delete(params, "format")
	_, result, err := m.metricq.QueryWithFormat(ql, q, format, getLanguage(ctx), nil, params)
	if err != nil {
		return nil, err
	}
	var data pb.QueryWithTableFormatResponse
	tableResponse := new(TableResponse)
	b, err := json.Marshal(&result)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(b, &tableResponse)
	if err != nil {
		return nil, err
	}
	var tableResult pb.TableResult
	tableResult.Interval = tableResponse.Interval
	var tableRows []*pb.TableRow
	for _, row := range tableResponse.Data {
		var tableRow pb.TableRow
		values := make(map[string]*anypb.Any)
		for k, v := range row {
			any, err := goany.Marshal(&v)
			if err != nil {
				return nil, err
			}
			values[k] = any
		}
		tableRow.Values = values
		tableRows = append(tableRows, &tableRow)
	}
	tableResult.Data = tableRows
	tableResult.Cols = tableResponse.Cols
	data.Data = &tableResult
	return &data, nil
}
