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
	"os"
	"sort"
	"strconv"
	"strings"
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
	Enabled       bool                    `file:"enabled" env:"AI_PROXY_MODEL_RETRY_ENABLED" default:"true"`
	Conditions    ModelRetryConditions    `file:"conditions"`
	Actions       ModelRetryActions       `file:"actions"`
	Observability ModelRetryObservability `file:"observability"`
}

type ModelRetryConditions struct {
	MaxLLMBackendRequestCount         int               `file:"max_llm_backend_request_count" env:"AI_PROXY_MODEL_RETRY_MAX_LLM_BACKEND_REQUEST_COUNT" default:"3"`
	Backoff                           ModelRetryBackoff `file:"backoff"`
	RetryableHTTPStatuses             []int             `file:"retryable_http_statuses"`
	RetryableHTTPStatusesRaw          string            `file:"-" env:"AI_PROXY_MODEL_RETRY_RETRYABLE_HTTP_STATUSES"`
	MatchNetworkIssueFromResponseBody bool              `file:"match_network_issue_from_response_body" env:"AI_PROXY_MODEL_RETRY_MATCH_NETWORK_ISSUE_FROM_RESPONSE_BODY" default:"false"`
}

type ModelRetryBackoff struct {
	// Base is the retry backoff base duration.
	// Retry #1 waits 1*Base, retry #2 waits 3*Base, retry #3 waits 7*Base, ...
	Base time.Duration `file:"base" env:"AI_PROXY_MODEL_RETRY_BACKOFF_BASE" default:"1s"`
	Max  time.Duration `file:"max" default:"10s"`
}

type ModelRetryActions struct {
	// ExcludeFailedInstance only affects the retry layer's per-request exclusion set.
	// It does not override model health filtering. When model health is enabled,
	// a failed instance may already be filtered out before the next attempt.
	ExcludeFailedInstance bool `file:"exclude_failed_instance" env:"AI_PROXY_MODEL_RETRY_EXCLUDE_FAILED_INSTANCE" default:"true"`
}

type ModelRetryObservability struct {
	ResponseHeaderMeta bool `file:"response_header_meta" default:"true"`
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
	if raw, ok := os.LookupEnv("AI_PROXY_MODEL_RETRY_RETRYABLE_HTTP_STATUSES"); ok {
		cfg.ModelRetry.Conditions.RetryableHTTPStatusesRaw = raw
	}
	if cfg.ModelRetry.Conditions.MaxLLMBackendRequestCount <= 0 {
		cfg.ModelRetry.Conditions.MaxLLMBackendRequestCount = 1
	}
	if cfg.ModelRetry.Conditions.Backoff.Base < 0 {
		cfg.ModelRetry.Conditions.Backoff.Base = 0
	}
	if cfg.ModelRetry.Conditions.Backoff.Max <= 0 {
		cfg.ModelRetry.Conditions.Backoff.Max = 10 * time.Second
	}
	cfg.ModelRetry.Conditions.RetryableHTTPStatuses = normalizeHTTPStatusCodes(resolveRetryableHTTPStatuses(cfg.ModelRetry.Conditions))
}

func resolveRetryableHTTPStatuses(cfg ModelRetryConditions) []int {
	if strings.TrimSpace(cfg.RetryableHTTPStatusesRaw) == "" {
		return cfg.RetryableHTTPStatuses
	}
	parts := strings.Split(cfg.RetryableHTTPStatusesRaw, ",")
	statuses := make([]int, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		code, err := strconv.Atoi(part)
		if err != nil {
			continue
		}
		statuses = append(statuses, code)
	}
	return statuses
}

func normalizeHTTPStatusCodes(codes []int) []int {
	if len(codes) == 0 {
		return nil
	}
	statuses := make([]int, 0, len(codes))
	seen := make(map[int]struct{})
	for _, code := range codes {
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
	return statuses
}
