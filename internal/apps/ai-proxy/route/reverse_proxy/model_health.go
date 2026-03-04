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
	"context"
	"errors"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"syscall"

	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/policy_group/health"
)

var networkErrorMarkers = []string{
	"connection reset by peer",
	"broken pipe",
	"dial tcp",
	"read tcp",
	"write tcp",
	"i/o timeout",
	"timeout awaiting response headers",
	"client timeout exceeded while awaiting headers",
	"connection refused",
	"no route to host",
	"network is unreachable",
	"tls handshake timeout",
}

func reportModelNetworkFailure(ctx context.Context, fallbackReq *http.Request, err error) {
	if !isNetworkFailureError(err) {
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
	if trusted, ok := ctxhelper.GetTrustedHealthProbe(ctx); ok && trusted {
		return
	}

	apiType, ok := health.ResolveAPIType(sourceReq.Method, sourceReq.URL.Path)
	if !ok {
		// Current phase only reports unhealthy for chat/responses;
		// embeddings and other APIs do not enter unhealthy lifecycle.
		return
	}

	model, ok := ctxhelper.GetModel(ctx)
	if !ok || model == nil || model.Id == "" {
		return
	}
	manager.MarkUnhealthy(ctx, model.Id, apiType, err.Error(), health.BuildProbeHeaders(sourceReq.Header))
}

func isNetworkFailureError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, context.Canceled) {
		return false
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}

	var urlErr *url.Error
	if errors.As(err, &urlErr) {
		if urlErr.Timeout() || isNetworkFailureError(urlErr.Err) {
			return true
		}
	}

	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return true
	}

	var opErr *net.OpError
	if errors.As(err, &opErr) {
		return true
	}

	var dnsErr *net.DNSError
	if errors.As(err, &dnsErr) {
		return true
	}

	var syscallErr *os.SyscallError
	if errors.As(err, &syscallErr) && isNetworkSyscallError(syscallErr.Err) {
		return true
	}

	if isNetworkSyscallError(err) {
		return true
	}

	lowerErr := strings.ToLower(err.Error())
	for _, marker := range networkErrorMarkers {
		if strings.Contains(lowerErr, marker) {
			return true
		}
	}

	return false
}

func isNetworkSyscallError(err error) bool {
	return errors.Is(err, syscall.ECONNRESET) ||
		errors.Is(err, syscall.EPIPE) ||
		errors.Is(err, syscall.ETIMEDOUT) ||
		errors.Is(err, syscall.ECONNREFUSED) ||
		errors.Is(err, syscall.EHOSTUNREACH) ||
		errors.Is(err, syscall.ENETUNREACH) ||
		errors.Is(err, syscall.ECONNABORTED)
}
