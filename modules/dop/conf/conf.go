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

import (
	"github.com/erda-project/erda/pkg/envconf"
)

// Conf define envs
type Conf struct {
	Debug          bool   `env:"DEBUG" default:"false"`
	ListenAddr     string `env:"LISTEN_ADDR" default:":9527"`
	UCClientID     string `default:"dice" env:"UC_CLIENT_ID"`
	UCClientSecret string `default:"secret" env:"UC_CLIENT_SECRET"`
	WildDomain     string `default:"dev.terminus.io" env:"DICE_ROOT_DOMAIN"`

	MonitorAddr      string `env:"MONITOR_ADDR"`
	GittarAddr       string `env:"GITTAR_ADDR"`
	BundleTimeoutSec int    `env:"BUNDLE_TIMEOUT_SECOND" default:"30"`

	ConsumerNum       int    `env:"CONSUMER_NUM" default:"5"`
	DiceClusterName   string `env:"DICE_CLUSTER_NAME" required:"true"`
	EventboxAddr      string `env:"EVENTBOX_ADDR"`
	CMDBAddr          string `env:"CMDB_ADDR"`
	PipelineAddr      string `env:"PIPELINE_ADDR"`
	NexusAddr         string `env:"NEXUS_ADDR" required:"true"`
	NexusUsername     string `env:"NEXUS_USERNAME" required:"true"`
	NexusPassword     string `env:"NEXUS_PASSWORD" required:"true"`
	SonarAddr         string `env:"SONAR_ADDR" required:"true"`
	SonarPublicURL    string `env:"SONAR_PUBLIC_URL" required:"true"`
	SonarAdminToken   string `env:"SONAR_ADMIN_TOKEN" required:"true"` // dice.yml 里依赖了 sonar，由工具链注入 SONAR_ADMIN_TOKEN
	GolangCILintImage string `env:"GOLANGCI_LINT_IMAGE" default:"registry.cn-hangzhou.aliyuncs.com/terminus/terminus-golangci-lint:1.27"`
}

var cfg Conf

// Load loads envs
func Load() {
	envconf.MustLoad(&cfg)
}

func Debug() bool {
	return cfg.Debug
}

func ListenAddr() string {
	return cfg.ListenAddr
}

func UCClientID() string {
	return cfg.UCClientID
}

func UCClientSecret() string {
	return cfg.UCClientSecret
}

func WildDomain() string {
	return cfg.WildDomain
}

func SuperUserID() string {
	return "1100"
}

func MonitorAddr() string {
	return cfg.MonitorAddr
}

func GittarAddr() string {
	return cfg.GittarAddr
}

func BundleTimeoutSecond() int {
	return cfg.BundleTimeoutSec
}

func ConsumerNum() int {
	return cfg.ConsumerNum
}

func NexusAddr() string {
	return cfg.NexusAddr
}

func NexusUsername() string {
	return cfg.NexusUsername
}

func NexusPassword() string {
	return cfg.NexusPassword
}

func DiceClusterName() string {
	return cfg.DiceClusterName
}

func SonarAddr() string {
	return cfg.SonarAddr
}

func SonarPublicURL() string {
	return cfg.SonarPublicURL
}

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
