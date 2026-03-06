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

package reverseproxy

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/erda-project/erda-infra/base/logs/logrusx"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/audit/audithelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/requestid"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/filter_define"
	httperror "github.com/erda-project/erda/internal/apps/ai-proxy/route/http_error"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/policy_group/health"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/reverse_proxy"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/router_define"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/transports"
	"github.com/erda-project/erda/internal/apps/ai-proxy/vars"
)

type OptionFunc func(context.Context, *httputil.ReverseProxy)

type transparentRetryPolicy struct {
	Enabled                      bool
	MaxAttempts                  int
	BackoffBase                  time.Duration
	RetryableHTTPStatuses        map[int]struct{}
	EnableResponseBodyIssueMatch bool
}

type proxyAttemptResult struct {
	Err         error
	Retryable   bool
	StatusCode  int
	Suppressed  bool
	InstanceID  string
	WroteHeader bool
}

func WithTransport(transport http.RoundTripper) OptionFunc {
	return func(_ context.Context, proxy *httputil.ReverseProxy) {
		proxy.Transport = transport
	}
}

func WithCtxHelperItems(putAnyFunctions ...func(context.Context)) OptionFunc {
	return func(ctx context.Context, _ *httputil.ReverseProxy) {
		for _, f := range putAnyFunctions {
			f(ctx)
		}
	}
}

func (p *provider) HandleReverseProxyAPI(options ...OptionFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// inject context
		ctx := ctxhelper.InitCtxMapIfNeed(r.Context())
		r = r.WithContext(ctx)

		// create a request-level Logger
		logger := logrusx.New().Sub("reverse-proxy-api")
		baseLogger := logrusx.New().Sub("reverse-proxy-api")
		reqID := requestid.GetOrGenRequestID(r)
		callID := requestid.GetCallID(r)
		logger.Set("req", reqID).Set("call", callID)
		baseLogger.Set("req", reqID).Set("call", callID)
		baseLogger.Infof("reverse proxy handler: %s %s", r.Method, r.URL.String())
		ctxhelper.PutLogger(ctx, logger)
		ctxhelper.PutLoggerBase(ctx, baseLogger)
		ctxhelper.PutRequestID(ctx, reqID)
		ctxhelper.PutGeneratedCallID(ctx, callID)

		// find best matched route using priority algorithm
		matched := p.Router.FindBestMatch(r.Method, r.URL.Path)
		var matchedRoute *router_define.Route
		if matched != nil {
			matchedRoute = matched.(*router_define.Route)
		}
		if matchedRoute == nil {
			httperror.NewHTTPError(r.Context(), http.StatusNotFound, "no matched route").WriteJSONHTTPError(w)
			return
		}

		// get all route filters
		var requestFilters []filter_define.FilterWithName[filter_define.ProxyRequestRewriter]
		for _, filterConfig := range matchedRoute.RequestFilters {
			creator, ok := filter_define.FilterFactory.RequestFilters[filterConfig.Name]
			if !ok {
				httperror.NewHTTPError(r.Context(), http.StatusInternalServerError, fmt.Sprintf("request filter %s not found", filterConfig.Name)).WriteJSONHTTPError(w)
				return
			}
			fc, err := filterConfig.GetConfigAsRawMessage()
			if err != nil {
				httperror.NewHTTPError(r.Context(), http.StatusInternalServerError, fmt.Sprintf("failed to convert config for filter %s: %v", filterConfig.Name, err)).WriteJSONHTTPError(w)
				return
			}
			f := creator(filterConfig.Name, fc)
			requestFilters = append(requestFilters, filter_define.FilterWithName[filter_define.ProxyRequestRewriter]{Name: filterConfig.Name, Instance: f})
		}
		var responseFilters []filter_define.FilterWithName[filter_define.ProxyResponseModifier]
		for _, filterConfig := range matchedRoute.ResponseFilters {
			creator, ok := filter_define.FilterFactory.ResponseFilters[filterConfig.Name]
			if !ok {
				httperror.NewHTTPError(r.Context(), http.StatusInternalServerError, fmt.Sprintf("response filter %s not found", filterConfig.Name)).WriteJSONHTTPError(w)
				return
			}
			fc, err := filterConfig.GetConfigAsRawMessage()
			if err != nil {
				httperror.NewHTTPError(r.Context(), http.StatusInternalServerError, fmt.Sprintf("failed to convert config for filter %s: %v", filterConfig.Name, err)).WriteJSONHTTPError(w)
				return
			}
			f := creator(filterConfig.Name, fc)
			responseFilters = append(responseFilters, filter_define.FilterWithName[filter_define.ProxyResponseModifier]{Name: filterConfig.Name, Instance: f})
		}

		ctxhelper.PutDBClient(ctx, p.Dao)
		ctxhelper.PutPathMatcher(ctx, matchedRoute.GetPathMatcher())
		ctxhelper.PutCacheManager(ctx, p.cacheManager)

		tw := reverse_proxy.NewTrackedResponseWriter(w)
		policy := p.resolveTransparentRetryPolicy(r)
		audithelper.Note(ctx, "reverse_proxy.retry.enabled", policy.Enabled)
		audithelper.Note(ctx, "reverse_proxy.retry.max_attempts", policy.MaxAttempts)
		if policy.Enabled {
			audithelper.Note(ctx, "reverse_proxy.retry.backoff_base", policy.BackoffBase.String())
			audithelper.Note(ctx, "reverse_proxy.retry.retryable_http_statuses", sortedStatusCodes(policy.RetryableHTTPStatuses))
			audithelper.Note(ctx, "reverse_proxy.retry.enable_response_body_issue_match", policy.EnableResponseBodyIssueMatch)
		}

		p.serveWithTransparentRetry(ctx, tw, r, requestFilters, responseFilters, options, policy)
	}
}

