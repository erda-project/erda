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

package health

import (
	"strings"
	"time"
)

type Config struct {
	Enabled bool        `file:"enabled" env:"MODEL_HEALTH_ENABLED" default:"true"`
	Probe  ProbeConfig  `file:"probe"`
	Rescue RescueConfig `file:"rescue"`
}

type ProbeConfig struct {
	BaseURL      string        `file:"base_url" env:"MODEL_HEALTH_PROBE_BASE_URL"`
	Timeout      time.Duration `file:"timeout" env:"MODEL_HEALTH_PROBE_TIMEOUT" default:"10s"`
	UnhealthyTTL time.Duration `file:"unhealthy_ttl" env:"MODEL_HEALTH_UNHEALTHY_TTL" default:"1h"`
}

type RescueConfig struct {
	InitialBackoff time.Duration `file:"initial_backoff" env:"MODEL_HEALTH_RESCUE_INITIAL_BACKOFF" default:"3s"`
	MaxBackoff     time.Duration `file:"max_backoff" env:"MODEL_HEALTH_RESCUE_MAX_BACKOFF" default:"2m"`
}

func (cfg *Config) normalize() {
	if !cfg.Enabled {
		return
	}
	if strings.TrimSpace(cfg.Probe.BaseURL) == "" {
		panic("model health probe base_url must not be empty")
	}
	if cfg.Probe.UnhealthyTTL <= 0 {
		cfg.Probe.UnhealthyTTL = time.Hour
	}
	if cfg.Probe.Timeout <= 0 {
		cfg.Probe.Timeout = 10 * time.Second
	}
	if cfg.Rescue.InitialBackoff <= 0 {
		panic("model health rescue initial_backoff must be > 0")
	}
	if cfg.Rescue.MaxBackoff <= 0 {
		panic("model health rescue max_backoff must be > 0")
	}
}
