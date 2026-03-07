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

type Policy struct {
	Enabled                           bool
	MaxLLMBackendRequestCount         int
	BackoffBase                       time.Duration
	BackoffMax                        time.Duration
	RetryableHTTPStatuses             map[int]struct{}
	MatchNetworkIssueFromResponseBody bool
	ExcludeFailedInstance             bool
	ResponseHeaderMetaEnabled         bool
}

func ResolvePolicy(r *http.Request, cfg Config) Policy {
	policy := Policy{
		Enabled:                           cfg.Enabled,
		MaxLLMBackendRequestCount:         cfg.Conditions.MaxLLMBackendRequestCount,
		BackoffBase:                       cfg.Conditions.Backoff.Base,
		BackoffMax:                        cfg.Conditions.Backoff.Max,
		RetryableHTTPStatuses:             toStatusCodeSet(cfg.Conditions.RetryableHTTPStatuses),
		MatchNetworkIssueFromResponseBody: cfg.Conditions.MatchNetworkIssueFromResponseBody,
		ExcludeFailedInstance:             cfg.Actions.ExcludeFailedInstance,
		ResponseHeaderMetaEnabled:         cfg.Observability.ResponseHeaderMeta,
	}
	if policy.MaxLLMBackendRequestCount <= 0 {
		policy.MaxLLMBackendRequestCount = 1
	}
	if r == nil {
		return policy
	}

	logger, _ := ctxhelper.GetLoggerBase(r.Context())

	if raw := strings.TrimSpace(r.Header.Get(vars.XAIProxyRetry)); raw != "" {
		if v, ok := parseHeaderBool(raw); ok {
			policy.Enabled = v
		} else if logger != nil {
			logger.Warnf("invalid %s=%q", vars.XAIProxyRetry, raw)
		}
	}
	if raw := strings.TrimSpace(r.Header.Get(vars.XAIProxyRetryDisabled)); raw != "" {
		if v, ok := parseHeaderBool(raw); ok {
			if v {
				policy.Enabled = false
			}
		} else if logger != nil {
			logger.Warnf("invalid %s=%q", vars.XAIProxyRetryDisabled, raw)
		}
	}
	if raw := strings.TrimSpace(r.Header.Get(vars.XAIProxyRetryMax)); raw != "" {
		if maxRequestCount, err := strconv.Atoi(raw); err == nil && maxRequestCount > 0 {
			policy.MaxLLMBackendRequestCount = maxRequestCount
		} else if logger != nil {
			logger.Warnf("invalid %s=%q", vars.XAIProxyRetryMax, raw)
		}
	}
	if health.IsHealthProbeRequest(r.Header) {
		policy.Enabled = false
	}

	return policy
}

func (p Policy) IsRetryableHTTPStatus(statusCode int) bool {
	_, ok := p.RetryableHTTPStatuses[statusCode]
	return ok
}

func (p Policy) NextBackoff(rawLLMBackendRequestCount int) time.Duration {
	if rawLLMBackendRequestCount <= 0 || p.BackoffBase <= 0 {
		return 0
	}
	if rawLLMBackendRequestCount > 60 {
		rawLLMBackendRequestCount = 60
	}
	multiplier := int64((1 << rawLLMBackendRequestCount) - 1)
	delay := time.Duration(multiplier) * p.BackoffBase
	if p.BackoffMax > 0 && delay > p.BackoffMax {
		return p.BackoffMax
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

func toStatusCodeSet(codes []int) map[int]struct{} {
	ret := make(map[int]struct{}, len(codes))
	for _, code := range codes {
		if code < 100 || code > 599 {
			continue
		}
		ret[code] = struct{}{}
	}
	return ret
}

func parseHeaderBool(raw string) (bool, bool) {
	v, err := strconv.ParseBool(strings.TrimSpace(raw))
	if err != nil {
		return false, false
	}
	return v, true
}
