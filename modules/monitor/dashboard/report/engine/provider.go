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

package engine

import (
	"context"
	"fmt"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda/pkg/discover"
	"github.com/go-playground/validator"
)

type define struct{}

func (d *define) Services() []string     { return []string{"report-engine"} }
func (d *define) Dependencies() []string { return []string{} }
func (d *define) Summary() string        { return "report engine" }
func (d *define) Description() string    { return d.Summary() }
func (d *define) Config() interface{} {
	return &config{
		MonitorAddr:  discover.Monitor(),
		EventboxAddr: discover.EventBox(),
	}
}
func (d *define) Creator() servicehub.Creator {
	return func() servicehub.Provider {
		return &provider{}
	}
}

type config struct {
	MonitorAddr  string `file:"monitor_addr"`
	EventboxAddr string `file:"eventbox_addr"`
	DomainAddr   string `file:"domain_addr" env:"ACTION_DOMAIN_ADDR" validate:"required"`
	ReportID     string `file:"report_id" env:"ACTION_REPORT_ID" validate:"required"`
	OrgName      string `file:"org_name" env:"ACTION_ORG_NAME" validate:"required"`
}

type provider struct {
	Cfg    *config
	Log    logs.Logger
	report *Report
}

func (p *provider) Init(ctx servicehub.Context) error {
	valid := validator.New()
	err := valid.Struct(p.Cfg)
	if err != nil {
		return fmt.Errorf("invalid config: %s", err)
	}
	p.report = New(p.Cfg)
	return nil
}

// Start .
func (p *provider) Start() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	p.Log.Info("reportengine start running ...")
	if err := p.report.Run(ctx); err != nil {
		return fmt.Errorf("reportengine failed: %s", err)
	}
	p.Log.Info("reportengine done successfully ...")
	return nil
}

func (p *provider) Close() error {
	p.Log.Debug("not support close report engine")
	return nil
}

func init() {
	servicehub.RegisterProvider("report-engine", &define{})
}
