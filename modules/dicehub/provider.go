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

const service = "dicehub"

type provider struct {
	Log         logs.Logger
	QueryClient query.MetricQuery
}

func init() { servicehub.RegisterProvider(service, &provider{}) }

func (p *provider) Service() []string      { return []string{service} }
func (p *provider) Dependencies() []string { return []string{"metricq-client"} }

func (p *provider) Init(ctx servicehub.Context) error {
	metrics.Client = p.QueryClient
	return nil
}

func (p *provider) Run(ctx context.Context) error { return Initialize(p) }
func (p *provider) Creator() servicehub.Creator {
	return func() servicehub.Provider { return &provider{} }
}
