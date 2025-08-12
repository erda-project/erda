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

package ai_proxy

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"reflect"
	"strings"

	"github.com/pkg/errors"

	"github.com/erda-project/erda-infra/base/logs/logrusx"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/requestid"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/body_util"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/filter_define"
	set_resp_body_chunk_splitter "github.com/erda-project/erda/internal/apps/ai-proxy/route/filters/request/set-resp-body-chunk-splitter"
	httperror "github.com/erda-project/erda/internal/apps/ai-proxy/route/http_error"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/router_define"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/transports"
)

type FilterWithName struct {
	Name     string
	Stage    string // "request" or "response"
	Instance any
}

func (p *provider) HandleReverseProxyAPI() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// inject context
		ctx := ctxhelper.InitCtxMapIfNeed(r.Context())
		r = r.WithContext(ctx)

		// create a request-level Logger
		logger := logrusx.New().Sub("reverse-proxy-api")
		reqID := requestid.GetOrSetRequestID(r)
		callID := requestid.GetCallID(r)
		logger.Set("req", reqID).Set("call", callID)
		logger.Infof("reverse proxy handler: %s %s", r.Method, r.URL.String())
		ctxhelper.PutLogger(ctx, logger)
		ctxhelper.PutRequestID(ctx, reqID)
		ctxhelper.PutGeneratedCallID(ctx, callID)

		// find best matched route using priority algorithm
		matched := p.Router.FindBestMatch(r.Method, r.URL.Path)
		var matchedRoute *router_define.Route
		if matched != nil {
			matchedRoute = matched.(*router_define.Route)
		}
		if matchedRoute == nil {
			httperror.NewHTTPError(http.StatusNotFound, "no matched route").WriteJSONHTTPError(w)
			return
		}

		// get all route filters
		var filters []FilterWithName
		for _, filterConfig := range matchedRoute.RequestFilters {
			creator, ok := filter_define.FilterFactory.RequestFilters[filterConfig.Name]
			if !ok {
				httperror.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("request filter %s not found", filterConfig.Name)).WriteJSONHTTPError(w)
				return
			}
			fc, err := filterConfig.GetConfigAsRawMessage()
			if err != nil {
				httperror.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("failed to convert config for filter %s: %v", filterConfig.Name, err)).WriteJSONHTTPError(w)
				return
			}
			f := creator(filterConfig.Name, fc)
			filters = append(filters, FilterWithName{Name: filterConfig.Name, Stage: "request", Instance: f})
		}
		for _, filterConfig := range matchedRoute.ResponseFilters {
			creator, ok := filter_define.FilterFactory.ResponseFilters[filterConfig.Name]
			if !ok {
				httperror.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("response filter %s not found", filterConfig.Name)).WriteJSONHTTPError(w)
				return
			}
			fc, err := filterConfig.GetConfigAsRawMessage()
			if err != nil {
				httperror.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("failed to convert config for filter %s: %v", filterConfig.Name, err)).WriteJSONHTTPError(w)
				return
			}
			f := creator(filterConfig.Name, fc)
			filters = append(filters, FilterWithName{Name: filterConfig.Name, Stage: "response", Instance: f})
		}

		ctxhelper.PutDBClient(ctx, p.Dao)
		ctxhelper.PutRichClientHandler(ctx, p.richClientHandler)
		ctxhelper.PutPathMatcher(ctx, matchedRoute.GetPathMatcher())

		// reverse proxy
		proxy := httputil.ReverseProxy{
			Rewrite: myRewrite(w, filters),
			Transport: &transports.RequestFilterGeneratedResponseTransport{
				Inner: &transports.CurlPrinterTransport{
					Inner: &transports.TimerTransport{},
				},
			},
			FlushInterval:  -1,
			ErrorLog:       nil,
			BufferPool:     nil,
			ModifyResponse: myResponseModify(w, filters),
			ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
				// check error at rewrite stage
				if rewriteErr, _ := ctxhelper.GetReverseProxyAtRewriteStage(r.Context()); rewriteErr != nil {
					err = rewriteErr
				}

				// Check if error is from response modifier
				if responseErr, _ := ctxhelper.GetResponseModifierError(r.Context()); responseErr != nil {
					err = responseErr
				}

				var httpError *httperror.HTTPError
				switch {
				case errors.As(err, &httpError):
					httpError.WriteJSONHTTPError(w)
					return
				default:
					httperror.NewHTTPError(http.StatusBadRequest, err.Error()).WriteJSONHTTPError(w)
					return
				}
			},
		}
		proxy.ServeHTTP(w, r)
	}
}

