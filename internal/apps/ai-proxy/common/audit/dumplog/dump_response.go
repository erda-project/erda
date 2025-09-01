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

package dumplog

import (
	"net/http"
	"net/http/httputil"

	"github.com/erda-project/erda/internal/apps/ai-proxy/common/audit/audithelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	httperrorutil "github.com/erda-project/erda/pkg/http/httputil"
)

func DumpResponseHeadersIn(resp *http.Response) {
	dumpBytes, err := httputil.DumpResponse(resp, false)
	if err != nil {
		ctxhelper.MustGetLoggerBase(resp.Request.Context()).Warnf("failed to dump response headers in: %v", err)
		return
	}
	ctxhelper.MustGetLoggerBase(resp.Request.Context()).Infof("dump proxy response headers in:\n%s", string(dumpBytes))
	// collect actual llm response info
	audithelper.Note(resp.Request.Context(), "status", resp.StatusCode)
	audithelper.Note(resp.Request.Context(), "actual_response_header", resp.Header.Clone())
	audithelper.Note(resp.Request.Context(), "actual_response_content_type", resp.Header.Get(httperrorutil.HeaderKeyContentType))
}

func DumpResponseHeadersOut(resp *http.Response) {
	dumpBytes, err := httputil.DumpResponse(resp, false)
	if err != nil {
		ctxhelper.MustGetLoggerBase(resp.Request.Context()).Warnf("failed to dump response headers out: %v", err)
		return
	}
	ctxhelper.MustGetLoggerBase(resp.Request.Context()).Infof("dump proxy response headers out:\n%s", string(dumpBytes))
	// collect handled response info
	audithelper.Note(resp.Request.Context(), "response_header", resp.Header.Clone())
	audithelper.Note(resp.Request.Context(), "response_content_type", resp.Header.Get(httperrorutil.HeaderKeyContentType))
}

func DumpResponseBodyChunk(resp *http.Response, chunk []byte, index int64) {
	contentType := resp.Header.Get("Content-Type")
	shouldDumpBody := ShouldDumpBody(contentType)

	if shouldDumpBody {
		ctxhelper.MustGetLoggerBase(resp.Request.Context()).Infof("dump proxy response body chunk (index=%d):\n%s", index, string(chunk))
	} else {
		ctxhelper.MustGetLoggerBase(resp.Request.Context()).Infof("dump proxy response body chunk (index=%d): [%d bytes omitted due to binary content-type: %s]",
			index, len(chunk), contentType)
	}
}

func DumpResponseComplete(resp *http.Response) {
	return
}
