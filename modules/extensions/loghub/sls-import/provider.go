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

package slsimport

import (
	"fmt"
	"sync"
	"time"

	"github.com/olivere/elastic"
	"github.com/recallsong/go-utils/errorx"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	writer "github.com/erda-project/erda-infra/pkg/parallel-writer"
	"github.com/erda-project/erda-infra/providers/elasticsearch"
	"github.com/erda-project/erda-infra/providers/kafka"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/pkg/httpclient"
)

type define struct{}

func (d *define) Service() []string      { return []string{"sls-import"} }
func (d *define) Dependencies() []string { return []string{"kafka", "elasticsearch"} }
func (d *define) Summary() string        { return "import logs from aliyun sls" }
func (d *define) Description() string    { return d.Summary() }
func (d *define) Config() interface{}    { return &config{} }
func (d *define) Creator() servicehub.Creator {
	return func() servicehub.Provider {
		return &provider{}
	}
}

type config struct {
	AccountsReloadInterval time.Duration `file:"accounts_reload_interval"`
	ProjectsReloadInterval time.Duration `file:"projects_reload_interval"`
	Projects               []string      `file:"projects"`
	LogFilters             []struct {
		Product string                 `file:"product"`
		Options map[string]interface{} `file:"options"`
	} `file:"log_filters"`
	Output struct {
		Elasticsearch struct {
			elasticsearch.WriterConfig `file:"writer_config"`
			IndexPrefix                string        `file:"index_prefix" default:"sls-"`
			IndexTemplateName          string        `file:"index_template_name" default:"spot_metric_template"`
			IndexCleanInterval         time.Duration `file:"index_clean_interval" default:"1h"`
			IndexTTL                   time.Duration `file:"index_ttl" default:"720h"`
			RequestTimeout             time.Duration `file:"request_time" default:"60s"`
		} `file:"elasticsearch"`
		Kafka kafka.ProducerConfig `file:"kafka"`
	} `file:"output"`
	Account struct {
		OrgID           string `file:"org_id"`
		OrgName         string `file:"org_name"`
		AccessKey       string `file:"ali_access_key"`
		AccessSecretKey string `file:"ali_access_secret_key"`
	} `file:"account"`
}

type provider struct {
	C         *config
	L         logs.Logger
	importers map[string]*Importer
	lock      sync.RWMutex
	closeCh   chan struct{}
	wg        sync.WaitGroup
	filters   Filters
	outputs   outputs
	bdl       *bundle.Bundle
	es        *elastic.Client
}

type outputs struct {
	es          writer.Writer
	indexPrefix string
	kafka       writer.Writer
}

func (p *provider) Init(ctx servicehub.Context) error {
	hc := httpclient.New(httpclient.WithTimeout(time.Second, time.Second*60))
	p.bdl = bundle.New(
		bundle.WithHTTPClient(hc),
		bundle.WithCMDB(),
		bundle.WithOps(),
	)
	filters, err := buildFilters(p.C.Projects)
	if err != nil {
		return err
	}
	p.filters = filters
	p.importers = make(map[string]*Importer)
	p.closeCh = make(chan struct{})

	es := ctx.Service("elasticsearch").(elasticsearch.Interface)
	err = p.initIndexTemplate(es.Client())
	if err != nil {
		return err
	}
	p.es = es.Client()
	p.outputs.es = es.NewBatchWriter(&p.C.Output.Elasticsearch.WriterConfig)
	p.outputs.indexPrefix = p.C.Output.Elasticsearch.IndexPrefix

	k, err := ctx.Service("kafka").(kafka.Interface).NewProducer(&p.C.Output.Kafka)
	if err != nil {
		return fmt.Errorf("fail to create kafka producer: %s", err)
	}
	p.outputs.kafka = k

	for _, pro := range p.C.LogFilters {
		initLogFilter(pro.Product, pro.Options)
	}
	return nil
}

// Start .
func (p *provider) Start() error {
	p.wg.Add(1)
	go p.startIndexManager()
	tick := time.Tick(p.C.AccountsReloadInterval)
	for {
		p.loadAll()
		select {
		case <-tick:
			continue
		case <-p.closeCh:
			return nil
		}
	}
}

func (p *provider) Close() error {
	close(p.closeCh)
	var errs errorx.Errors
	for _, item := range p.importers {
		err := item.Close()
		if err != nil {
			errs = append(errs, err)
		}
	}
	p.wg.Wait()
	return errs.MaybeUnwrap()
}

func init() {
	servicehub.RegisterProvider("sls-import", &define{})
}