var myRewrite = func(w http.ResponseWriter, filters []FilterWithName) func(*httputil.ProxyRequest) {
	return func(pr *httputil.ProxyRequest) {
		var brokenInErr error
		defer func() {
			if brokenInErr == nil {
				return
			}
			ctxhelper.PutReverseProxyAtRewriteStage(pr.In.Context(), brokenInErr)
			// force make RoundTrip failed and handle error at ErrorHandler
			pr.Out.URL.Host = ""
			pr.Out.URL.Scheme = ""
			return
		}()
		// Enhance ProxyRequest
		// Since pr.In and pr.Out share the same body reader, need to save data first and give each one a copy
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

		//pr.Out.Header.Set("Accept-Encoding", "identity") // Disable compression to avoid affecting body reading
		for _, filter := range filters {
			if filter.Stage != "request" {
				continue
			}
			// Inject logger with package name into the context
			injectLogger(pr.In.Context(), filter, "request")

			if rewriter, ok := filter.Instance.(filter_define.ProxyRequestRewriter); ok {
				pr.In = &inSnap // pr.In cannot be modified, so we always put original pr.In, discard changes by previous filters
				if err := rewriter.OnProxyRequest(pr); err != nil {
					brokenInErr = err
					break
				}
			}
		}
	}
}

