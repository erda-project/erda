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
	"testing"

	"bou.ke/monkey"
	"github.com/IBM/sarama"

	"github.com/erda-project/erda-infra/base/logs/logrusx"
)

func TestMain(m *testing.M) {
	mockProvider = &provider{
		Log: logrusx.New(),
		Cfg: &config{
			Producer:        &globalProducerConfig{},
			ProtocolVersion: "1.1.0",
			Servers:         "localhost:9092",
			ClientID:        "test",
		},
	}
	mockProvider.Init(nil)
	monkey.Patch(sarama.NewConsumerGroup, func(addrs []string, groupID string, config *sarama.Config) (sarama.ConsumerGroup, error) {
		return &mockConsumerGroup{done: make(chan struct{})}, nil
	})
	monkey.Patch(sarama.NewAsyncProducer, func(addrs []string, conf *sarama.Config) (sarama.AsyncProducer, error) {
		return newMockAsyncProducer(), nil
	})
	m.Run()
}
