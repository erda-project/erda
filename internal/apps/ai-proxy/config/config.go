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

package config

import (
	"embed"
	"fmt"

	"github.com/sirupsen/logrus"
)

type Config struct {
	LogLevelStr string       `file:"log_level" default:"info" env:"LOG_LEVEL"`
	LogLevel    logrus.Level `json:"-" yaml:"-"`

	SelfURL           string `file:"self_url" env:"SELF_URL" required:"true"`
	McpProxyPublicURL string `file:"mcp_proxy_public_url" env:"McpProxyPublicURL"`

	EmbedRoutesFS embed.FS
}

var embedRoutesFS embed.FS

func InjectEmbedRoutesFS(in embed.FS) {
	embedRoutesFS = in
}

// DoPost do some post process after config loaded
func (cfg *Config) DoPost() error {
	// routes fs
	cfg.EmbedRoutesFS = embedRoutesFS

	// parse log level
	level, err := logrus.ParseLevel(cfg.LogLevelStr)
	if err != nil {
		return fmt.Errorf("failed to parse log level, level: %s, err: %v", cfg.LogLevel, err)
	}
	logrus.SetLevel(level)
	cfg.LogLevel = level

	return nil
}
