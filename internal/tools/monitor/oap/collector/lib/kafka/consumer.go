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

	"github.com/IBM/sarama"

	"github.com/erda-project/erda-infra/base/logs"
)

// deprecated
type ConsumerFunc func(key []byte, value []byte, topic *string, timestamp time.Time) error
type ConsumerFuncV2 func(msg *sarama.ConsumerMessage) error

// ConsumerConfig .
type ConsumerConfig struct {
	Topics []string `file:"topics" desc:"topics"`
	// group related
	Group             string        `file:"group" desc:"consumer group id"`
	SessionTimeout    time.Duration `file:"session_timeout" default:"2m" env:"KAFKA_C_G_SESSION_TIMEOUT"`
	MaxProcessingTime time.Duration `file:"max_processing_time" default:"30s" env:"KAFKA_C_MAX_PROCESSING_TIME"`
	ChannelBufferSize int           `file:"channel_buffer_size" default:"500" env:"KAFKA_CHANNEL_BUFFER_SIZE"`

	Offsets struct {
		AutoCommit struct {
			Enable   bool          `file:"enable" default:"true"`
			Interval time.Duration `file:"interval" default:"1s"`
		} `file:"auto_commit"`
		Initial string `file:"initial" default:"latest"`
	} `file:"offsets"`
}

func (p *provider) newConsumerGroup(c *ConsumerConfig) (sarama.ConsumerGroup, error) {
	cfg, err := p.customConfig(c)
	if err != nil {
		return nil, fmt.Errorf("custom config: %w", err)
	}
	cg, err := sarama.NewConsumerGroup(p.Brokers(), c.Group, cfg)
	if err != nil {
		return nil, fmt.Errorf("creat consumer group: %w", err)
	}
	return cg, nil
}

func (p *provider) customConfig(c *ConsumerConfig) (*sarama.Config, error) {
	cfg := sarama.NewConfig()
	cfg.Version = p.protoVersion
	cfg.ClientID = p.Cfg.ClientID
	cfg.ChannelBufferSize = c.ChannelBufferSize

	cfg.Consumer.MaxProcessingTime = c.MaxProcessingTime
	cfg.Consumer.Fetch.Min = 512 * 1024
	cfg.Consumer.Fetch.Default = 5 * 1024 * 1024

	cfg.Consumer.Group.Session.Timeout = c.SessionTimeout
	cfg.Consumer.Group.Heartbeat.Interval = c.SessionTimeout / 4

	if v, err := parserOffsetInitial(c.Offsets.Initial); err != nil {
		return nil, err
	} else {
		cfg.Consumer.Offsets.Initial = v
	}
	cfg.Consumer.Offsets.AutoCommit.Enable = c.Offsets.AutoCommit.Enable
	cfg.Consumer.Offsets.AutoCommit.Interval = c.Offsets.AutoCommit.Interval

	cfg.Net.ReadTimeout = cfg.Consumer.Group.Rebalance.Timeout + 30*time.Second
	return cfg, nil
}

func (p *provider) NewConsumerGroup(c *ConsumerConfig, handler ConsumerFuncV2) (*ConsumerGroupManager, error) {
	cg, err := p.newConsumerGroup(c)
	if err != nil {
		return nil, err
	}
	return NewConsumerGroupManager(
		cg,
		&defaultHandler{fn: handler, log: p.Log.Sub(c.Group).Sub("default-handler")},
		c.Topics,
	), nil
}

// deprecated
func (p *provider) NewConsumer(c *ConsumerConfig, handler ConsumerFunc) error {
	cg, err := p.newConsumerGroup(c)
	if err != nil {
		return err
	}
	handlerv2 := ConsumerFuncV2(func(msg *sarama.ConsumerMessage) error {
		return handler(msg.Key, msg.Value, &msg.Topic, msg.Timestamp)
	})
	_ = NewConsumerGroupManager(
		cg,
		&defaultHandler{fn: handlerv2, log: p.Log.Sub(c.Group).Sub("default-handler")},
		c.Topics,
	)

	return nil
}

func (p *provider) NewConsumerWitchCreator(c *ConsumerConfig, handlerFactory func(i int) (ConsumerFunc, error)) error {
	handler, err := handlerFactory(0)
	if err != nil {
		return fmt.Errorf("create handler failed: %w", err)
	}
	err = p.NewConsumer(c, handler)
	if err != nil {
		return fmt.Errorf("failed NewConsumer: %w", err)
	}
	return nil
}

type ConsumerGroupManager struct {
	cg      sarama.ConsumerGroup
	handler sarama.ConsumerGroupHandler
	topics  []string
	cancel  context.CancelFunc
}

func NewConsumerGroupManager(cg sarama.ConsumerGroup, handler sarama.ConsumerGroupHandler, topics []string) *ConsumerGroupManager {
	ctx, cancel := context.WithCancel(context.Background())
	cgm := &ConsumerGroupManager{
		cg:      cg,
		cancel:  cancel,
		handler: handler,
		topics:  topics,
	}
	go cgm.consume(ctx)
	return cgm
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
	fn  ConsumerFuncV2
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
			err := d.fn(msg)
			if err != nil {
				d.log.Errorf("consume err: %s", err)
			}
			session.MarkMessage(msg, "")
		case <-session.Context().Done():
			return nil
		}
	}
}

func parserOffsetInitial(s string) (int64, error) {
	switch s {
	case "latest":
		return sarama.OffsetNewest, nil
	case "earliest":
		return sarama.OffsetOldest, nil
	default:
		return 0, fmt.Errorf("invalid offset initial: %q", s)
	}
}
