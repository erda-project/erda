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
	"context"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/policy_group/health"
	"github.com/erda-project/erda/internal/apps/ai-proxy/vars"
)

func (cfg Config) WithRequestOverrides(r *http.Request) Config {
	cfg.Normalize()
	if r == nil {
		return cfg
	}

	logger := ctxhelper.MustGetLoggerBase(r.Context())

	if raw := strings.TrimSpace(r.Header.Get(vars.XAIProxyModelRetry)); raw != "" {
		if v, ok := parseHeaderBool(raw); ok {
			cfg.Enabled = v
		} else {
			logger.Warnf("invalid %s=%q", vars.XAIProxyModelRetry, raw)
		}
	}
	if raw := strings.TrimSpace(r.Header.Get(vars.XAIProxyModelRetryDisabled)); raw != "" {
		if v, ok := parseHeaderBool(raw); ok {
			if v {
				cfg.Enabled = false
			}
		} else {
			logger.Warnf("invalid %s=%q", vars.XAIProxyModelRetryDisabled, raw)
		}
	}
	if raw := strings.TrimSpace(r.Header.Get(vars.XAIProxyModelRetryMax)); raw != "" {
		if maxRequestCount, err := strconv.Atoi(raw); err == nil && maxRequestCount > 0 {
			cfg.Conditions.MaxLLMBackendRequestCount = maxRequestCount
		} else {
			logger.Warnf("invalid %s=%q", vars.XAIProxyModelRetryMax, raw)
		}
	}
	if health.IsHealthProbeRequest(r.Header) {
		cfg.Enabled = false
	}

	return cfg
}

func (cfg Config) IsRetryableHTTPStatus(statusCode int) bool {
	for _, code := range cfg.Conditions.RetryableHTTPStatuses {
		if code == statusCode {
			return true
		}
	}
	return false
}

func (cfg Config) NextBackoff(rawLLMBackendRequestCount int) time.Duration {
	if rawLLMBackendRequestCount <= 0 || cfg.Conditions.Backoff.Base <= 0 {
		return 0
	}
	if rawLLMBackendRequestCount > 60 {
		rawLLMBackendRequestCount = 60
	}
	multiplier := int64((1 << rawLLMBackendRequestCount) - 1)
	delay := time.Duration(multiplier) * cfg.Conditions.Backoff.Base
	if cfg.Conditions.Backoff.Max > 0 && delay > cfg.Conditions.Backoff.Max {
		return cfg.Conditions.Backoff.Max
	}
	return delay
}

func AddExcludedModelID(ctx context.Context, modelID string) {
	if modelID == "" {
		return
	}
	excluded, _ := GetExcludedModelIDs(ctx)
	cloned := make(map[string]struct{}, len(excluded)+1)
	for id := range excluded {
		cloned[id] = struct{}{}
	}
	cloned[modelID] = struct{}{}
	ctxhelper.PutModelRetryExcludedModelIDs(ctx, cloned)
}

func GetExcludedModelIDs(ctx context.Context) (map[string]struct{}, bool) {
	excluded, ok := ctxhelper.GetModelRetryExcludedModelIDs(ctx)
	if !ok || excluded == nil {
		return nil, false
	}
	ret, ok := excluded.(map[string]struct{})
	if !ok || ret == nil {
		return nil, false
	}
	return ret, true
}

func SortedStatusCodes(codes map[int]struct{}) []int {
	ret := make([]int, 0, len(codes))
	for code := range codes {
		ret = append(ret, code)
	}
	sort.Ints(ret)
	return ret
}

func (cfg Config) SortedRetryableHTTPStatuses() []int {
	ret := append([]int(nil), cfg.Conditions.RetryableHTTPStatuses...)
	sort.Ints(ret)
	return ret
}

func parseHeaderBool(raw string) (bool, bool) {
	v, err := strconv.ParseBool(strings.TrimSpace(raw))
	if err != nil {
		return false, false
	}
	return v, true
}
