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
	"sync"
	"sync/atomic"
	"time"

	"github.com/IBM/sarama"

	"github.com/erda-project/erda/pkg/strutil"
)

var (
	mu             sync.Mutex
	sharedProducer *AsyncProducer // producer is thread safe, so we shared
)

type globalProducerConfig struct {
	MaxBlockMs        time.Duration `file:"max.block.ms" default:"30s"`
	LingerMs          time.Duration `file:"linger.ms" default:"50ms"`
	RequestTimeoutMs  time.Duration `file:"request.timeout.ms" default:"30s"`
	BatchSize         int           `file:"batch.size" default:"16384"`
	BufferMemory      int           `file:"buffer.memory" default:"33554432"`
	ChannelBufferSize int           `file:"channel_buffer_size" default:"500" env:"KAFKA_P_CHANNEL_BUFFER_SIZE"`
}

type ProducerConfig struct {
	Topic string `file:"topic" env:"KAFKA_P_TOPIC" desc:"topic"`
}

type ProducerOption interface{}

// NewProducer default shared
func (p *provider) NewProducer(pc *ProducerConfig) (*AsyncProducer, error) {
	mu.Lock()
	defer mu.Unlock()
	if sharedProducer != nil {
		atomic.AddUint32(sharedProducer.ref, 1)
		return sharedProducer, nil
	}

	cfg := sarama.NewConfig()
	cfg.Version = p.protoVersion
	cfg.ClientID = p.Cfg.ClientID
	cfg.Producer.RequiredAcks = sarama.WaitForLocal
	cfg.ChannelBufferSize = p.Cfg.Producer.ChannelBufferSize
	cfg.Producer.Flush.MaxMessages = p.Cfg.Producer.BatchSize
	cfg.Producer.Flush.Frequency = p.Cfg.Producer.LingerMs
	cfg.Producer.Flush.Bytes = p.Cfg.Producer.BufferMemory

	producer, err := sarama.NewAsyncProducer(p.Brokers(), cfg)
	if err != nil {
		return nil, fmt.Errorf("create async producer: %w", err)
	}

	ref := uint32(1)
	sharedProducer = &AsyncProducer{
		topic:        pc.Topic,
		ref:          &ref,
		producer:     producer,
		blockTimeout: p.Cfg.Producer.MaxBlockMs,
	}

	go func() {
		for pe := range producer.Errors() {
			if pe.Err == nil {
				continue
			}
			p.Log.Sub("producer").Errorf("failed to produce, topic: %s, err: %s", pe.Msg.Topic, pe.Err)
		}
	}()
	return sharedProducer, nil
}

type AsyncProducer struct {
	topic        string
	ref          *uint32
	blockTimeout time.Duration
	producer     sarama.AsyncProducer
}

// Message .
type Message struct {
	Topic *string
	Data  []byte
	Key   []byte
}

func (a *AsyncProducer) Write(data interface{}) error {
	switch val := data.(type) {
	case *sarama.ProducerMessage:
		return a.send(val)
	case []byte:
		return a.send(&sarama.ProducerMessage{Topic: a.topic, Value: sarama.ByteEncoder(val)})
	case string:
		return a.send(&sarama.ProducerMessage{Topic: a.topic, Value: sarama.ByteEncoder(strutil.NoCopyStringToBytes(val))})
	case *Message:
		return a.send(&sarama.ProducerMessage{Topic: *val.Topic, Value: sarama.ByteEncoder(val.Data)})
	default:
		buf, err := json.Marshal(data)
		if err != nil {
			return err
		}
		return a.send(&sarama.ProducerMessage{Topic: a.topic, Value: sarama.ByteEncoder(buf)})
	}
}

func (a *AsyncProducer) send(pmsg *sarama.ProducerMessage) error {
	select {
	case a.producer.Input() <- pmsg:
	case <-time.After(a.blockTimeout):
		return fmt.Errorf("produce message block timout")
	}
	return nil
}

func (a *AsyncProducer) WriteN(data ...interface{}) (int, error) {
	offset := 0
	for ; offset < len(data); offset++ {
		err := a.Write(data[offset])
		if err != nil {
			return offset, err
		}
	}
	return offset, nil
}

func (a *AsyncProducer) Close() error {
	atomic.AddUint32(a.ref, ^uint32(0))
	if atomic.LoadUint32(a.ref) > 0 {
		return nil
	}
	err := a.producer.Close()
	if err != nil {
		return fmt.Errorf("close produer: %w", err)
	}
	return nil
}
