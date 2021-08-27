// Copyright (c) 2021 Terminus, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package cmp Core components of multi-cloud management platform
package cmp

import (
	"context"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-proto-go/core/monitor/metric/pb"
	"github.com/erda-project/erda/modules/cmp/metrics"
)

type provider struct {
	Server pb.MetricServiceServer `autowired:"erda.core.monitor.metric.MetricService"`

	Metrics *metrics.Metric
}

// Run Run the provider
func (p *provider) Run(ctx context.Context) error {
	newCtx := context.WithValue(ctx, "metrics", p.Metrics)
	logrus.Info("cmp provider is running...")
	return initialize(newCtx)
}

func (p *provider) Init(ctx servicehub.Context) error {
	p.Metrics = &metrics.Metric{
		Metricq: p.Server,
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
