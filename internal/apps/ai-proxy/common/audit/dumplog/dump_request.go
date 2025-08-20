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

	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
)

func DumpRequestIn(in *http.Request) {
	contentType := in.Header.Get("Content-Type")
	shouldDumpBody := ShouldDumpBody(contentType)

	dumpBytesIn, err := httputil.DumpRequest(in, shouldDumpBody)
	if err != nil {
		ctxhelper.MustGetLoggerBase(in.Context()).Warnf("failed to dump request in, err: %v", err)
		return
	}

	if shouldDumpBody {
		ctxhelper.MustGetLoggerBase(in.Context()).Infof("dump proxy request in:\n%s", string(dumpBytesIn))
	} else {
		ctxhelper.MustGetLoggerBase(in.Context()).Infof("dump proxy request in (body omitted due to binary content-type: %s):\n%s", contentType, string(dumpBytesIn))
	}

	// save to sink
	if sink, ok := ctxhelper.GetAuditSink(in.Context()); ok {
		sink.Note("request_body", string(dumpBytesIn))
	}
}

func DumpRequestOut(out *http.Request) {
	contentType := out.Header.Get("Content-Type")
	shouldDumpBody := ShouldDumpBody(contentType)

	dumpBytesOut, err := httputil.DumpRequestOut(out, shouldDumpBody)
	if err != nil {
		ctxhelper.MustGetLoggerBase(out.Context()).Warnf("failed to dump request out, err: %v", err)
		return
	}

	if shouldDumpBody {
		ctxhelper.MustGetLoggerBase(out.Context()).Infof("dump proxy request out:\n%s", dumpBytesOut)
	} else {
		ctxhelper.MustGetLoggerBase(out.Context()).Infof("dump proxy request out (body omitted due to binary content-type: %s):\n%s", contentType, dumpBytesOut)
	}

	// save to sink
	if sink, ok := ctxhelper.GetAuditSink(out.Context()); ok {
		sink.Note("actual_request_body", string(dumpBytesOut))
		sink.Note("actual_request_url", out.URL.String())
	}
}
