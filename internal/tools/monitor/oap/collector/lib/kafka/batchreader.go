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
	"github.com/erda-project/erda/internal/tools/monitor/core/storekit"
)

// BatchReaderOption .
type BatchReaderOption interface{}

// WithReaderDecoder .
func WithReaderDecoder(dec Decoder) BatchReaderOption {
	return BatchReaderOption(dec)
}

type Decoder func(key, value []byte, topic *string, timestamp time.Time) (interface{}, error)

type BatchReaderConfig ConsumerConfig

func (p *provider) NewBatchReader(c *BatchReaderConfig, options ...BatchReaderOption) (storekit.BatchReader, error) {
	c.Offsets.AutoCommit.Enable = false

	cg, err := p.newConsumerGroup((*ConsumerConfig)(c))
	if err != nil {
		return nil, err
	}

	var dec Decoder
	for _, opt := range options {
		switch v := opt.(type) {
		case Decoder:
			dec = v
		}
	}
	if dec == nil {
		dec = func(key, value []byte, topic *string, timestamp time.Time) (interface{}, error) {
			return value, nil
		}
	}

	return newKafkaBatchReader(cg, dec, p.Log.Sub(c.Group), c.Topics), nil
}

type kafkaBatchReader struct {
	log    logs.Logger
	cg     sarama.ConsumerGroup
	topics []string
	dec    Decoder

	cancel  context.CancelFunc
	handler *batchHandler
}

func newKafkaBatchReader(cg sarama.ConsumerGroup, dec Decoder, log logs.Logger, topics []string) *kafkaBatchReader {
	ctx, cancel := context.WithCancel(context.Background())
	obj := &kafkaBatchReader{
		cg:      cg,
		topics:  topics,
		dec:     dec,
		log:     log,
		handler: newBatchHandler(dec, log.Sub("batch-handler")),
		cancel:  cancel,
	}
	go obj.consume(ctx)
	return obj
}

func (kbr *kafkaBatchReader) ReadN(buf []storekit.Data, timeout time.Duration) (int, error) {
	maxWaitTimer := time.NewTimer(timeout * 3)
	defer maxWaitTimer.Stop()
	offset, size := 0, len(buf)
loop:
	for ; offset < size; offset++ {
		select {
		case <-maxWaitTimer.C:
			break loop
		case msg, ok := <-kbr.handler.msgC:
			if !ok {
				return offset, nil
			}
			if msg.err != nil {
				return offset, msg.err
			}
			buf[offset] = msg.data
		}
	}
	return offset, nil
}

func (kbr *kafkaBatchReader) Confirm() error {
	kbr.handler.commit()
	return nil
}

func (kbr *kafkaBatchReader) Close() error {
	kbr.cancel()
	return kbr.cg.Close()
}

func (kbr *kafkaBatchReader) consume(ctx context.Context) {
	for {
		// TODO. may consume multi time, because of rebalance
		err := kbr.cg.Consume(ctx, kbr.topics, kbr.handler)
		if err != nil {
			kbr.log.Fatalf("consumer group consume failed: %s", err)
		}
		// check ctx is canceled
		if ctx.Err() != nil {
			return
		}
		kbr.handler = newBatchHandler(kbr.dec, kbr.log.Sub("batch-handler"))
	}
}

type batchHandler struct {
	fn      Decoder
	log     logs.Logger
	msgC    chan *message
	session sarama.ConsumerGroupSession
}

func newBatchHandler(fn Decoder, log logs.Logger) *batchHandler {
	return &batchHandler{
		fn:   fn,
		log:  log,
		msgC: make(chan *message, 100),
	}
}

func (d *batchHandler) Setup(session sarama.ConsumerGroupSession) error {
	d.log.Infof("session setup: member-id: %q, claims: %v", session.MemberID(), session.Claims())
	d.session = session
	return nil
}

func (d *batchHandler) Cleanup(session sarama.ConsumerGroupSession) error {
	d.log.Infof("session cleanup: member-id: %q, claims: %v", session.MemberID(), session.Claims())
	close(d.msgC)
	session.Commit()
	return nil
}

func (d *batchHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case msg, ok := <-claim.Messages():
			if !ok {
				return fmt.Errorf("empty received")
			}
			data, err := d.fn(msg.Key, msg.Value, &msg.Topic, msg.Timestamp)
			if err != nil {
				d.msgC <- &message{err: fmt.Errorf("decode failed: %w", err)}
			} else {
				d.msgC <- &message{data: data}
			}
			session.MarkMessage(msg, "")
		case <-session.Context().Done():
			return nil
		}
	}
}

func (d *batchHandler) commit() {
	if d.session != nil {
		d.session.Commit()
	}
}

type message struct {
	data storekit.Data
	err  error
}
