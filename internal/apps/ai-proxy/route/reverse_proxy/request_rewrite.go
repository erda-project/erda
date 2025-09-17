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
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"runtime/debug"
	"strings"
	"time"

	"github.com/labstack/echo"

	"github.com/erda-project/erda/internal/apps/ai-proxy/common/audit/audithelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/audit/dumplog"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/audit/skeleton"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/body_util"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/filter_define"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/filters/common/logutil"
	"github.com/erda-project/erda/internal/apps/ai-proxy/vars"
)

var MyRewrite = func(w http.ResponseWriter, requestFilters []filter_define.FilterWithName[filter_define.ProxyRequestRewriter]) func(*httputil.ProxyRequest) {
	return func(pr *httputil.ProxyRequest) {
		var currentFilterName string
		var brokenInErr error

		// record accurate request time before audit created
		requestBeginAt := time.Now()
		defer func() {
			audithelper.Note(pr.In.Context(), "request_begin_at", requestBeginAt)
			audithelper.Flush(pr.In.Context())
		}()

		defer func() {
			if r := recover(); r != nil {
				debug.PrintStack()
				brokenInErr = fmt.Errorf("panic: %v", r)
			}
			if brokenInErr == nil {
				return
			}
			ctxhelper.PutReverseProxyRequestRewriteError(pr.In.Context(), &ctxhelper.ReverseProxyFilterError{
				FilterName: currentFilterName,
				Error:      brokenInErr,
			})
			// force make RoundTrip failed and handle error at ErrorHandler
			pr.Out.URL.Host = ""
			pr.Out.URL.Scheme = ""
			return
		}()

		// enhance ProxyRequest
		// since pr.In and pr.Out share the same body reader, need to save data first and give each one a copy
		if pr.In.Body != nil {
			save, err := io.ReadAll(pr.In.Body)
			if err != nil {
				brokenInErr = err
				panic(brokenInErr)
			}
			_ = pr.In.Body.Close()
			pr.In.Body = io.NopCloser(bytes.NewReader(save))
			pr.Out.Body = io.NopCloser(bytes.NewReader(bytes.Clone(save)))
		}

		inSnap, err := body_util.SafeCloneRequest(pr.In, body_util.MaxSample)
		if err != nil {
			brokenInErr = err
			panic(brokenInErr)
		}
		ctxhelper.PutReverseProxyRequestInSnapshot(pr.In.Context(), &inSnap)

		// create audit skeleton
		if err := skeleton.CreateSkeleton(pr.In); err != nil {
			ctxhelper.MustGetLoggerBase(pr.In.Context()).Warnf("failed to create audit skeleton: %v", err)
		}

		// dump request in
		dumplog.DumpRequestIn(pr.In)

		// handle ai-proxy request header
		handleAIProxyRequestHeader(pr)

		for _, filter := range requestFilters {
			currentFilterName = filter.Name
			// inject logger with package name into the context
			logutil.InjectLoggerWithFilterInfo(pr.In.Context(), filter)
			if rewriter, ok := filter.Instance.(filter_define.ProxyRequestRewriter); ok {
				pr.In = &inSnap // pr.In cannot be modified, so we always put original pr.In, discard changes by previous filters
				if err := rewriter.OnProxyRequest(pr); err != nil {
					brokenInErr = err
					break
				}
			}
		}
		currentFilterName = ""

		// only dump request out when no error
		if brokenInErr == nil {
			dumplog.DumpRequestOut(pr.Out)
		}
	}
}

func handleAIProxyRequestHeader(pr *httputil.ProxyRequest) {
	// del all X-AI-Proxy-* headers before invoking llm
	for k := range pr.Out.Header {
		if strings.HasPrefix(strings.ToUpper(k), strings.ToUpper(vars.XAIProxyHeaderPrefix)) {
			pr.Out.Header.Del(k)
		}
	}
	// del Origin to avoid upstream CORS header
	pr.Out.Header.Del(echo.HeaderOrigin)
}