func (p *provider) serveWithTransparentRetry(
	ctx context.Context,
	tw *reverse_proxy.TrackedResponseWriter,
	r *http.Request,
	requestFilters []filter_define.FilterWithName[filter_define.ProxyRequestRewriter],
	responseFilters []filter_define.FilterWithName[filter_define.ProxyResponseModifier],
	options []OptionFunc,
	policy transparentRetryPolicy,
) {
	logger := ctxhelper.MustGetLoggerBase(ctx)
	defaultErrorHandler := reverse_proxy.MyErrorHandler()

	totalAttempts := 1
	if policy.Enabled && policy.MaxAttempts > 1 {
		totalAttempts = policy.MaxAttempts
	}

	for attempt := 1; attempt <= totalAttempts; attempt++ {
		// isolate context per attempt to avoid audit/state contamination
		r = prepareRequestForAttempt(r.Context(), r, attempt)
		ctx = r.Context()
		ctxhelper.PutReverseProxyRetryAttempt(ctx, attempt)

		// add cancel layer for this attempt to prevent leaked goroutines (e.g. asyncHandleRespBody)
		retryCtx, attemptCancel := context.WithCancel(ctx)
		reqForAttempt := r.WithContext(retryCtx)

		result := proxyAttemptResult{}
		proxy := httputil.ReverseProxy{
			Rewrite: reverse_proxy.MyRewrite(tw, requestFilters),
			Transport: transports.NewRequestFilterGeneratedResponseTransport(&transports.CurlPrinterTransport{
				Inner: &transports.TimerTransport{},
			}),
			FlushInterval:  -1,
			ErrorLog:       nil,
			BufferPool:     nil,
			ModifyResponse: reverse_proxy.MyResponseModify(tw, responseFilters),
		}
		proxy.ErrorHandler = func(w http.ResponseWriter, req *http.Request, err error) {
			result.Err = err
			result.StatusCode = extractHTTPErrorStatus(err)
			result.Retryable = isRetryableProxyError(err, result.StatusCode, policy)
			if policy.Enabled && attempt < totalAttempts && !tw.WroteHeader() && result.Retryable {
				result.Suppressed = true
				if result.StatusCode != 0 && isRetryableHTTPStatus(result.StatusCode, policy) {
					reportRetryableStatusFailure(ctx, req, result.StatusCode)
				}
				if result.StatusCode == 0 {
					health.ReportModelNetworkFailure(ctx, req, err)
				}

				// Ensure audit data of the suppressed attempt is persisted before retry.
				reverse_proxy.NoteAttemptCompleted(req.Context())
				audithelper.Note(req.Context(), "retry.attempt_failed_reason", retryReason(result))
				audithelper.Flush(req.Context())
				return
			}
			defaultErrorHandler(w, req, err)
		}

		for _, option := range options {
			option(ctx, &proxy)
		}

		proxy.ServeHTTP(tw, reqForAttempt)

		// cancel background goroutines of this attempt if we are going to retry
		if result.Suppressed {
			attemptCancel()
		} else {
			defer attemptCancel()
		}

		result.WroteHeader = tw.WroteHeader()
		result.InstanceID = getCurrentInstanceID(retryCtx)
		noteRetryAttempt(retryCtx, attempt, result)

		if result.Err == nil {
			noteRetryFinal(ctx, attempt, result.InstanceID)
			logger.Infof("transparent retry finished, attempt=%d, instance=%s", attempt, result.InstanceID)
			return
		}

		if !result.Suppressed {
			noteRetryFinal(ctx, attempt, result.InstanceID)
			logger.Warnf("transparent retry stop at attempt=%d, instance=%s, wrote_header=%v, retryable=%v, err=%v", attempt, result.InstanceID, result.WroteHeader, result.Retryable, result.Err)
			return
		}
		if result.InstanceID != "" {
			ctxhelper.AddReverseProxyRetryExcludedModelID(ctx, result.InstanceID)
		}

		delay := nextRetryBackoff(policy, attempt)
		audithelper.NoteAppend(ctx, "reverse_proxy.retry.events", map[string]any{
			"attempt":      attempt,
			"next_attempt": attempt + 1,
			"sleep":        delay.String(),
			"reason":       retryReason(result),
		})
		logger.Warnf("transparent retry trigger, attempt=%d, next_attempt=%d, sleep=%s, instance=%s, reason=%s", attempt, attempt+1, delay.String(), result.InstanceID, retryReason(result))
		if delay > 0 {
			time.Sleep(delay)
		}
	}
}

