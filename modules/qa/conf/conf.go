// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package conf

import "github.com/erda-project/erda/pkg/envconf"

var cfg Conf

// Conf qa server config info
type Conf struct {
	Debug       bool   `env:"DEBUG" default:"false"`
	ListenAddr  string `env:"LISTEN_ADDR" default:":3033"`
	ConsumerNum int    `env:"CONSUMER_NUM" default:"5"`

	DiceClusterName string `env:"DICE_CLUSTER_NAME" required:"true"`

	SelfAddr     string `env:"SELF_ADDR"`
	GittarAddr   string `env:"GITTAR_ADDR"`
	EventboxAddr string `env:"EVENTBOX_ADDR"`
	CMDBAddr     string `env:"CMDB_ADDR"`
	PipelineAddr string `env:"PIPELINE_ADDR"`

	NexusAddr     string `env:"NEXUS_ADDR" required:"true"`
	NexusUsername string `env:"NEXUS_USERNAME" required:"true"`
	NexusPassword string `env:"NEXUS_PASSWORD" required:"true"`

	SonarAddr       string `env:"SONAR_ADDR" required:"true"`
	SonarPublicURL  string `env:"SONAR_PUBLIC_URL" required:"true"`
	SonarAdminToken string `env:"SONAR_ADMIN_TOKEN" required:"true"` // dice.yml 里依赖了 sonar，由工具链注入 SONAR_ADMIN_TOKEN

	GolangCILintImage string `env:"GOLANGCI_LINT_IMAGE" default:"registry.cn-hangzhou.aliyuncs.com/terminus/terminus-golangci-lint:1.27"`
}

// Load load qa config
func Load() {
	envconf.MustLoad(&cfg)
}

// Debug enable log debug level
func Debug() bool {
	return cfg.Debug
}

// ListenAddr
func ListenAddr() string {
	return cfg.ListenAddr
}

// SelfAddr
func SelfAddr() string {
	return cfg.SelfAddr
}

// GittarAddr
func GittarAddr() string {
	return cfg.GittarAddr
}

// ConsumerNum
func ConsumerNum() int {
	return cfg.ConsumerNum
}

// NexusAddr
func NexusAddr() string {
	return cfg.NexusAddr
}

// NexusUsername
func NexusUsername() string {
	return cfg.NexusUsername
}

// NexusPassword
func NexusPassword() string {
	return cfg.NexusPassword
}

// DiceClusterName 返回 qa 组件运行的中心集群名
func DiceClusterName() string {
	return cfg.DiceClusterName
}

// SonarAddr
func SonarAddr() string {
	return cfg.SonarAddr
}

// SonarPublicURL
func SonarPublicURL() string {
	return cfg.SonarPublicURL
}

// SonarAdminToken
func SonarAdminToken() string {
	return cfg.SonarAdminToken
}

func EventboxAddr() string {
	return cfg.EventboxAddr
}

func GolangCILintImage() string {
	return cfg.GolangCILintImage
}

func CMDBAddr() string {
	return cfg.CMDBAddr
}
