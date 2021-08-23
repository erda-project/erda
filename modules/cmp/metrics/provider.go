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

package metrics
//
//import (
//	"github.com/erda-project/erda-infra/base/servicehub"
//	"github.com/erda-project/erda-proto-go/core/monitor/metric/pb"
//	"github.com/erda-project/erda/modules/cmp/cache"
//)
//
//type provider struct {
//	server pb.MetricServiceServer `autowired:"erda.core.monitor.metric.MetricService"`
//
//	Metrics *Metric
//}
////
////func (p* provider)Run(ctx context.Context)error{
////	return nil
////}
//
//func (p *provider) Init(ctx servicehub.Context)error {
//	c, err := cache.New(1<<20, 1<<10)
//	if err != nil {
//		return err
//	}
//
//	p.Metrics = &Metric{
//		Metricq: p.server,
//		Cache:   c,
//	}
//	return nil
//}
//
//func init() {
//	servicehub.Register("cmp.metrics", &servicehub.Spec{
//		Services:    []string{"cmp.metrics"},
//		Description: "cmp metrics.",
//		Creator:     func() servicehub.Provider { return &provider{} },
//	})
//}
