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

// Author: recallsong
// Email: ruiguo.srg@alibaba-inc.com

package elasticsearch

import (
	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	writer "github.com/erda-project/erda-infra/pkg/parallel-writer"
	"github.com/erda-project/erda-infra/providers/elasticsearch"
	"github.com/erda-project/erda/modules/extensions/loghub/exporter"
)

type define struct{}

func (d *define) Service() []string      { return []string{"logs-exporter-elasticsearch"} }
func (d *define) Dependencies() []string { return []string{"logs-exporter-base", "elasticsearch@logs"} }
func (d *define) Summary() string        { return "logs export to elasticsearch" }
func (d *define) Description() string    { return d.Summary() }
func (d *define) Config() interface{}    { return &config{} }
func (d *define) Creator() servicehub.Creator {
	return func() servicehub.Provider {
		return &provider{}
	}
}

type config struct {
	WriterConfig elasticsearch.WriterConfig `file:"writer_config"`
	Index        struct {
		Prefix string `file:"prefix"`
		Index  string `file:"index"`
	} `file:"index"`
}

type provider struct {
	C      *config
	L      logs.Logger
	exp    exporter.Interface
	es     elasticsearch.Interface
	output writer.Writer
}

func (p *provider) Init(ctx servicehub.Context) error {
	p.exp = ctx.Service("logs-exporter-base").(exporter.Interface)
	p.es = ctx.Service("elasticsearch").(elasticsearch.Interface)
	p.output = p.es.NewBatchWriter(&p.C.WriterConfig)
	return nil
}

func (p *provider) Start() error {
	return p.exp.NewConsumer(p.newOutput)
}

func (p *provider) Close() error { return nil }

func (p *provider) newOutput(i int) (exporter.Output, error) {
	return &esOutput{
		prefix: p.C.Index.Prefix,
		index:  p.C.Index.Index,
		writer: p.output,
	}, nil
}

type esOutput struct {
	prefix string
	index  string
	writer writer.Writer
}

func (o *esOutput) Write(key string, data []byte) error {
	index := o.index
	if len(index) <= 0 {
		index = o.prefix + key
	}
	return o.writer.Write(&elasticsearch.Document{
		Index: index,
		Data:  string(data),
	})
}

func init() {
	servicehub.RegisterProvider("logs-exporter-elasticsearch", &define{})
}
