package storage

import (
	"fmt"
	"time"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	writer "github.com/erda-project/erda-infra/pkg/parallel-writer"
	"github.com/erda-project/erda-infra/providers/cassandra"
	"github.com/erda-project/erda-infra/providers/kafka"
	"github.com/gocql/gocql"
)

type define struct{}

func (d *define) Service() []string      { return []string{"trace-storage"} }
func (d *define) Dependencies() []string { return []string{"kafka", "cassandra"} }
func (d *define) Summary() string        { return "trace storage" }
func (d *define) Description() string    { return d.Summary() }
func (d *define) Config() interface{}    { return &config{} }
func (d *define) Creator() servicehub.Creator {
	return func() servicehub.Provider {
		return &provider{}
	}
}

type config struct {
	Input  kafka.ConsumerConfig `file:"input"`
	Output struct {
		Cassandra struct {
			cassandra.WriterConfig  `file:"writer_config"`
			cassandra.SessionConfig `file:"session_config"`
			GCGraceSeconds          int           `file:"gc_grace_seconds" default:"86400"`
			TTL                     time.Duration `file:"ttl" default:"168h"`
		} `file:"cassandra"`
		Kafka kafka.ProducerConfig `file:"kafka"`
	} `file:"output"`
}

type provider struct {
	C      *config
	L      logs.Logger
	ttlSec int
	kafka  kafka.Interface
	output struct {
		cassandra writer.Writer
		kafka     writer.Writer
	}
	cassandraSession *gocql.Session
}

func (p *provider) Init(ctx servicehub.Context) error {
	p.ttlSec = int(p.C.Output.Cassandra.TTL.Seconds())
	cassandra := ctx.Service("cassandra").(cassandra.Interface)
	session, err := cassandra.Session(&p.C.Output.Cassandra.SessionConfig)
	p.cassandraSession = session
	if err != nil {
		return fmt.Errorf("fail to create cassandra session: %s", err)
	}
	err = p.initCassandra(session)
	if err != nil {
		return err
	}
	p.output.cassandra = cassandra.NewBatchWriter(session, &p.C.Output.Cassandra.WriterConfig, p.createTraceStatement)

	p.kafka = ctx.Service("kafka").(kafka.Interface)
	w, err := p.kafka.NewProducer(&p.C.Output.Kafka)
	if err != nil {
		return fmt.Errorf("fail to create kafka producer: %s", err)
	}
	p.output.kafka = w
	return nil
}

// Start .
func (p *provider) Start() error {
	return p.kafka.NewConsumer(&p.C.Input, p.invoke)
}

func (p *provider) Close() error {
	p.L.Debug("not support close kafka consumer")
	return nil
}

func init() {
	servicehub.RegisterProvider("trace-storage", &define{})
}
