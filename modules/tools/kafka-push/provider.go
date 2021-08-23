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

package kafka_push

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	writer "github.com/erda-project/erda-infra/pkg/parallel-writer"
	"github.com/erda-project/erda-infra/providers/kafka"
)

type define struct{}

func (d *define) Service() []string      { return []string{"kafka-push"} }
func (d *define) Dependencies() []string { return []string{"kafka"} }
func (d *define) Summary() string        { return "push data to kafka" }
func (d *define) Description() string    { return d.Summary() }
func (d *define) Config() interface{}    { return &config{} }
func (d *define) Creator() servicehub.Creator {
	return func() servicehub.Provider {
		return &provider{}
	}
}

type config struct {
	Kafka kafka.ProducerConfig `file:"kafka"`
}

type provider struct {
	C      *config
	L      logs.Logger
	kafka  kafka.Interface
	output writer.Writer
	data   map[string]interface{}
}

var pushData = []byte(`{
		"name":"ads_rfm_category_forecast",
		"fields":{
			"createdat":"20200918",
			"quantity":0,
			"category":3
		},
		"tags":{
			"createdat":"20200918",
			"cluster_name":"terminus-captain",
			"quantity":0,
			"_metric_scope_id":"terminus",
			"_metric_scope":"bigdata",
			"_meta":"true",
			"state":"running",
			"category":3
		},
		"timestamp":1600413368836000000
	}`)

func (p *provider) Init(ctx servicehub.Context) error {
	p.kafka = ctx.Service("kafka").(kafka.Interface)
	w, err := p.kafka.NewProducer(&p.C.Kafka)
	if err != nil {
		return fmt.Errorf("fail to create kafka producer: %s", err)
	}
	p.output = w
	p.data = make(map[string]interface{})
	return json.Unmarshal(pushData, &p.data)
}

// Start .
func (p *provider) Start() error {
	for {
		p.data["timestamp"] = time.Now().UnixNano()
		p.output.Write(p.data)
		time.Sleep(3 * time.Second)
	}
}

func (p *provider) Close() error { return nil }

func init() {
	servicehub.RegisterProvider("kafka-push", &define{})
}
