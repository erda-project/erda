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
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda-infra/base/logs/logrusx"
	"github.com/erda-project/erda/internal/tools/monitor/core/storekit"
)

var mockProvider *provider

func Test_kafkaBatchReader(t *testing.T) {
	mcg := &mockConsumerGroup{done: make(chan struct{})}
	kbr := newKafkaBatchReader(mcg, func(key, value []byte, topic *string, timestamp time.Time) (interface{}, error) {
		return value, nil
	}, logrusx.New(), []string{"my-topic"})
	assert.NotNil(t, kbr)

	go func() {
		for i := 0; i < 3; i++ {
			kbr.handler.msgC <- &message{err: nil, data: "hello"}
		}
	}()

	buf := make([]storekit.Data, 3)
	n, err := kbr.ReadN(buf, time.Second)
	assert.Nil(t, err)
	assert.Equal(t, 3, n)
	assert.Equal(t, "hello", buf[0].(string))
}

func Test_provider_NewBatchReader(t *testing.T) {
	cgcfg := &ConsumerConfig{}
	cgcfg.SessionTimeout = time.Minute * 3
	cgcfg.Topics = []string{"my-topic"}
	cgcfg.Group = "myu-group"
	cgcfg.ChannelBufferSize = 10
	cgcfg.MaxProcessingTime = time.Second * 30
	cgcfg.Offsets.Initial = "latest"
	cgcfg.Offsets.AutoCommit.Enable = true
	cgcfg.Offsets.AutoCommit.Interval = time.Second
	_, err := mockProvider.NewBatchReader((*BatchReaderConfig)(cgcfg))
	assert.Nil(t, err)
}