func (p *provider) resolveTransparentRetryPolicy(r *http.Request) transparentRetryPolicy {
	policy := transparentRetryPolicy{
		Enabled:                      p.Config.ModelRetry.Enabled,
		MaxAttempts:                  p.Config.ModelRetry.MaxAttempts,
		BackoffBase:                  p.Config.ModelRetry.Backoff.Base,
		RetryableHTTPStatuses:        toStatusCodeSet(p.Config.ModelRetry.RetryableHTTPStatuses),
		EnableResponseBodyIssueMatch: p.Config.ModelRetry.EnableResponseBodyNetworkIssueMatch,
	}
	if policy.MaxAttempts <= 0 {
		policy.MaxAttempts = 1
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
		if maxAttempts, err := strconv.Atoi(raw); err == nil && maxAttempts > 0 {
			policy.MaxAttempts = maxAttempts
		} else if logger != nil {
			logger.Warnf("invalid %s=%q", vars.XAIProxyRetryMax, raw)
		}
	}
	if health.IsHealthProbeRequest(r.Header) {
		policy.Enabled = false
	}

	return policy
}

func prepareRequestForAttempt(ctx context.Context, r *http.Request, attempt int) *http.Request {
	// Always reset ctx to isolate per-attempt state (model, audit sink, filters).
	newCtx := ctxhelper.ResetForRetry(ctx)
	newReq := r.WithContext(newCtx)

	if attempt > 1 {
		ctxhelper.PutGeneratedCallID(newCtx, requestid.GetCallID(newReq))
	}

	if attempt > 1 {
		bodyValue, ok := ctxhelper.GetReverseProxyRequestBodyBytes(newCtx)
		if ok {
			if bodyBytes, ok := bodyValue.([]byte); ok {
				copied := bytes.Clone(bodyBytes)
				newReq.Body = io.NopCloser(bytes.NewReader(copied))
				newReq.GetBody = func() (io.ReadCloser, error) {
					return io.NopCloser(bytes.NewReader(bytes.Clone(bodyBytes))), nil
				}
				newReq.ContentLength = int64(len(copied))
			}
		}
		newReq.Form = nil
		newReq.PostForm = nil
		newReq.MultipartForm = nil
	}

	ctxhelper.PutReverseProxyRequestRewriteError(newCtx, nil)
	ctxhelper.PutReverseProxyResponseModifyError(newCtx, nil)
	return newReq
}

func isRetryableProxyError(err error, statusCode int, policy transparentRetryPolicy) bool {
	if err == nil {
		return false
	}
	if statusCode == 0 {
		statusCode = extractHTTPErrorStatus(err)
	}
	if statusCode != 0 {
		if isRetryableHTTPStatus(statusCode, policy) {
			return true
		}
		return policy.EnableResponseBodyIssueMatch && hasRetryableNetworkIssueInHTTPErrorBody(err)
	}
	if health.IsNetworkFailureError(err) {
		return true
	}
	return false
}

