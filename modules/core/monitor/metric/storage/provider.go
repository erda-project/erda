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

package storage

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	writer "github.com/erda-project/erda-infra/pkg/parallel-writer"
	"github.com/erda-project/erda-infra/providers/elasticsearch"
	"github.com/erda-project/erda-infra/providers/kafka"
	indexmanager "github.com/erda-project/erda/modules/core/monitor/metric/index"
)

const serviceIndexManager = "erda.core.monitor.metric.index"

type config struct {
	Inputs struct {
		Metric              kafka.ConsumerConfig `file:"metric"`
		CreatingIndexMetric kafka.ConsumerConfig `file:"creating_index_metric"`
	} `file:"inputs"`
	Output struct {
		Features struct {
			GenerateMeta   bool   `file:"generate_meta" default:"true"`
			Counter        bool   `file:"counter" default:"true"`
			MachineSummary bool   `file:"machine_summary" default:"false"` // This code will be removed later.
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
	p.index = ctx.Service(serviceIndexManager).(indexmanager.Index)

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
	servicehub.Register("metrics-storage", &servicehub.Spec{
		Services:     []string{"metrics-storage"},
		Dependencies: []string{"kafka", "elasticsearch", serviceIndexManager},
		Description:  "metrics store",
		ConfigFunc:   func() interface{} { return &config{} },
		Creator: func() servicehub.Provider {
			return &provider{}
		},
	})
}
