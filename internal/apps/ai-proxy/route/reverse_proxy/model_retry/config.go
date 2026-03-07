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

package model_retry

import (
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Enabled       bool          `file:"enabled" env:"AI_PROXY_MODEL_RETRY_ENABLED" default:"true"`
	Conditions    Conditions    `file:"conditions"`
	Actions       Actions       `file:"actions"`
	Observability Observability `file:"observability"`
}

type Conditions struct {
	MaxLLMBackendRequestCount         int     `file:"max_llm_backend_request_count" env:"AI_PROXY_MODEL_RETRY_MAX_LLM_BACKEND_REQUEST_COUNT" default:"3"`
	Backoff                           Backoff `file:"backoff"`
	RetryableHTTPStatuses             []int   `file:"retryable_http_statuses"`
	RetryableHTTPStatusesRaw          string  `file:"-" env:"AI_PROXY_MODEL_RETRY_RETRYABLE_HTTP_STATUSES"`
	MatchNetworkIssueFromResponseBody bool    `file:"match_network_issue_from_response_body" env:"AI_PROXY_MODEL_RETRY_MATCH_NETWORK_ISSUE_FROM_RESPONSE_BODY" default:"false"`
}

type Backoff struct {
	// Base is the retry backoff base duration.
	// Retry #1 waits 1*Base, retry #2 waits 3*Base, retry #3 waits 7*Base, ...
	Base time.Duration `file:"base" env:"AI_PROXY_MODEL_RETRY_BACKOFF_BASE" default:"1s"`
	Max  time.Duration `file:"max" default:"10s"`
}

type Actions struct {
	// ExcludeFailedInstance only affects the retry layer's per-request exclusion set.
	// It does not override model health filtering. When model health is enabled,
	// a failed instance may already be filtered out before the next attempt.
	ExcludeFailedInstance bool `file:"exclude_failed_instance" env:"AI_PROXY_MODEL_RETRY_EXCLUDE_FAILED_INSTANCE" default:"true"`
}

type Observability struct {
	ResponseHeaderMeta bool `file:"response_header_meta" default:"true"`
}

func (cfg *Config) Normalize() {
	if raw, ok := os.LookupEnv("AI_PROXY_MODEL_RETRY_RETRYABLE_HTTP_STATUSES"); ok {
		cfg.Conditions.RetryableHTTPStatusesRaw = raw
	}
	if cfg.Conditions.MaxLLMBackendRequestCount <= 0 {
		cfg.Conditions.MaxLLMBackendRequestCount = 1
	}
	if cfg.Conditions.Backoff.Base < 0 {
		cfg.Conditions.Backoff.Base = 0
	}
	if cfg.Conditions.Backoff.Max <= 0 {
		cfg.Conditions.Backoff.Max = 10 * time.Second
	}
	cfg.Conditions.RetryableHTTPStatuses = normalizeHTTPStatusCodes(resolveRetryableHTTPStatuses(cfg.Conditions))
}

func resolveRetryableHTTPStatuses(cfg Conditions) []int {
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
