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
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"runtime/debug"
	"time"

	"github.com/erda-project/erda/internal/apps/ai-proxy/common/audit/audithelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/audit/dumplog"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/filter_define"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/filters/common/logutil"
	set_resp_body_chunk_splitter "github.com/erda-project/erda/internal/apps/ai-proxy/route/filters/request/set-resp-body-chunk-splitter"
	httperror "github.com/erda-project/erda/internal/apps/ai-proxy/route/http_error"
	"github.com/erda-project/erda/internal/apps/ai-proxy/vars"
)

// MyResponseModify will be assigned to ReverseProxy.ModifyResponse inside p.ServeHTTP.
//
// Key flow:
//  1. First run all ProxyResponseModifier.OnHeaders(), allowing addition/removal of response headers and Trailer declaration.
//  2. Loop read chunks according to ChunkSplitter in ctx (default FixedSize);
//     each chunk flows through all Modifier.OnBodyChunk() in sequence; write to downstream immediately after getting out.
//  3. When stream ends or errors, call each Modifier.OnComplete();
//     they can write trailer headers to resp.Trailer.
//  4. Return nil; ReverseProxy will continue copying resp to client.
var MyResponseModify = func(w http.ResponseWriter, filters []filter_define.FilterWithName[filter_define.ProxyResponseModifier]) func(*http.Response) error {
	return func(resp *http.Response) (_err error) {
		var brokenFilterName string
		defer func() {
			if r := recover(); r != nil {
				debug.PrintStack()
				_err = fmt.Errorf("panic from response modify: %v", r)
			}
			if _err == nil {
				return
			}
			ctxhelper.PutReverseProxyResponseModifyError(resp.Request.Context(), &ctxhelper.ReverseProxyFilterError{
				FilterName: brokenFilterName,
				Error:      _err,
			})
			// force make ReverseProxy failed and handle error at ErrorHandler
			return
		}()

		// audit response begin
		audithelper.NoteOnce(resp.Request.Context(), "response_at", time.Now()) // table field

		// handle error status firstly
		if _err = errorStatusHandler(resp); _err != nil {
			return
		}

		upstream := resp.Body
		pr, pw := io.Pipe()
		resp.Body = pr // replace with read end

		// 1) choose chunk splitting strategy -- get from ctx; fallback to FixedSize if none
		// determine splitter based on original response header
		splitter := set_resp_body_chunk_splitter.GetSplitterByResp(resp)
		if splitter == nil {
			panic("no RespBodyChunkSplitter found in context")
		}
		// wrap as auto-decompressing splitter
		autoDecompressSplitter := set_resp_body_chunk_splitter.NewDecompressingChunkSplitter(splitter, resp)
		ctxhelper.MustGetLoggerBase(resp.Request.Context()).Debugf("splitter: %s", reflect.TypeOf(splitter))
		audithelper.NoteOnce(resp.Request.Context(), "response_body_splitter", reflect.TypeOf(splitter).String())

		// ---------------------------------------------------------------------
		// 2) let all Modifiers run OnHeaders first
		dumplog.DumpResponseHeadersIn(resp)

		// handle ai-proxy response header
		handleAIProxyResponseHeader(resp)

		for _, filter := range filters {
			brokenFilterName = filter.Name
			logutil.InjectLoggerWithFilterInfo(resp.Request.Context(), filter)
			if err := filter.Instance.OnHeaders(resp); err != nil {
				return err
			}
		}
		brokenFilterName = ""

		dumplog.DumpResponseHeadersOut(resp)

		// ---------------------------------------------------------------------
		// 3) write upstream Body-Splitter-Modifier chain results to Pipe, let ReverseProxy copy to client

		go asyncHandleRespBody(upstream, autoDecompressSplitter, pw, resp, filters)

		return nil
	}
}

