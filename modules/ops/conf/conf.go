// Package conf Define the configuration
package conf

import (
	"github.com/erda-project/erda/pkg/envconf"
)

// Conf Define the configuration
type Conf struct {
	Debug         bool   `env:"DEBUG" default:"false"`
	ListenAddr    string `env:"LISTEN_ADDR" default:":9027"`
	SoldierAddr   string `env:"SOLDIER_ADDR"`
	SchedulerAddr string `env:"SCHEDULER_ADDR"`
}

var cfg Conf

// Load Load envs
func Load() {
	envconf.MustLoad(&cfg)
}

// ListenAddr return the address of listen.
func ListenAddr() string {
	return cfg.ListenAddr
}

// SoldierAddr return the address of soldier.
func SoldierAddr() string {
	return cfg.SoldierAddr
}

// SchedulerAddr Return the address of scheduler.
func SchedulerAddr() string {
	return cfg.SchedulerAddr
}

// Debug Return the switch of debug.
func Debug() bool {
	return cfg.Debug
}
