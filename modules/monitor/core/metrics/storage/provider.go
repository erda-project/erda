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

package storage

import (
	"fmt"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	writer "github.com/erda-project/erda-infra/pkg/parallel-writer"
	"github.com/erda-project/erda-infra/providers/elasticsearch"
	"github.com/erda-project/erda-infra/providers/kafka"
	indexmanager "github.com/erda-project/erda/modules/monitor/core/metrics/index"
	"github.com/prometheus/client_golang/prometheus"
)

type define struct{}

func (d *define) Services() []string { return []string{"metrics-storage"} }
func (d *define) Dependencies() []string {
	return []string{"kafka", "elasticsearch", "metrics-index-manager"}
}
func (d *define) Summary() string     { return "metrics store" }
func (d *define) Description() string { return d.Summary() }
func (d *define) Config() interface{} { return &config{} }
func (d *define) Creator() servicehub.Creator {
	return func() servicehub.Provider {
		return &provider{}
	}
}

type config struct {
	Inputs struct {
		Metric              kafka.ConsumerConfig `file:"metric"`
		CreatingIndexMetric kafka.ConsumerConfig `file:"creating_index_metric"`
	} `file:"inputs"`
	Output struct {
		Features struct {
			GenerateMeta   bool   `file:"generate_meta" default:"true"`
			Counter        bool   `file:"counter" default:"true"`
			MachineSummary bool   `file:"machine_summary" default:"false"` // 后面要移除这段代码
			FilterPrefix   string `file:"filter_prefix" default:"go_" env:"METRIC_FILTER_PREFIX"`
		} `file:"features"`
		Elasticsearch struct {
			elasticsearch.WriterConfig `file:"writer_config"`
		} `file:"elasticsearch"`
		Kafka kafka.ProducerConfig `file:"kafka"`
	} `file:"output"`
}

type provider struct {
	C      *config
	L      logs.Logger
	kafka  kafka.Interface
	index  indexmanager.Index
	output struct {
		es    writer.Writer
		kafka writer.Writer
	}
	counter       *prometheus.Counter
	metaProcessor *metaProcessor
}

func (p *provider) Init(ctx servicehub.Context) error {
	es := ctx.Service("elasticsearch").(elasticsearch.Interface)
	p.output.es = es.NewBatchWriter(&p.C.Output.Elasticsearch.WriterConfig)
	p.index = ctx.Service("metrics-index-manager").(indexmanager.Index)

	p.kafka = ctx.Service("kafka").(kafka.Interface)
	if p.index.EnableRollover() {
		w, err := p.kafka.NewProducer(&p.C.Output.Kafka)
		if err != nil {
			return fmt.Errorf("fail to create kafka producer: %s", err)
		}
		p.output.kafka = w
	}

	// if p.C.Output.Features.Counter {
	// 	p.counter = promxp.RegisterAutoResetCounterVec(
	// 		"metric_store",
	// 		"metric写入统计",
	// 		map[string]string{"instance_id": common2.InstanceID()},
	// 		[]string{metricNameKey /*metricScopeKey, metricScopeIDKey*/, srcOrgNameKey, srcClusterNameKey},
	// 		"dice", "monitor",
	// 	)
	// }
	if p.C.Output.Features.GenerateMeta {
		p.metaProcessor = createMetaProcess(p.output.es, p.index, p.counter)
	}
	return nil
}

// Start .
func (p *provider) Start() error {
	p.kafka.NewConsumer(&p.C.Inputs.Metric, p.invoke)
	if p.index.EnableRollover() && len(p.C.Inputs.CreatingIndexMetric.Topics) > 0 {
		p.kafka.NewConsumer(&p.C.Inputs.CreatingIndexMetric, p.handleCreatingIndexMetric)
	}
	return nil
}

func (p *provider) Close() error {
	p.L.Debug("not support close kafka consumer")
	return nil
}

func init() {
	servicehub.RegisterProvider("metrics-storage", &define{})
}
