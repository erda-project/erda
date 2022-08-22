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
	"context"
	"fmt"
	"time"

	"github.com/Shopify/sarama"

	"github.com/erda-project/erda-infra/base/logs"
)

// ConsumerConfig .
type ConsumerConfig struct {
	Topics []string `file:"topics" desc:"topics"`
	// group related
	Group             string        `file:"group" desc:"consumer group id"`
	SessionTimeout    time.Duration `file:"session_timeout" default:"3m" env:"KAFKA_C_G_SESSION_TIMEOUT"`
	MaxProcessingTime time.Duration `file:"max_processing_time" default:"30s" env:"KAFKA_C_MAX_PROCESSING_TIME"`
	ChannelBufferSize int           `file:"channel_buffer_size" default:"500" env:"KAFKA_C_G_CHANNEL_BUFFER_SIZE"`
}

type ConsumerOption interface{}

// ConsumerFunc .
type ConsumerFunc func(key []byte, value []byte, topic string, timestamp time.Time) error

func (p *provider) NewConsumerGroup(c *ConsumerConfig, handler ConsumerFunc, options ...ConsumerOption) (*ConsumerGroupManager, error) {
	cfg := sarama.NewConfig()
	cfg.Version = p.protoVersion
	cfg.ClientID = p.Cfg.ClientID
	cfg.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategySticky
	cfg.Consumer.Group.Session.Timeout = c.SessionTimeout
	cfg.Consumer.Group.Heartbeat.Interval = c.SessionTimeout / 4
	cfg.Consumer.MaxProcessingTime = c.MaxProcessingTime
	cfg.ChannelBufferSize = c.ChannelBufferSize
	cfg.Consumer.Fetch.Min = 512 * 1024
	cfg.Consumer.Fetch.Default = 5 * 1024 * 1024
	cfg.Consumer.Offsets.Initial = sarama.OffsetNewest

	cfg.Net.ReadTimeout = cfg.Consumer.Group.Rebalance.Timeout + 30*time.Second
	cg, err := sarama.NewConsumerGroup(p.Brokers(), c.Group, cfg)
	if err != nil {
		return nil, fmt.Errorf("creat consumer group: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cgm := &ConsumerGroupManager{
		cancel:  cancel,
		cg:      cg,
		handler: &defaultHandler{fn: handler, log: p.Log.Sub(c.Group).Sub("handler")},
		topics:  c.Topics,
	}

	go cgm.consume(ctx)
	return cgm, nil
}

type ConsumerGroupManager struct {
	cg      sarama.ConsumerGroup
	handler sarama.ConsumerGroupHandler
	topics  []string
	cancel  context.CancelFunc
}

func (cgm *ConsumerGroupManager) Close() error {
	cgm.cancel()
	return cgm.cg.Close()
}

func (cgm *ConsumerGroupManager) consume(ctx context.Context) {
	for {
		err := cgm.cg.Consume(ctx, cgm.topics, cgm.handler)
		if err != nil {
			panic(fmt.Sprintf("consumer group consume failed: %s", err))
		}
		// check ctx is canceled
		if ctx.Err() != nil {
			return
		}
	}
}

type defaultHandler struct {
	fn  ConsumerFunc
	log logs.Logger
}

func (d *defaultHandler) Setup(session sarama.ConsumerGroupSession) error {
	d.log.Infof("session setup: member-id: %q, claims: %v", session.MemberID(), session.Claims())
	return nil
}

func (d *defaultHandler) Cleanup(session sarama.ConsumerGroupSession) error {
	d.log.Infof("session cleanup: member-id: %q, claims: %v", session.MemberID(), session.Claims())
	return nil
}

func (d *defaultHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case msg, ok := <-claim.Messages():
			if !ok {
				return fmt.Errorf("empty received")
			}
			err := d.fn(msg.Key, msg.Value, msg.Topic, msg.Timestamp)
			if err != nil {
				d.log.Errorf("consume err: %s", err)
			}
			session.MarkMessage(msg, "")
		case <-session.Context().Done():
			return nil
		}
	}
}
