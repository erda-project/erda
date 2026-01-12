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
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/usage/token_usage"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/filter_define"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/filters/common/logutil"
	set_resp_body_chunk_splitter "github.com/erda-project/erda/internal/apps/ai-proxy/route/filters/request/set-resp-body-chunk-splitter"
	httperror "github.com/erda-project/erda/internal/apps/ai-proxy/route/http_error"
	"github.com/erda-project/erda/internal/apps/ai-proxy/vars"
)

const (
	FirstChunkIndex = 0
)

// MyResponseModify will be assigned to ReverseProxy.ModifyResponse inside p.ServeHTTP.
//
// Key flow:
//  0. If any filter implements ProxyResponseModifierBodyPeeker and is enabled,
//     peek the first chunk BEFORE sending headers, allowing filters to inspect body content
//     and modify headers accordingly via OnPeekChunkBeforeHeaders().
//  1. Run all ProxyResponseModifier.OnHeaders(), allowing addition/removal of response headers
//     and Trailer declaration.
//  2. Loop read chunks according to ChunkSplitter in ctx (default FixedSize);
//     each chunk flows through all Modifier.OnBodyChunk() in sequence;
//     write to downstream immediately after getting out.
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
				Stage:      "response",
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

		// get splitter BEFORE modifying headers (Content-Type may be changed by handleAIProxyResponseHeader);
		// this ensures we use the original Content-Type to select the correct splitter
		splitter := set_resp_body_chunk_splitter.GetSplitterByResp(resp)
		if splitter == nil {
			panic("no RespBodyChunkSplitter found in context")
		}
		autoDecompressSplitter := set_resp_body_chunk_splitter.NewDecompressingChunkSplitter(splitter, resp)
		ctxhelper.MustGetLoggerBase(resp.Request.Context()).Debugf("splitter type: %s", reflect.TypeOf(splitter))
		audithelper.NoteOnce(resp.Request.Context(), "response_body_splitter", reflect.TypeOf(splitter).String())

		// run modifiers before creating the pipe
		dumplog.DumpResponseHeadersIn(resp)
		handleAIProxyResponseHeader(resp)

		// peek the first chunk BEFORE sending headers (so filters can modify headers based on body content)
		peekedChunk, peekedChunkErr := autoDecompressSplitter.NextChunk(resp.Body)
		if peekedChunkErr != nil && peekedChunkErr != io.EOF {
			return peekedChunkErr
		}

		// call OnPeekChunkBeforeHeaders for filters BEFORE OnHeaders
		for _, filter := range filters {
			brokenFilterName = filter.Name
			logutil.InjectLoggerWithFilterInfo(resp.Request.Context(), filter)
			if fe, ok := filter.Instance.(filter_define.ProxyResponseModifierEnabler); ok && !fe.Enable(resp) {
				continue
			}
			peeker, ok := filter.Instance.(filter_define.ProxyResponseModifierBodyPeeker)
			if !ok {
				continue
			}
			if err := peeker.OnPeekChunkBeforeHeaders(resp, peekedChunk); err != nil {
				return err
			}
		}
		brokenFilterName = ""

		// now call OnHeaders (after peek, so filter can modify headers based on peeked body)
		for _, filter := range filters {
			brokenFilterName = filter.Name
			logutil.InjectLoggerWithFilterInfo(resp.Request.Context(), filter)
			if fe, ok := filter.Instance.(filter_define.ProxyResponseModifierEnabler); ok && !fe.Enable(resp) {
				continue
			}
			if err := filter.Instance.OnHeaders(resp); err != nil {
				return err
			}
		}
		brokenFilterName = ""

		dumplog.DumpResponseHeadersOut(resp)

		// prepare body pipe for async processing
		upstream := resp.Body
		pr, pw := io.Pipe()
		resp.Body = pr

		// write upstream to pipe asynchronously, passing peeked chunk to be processed first
		go asyncHandleRespBody(upstream, autoDecompressSplitter, pw, resp, filters, peekedChunk, peekedChunkErr == io.EOF)

		return nil
	}
}

