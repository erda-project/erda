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
	"sort"
	"time"

	"github.com/sirupsen/logrus"
)

type Config struct {
	LogLevelStr string       `file:"log_level" default:"info" env:"LOG_LEVEL"`
	LogLevel    logrus.Level `json:"-" yaml:"-"`

	SelfURL string `file:"self_url" env:"SELF_URL" required:"true"`

	ModelRetry ModelRetryConfig `file:"model_retry"`

	EmbedRoutesFS    embed.FS
	EmbedTemplatesFS embed.FS
}

type ModelRetryConfig struct {
	Enabled                             bool              `file:"enabled" env:"AI_PROXY_MODEL_RETRY_ENABLED" default:"true"`
	MaxAttempts                         int               `file:"max_attempts" env:"AI_PROXY_MODEL_RETRY_MAX_ATTEMPTS" default:"3"`
	Backoff                             ModelRetryBackoff `file:"backoff"`
	RetryableHTTPStatuses               []int             `file:"retryable_http_statuses"`
	EnableResponseBodyNetworkIssueMatch bool              `file:"enable_response_body_network_issue_match" env:"AI_PROXY_MODEL_RETRY_ENABLE_RESPONSE_BODY_NETWORK_ISSUE_MATCH" default:"false"`
}

type ModelRetryBackoff struct {
	// Base is the retry backoff base duration.
	// Retry #1 waits 1*Base, retry #2 waits 3*Base, retry #3 waits 7*Base, ...
	Base time.Duration `file:"base" env:"AI_PROXY_MODEL_RETRY_BACKOFF_BASE" default:"1s"`
}

var (
	EmbedRoutesFS    embed.FS
	EmbedTemplatesFS embed.FS
)

func InjectEmbedFS(routesFS, TemplatesFS *embed.FS) {
	if routesFS != nil {
		EmbedRoutesFS = *routesFS
	}
	if TemplatesFS != nil {
		EmbedTemplatesFS = *TemplatesFS
	}
}

// DoPost do some post process after config loaded
func (cfg *Config) DoPost() error {
	// routes fs
	cfg.EmbedRoutesFS = EmbedRoutesFS
	cfg.EmbedTemplatesFS = EmbedTemplatesFS

	// parse log level
	level, err := logrus.ParseLevel(cfg.LogLevelStr)
	if err != nil {
		return fmt.Errorf("failed to parse log level, level: %s, err: %v", cfg.LogLevel, err)
	}
	logrus.SetLevel(level)
	cfg.LogLevel = level

	cfg.normalizeModelRetry()

	return nil
}

func (cfg *Config) normalizeModelRetry() {
	if cfg.ModelRetry.MaxAttempts <= 0 {
		cfg.ModelRetry.MaxAttempts = 1
	}
	if cfg.ModelRetry.Backoff.Base < 0 {
		cfg.ModelRetry.Backoff.Base = 0
	}
	if len(cfg.ModelRetry.RetryableHTTPStatuses) == 0 {
		return
	}
	statuses := make([]int, 0, len(cfg.ModelRetry.RetryableHTTPStatuses))
	seen := make(map[int]struct{})
	for _, code := range cfg.ModelRetry.RetryableHTTPStatuses {
		if code < 100 || code > 599 {
			continue
		}
		if _, ok := seen[code]; ok {
			continue
		}
		seen[code] = struct{}{}
		statuses = append(statuses, code)
	}
	sort.Ints(statuses)
	cfg.ModelRetry.RetryableHTTPStatuses = statuses
}
