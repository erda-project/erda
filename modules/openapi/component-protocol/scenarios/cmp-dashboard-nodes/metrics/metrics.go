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

import (
	"context"

<<<<<<< Updated upstream:modules/openapi/component-protocol/scenarios/cmp-dashboard-nodes/metrics/metrics.go
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-proto-go/core/monitor/metric/pb"
)

type MetricsImpl struct {
	Metric pb.MetricServiceServer
	ctx context.Context
=======
	"github.com/erda-project/erda-proto-go/core/monitor/log/query/pb"
	"github.com/erda-project/erda/pkg/common"
)

type Log struct {
	ctx   context.Context
	Log   pb.LogQueryServiceServer
>>>>>>> Stashed changes:modules/cmp/log/log.go
}

type provider struct {
	Metric pb.MetricServiceServer  `autowired:"erda.core.monitor.metric.MetricService"`
	MetricsImpl *MetricsImpl
}

func (p *provider) Init(ctx servicehub.Context) error {
	p.MetricsImpl = &MetricsImpl{
		Metric: p.Metric,
		ctx: context.Background(),
	}
	return nil
}