func asyncHandleRespBody(upstream io.ReadCloser, splitter filter_define.RespBodyChunkSplitter, pw *io.PipeWriter, resp *http.Response, filters []filter_define.FilterWithName[filter_define.ProxyResponseModifier], peekedChunk []byte, peekedChunkIsEOF bool) {
	defer func() {
		// sink audit
		audithelper.Flush(resp.Request.Context())
		// sink token usage
		token_usage.Collect(resp)
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

	const (
		bodyHeadPersistBytes = 1024 * 30 // 30K
		bodyTailPersistBytes = 1024 * 2  // 2K
	)

	// audit response body
	defer func() {
		// put before cut off
		ctxhelper.PutReverseProxyWholeHandledResponseBodyStr(resp.Request.Context(), string(wholeHandledBody))
		wholeReceivedBody = truncateBodyForAudit(wholeReceivedBody, bodyHeadPersistBytes, bodyTailPersistBytes)
		wholeHandledBody = truncateBodyForAudit(wholeHandledBody, bodyHeadPersistBytes, bodyTailPersistBytes)
		audithelper.NoteAppend(resp.Request.Context(), "actual_response_body", string(wholeReceivedBody))
		audithelper.NoteAppend(resp.Request.Context(), "response_body", string(wholeHandledBody))
	}()

	// audit response time
	var responseChunkBeginAt time.Time
	defer func() {
		responseChunkDoneAt := time.Now()
		audithelper.NoteOnce(resp.Request.Context(), "response_chunk_done_at", responseChunkDoneAt)
	}()

	// processChunk handles a single chunk through all filters and writes to output
	processChunk := func(chunk []byte, chunkIndex int64, rerr error) bool {
		if rerr != nil && rerr == io.EOF {
			ctxhelper.PutIsLastBodyChunk(resp.Request.Context(), true)
		}
		// begin response here
		if responseChunkBeginAt.IsZero() {
			responseChunkBeginAt = time.Now()
			audithelper.NoteOnce(resp.Request.Context(), "response_chunk_begin_at", responseChunkBeginAt) // table metadata
		}
		if len(chunk) == 0 && rerr == nil {
			// indicates splitter violated contract, avoid infinite loop
			writeAndCloseWithErr(resp, pw, errors.New("splitter returned empty chunk without error"))
			return false // stop
		}
		if len(chunk) > 0 {
			out := chunk
			// same chunk flows through all Modifiers in sequence
			dumpReceivedOut := dumplog.DumpResponseBodyChunk(resp, out, chunkIndex)
			wholeReceivedBody = append(wholeReceivedBody, dumpReceivedOut...)
			for _, filter := range filters {
				currentFilterName = filter.Name
				m := filter.Instance
				logutil.InjectLoggerWithFilterInfo(resp.Request.Context(), filter)
				if fe, ok := m.(filter_define.ProxyResponseModifierEnabler); ok && !fe.Enable(resp) {
					continue
				}
				o, err := m.OnBodyChunk(resp, out, chunkIndex)
				if err != nil {
					writeAndCloseWithErr(resp, pw, err)
					return false // stop
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
					return false // stop
				}
			}
		}

		if rerr != nil { // EOF or real error
			if rerr != io.EOF && !errors.Is(rerr, context.Canceled) {
				writeAndCloseWithErr(resp, pw, rerr)
			}
			// cleanup: call OnComplete; ensure call even if previous errors
			dumplog.DumpResponseComplete(resp)
			for _, filter := range filters {
				currentFilterName = filter.Name
				if m, ok := filter.Instance.(filter_define.ProxyResponseModifier); ok {
					logutil.InjectLoggerWithFilterInfo(resp.Request.Context(), filter)
					if fe, ok := m.(filter_define.ProxyResponseModifierEnabler); ok && !fe.Enable(resp) {
						continue
					}
					out, err := m.OnComplete(resp)
					if err != nil {
						audithelper.Note(resp.Request.Context(), fmt.Sprintf("myResponseModify.%s.OnComplete.error", filter.Name), err.Error())
					}
					if len(out) > 0 {
						wholeHandledBody = append(wholeHandledBody, out...)
						_, _ = pw.Write(out)
					}
				}
			}
			currentFilterName = ""
			_ = pw.Close()
			return false // stop
		}
		return true // continue
	}

	var chunkIndex int64 = FirstChunkIndex - 1

	// process peeked chunk first (if any)
	if len(peekedChunk) > 0 {
		chunkIndex++
		var inputErr error
		if peekedChunkIsEOF {
			inputErr = io.EOF
		}
		if !processChunk(peekedChunk, chunkIndex, inputErr) {
			return
		}
		// if peekedChunkErr was io.EOF, we're done
		if peekedChunkIsEOF {
			return
		}
	}

	// continue reading remaining chunks from upstream
	for {
		chunkIndex++
		chunk, rerr := splitter.NextChunk(upstream)
		if !processChunk(chunk, chunkIndex, rerr) {
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
		"type":                   "llm-backend-error",
		"raw_llm_backend_status": fmt.Sprintf("%d (%s)", resp.StatusCode, http.StatusText(resp.StatusCode)),
	}

	// Content-Length may be -1, so only check body
	if resp.Body != nil {
		wholeStreamSplitter := set_resp_body_chunk_splitter.WholeStreamSplitter{}
		autoDecompressSplitter := set_resp_body_chunk_splitter.NewDecompressingChunkSplitter(wholeStreamSplitter, resp)
		respBodyBytes, err := autoDecompressSplitter.NextChunk(resp.Body)
		if err != nil && err != io.EOF {
			ctxhelper.MustGetLoggerBase(resp.Request.Context()).Warnf("failed to read resp body at error-status-handler: %v", err)
		} else if len(respBodyBytes) > 0 {
			// try to parse backend response as JSON, if so preserve it
			var jsonResp interface{}
			if json.Unmarshal(respBodyBytes, &jsonResp) == nil {
				errCtx["raw_llm_backend_response"] = jsonResp
			} else {
				errCtx["raw_llm_backend_response"] = string(respBodyBytes)
			}
			audithelper.Note(resp.Request.Context(), "actual_response_body", string(respBodyBytes))
		}
	}

	httpErr := httperror.NewHTTPErrorWithCtx(
		resp.Request.Context(),
		resp.StatusCode,
		"LLM Backend Error",
		errCtx,
	)
	httpErr.AIProxyMeta = map[string]any{
		vars.XAIProxyModel: func() string {
			model, _ := ctxhelper.GetModel(resp.Request.Context())
			if model != nil {
				return model.Name
			}
			return ""
		}(),
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

func truncateBodyForAudit(body []byte, headLimit, tailLimit int) []byte {
	if len(body) == 0 {
		return body
	}
	if headLimit < 0 {
		headLimit = 0
	}
	if tailLimit < 0 {
		tailLimit = 0
	}
	if headLimit+tailLimit >= len(body) {
		return body
	}
	if tailLimit > len(body) {
		tailLimit = len(body)
	}
	if headLimit > len(body)-tailLimit {
		headLimit = len(body) - tailLimit
	}
	omittedBytes := len(body) - headLimit - tailLimit
	placeholder := []byte(fmt.Sprintf("...[omitted %d bytes]...", omittedBytes))
	truncated := make([]byte, 0, headLimit+len(placeholder)+tailLimit)
	truncated = append(truncated, body[:headLimit]...)
	truncated = append(truncated, placeholder...)
	if tailLimit > 0 {
		truncated = append(truncated, body[len(body)-tailLimit:]...)
	}
	return truncated
}
