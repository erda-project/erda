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

	"github.com/IBM/sarama"
	"github.com/stretchr/testify/assert"
)

func TestAsyncProducer(t *testing.T) {
	ref := uint32(0)
	ap := &AsyncProducer{
		producer:     newMockAsyncProducer(),
		blockTimeout: time.Second * 30,
		topic:        "default",
		ref:          &ref,
	}
	assert.Nil(t, ap.Write(&sarama.ProducerMessage{Topic: "my-topic", Value: sarama.ByteEncoder(`hello1`)}))
	assert.Nil(t, ap.Write("hello2"))
	assert.Nil(t, ap.Write([]byte("hello3")))
	assert.Nil(t, ap.Write(&Message{Topic: pointerString("my-topic"), Data: []byte("hello4")}))
	assert.Nil(t, ap.Write(1))
	assert.NotNil(t, ap.Write(make(chan struct{})))

	n, err := ap.WriteN(1, 2, 3)
	assert.Nil(t, err)
	assert.Equal(t, 3, n)

	assert.Nil(t, ap.Close())
}

func Test_provider_NewProducer(t *testing.T) {
	_, err := mockProvider.NewProducer(&ProducerConfig{Topic: "my-topic"})
	assert.Nil(t, err)
}

type mockAsyncProducer struct {
	ch      chan *sarama.ProducerMessage
	errorsC chan *sarama.ProducerError
}

func newMockAsyncProducer() *mockAsyncProducer {
	ap := &mockAsyncProducer{ch: make(chan *sarama.ProducerMessage, 1), errorsC: make(chan *sarama.ProducerError)}
	go func() {
		for range ap.ch {
		}
	}()
	return ap
}

func (m *mockAsyncProducer) getMessage() *sarama.ProducerMessage {
	select {
	case msg := <-m.ch:
		return msg
	default:
		return nil
	}
}

func (m *mockAsyncProducer) AsyncClose() {
	//TODO implement me
	panic("implement me")
}

func (m *mockAsyncProducer) Close() error {
	return nil
}

func (m *mockAsyncProducer) Input() chan<- *sarama.ProducerMessage {
	return m.ch
}

func (m *mockAsyncProducer) Successes() <-chan *sarama.ProducerMessage {
	//TODO implement me
	panic("implement me")
}

func (m *mockAsyncProducer) Errors() <-chan *sarama.ProducerError {
	return m.errorsC
}

func pointerString(s string) *string {
	var ps string
	ps = s
	return &ps
}
