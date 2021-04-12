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

// Package conf Dicehub 环境配置
package conf

import (
	"github.com/erda-project/erda/pkg/envconf"
)

var cfg Conf

// Conf DiceHub 环境变量配置项
type Conf struct {
	ListenAddr      string              `env:"LISTEN_ADDR" default:":10000"`
	MonitorAddr     string              `env:"MONITOR_ADDR" default:"monitor.default.svc.cluster.local:7096"`
	Debug           bool                `env:"DEBUG" default:"false"`
	OpsAddr         string              `env:"OPS_ADDR"`
	EventBoxAddr    string              `env:"EVENTBOX_ADDR"`
	GCSwitch        bool                `env:"RELEASE_GC_SWITCH" default:"true"`
	MaxTimeReserved string              `env:"RELEASE_MAX_TIME_RESERVED" default:"72"` // default: 72h
	ExtensionMenu   map[string][]string `env:"EXTENSION_MENU" default:"{}"`
	SiteUrl         string              `env:"SITE_URL"`
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

// OpsAddr 返回ops地址
func OpsAddr() string {
	return cfg.OpsAddr
}

func SiteUrl() string {
	return cfg.SiteUrl
}

// EventBoxAddr 返回eventbox地址
func EventBoxAddr() string {
	return cfg.EventBoxAddr
}

// GCSwitch release自动回收开关
func GCSwitch() bool {
	return cfg.GCSwitch
}

// MaxTimeReserved 未使用的release保留最长时间
func MaxTimeReserved() string {
	return cfg.MaxTimeReserved
}

// ExtensionMenu 服务扩展菜单配置
func ExtensionMenu() map[string][]string {
	return cfg.ExtensionMenu
}

// MonitorAddr monitor
func MonitorAddr() string {
	return cfg.MonitorAddr
}
