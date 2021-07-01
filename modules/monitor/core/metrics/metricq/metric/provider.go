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
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/pkg/transport"
	transhttp "github.com/erda-project/erda-infra/pkg/transport/http"
	"github.com/erda-project/erda-infra/pkg/transport/http/encoding"
	"github.com/erda-project/erda-proto-go/core/monitor/metric/pb"
	"github.com/erda-project/erda/modules/monitor/core/metrics/metricq"
	"github.com/erda-project/erda/pkg/common/apis"
)

type config struct {
}

// +provider
type provider struct {
	Cfg           *config
	Log           logs.Logger
	Register      transport.Register `autowired:"service-register" optional:"true"`
	Metricq       metricq.Queryer    `autowired:"metrics-query"`
	metricService *metricService
}

type metricService struct {
	metricq metricq.Queryer
}

func (p *provider) Init(ctx servicehub.Context) error {
	p.metricService = &metricService{
		metricq: p.Metricq,
	}
	if p.Register != nil {
		pb.RegisterMetricServiceImp(p.Register, p.metricService, apis.Options(),
			transport.WithHTTPOptions(
				transhttp.WithDecoder(func(r *http.Request, data interface{}) error {
					option := make(map[string]string)
					if err := r.ParseForm(); err != nil {
						return err
					}
					b, err := ioutil.ReadAll(r.Body)
					if err != nil {
						return err
					}
					for k, v := range r.Form {
						option[k] = v[0]
					}
					option["body"] = string(b)
					if resp, ok := data.(*pb.QueryWithInfluxFormatRequest); ok && resp != nil {
						resp.Options = option
						data = resp
					}
					if resp, ok := data.(*pb.QueryWithTableFormatRequest); ok && resp != nil {
						resp.Options = option
						data = resp
					}
					return encoding.DecodeRequest(r, data)
				}),
				transhttp.WithEncoder(func(rw http.ResponseWriter, r *http.Request, data interface{}) error {
					if resp, ok := data.(*apis.Response); ok && resp != nil {
						if r, ok := resp.Data.(*pb.QueryWithInfluxFormatResponse); ok {
							var influxqlRespone InfluxqlRespone
							var results []*Results
							for _, v := range r.Results {
								var result Results
								i, err := strconv.Atoi(v.StatementId)
								if err != nil {
									return err
								}
								result.StatementId = i
								var series []*Series
								for _, s := range v.Series {
									var serie Series
									serie.Name = s.Name
									serie.Columns = s.Columns
									for _, row := range s.Rows {
										var interfaces []interface{}
										for _, rw := range row.Values {
											interfaces = append(interfaces, rw.AsInterface())
										}
										serie.Values = append(serie.Values, interfaces)
									}
									series = append(series, &serie)
								}
								result.Series = series
								results = append(results, &result)
							}
							influxqlRespone.Result = results
							data = influxqlRespone
						}
						if r, ok := resp.Data.(*pb.TableResult); ok {
							var tableResponse TableResponse
							var tableRows []map[string]interface{}
							tableResponse.Interval = r.Interval
							tableResponse.Cols = r.Cols
							for _, v := range r.Data {
								tableRow := make(map[string]interface{})
								for k, value := range v.Values {
									tableRow[k] = value.AsInterface()
								}
								tableRows = append(tableRows, tableRow)
							}
							tableResponse.Data = tableRows
							data = tableResponse
						}
					}
					return encoding.EncodeResponse(rw, r, data)
				})))
	}
	return nil
}

func (p *provider) Provide(ctx servicehub.DependencyContext, args ...interface{}) interface{} {
	if ctx.Service() == "erda.core.monitor.metric.MetricService" || ctx.Type() == pb.MetricServiceServerType() || ctx.Type() == pb.MetricServiceHandlerType() {
		return p.metricService
	}
	return p
}

func init() {
	servicehub.Register("erda.core.monitor.metric", &servicehub.Spec{
		Services:             pb.ServiceNames(),
		Types:                pb.Types(),
		OptionalDependencies: []string{"service-register"},
		Description:          "",
		ConfigFunc: func() interface{} {
			return &config{}
		},
		Creator: func() servicehub.Provider {
			return &provider{}
		},
	})
}
