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

package manager

import (
	"sync/atomic"
	"time"

	"github.com/olivere/elastic"
	"github.com/robfig/cron"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/elasticsearch"
	"github.com/erda-project/erda-infra/providers/httpserver"
)

type define struct{}

func (d *define) Service() []string      { return []string{"logs-index-manager"} }
func (d *define) Dependencies() []string { return []string{"elasticsearch@logs", "http-server"} }
func (d *define) Summary() string        { return "logs index manager" }
func (d *define) Description() string    { return d.Summary() }
func (d *define) Config() interface{}    { return &config{} }
func (d *define) Creator() servicehub.Creator {
	return func() servicehub.Provider {
		return &provider{
			timeRanges: make(map[string]*timeRange),
			reload:     make(chan struct{}),
		}
	}
}

type config struct {
	IndexPrefix       string `file:"index_prefix" default:"rlogs-"`
	IndexTemplateName string `file:"index_template_name"`
	IndexTemplateFile string `file:"index_template_file"`

	// reload
	RequestTimeout time.Duration `file:"request_timeout" default:"120s"`
	ReloadInterval time.Duration `file:"reload_interval" default:"2m"`

	// clean
	IndexTTL           time.Duration `file:"index_ttl" default:"168h"`
	EnableIndexClean   bool          `file:"enable_index_clean"`
	IndexCheckInterval string        `file:"index_check_interval"`

	// rollover
	EnableIndexRollover bool   `file:"enable_index_rollover"`
	RolloverBodyFile    string `file:"rollover_body_file"`
	RolloverInterval    string `file:"rollover_interval"`
}

type provider struct {
	C      *config
	L      logs.Logger
	client *elastic.Client
	cron   *cron.Cron

	rolloverBody string

	reload     chan struct{}
	indices    atomic.Value
	timeRanges map[string]*timeRange
}

func (p *provider) Init(ctx servicehub.Context) error {
	es := ctx.Service("elasticsearch@logs").(elasticsearch.Interface)
	p.client = es.Client()
	err := p.setupIndexTemplate(p.client)
	if err != nil {
		return err
	}
	err = p.loadRolloverBody()
	if err != nil {
		return err
	}
	p.cron = cron.New()
	routes := ctx.Service("http-server").(httpserver.Router)
	return p.intRoutes(routes)
}

// Start .
func (p *provider) Start() error {
	// load
	go func() {
		tick := time.Tick(p.C.ReloadInterval)
		for {
			p.reloadAllIndices()
			select {
			case <-tick:
			case _, ok := <-p.reload:
				if !ok {
					return
				}
			}
		}
	}()

	// clean
	if p.C.EnableIndexClean && len(p.C.IndexCheckInterval) > 0 {
		p.cleanIndices()
		p.cron.AddFunc(p.C.IndexCheckInterval, p.cleanIndices)
	}

	// rollover
	if p.C.EnableIndexRollover && len(p.C.RolloverInterval) > 0 && len(p.rolloverBody) > 0 {
		p.doRolloverAlias()
		p.cron.AddFunc(p.C.RolloverInterval, p.doRolloverAlias)
	}
	p.cron.Run()
	return nil
}

func (p *provider) Close() error {
	p.cron.Stop()
	close(p.reload)
	return nil
}

func init() {
	servicehub.RegisterProvider("logs-index-manager", &define{})
}
