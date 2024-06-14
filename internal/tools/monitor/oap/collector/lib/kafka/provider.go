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
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/IBM/sarama"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda/internal/tools/monitor/core/storekit"
)

type Interface interface {
	NewProducer(c *ProducerConfig) (*AsyncProducer, error)
	NewConsumerGroup(c *ConsumerConfig, handler ConsumerFuncV2) (*ConsumerGroupManager, error)
	NewAdminClient() (AdminInterface, error)
	Brokers() []string
	//deprecated. Please use NewConsumerGroup
	NewBatchReader(c *BatchReaderConfig, options ...BatchReaderOption) (storekit.BatchReader, error)
	//deprecated. Please use NewConsumerGroup
	NewConsumer(c *ConsumerConfig, handler ConsumerFunc) error
	//deprecated. Please use NewConsumerGroup
	NewConsumerWitchCreator(c *ConsumerConfig, creator func(i int) (ConsumerFunc, error)) error
}

type config struct {
	Servers         string `file:"servers" env:"BOOTSTRAP_SERVERS" default:"localhost:9092" desc:"kafka servers"`
	ClientID        string `file:"client_id" env:"COMPONENT_NAME" default:"sarama" desc:"kafka client name"`
	DebugClient     bool   `file:"debug_client" env:"KAFKA_DEBUG_CLIENT" desc:"log sarama client log to console"`
	ProtocolVersion string `file:"protocol_version" default:"1.1.0" desc:"kafka broker protocol version"`

	Producer *globalProducerConfig `file:"producer"`
}

var _ Interface = (*provider)(nil)

// +provider
type provider struct {
	Cfg          *config
	Log          logs.Logger
	protoVersion sarama.KafkaVersion

	producer *AsyncProducer
}

func (p *provider) Init(ctx servicehub.Context) error {
	if p.Cfg.DebugClient {
		sarama.Logger = log.New(os.Stdout, "[Sarama] ", log.LstdFlags)
	}
	v, err := sarama.ParseKafkaVersion(p.Cfg.ProtocolVersion)
	if err != nil {
		return fmt.Errorf("parser version: %w", err)
	}
	p.protoVersion = v
	return nil
}

func (p *provider) NewAdminClient() (AdminInterface, error) {
	panic("implement me")
}

func (p *provider) Brokers() []string {
	return strings.Split(p.Cfg.Servers, ",")
}

// Provide .
func (p *provider) Provide(ctx servicehub.DependencyContext, options ...interface{}) interface{} {
	return p
}

func init() {
	servicehub.Register("kafkago", &servicehub.Spec{
		Services: []string{"kafkago"},
		ConfigFunc: func() interface{} {
			return &config{}
		},
		Creator: func() servicehub.Provider {
			return &provider{}
		},
	})
}
