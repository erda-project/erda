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

package kafka

import (
	"encoding/json"
	"fmt"

	"github.com/IBM/sarama"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	writer "github.com/erda-project/erda-infra/pkg/parallel-writer"
	"github.com/erda-project/erda/internal/apps/msp/apm/trace"
	"github.com/erda-project/erda/internal/tools/monitor/core/log"
	"github.com/erda-project/erda/internal/tools/monitor/core/metric"
	"github.com/erda-project/erda/internal/tools/monitor/core/profile"
	"github.com/erda-project/erda/internal/tools/monitor/oap/collector/core/model"
	"github.com/erda-project/erda/internal/tools/monitor/oap/collector/core/model/odata"
	"github.com/erda-project/erda/internal/tools/monitor/oap/collector/lib/kafka"
	"github.com/erda-project/erda/internal/tools/monitor/oap/collector/plugins"
)

var providerName = plugins.WithPrefixExporter("kafka")

type config struct {
	Keypass    map[string][]string `file:"keypass"`
	Keydrop    map[string][]string `file:"keydrop"`
	Keyinclude []string            `file:"keyinclude"`
	Keyexclude []string            `file:"keyexclude"`

	MetadataKeyOfTopic string `file:"metadata_key_of_topic" desc:"note: only for raw data type"`
	// produce raw data with key for same partition in kafka detail: https://erda.cloud/erda/dop/projects/387/issues/all?id=611023&iterationID=12783&tab=BUG&type=BUG
	ProduceRawWithKey bool   `file:"produce_raw_with_key" default:"false"`
	Topic             string `file:"topic"`
}

var _ model.Exporter = (*provider)(nil)

type produceRawFunc func(item *odata.Raw) *sarama.ProducerMessage

// +provider
type provider struct {
	Cfg     *config
	Log     logs.Logger
	Kafka   kafka.Interface `autowired:"kafkago"`
	writer  writer.Writer
	rawFunc produceRawFunc
}

func (p *provider) ComponentClose() error {
	return p.writer.Close()
}

func (p *provider) ExportMetric(items ...*metric.Metric) error {
	for _, item := range items {
		data, err := json.Marshal(item)
		if err != nil {
			p.Log.Errorf("serialize err: %s", err)
			continue
		}
		err = p.writer.Write(&sarama.ProducerMessage{
			Topic: p.Cfg.Topic,
			Value: sarama.ByteEncoder(data),
		})
		if err != nil {
			p.Log.Errorf("write data to %s err: %s", p.Cfg.Topic, err)
			continue
		}
	}
	return nil
}

func (p *provider) ExportLog(items ...*log.Log) error            { return nil }
func (p *provider) ExportSpan(items ...*trace.Span) error        { return nil }
func (p *provider) ExportProfile(items ...*profile.Output) error { return nil }
func (p *provider) ExportRaw(items ...*odata.Raw) error {
	for _, item := range items {
		msg := p.rawFunc(item)

		if err := p.writer.Write(msg); err != nil {
			p.Log.Errorf("write data to topic %s err: %s", msg.Topic, err)
		}

	}
	return nil
}

func (p *provider) produceRawMessage(item *odata.Raw) *sarama.ProducerMessage {
	topic := p.Cfg.Topic
	if p.Cfg.MetadataKeyOfTopic != "" {
		tmp, ok := item.Meta[p.Cfg.MetadataKeyOfTopic]
		if ok {
			topic = tmp
		}
	}
	msg := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.ByteEncoder(item.Data),
	}
	return msg
}

type key struct {
	ID string `json:"id"`
}

func (p *provider) produceRawMessageWithKey(item *odata.Raw) *sarama.ProducerMessage {
	msg := p.produceRawMessage(item)
	var k key
	if err := json.Unmarshal(item.Data, &k); err == nil && len(k.ID) > 0 {
		msg.Key = sarama.ByteEncoder(k.ID)
	}
	return msg
}

func (p *provider) ComponentConfig() interface{} {
	return p.Cfg
}

func (p *provider) Connect() error {
	return nil
}

// Run this is optional
func (p *provider) Init(ctx servicehub.Context) error {
	if p.Cfg.MetadataKeyOfTopic == "" && p.Cfg.Topic == "" {
		return fmt.Errorf("must specify metadata_key_of_topic or producer.topic")
	}
	producer, err := p.Kafka.NewProducer(&kafka.ProducerConfig{Topic: p.Cfg.Topic})
	if err != nil {
		return err
	}
	p.writer = producer
	p.rawFunc = p.produceRawMessage
	if p.Cfg.ProduceRawWithKey {
		p.rawFunc = p.produceRawMessageWithKey
	}
	return nil
}

func init() {
	servicehub.Register(providerName, &servicehub.Spec{
		Services: []string{
			providerName,
		},
		Description: "here is description of erda.oap.collector.exporter.kafka",
		ConfigFunc: func() interface{} {
			return &config{}
		},
		Creator: func() servicehub.Provider {
			return &provider{}
		},
	})
}
