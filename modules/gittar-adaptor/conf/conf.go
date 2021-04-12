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

// Package conf 配置文件信息
package conf

import (
	"github.com/erda-project/erda/pkg/envconf"
)

var cfg Conf

// Conf Adaptor 环境变量配置项
type Conf struct {
	ListenAddr  string `env:"LISTEN_ADDR" default:":1086"`
	Debug       bool   `env:"DEBUG" default:"false"`
	SelfAddr    string `env:"SELF_ADDR"`
	MonitorAddr string `env:"MONITOR_ADDR"`
	GittarAddr  string `env:"GITTAR_ADDR"`
}

// Load 加载环境变量配置.
func Load() {
	envconf.MustLoad(&cfg)
}

// ListenAddr 返回 ListenAddr 选项.
func ListenAddr() string {
	return cfg.ListenAddr
}

// Debug 是否处于调试模式
func Debug() bool {
	return cfg.Debug
}

// SelfAddr 返回 SELF_ADDR
func SelfAddr() string {
	return cfg.SelfAddr
}

// MonitorAddr 返回 monitor 地址
func MonitorAddr() string {
	return cfg.MonitorAddr
}

// GittarAddr 返回 gittar 的集群内部地址.
func GittarAddr() string {
	return cfg.GittarAddr
}
