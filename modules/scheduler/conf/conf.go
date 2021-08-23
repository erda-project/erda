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

package conf

import (
	"github.com/erda-project/erda/pkg/envconf"
)

// Conf scheduler conf, use envconf to load configuration
type Conf struct {
	// Debug Control log level
	Debug bool `env:"DEBUG" default:"false"`
	// PoolSize goroutine pool size
	PoolSize int `env:"POOL_SIZE" default:"50"`
	// ListenAddr scheduler listening address , eg: ":9091"
	ListenAddr             string `env:"LISTEN_ADDR" default:":9091"`
	DefaultRuntimeExecutor string `env:"DEFAULT_RUNTIME_EXECUTOR" default:"MARATHON"`
	// TraceLogEnv shows the key of environment variable defined for tracing log
	TraceLogEnv string `env:"TRACELOGENV" default:"TERMINUS_DEFINE_TAG"`
	// PlaceHolderImage Image used to occupy the seat when disassembling the service deployment
	PlaceHolderImage string `env:"PLACEHOLDER_IMAGE" default:"registry.cn-hangzhou.aliyuncs.com/terminus/busybox"`

	KafkaBrokers        string `env:"BOOTSTRAP_SERVERS"`
	KafkaContainerTopic string `env:"CMDB_CONTAINER_TOPIC"`
	KafkaGroup          string `env:"CMDB_GROUP"`

	TerminalSecurity bool `env:"TERMINAL_SECURITY" default:"false"`

	WsDiceRootDomain string `env:"WS_DICE_ROOT_DOMAIN" default:"app.terminus.io,erda.cloud"`
}

var cfg Conf

// Load environment variable
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

func WsDiceRootDomain() string {
	return cfg.WsDiceRootDomain
}
