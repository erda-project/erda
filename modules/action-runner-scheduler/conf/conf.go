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