func extractHTTPErrorStatus(err error) int {
	if err == nil {
		return 0
	}
	var httpErr *httperror.HTTPError
	if errors.As(err, &httpErr) {
		return httpErr.StatusCode
	}
	return 0
}

func isRetryableHTTPStatus(statusCode int, policy transparentRetryPolicy) bool {
	_, ok := policy.RetryableHTTPStatuses[statusCode]
	return ok
}

func reportRetryableStatusFailure(ctx context.Context, fallbackReq *http.Request, statusCode int) {
	if trusted, ok := ctxhelper.GetTrustedHealthProbe(ctx); ok && trusted {
		return
	}
	manager := health.GetManager()
	if manager == nil {
		return
	}

	sourceReq := fallbackReq
	if reqInSnap, ok := ctxhelper.GetReverseProxyRequestInSnapshot(ctx); ok && reqInSnap != nil {
		sourceReq = reqInSnap
	}
	if sourceReq == nil || sourceReq.URL == nil {
		return
	}
	apiType, ok := health.ResolveAPIType(sourceReq.Method, sourceReq.URL.Path)
	if !ok {
		return
	}
	model, ok := ctxhelper.GetModel(ctx)
	if !ok || model == nil || model.Id == "" {
		return
	}
	clientID, ok := ctxhelper.GetClientId(ctx)
	if !ok || clientID == "" {
		return
	}
	manager.MarkUnhealthy(ctx, clientID, model.Id, apiType, fmt.Sprintf("retryable status code: %d", statusCode), health.BuildProbeHeaders(sourceReq.Header))
}

func hasRetryableNetworkIssueInHTTPErrorBody(err error) bool {
	var httpErr *httperror.HTTPError
	if !errors.As(err, &httpErr) || httpErr == nil || len(httpErr.ErrorCtx) == 0 {
		return false
	}
	raw, ok := httpErr.ErrorCtx["raw_llm_backend_response"]
	if !ok || raw == nil {
		return false
	}
	switch v := raw.(type) {
	case string:
		return health.IsNetworkFailureError(errors.New(v))
	default:
		b, marshalErr := json.Marshal(v)
		if marshalErr != nil {
			return false
		}
		return health.IsNetworkFailureError(errors.New(string(b)))
	}
}

func getCurrentInstanceID(ctx context.Context) string {
	model, ok := ctxhelper.GetModel(ctx)
	if !ok || model == nil {
		return ""
	}
	return model.Id
}

func retryReason(result proxyAttemptResult) string {
	if result.StatusCode != 0 {
		return fmt.Sprintf("http_status_%d", result.StatusCode)
	}
	if result.Err != nil {
		return result.Err.Error()
	}
	return "unknown"
}

func noteRetryAttempt(ctx context.Context, attempt int, result proxyAttemptResult) {
	audithelper.NoteAppend(ctx, "reverse_proxy.retry.attempts", map[string]any{
		"attempt":      attempt,
		"instance_id":  result.InstanceID,
		"retryable":    result.Retryable,
		"status_code":  result.StatusCode,
		"wrote_header": result.WroteHeader,
		"suppressed":   result.Suppressed,
		"error": func() string {
			if result.Err == nil {
				return ""
			}
			return result.Err.Error()
		}(),
	})
}

func noteRetryFinal(ctx context.Context, attempt int, instanceID string) {
	audithelper.Note(ctx, "reverse_proxy.retry.final_attempt", attempt)
	if instanceID != "" {
		audithelper.Note(ctx, "reverse_proxy.retry.final_instance_id", instanceID)
	}
}

func nextRetryBackoff(policy transparentRetryPolicy, attempt int) time.Duration {
	if attempt <= 0 || policy.BackoffBase <= 0 {
		return 0
	}
	if attempt > 60 {
		attempt = 60
	}
	// retry #1 => 1*base, retry #2 => 3*base, retry #3 => 7*base
	multiplier := int64((1 << attempt) - 1)
	return time.Duration(multiplier) * policy.BackoffBase
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

func sortedStatusCodes(codes map[int]struct{}) []int {
	ret := make([]int, 0, len(codes))
	for code := range codes {
		ret = append(ret, code)
	}
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
