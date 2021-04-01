package conf

import (
	"github.com/erda-project/erda/pkg/envconf"
)

// Conf scheduler conf, 使用 envconf 加载配置
type Conf struct {
	// Debug 控制日志等级
	Debug bool `env:"DEBUG" default:"false"`
	// PoolSize goroutine pool size
	PoolSize int `env:"POOL_SIZE" default:"50"`
	// ListenAddr scheduler 监听地址, eg: ":9091"
	ListenAddr             string `env:"LISTEN_ADDR" default:":9091"`
	DefaultRuntimeExecutor string `env:"DEFAULT_RUNTIME_EXECUTOR" default:"MARATHON"`
	// TraceLogEnv shows the key of environment variable defined for tracing log
	TraceLogEnv string `env:"TRACELOGENV" default:"TERMINUS_DEFINE_TAG"`
	// PlaceHolderImage 打散部署service时，用于占位的镜像
	PlaceHolderImage string `env:"PLACEHOLDER_IMAGE" default:"registry.cn-hangzhou.aliyuncs.com/terminus/busybox"`

	KafkaBrokers        string `env:"BOOTSTRAP_SERVERS"`
	KafkaContainerTopic string `env:"CMDB_CONTAINER_TOPIC"`
	KafkaGroup          string `env:"CMDB_GROUP"`

	TerminalSecurity bool `env:"TERMINAL_SECURITY" default:"false"`
}

var cfg Conf

// Load 加载环境变量配置.
func Load() {
	envconf.MustLoad(&cfg)
}

// Debug return cfg.Debug
func Debug() bool {
	return cfg.Debug
}

// PoolSize return cfg.PoolSize
func PoolSize() int {
	return cfg.PoolSize
}

// ListenAddr return cfg.ListenAddr
func ListenAddr() string {
	return cfg.ListenAddr
}

// DefaultRuntimeExecutor return cfg.DefaultRuntimeExecutor
func DefaultRuntimeExecutor() string {
	return cfg.DefaultRuntimeExecutor
}

// TraceLogEnv return cfg.TraceLogEnv
func TraceLogEnv() string {
	return cfg.TraceLogEnv
}

// PlaceHolderImage return cfg.PlaceHolderImage
func PlaceHolderImage() string {
	return cfg.PlaceHolderImage
}

func KafkaBrokers() string {
	return cfg.KafkaBrokers
}
func KafkaContainerTopic() string {
	return cfg.KafkaContainerTopic
}
func KafkaGroup() string {
	return cfg.KafkaGroup
}

// TerminalSecurity return cfg.TerminalSecurity
func TerminalSecurity() bool {
	return cfg.TerminalSecurity
}
