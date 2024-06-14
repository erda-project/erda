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
	"testing"
	"time"

	"github.com/IBM/sarama"
	"github.com/stretchr/testify/assert"
)

func Test_provider_NewConsumerGroup(t *testing.T) {
	cgcfg := &ConsumerConfig{}
	cgcfg.SessionTimeout = time.Minute * 3
	cgcfg.Topics = []string{"my-topic"}
	cgcfg.Group = "myu-group"
	cgcfg.ChannelBufferSize = 10
	cgcfg.MaxProcessingTime = time.Second * 30
	cgcfg.Offsets.Initial = "latest"
	cgcfg.Offsets.AutoCommit.Enable = true
	cgcfg.Offsets.AutoCommit.Interval = time.Second
	_, err := mockProvider.customConfig(cgcfg)
	assert.Nil(t, err)

	cgcfg.Offsets.Initial = "invalid"
	_, err = mockProvider.customConfig(cgcfg)
	assert.NotNil(t, err)

	cgcfg.Offsets.Initial = "latest"
	_, err = mockProvider.NewConsumerGroup(cgcfg, func(msg *sarama.ConsumerMessage) error {
		return nil
	})
	assert.Nil(t, err)

	err = mockProvider.NewConsumer(cgcfg, func(key []byte, value []byte, topic *string, timestamp time.Time) error {
		return nil
	})
	assert.Nil(t, err)
}

func TestNewConsumerGroupManager(t *testing.T) {
	mcg := &mockConsumerGroup{done: make(chan struct{})}
	cm := NewConsumerGroupManager(mcg, &defaultHandler{}, []string{"my-topic"})
	defer close(mcg.done)
	assert.NotNil(t, t, cm)
}

type mockConsumerGroup struct {
	done chan struct{}
}

func (m *mockConsumerGroup) Consume(ctx context.Context, topics []string, handler sarama.ConsumerGroupHandler) error {
	<-m.done
	return nil
}

func (m *mockConsumerGroup) Errors() <-chan error {
	return nil
}

func (m *mockConsumerGroup) Close() error {
	return nil
}

func (m *mockConsumerGroup) Pause(partitions map[string][]int32) {
	return
}

func (m *mockConsumerGroup) Resume(partitions map[string][]int32) {
	return
}

func (m *mockConsumerGroup) PauseAll() {
	return
}

func (m *mockConsumerGroup) ResumeAll() {
	return
}