func asyncHandleRespBody(upstream io.ReadCloser, splitter filter_define.RespBodyChunkSplitter, pw *io.PipeWriter, resp *http.Response, filters []filter_define.FilterWithName[filter_define.ProxyResponseModifier]) {
	defer func() {
		// sink audit
		audithelper.Flush(resp.Request.Context())
	}()

	var currentFilterName string
	defer func() {
		if r := recover(); r != nil {
			debug.PrintStack()
			err := fmt.Errorf("panic: %v", r)
			if currentFilterName != "" {
				err = fmt.Errorf("%s: %v", currentFilterName, err)
			}
			writeAndCloseWithErr(resp, pw, fmt.Errorf("panic from response modify: %v", err))
		}
		_ = pw.Close()
	}()

	defer func() { _ = upstream.Close() }()

	var (
		wholeReceivedBody []byte
		wholeHandledBody  []byte
	)

	const bodyLimitBytes = 1024 * 32 // 32K

	// audit response body
	defer func() {
		if len(wholeReceivedBody) > bodyLimitBytes {
			wholeReceivedBody = []byte(string(wholeReceivedBody[:bodyLimitBytes]) + fmt.Sprintf(".....[omitted due to length: %d]", len(wholeReceivedBody)))
		}
		if len(wholeHandledBody) > bodyLimitBytes {
			wholeHandledBody = []byte(string(wholeHandledBody[:bodyLimitBytes]) + fmt.Sprintf(".....[omitted due to length: %d]", len(wholeHandledBody)))
		}
		audithelper.NoteAppend(resp.Request.Context(), "actual_response_body", string(wholeReceivedBody))
		audithelper.NoteAppend(resp.Request.Context(), "response_body", string(wholeHandledBody))
	}()

	// audit response time
	var responseChunkBeginAt time.Time
	defer func() {
		responseChunkDoneAt := time.Now()
		audithelper.NoteOnce(resp.Request.Context(), "response_chunk_done_at", responseChunkDoneAt)
	}()

	var chunkIndex int64 = -1
	for {
		chunkIndex++
		chunk, rerr := splitter.NextChunk(upstream)
		// begin response here
		if responseChunkBeginAt.IsZero() {
			responseChunkBeginAt = time.Now()
			audithelper.NoteOnce(resp.Request.Context(), "response_chunk_begin_at", responseChunkBeginAt) // table metadata
		}
		if len(chunk) == 0 && rerr == nil {
			// indicates splitter violated contract, avoid infinite loop
			writeAndCloseWithErr(resp, pw, errors.New("splitter returned empty chunk without error"))
			return
		}
		if len(chunk) > 0 {
			out := chunk
			// 3.2 same chunk flows through all Modifiers in sequence
			dumpReceivedOut := dumplog.DumpResponseBodyChunk(resp, out, chunkIndex)
			wholeReceivedBody = append(wholeReceivedBody, dumpReceivedOut...)
			for _, filter := range filters {
				currentFilterName = filter.Name
				m := filter.Instance
				logutil.InjectLoggerWithFilterInfo(resp.Request.Context(), filter)
				o, err := m.OnBodyChunk(resp, out, chunkIndex)
				if err != nil {
					writeAndCloseWithErr(resp, pw, err)
					return
				}
				out = o
				if len(out) == 0 { // Swallowed
					break
				}
			}
			currentFilterName = ""
			dumpHandledOut := dumplog.DumpResponseBodyChunk(resp, out, chunkIndex)
			wholeHandledBody = append(wholeHandledBody, dumpHandledOut...)
			if len(out) > 0 {
				if _, err := pw.Write(out); err != nil {
					// Downstream disconnected
					return
				}
			}
		}

		if rerr != nil { // EOF or real error
			if rerr != io.EOF && rerr != context.Canceled {
				writeAndCloseWithErr(resp, pw, rerr)
			}
			// cleanup: call OnComplete; ensure call even if previous errors
			dumplog.DumpResponseComplete(resp)
			for _, filter := range filters {
				currentFilterName = filter.Name
				if m, ok := filter.Instance.(filter_define.ProxyResponseModifier); ok {
					logutil.InjectLoggerWithFilterInfo(resp.Request.Context(), filter)
					out, _ := m.OnComplete(resp)
					if len(out) > 0 {
						wholeHandledBody = append(wholeHandledBody, out...)
						_, _ = pw.Write(out)
					}
				}
			}
			currentFilterName = ""
			_ = pw.Close()
			return
		}
	}
}

func errorStatusHandler(resp *http.Response) error {

	if resp.StatusCode < 400 {
		return nil
	}

	// generate standardized error response
	errCtx := map[string]any{
		"type": "llm-backend-error",
	}

	// Content-Length may be -1, so only check body
	if resp.Body != nil {
		respBodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			ctxhelper.MustGetLoggerBase(resp.Request.Context()).Warnf("failed to read resp body at error-status-handler: %v", err)
		} else {
			// try to parse backend response as JSON, if so preserve it
			var jsonResp interface{}
			if json.Unmarshal(respBodyBytes, &jsonResp) == nil {
				errCtx["raw_llm_backend_response"] = jsonResp
			} else {
				errCtx["raw_llm_backend_response"] = string(respBodyBytes)
			}
		}
	}

	httpErr := httperror.NewHTTPErrorWithCtx(
		resp.Request.Context(),
		resp.StatusCode,
		"LLM Backend Error",
		errCtx,
	)
	httpErr.AIProxyMeta = map[string]any{
		vars.XAIProxyModel:           ctxhelper.MustGetModel(resp.Request.Context()).Name,
		vars.XRequestId:              ctxhelper.MustGetRequestID(resp.Request.Context()),
		vars.XAIProxyGeneratedCallId: ctxhelper.MustGetGeneratedCallID(resp.Request.Context()),
	}

	// modify response headers
	resp.Header.Del("Content-Length") // delete original length, let system recalculate

	return httpErr
}

func writeAndCloseWithErr(resp *http.Response, pw *io.PipeWriter, err error) {
	_, _ = pw.Write([]byte(err.Error()))
	_ = pw.CloseWithError(err)
	audithelper.Note(resp.Request.Context(), "myResponseModify.error", err)
}