// myResponseModify will be assigned to ReverseProxy.ModifyResponse inside p.ServeHTTP.
//
// Key flow:
//  1. First run all ProxyResponseModifier.OnHeaders(), allowing addition/removal of response headers and Trailer declaration.
//  2. Loop read chunks according to ChunkSplitter in ctx (default FixedSize);
//     each chunk flows through all Modifier.OnBodyChunk() in sequence; write to downstream immediately after getting out.
//  3. When stream ends or errors, call each Modifier.OnComplete();
//     they can write trailer headers to resp.Trailer.
//  4. Return nil; ReverseProxy will continue copying resp to client.
var myResponseModify = func(w http.ResponseWriter, filters []FilterWithName) func(*http.Response) error {
	return func(resp *http.Response) (_err error) {
		upstream := resp.Body
		pr, pw := io.Pipe()
		resp.Body = pr // Replace with read end

		defer func() {
			if r := recover(); r != nil {
				// ① Close pipe -> client immediately receives 502
				pw.CloseWithError(fmt.Errorf("panic: %v", r))
				// ② Let ReverseProxy go to ErrorHandler instead of continuing normal flow
				_err = fmt.Errorf("panic from response modify: %v", r)
			}
		}()
		// Filter out Modifiers with stage != "response"
		var nfs []FilterWithName
		for _, filter := range filters {
			if filter.Stage != "response" {
				continue
			}
			nfs = append(nfs, filter)
		}
		filters = nfs

		// 1) Choose chunk splitting strategy -- get from ctx; fallback to FixedSize if none
		// Determine splitter based on original response header
		splitter := set_resp_body_chunk_splitter.GetSplitterByResp(resp)
		if splitter == nil {
			panic("no RespBodyChunkSplitter found in context")
		}
		// Wrap as auto-decompressing splitter
		autoDecompressSplitter := set_resp_body_chunk_splitter.NewDecompressingChunkSplitter(splitter, resp)
		ctxhelper.MustGetLogger(resp.Request.Context()).Infof("splitter: %s", reflect.TypeOf(splitter))

		// ---------------------------------------------------------------------
		// 2) Let all Modifiers run OnHeaders first
		for _, filter := range filters {
			injectLogger(resp.Request.Context(), filter, "resp.OnHeaders")
			if err := filter.Instance.(filter_define.ProxyResponseModifier).OnHeaders(resp); err != nil {
				return err
			}
		}

		// Force chunked transfer, worry-free
		resp.Header.Del("Content-Length")
		if ctxhelper.MustGetIsStream(resp.Request.Context()) && resp.Header.Get("Content-Type") == "" {
			resp.Header.Set("Content-Type", "text/event-stream; charset=utf-8")
		}

		// ---------------------------------------------------------------------
		// 3) Write upstream Body-Splitter-Modifier chain results to Pipe, let ReverseProxy copy to client

		go func() {
			defer upstream.Close()

			if len(filters) == 0 {
				// If no Modifiers, directly write upstream content to downstream as-is
				if _, err := io.Copy(pw, upstream); err != nil {
					pw.CloseWithError(err)
					return
				}
				pw.Close()
				return
			}

			chunkIndex := -1
			for {
				chunkIndex++
				ctxhelper.PutResponseChunkIndex(resp.Request.Context(), chunkIndex)
				chunk, rerr := autoDecompressSplitter.NextChunk(upstream)
				if len(chunk) == 0 && rerr == nil {
					// Indicates Splitter violated contract, avoid infinite loop
					pw.CloseWithError(errors.New("splitter returned empty chunk without error"))
					return
				}
				if len(chunk) > 0 {
					out := chunk
					// 3.2 Same chunk flows through all Modifiers in sequence
					for _, filter := range filters {
						m := filter.Instance.(filter_define.ProxyResponseModifier)
						injectLogger(resp.Request.Context(), filter, "resp.OnBodyChunk")
						o, err := m.OnBodyChunk(resp, out)
						if err != nil {
							pw.CloseWithError(err)
							return
						}
						out = o
						if len(out) == 0 { // Swallowed
							break
						}
					}
					if len(out) > 0 {
						if _, err := pw.Write(out); err != nil {
							// Downstream disconnected
							return
						}
					}
				}

				if rerr != nil { // EOF or real error
					if rerr != io.EOF && rerr != context.Canceled {
						pw.CloseWithError(rerr)
					}
					// Cleanup: all Modifiers call OnComplete; ensure call even if previous errors
					for _, filter := range filters {
						if m, ok := filter.Instance.(filter_define.ProxyResponseModifier); ok {
							injectLogger(resp.Request.Context(), filter, "resp.OnComplete")
							out, _ := m.OnComplete(resp)
							if len(out) > 0 {
								pw.Write(out)
							}
						}
					}
					pw.Close()
					return
				}
			}
		}()

		return nil // Tell ReverseProxy: taken over resp.Body, continue with its own copy
	}
}

// injectLogger injects a sub-logger with package name into the context and returns the new context
func injectLogger(ctx context.Context, filter FilterWithName, stage string) {
	// Use reflection to get package name from filter type
	pkgPath := reflect.TypeOf(filter.Instance).Elem().PkgPath()
	// Extract two levels of package path (e.g., "filters/after-audit" from "github.com/.../filters/after-audit")
	parts := strings.Split(pkgPath, "/")
	var packageName string
	if len(parts) >= 2 {
		packageName = parts[len(parts)-2] + "/" + parts[len(parts)-1]
	} else {
		packageName = pkgPath[strings.LastIndex(pkgPath, "/")+1:]
	}
	// Add filter name
	fullFilterPath := packageName + "@" + filter.Name

	logger := ctxhelper.MustGetLogger(ctx)
	logger.Set("filter", fullFilterPath).Set("stage", stage)
	ctxhelper.PutLogger(ctx, logger)
}
