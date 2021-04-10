//  Copyright (c) 2021 Terminus, Inc.
//
//  This program is free software: you can use, redistribute, and/or modify
//  it under the terms of the GNU Affero General Public License, version 3
//  or later ("AGPL"), as published by the Free Software Foundation.
//
//  This program is distributed in the hope that it will be useful, but WITHOUT
//  ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
//  FITNESS FOR A PARTICULAR PURPOSE.
//
//  You should have received a copy of the GNU Affero General Public License
//  along with this program. If not, see <http://www.gnu.org/licenses/>.

package conf

import (
	"github.com/erda-project/erda/pkg/envconf"
)

// Conf defines env configs.
type Conf struct {
	Debug        bool   `env:"DEBUG" default:"false"`
	ListenAddr   string `env:"LISTEN_ADDR" default:":9500"`
	ClientID     string `env:"CLIENT_ID" default:"action-runner"`
	ClientSecret string `env:"CLIENT_SECRET" default:"devops/action-runner"`
	RunnerUserID string `env:"RUNNER_USER_ID" default:"1111"`
}

var cfg Conf

// Load load env configs.
func Load() {
	envconf.MustLoad(&cfg)
}

// ListenAddr return ListenAddr .
func ListenAddr() string {
	return cfg.ListenAddr
}

func ClientID() string {
	return cfg.ClientID
}

func ClientSecret() string {
	return cfg.ClientSecret
}

func RunnerUserID() string {
	return cfg.RunnerUserID
}
