package kafka

import (
	"fmt"
	"time"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	kafkaInf "github.com/erda-project/erda-infra/providers/kafka"
	"github.com/erda-project/erda/modules/msp/apm/trace"
	"github.com/erda-project/erda/modules/oap/collector/core/model"
	"github.com/erda-project/erda/modules/oap/collector/lib/protoparser/oapspan"
	"github.com/erda-project/erda/modules/oap/collector/lib/protoparser/spotspan"
)

type parserName string

const (
	oapSpan  parserName = "oapspan"
	spotSpan parserName = "spotspan"
)

type config struct {
	ProtoParser string                   `file:"proto_parser"`
	Concurrency int                      `file:"concurrency" default:"9"`
	BufferSize  int                      `file:"buffer_size" default:"512"`
	ReadTimeout time.Duration            `file:"read_timeout" default:"10s"`
	Consumer    *kafkaInf.ConsumerConfig `file:"consumer"`
}

// +provider
type provider struct {
	Cfg      *config
	parser   parserName
	Log      logs.Logger
	Kafka    kafkaInf.Interface `autowired:"kafka"`
	consumer model.ObservableDataConsumerFunc
}

func (p *provider) ComponentConfig() interface{} {
	return p.Cfg
}

func (p *provider) RegisterConsumer(consumer model.ObservableDataConsumerFunc) {
	p.consumer = consumer
}

// Run this is optional
func (p *provider) Init(ctx servicehub.Context) error {
	if p.Cfg.ProtoParser == "" {
		return fmt.Errorf("proto_parser required")
	}

	p.parser = parserName(p.Cfg.ProtoParser)

	var invokeFunc kafkaInf.ConsumerFunc
	switch p.parser {
	case oapSpan:
		invokeFunc = p.parseOapSpan()
	case spotSpan:
		invokeFunc = p.parseSpotSpan()
	default:
		return fmt.Errorf("invalide parser: %q", p.parser)
	}

	err := p.Kafka.NewConsumer(p.Cfg.Consumer, invokeFunc)
	if err != nil {
		return fmt.Errorf("failed create consumer: %w", err)

	}
	return nil
}

func (p *provider) parseOapSpan() kafkaInf.ConsumerFunc {
	return func(key []byte, value []byte, topic *string, timestamp time.Time) error {
		return oapspan.ParseOapSpan(value, func(span *trace.Span) error {
			p.consumer(span)
			return nil
		})
	}
}

func (p *provider) parseSpotSpan() kafkaInf.ConsumerFunc {
	return func(key []byte, value []byte, topic *string, timestamp time.Time) error {
		return spotspan.ParseSpotSpan(value, func(span *trace.Span) error {
			p.consumer(span)
			return nil
		})
	}
}

func init() {
	servicehub.Register("erda.oap.collector.receiver.kafka", &servicehub.Spec{
		Services: []string{
			"erda.oap.collector.receiver.kafka",
		},
		Description: "here is description of erda.oap.collector.receiver.kafka",
		ConfigFunc: func() interface{} {
			return &config{}
		},
		Creator: func() servicehub.Provider {
			return &provider{}
		},
	})
}
