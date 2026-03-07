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
	"time"

	"github.com/erda-project/erda-infra/base/logs/logrusx"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/audit/audithelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/requestid"
	modelretry "github.com/erda-project/erda/internal/apps/ai-proxy/providers/reverseproxy/retry/model_retry"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/filter_define"
	httperror "github.com/erda-project/erda/internal/apps/ai-proxy/route/http_error"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/policy_group/health"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/reverse_proxy"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/router_define"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/transports"
)

type OptionFunc func(context.Context, *httputil.ReverseProxy)

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
		policy := modelretry.ResolvePolicy(r, p.Config.ModelRetry)
		ctxhelper.PutModelRetryResponseHeaderMetaEnabled(ctx, policy.ResponseHeaderMetaEnabled)
		audithelper.Note(ctx, "reverse_proxy.retry.enabled", policy.Enabled)
		audithelper.Note(ctx, "reverse_proxy.retry.max_llm_backend_request_count", policy.MaxLLMBackendRequestCount)
		if policy.Enabled {
			audithelper.Note(ctx, "reverse_proxy.retry.backoff_base", policy.BackoffBase.String())
			audithelper.Note(ctx, "reverse_proxy.retry.backoff_max", policy.BackoffMax.String())
			audithelper.Note(ctx, "reverse_proxy.retry.retryable_http_statuses", modelretry.SortedStatusCodes(policy.RetryableHTTPStatuses))
			audithelper.Note(ctx, "reverse_proxy.retry.match_network_issue_from_response_body", policy.MatchNetworkIssueFromResponseBody)
			audithelper.Note(ctx, "reverse_proxy.retry.exclude_failed_instance", policy.ExcludeFailedInstance)
			audithelper.Note(ctx, "reverse_proxy.retry.response_header_meta", policy.ResponseHeaderMetaEnabled)
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
	policy modelretry.Policy,
) {
	logger := ctxhelper.MustGetLoggerBase(ctx)
	defaultErrorHandler := reverse_proxy.MyErrorHandler()

	totalAttempts := 1
	if policy.Enabled && policy.MaxLLMBackendRequestCount > 1 {
		totalAttempts = policy.MaxLLMBackendRequestCount
	}

	for attempt := 1; attempt <= totalAttempts; attempt++ {
		// isolate context per attempt to avoid audit/state contamination
		r = prepareRequestForAttempt(r.Context(), r, attempt)
		ctx = r.Context()
		ctxhelper.PutModelRetryRawLLMBackendRequestCount(ctx, attempt)

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
				if result.StatusCode != 0 && policy.IsRetryableHTTPStatus(result.StatusCode) {
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

		if result.Err == nil {
			logger.Infof("transparent retry finished, attempt=%d, instance=%s", attempt, result.InstanceID)
			return
		}

		if !result.Suppressed {
			logger.Warnf("transparent retry stop at attempt=%d, instance=%s, wrote_header=%v, retryable=%v, err=%v", attempt, result.InstanceID, result.WroteHeader, result.Retryable, result.Err)
			return
		}
		// Retry-layer exclusion is request-scoped and only supplements routing.
		// It does not override model health filtering, which may already remove
		// the failed instance before the next attempt.
		if policy.ExcludeFailedInstance && result.InstanceID != "" {
			modelretry.AddExcludedModelID(ctx, result.InstanceID)
		}

		delay := policy.NextBackoff(attempt)
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

func isRetryableProxyError(err error, statusCode int, policy modelretry.Policy) bool {
	if err == nil {
		return false
	}
	if statusCode == 0 {
		statusCode = extractHTTPErrorStatus(err)
	}
	if statusCode != 0 {
		if policy.IsRetryableHTTPStatus(statusCode) {
			return true
		}
		return policy.MatchNetworkIssueFromResponseBody && hasRetryableNetworkIssueInHTTPErrorBody(err)
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
