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

package dicehub

import (
	"context"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda/modules/dicehub/metrics"
	"github.com/erda-project/erda/providers/metrics/query"
)

type provider struct {
	Log         logs.Logger
	QueryClient query.MetricQuery `autowired:"metricq-client"`
}

func (p *provider) Init(ctx servicehub.Context) error {
	metrics.Client = p.QueryClient
	return nil
}

func (p *provider) Run(ctx context.Context) error { return Initialize(p) }

func init() {
	servicehub.Register("dicehub", &servicehub.Spec{
		Services: []string{"dicehub"},
		Creator:  func() servicehub.Provider { return &provider{} },
	})
}
