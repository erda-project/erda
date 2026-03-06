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
	"testing"
	"time"
)

func TestDoPost_NormalizeModelRetry(t *testing.T) {
	t.Setenv("AI_PROXY_MODEL_RETRY_RETRYABLE_HTTP_STATUSES", "")

	cfg := &Config{
		LogLevelStr: "info",
		SelfURL:     "http://127.0.0.1:8081",
		ModelRetry: ModelRetryConfig{
			Conditions: ModelRetryConditions{
				MaxLLMBackendRequestCount: 0,
				Backoff: ModelRetryBackoff{
					Base: -1 * time.Second,
					Max:  0,
				},
				RetryableHTTPStatuses: []int{504, 429, 504, 99, 600},
			},
		},
	}

	if err := cfg.DoPost(); err != nil {
		t.Fatalf("DoPost failed: %v", err)
	}

	if cfg.ModelRetry.Conditions.MaxLLMBackendRequestCount != 1 {
		t.Fatalf("expected max request count normalized to 1, got %d", cfg.ModelRetry.Conditions.MaxLLMBackendRequestCount)
	}
	if cfg.ModelRetry.Conditions.Backoff.Base != 0 {
		t.Fatalf("expected negative backoff base normalized to 0, got %s", cfg.ModelRetry.Conditions.Backoff.Base)
	}
	if cfg.ModelRetry.Conditions.Backoff.Max != 10*time.Second {
		t.Fatalf("expected backoff max default 10s, got %s", cfg.ModelRetry.Conditions.Backoff.Max)
	}
	if got, want := cfg.ModelRetry.Conditions.RetryableHTTPStatuses, []int{429, 504}; len(got) != len(want) || got[0] != want[0] || got[1] != want[1] {
		t.Fatalf("expected normalized statuses %v, got %v", want, got)
	}
}

func TestDoPost_ModelRetryRetryableHTTPStatusesEnvOverride(t *testing.T) {
	t.Setenv("AI_PROXY_MODEL_RETRY_RETRYABLE_HTTP_STATUSES", "503, 429,abc,429,,600,502")

	cfg := &Config{
		LogLevelStr: "info",
		SelfURL:     "http://127.0.0.1:8081",
		ModelRetry: ModelRetryConfig{
			Conditions: ModelRetryConditions{
				RetryableHTTPStatuses: []int{504},
			},
		},
	}

	if err := cfg.DoPost(); err != nil {
		t.Fatalf("DoPost failed: %v", err)
	}

	want := []int{429, 502, 503}
	got := cfg.ModelRetry.Conditions.RetryableHTTPStatuses
	if len(got) != len(want) {
		t.Fatalf("expected %v, got %v", want, got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("expected %v, got %v", want, got)
		}
	}
}
