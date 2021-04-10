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

// Package conf 定义配置选项
package conf

import (
	"github.com/erda-project/erda/pkg/envconf"
)

// Conf 定义基于环境变量的配置项
type Conf struct {
	Debug      bool   `env:"DEBUG" default:"false"`
	ListenAddr string `env:"LISTEN_ADDR" default:":3083"`

	UCClientID     string `default:"dice" env:"UC_CLIENT_ID"`
	UCClientSecret string `default:"secret" env:"UC_CLIENT_SECRET"`

	WildDomain string `default:"dev.terminus.io" env:"DICE_ROOT_DOMAIN"`
}

var cfg Conf

// Load 加载环境变量配置.
func Load() {
	envconf.MustLoad(&cfg)
}

// ListenAddr 返回 ListenAddr 选项.
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

func Debug() bool {
	return cfg.Debug
}
