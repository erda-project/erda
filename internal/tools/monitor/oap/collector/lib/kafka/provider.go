package kafka

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/Shopify/sarama"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
)

type Interface interface {
	NewProducer() (*AsyncProducer, error)
	NewConsumerGroup(c *ConsumerConfig, handler ConsumerFunc, options ...ConsumerOption) (*ConsumerGroupManager, error)
	NewAdminClient() (AdminInterface, error)
	Brokers() []string
}

type config struct {
	Servers         string `file:"servers" env:"BOOTSTRAP_SERVERS" default:"localhost:9092" desc:"kafka servers"`
	ClientID        string `file:"client_id" env:"COMPONENT_NAME" default:"sarama" desc:"kafka client name"`
	DebugClient     bool   `file:"debug_client" env:"KAFKA_DEBUG_CLIENT" desc:"log sarama client log to console"`
	ProtocolVersion string `file:"protocol_version" default:"1.1.0" desc:"kafka broker protocol version"`

	Producer *producerConfig `file:"producer"`
}

var _ Interface = (*provider)(nil)

// +provider
type provider struct {
	Cfg          *config
	Log          logs.Logger
	protoVersion sarama.KafkaVersion

	producer *AsyncProducer
}

func (p *provider) Init(ctx servicehub.Context) error {
	if p.Cfg.DebugClient {
		sarama.Logger = log.New(os.Stdout, "[Sarama] ", log.LstdFlags)
	}
	switch p.Cfg.ProtocolVersion {
	case "1.1.0":
		p.protoVersion = sarama.V1_1_0_0
	default:
		return fmt.Errorf("invalid version: %q", p.Cfg.ProtocolVersion)
	}
	return nil
}

func (p *provider) NewAdminClient() (AdminInterface, error) {
	panic("implement me")
}

func (p *provider) Brokers() []string {
	return strings.Split(p.Cfg.Servers, ",")
}

// Provide .
func (p *provider) Provide(ctx servicehub.DependencyContext, options ...interface{}) interface{} {
	return p
}

func init() {
	servicehub.Register("kafkago", &servicehub.Spec{
		Services: []string{"kafkago"},
		ConfigFunc: func() interface{} {
			return &config{}
		},
		Creator: func() servicehub.Provider {
			return &provider{}
		},
	})
}
