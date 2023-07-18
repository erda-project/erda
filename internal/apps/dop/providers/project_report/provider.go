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

package project_report

import (
	"context"
	"time"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	election "github.com/erda-project/erda-infra/providers/etcd-election"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/internal/apps/dop/providers/issue/core/query"
	"github.com/erda-project/erda/internal/core/org"
	"github.com/erda-project/erda/internal/pkg/metrics/report"
)

type config struct {
	ProjectReportMetricDuration time.Duration `file:"project_report_metric_duration" env:"PROJECT_REPORT_METRIC_DURATION" default:"2h"`
}

// +provider
type provider struct {
	Cfg *config
	Log logs.Logger
	bdl *bundle.Bundle

	ReportClient report.MetricReport `autowired:"metric-report-client"`
	Org          org.Interface
	IssueSvc     query.Interface
	Election     election.Interface `autowired:"etcd-election@project-management-report"`
}

func (p *provider) Init(ctx servicehub.Context) error {
	p.bdl = bundle.New(bundle.WithErdaServer())
	return nil
}

func (p *provider) Run(ctx context.Context) error {
	p.Election.OnLeader(p.doReportMetricTask)
	return nil
}

func init() {
	servicehub.Register("project-report", &servicehub.Spec{
		Services:     []string{"project-report"},
		Dependencies: []string{"metric-report-client"},
		Description:  "project delivery management report",
		ConfigFunc:   func() interface{} { return &config{} },
		Creator:      func() servicehub.Provider { return &provider{} },
	})
}
