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
