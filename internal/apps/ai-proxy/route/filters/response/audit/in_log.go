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

package audit

import (
	"net/http"
	"net/http/httputil"

	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/filters/request/audit/audit_util"
)

func (f *AuditResponse) inLogOnHeaders(resp *http.Response) error {
	logger := ctxhelper.MustGetLogger(resp.Request.Context())
	dumpBytes, err := httputil.DumpResponse(resp, false)
	if err != nil {
		logger.Warnf("failed to dump response headers (stage=%s): %v", f.Stage, err)
		return nil
	}
	logger.Infof("audit response headers (stage=%s):\n%s", f.Stage, string(dumpBytes))
	return nil
}

func (f *AuditResponse) inLogOnBodyChunk(resp *http.Response, chunk []byte) (out []byte, err error) {
	logger := ctxhelper.MustGetLogger(resp.Request.Context())

	// Decide whether to dump body based on content-type
	contentType := resp.Header.Get("Content-Type")
	shouldDumpBody := audit_util.ShouldDumpBody(contentType)

	if shouldDumpBody {
		logger.Infof("audit response body chunk (stage=%s, index=%d):\n%s",
			f.Stage, ctxhelper.MustGetResponseChunkIndex(resp.Request.Context()), string(chunk))
	} else {
		logger.Infof("audit response body chunk (stage=%s, index=%d): [%d bytes omitted due to binary content-type: %s]",
			f.Stage, ctxhelper.MustGetResponseChunkIndex(resp.Request.Context()), len(chunk), contentType)
	}
	return chunk, nil
}

func (f *AuditResponse) inLogOnComplete(resp *http.Response) ([]byte, error) {
	return nil, nil
}
