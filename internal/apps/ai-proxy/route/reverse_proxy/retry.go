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

package reverse_proxy

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

	"github.com/erda-project/erda/internal/apps/ai-proxy/common/audit/audithelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/requestid"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/filter_define"
	httperror "github.com/erda-project/erda/internal/apps/ai-proxy/route/http_error"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/policy_group/health"
	modelretry "github.com/erda-project/erda/internal/apps/ai-proxy/route/reverse_proxy/model_retry"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/transports"
	"github.com/erda-project/erda/internal/apps/ai-proxy/vars"
)

type OptionFunc func(context.Context, *httputil.ReverseProxy)

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

type proxyAttemptResult struct {
	Err         error
	Retryable   bool
	StatusCode  int
	Suppressed  bool
	InstanceID  string
	WroteHeader bool
}

type modelRetryMetaHeader struct {
	RawLLMBackendRequestCount int    `json:"raw_llm_backend_request_count"`
	RawLLMBackendRetryCount   int    `json:"raw_llm_backend_retry_count"`
	FinalModelInstanceID      string `json:"final_model_instance_id,omitempty"`
}

func ServeWithRetry(
	ctx context.Context,
	tw *TrackedResponseWriter,
	r *http.Request,
	requestFilters []filter_define.FilterWithName[filter_define.ProxyRequestRewriter],
	responseFilters []filter_define.FilterWithName[filter_define.ProxyResponseModifier],
	options []OptionFunc,
	policy modelretry.Config,
) {
	if policy.Enabled {
		noteModelRetryPolicy(ctx, policy)
	}

	logger := ctxhelper.MustGetLoggerBase(ctx)
	defaultErrorHandler := MyErrorHandler()

	totalAttempts := 1
	if policy.Enabled && policy.Conditions.MaxLLMBackendRequestCount > 1 {
		totalAttempts = policy.Conditions.MaxLLMBackendRequestCount
	}

	for attempt := 1; attempt <= totalAttempts; attempt++ {
		r = prepareRequestForAttempt(r.Context(), r, attempt)
		ctx = r.Context()
		ctxhelper.PutModelRetryRawLLMBackendRequestCount(ctx, attempt)

		retryCtx, attemptCancel := context.WithCancel(ctx)
		reqForAttempt := r.WithContext(retryCtx)

		result := proxyAttemptResult{}
		proxy := httputil.ReverseProxy{
			Rewrite: MyRewrite(tw, requestFilters),
			Transport: transports.NewRequestFilterGeneratedResponseTransport(&transports.CurlPrinterTransport{
				Inner: &transports.TimerTransport{},
			}),
			FlushInterval:  -1,
			ErrorLog:       nil,
			BufferPool:     nil,
			ModifyResponse: MyResponseModify(tw, responseFilters),
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

				NoteAttemptCompleted(req.Context())
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

		if policy.Actions.ExcludeFailedInstance && result.InstanceID != "" {
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

func noteModelRetryPolicy(ctx context.Context, policy modelretry.Config) {
	ctxhelper.PutModelRetryResponseHeaderMetaEnabled(ctx, policy.Observability.ResponseHeaderMeta)
	audithelper.Note(ctx, "reverse_proxy.retry.enabled", true)
	audithelper.Note(ctx, "reverse_proxy.retry.max_llm_backend_request_count", policy.Conditions.MaxLLMBackendRequestCount)
	audithelper.Note(ctx, "reverse_proxy.retry.backoff_base", policy.Conditions.Backoff.Base.String())
	audithelper.Note(ctx, "reverse_proxy.retry.backoff_max", policy.Conditions.Backoff.Max.String())
	audithelper.Note(ctx, "reverse_proxy.retry.retryable_http_statuses", policy.SortedRetryableHTTPStatuses())
	audithelper.Note(ctx, "reverse_proxy.retry.match_network_issue_from_response_body", policy.Conditions.MatchNetworkIssueFromResponseBody)
	audithelper.Note(ctx, "reverse_proxy.retry.exclude_failed_instance", policy.Actions.ExcludeFailedInstance)
	audithelper.Note(ctx, "reverse_proxy.retry.response_header_meta", policy.Observability.ResponseHeaderMeta)
}

func prepareRequestForAttempt(ctx context.Context, r *http.Request, attempt int) *http.Request {
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

func isRetryableProxyError(err error, statusCode int, policy modelretry.Config) bool {
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
		return policy.Conditions.MatchNetworkIssueFromResponseBody && hasRetryableNetworkIssueInHTTPErrorBody(err)
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

func noteRetryAuditMetadata(ctx context.Context) {
	rawLLMBackendRequestCount, ok := ctxhelper.GetModelRetryRawLLMBackendRequestCount(ctx)
	if !ok || rawLLMBackendRequestCount <= 1 {
		return
	}
	audithelper.Note(ctx, "model_retry.raw_llm_backend_request_count", rawLLMBackendRequestCount)
	audithelper.Note(ctx, "model_retry.raw_llm_backend_retry_count", rawLLMBackendRequestCount-1)
}

func _handleModelRetryMetaHeader(resp *http.Response) {
	if enabled, ok := ctxhelper.GetModelRetryResponseHeaderMetaEnabled(resp.Request.Context()); ok && !enabled {
		return
	}
	rawLLMBackendRequestCount, ok := ctxhelper.GetModelRetryRawLLMBackendRequestCount(resp.Request.Context())
	if !ok || rawLLMBackendRequestCount <= 1 {
		return
	}
	payload := modelRetryMetaHeader{
		RawLLMBackendRequestCount: rawLLMBackendRequestCount,
		RawLLMBackendRetryCount:   rawLLMBackendRequestCount - 1,
	}
	if model, ok := ctxhelper.GetModel(resp.Request.Context()); ok && model != nil && model.Id != "" {
		payload.FinalModelInstanceID = model.Id
	}
	b, err := json.Marshal(payload)
	if err != nil {
		return
	}
	resp.Header.Set(vars.XAIProxyModelRetryMeta, string(b))
}
