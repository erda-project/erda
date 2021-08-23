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

// Package cmp Core components of multi-cloud management platform
package cmp

import (
	"context"
	"fmt"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-proto-go/core/monitor/metric/pb"
	"github.com/erda-project/erda/modules/cmp/cache"
	"github.com/erda-project/erda/modules/cmp/metrics"
)

type provider struct {
	Server pb.MetricServiceServer `autowired:"erda.core.monitor.metric.MetricService"`

	Metrics *metrics.Metric
}

// Run Run the provider
func (p *provider) Run(ctx context.Context) error {
	fmt.Println("isisisiis",p.Server == nil)
	newCtx := context.WithValue(ctx,"metrics",p.Metrics)
	logrus.Info("cmp provider is running...")
	return initialize(newCtx)
}

func (p *provider) Init(ctx servicehub.Context)error {
	fmt.Println("init provider",p.Server == nil)
	c, err := cache.New(1<<20, 1<<10)
	if err != nil {
		return err
	}

	p.Metrics = &metrics.Metric{
		Metricq: p.Server,
		Cache:   c,
	}
	return nil
}

func init() {
	servicehub.Register("cmp", &servicehub.Spec{
		Services:    []string{"cmp"},
		Description: "Core components of multi-cloud management platform.",
		Creator:     func() servicehub.Provider { return &provider{} },
	})
}
