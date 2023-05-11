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

package log_http

import (
	"context"
	"encoding/json"
	"net/http"
	"os"

	"github.com/dspo/roundtrip"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda/internal/apps/ai-proxy/filters"
	"github.com/erda-project/erda/pkg/http/httputil"
	"github.com/erda-project/erda/pkg/strutil"
)

const (
	Name = "log-http"
)

var (
	_ roundtrip.RequestFilter        = (*LogHttp)(nil)
	_ roundtrip.ResponseStreamFilter = (*LogHttp)(nil)
)

func init() {
	filters.RegisterFilterCreator(Name, New)
}

type LogHttp struct {
	*roundtrip.DefaultResponseFilter

	headerPrinted bool
	lineCount     int
}

func New(_ json.RawMessage) (roundtrip.Filter, error) {
	return &LogHttp{DefaultResponseFilter: roundtrip.NewDefaultResponseFilter()}, nil
}

func (f *LogHttp) OnRequest(ctx context.Context, w http.ResponseWriter, infor roundtrip.HttpInfor) (signal roundtrip.Signal, err error) {
	if !strutil.Equal(os.Getenv("LOG_LEVEL"), "debug") {
		return roundtrip.Continue, nil
	}
	var l = ctx.Value(roundtrip.CtxKeyLogger{}).(logs.Logger)
	var url = infor.URL()
	var m = map[string]any{
		"scheme":        url.Scheme,
		"host":          infor.Host(),
		"url.host":      url.Host,
		"uri":           url.RequestURI(),
		"headers":       infor.Header(),
		"remoteAddr":    infor.RemoteAddr(),
		"contentType":   infor.Header().Get("Content-Type"),
		"contentLength": infor.ContentLength(),
	}
	if body := infor.BodyBuffer(); body == nil {
		m["body"] = json.RawMessage("null")
	} else {
		if httputil.HeaderContains(infor.Header(), httputil.ApplicationJson) {
			m["body"] = json.RawMessage(body.Bytes())
		} else {
			m["body"] = body.String()
		}
	}
	l.Debugf("request info: %s", strutil.TryGetJsonStr(m))
	return roundtrip.Continue, nil
}

func (f *LogHttp) OnResponseChunk(ctx context.Context, infor roundtrip.HttpInfor, w roundtrip.Writer, chunk []byte) (signal roundtrip.Signal, err error) {
	if strutil.Equal(os.Getenv("LOG_LEVEL"), "debug") && !f.headerPrinted {
		var l = ctx.Value(roundtrip.CtxKeyLogger{}).(logs.Logger)
		var m = map[string]any{
			"headers":     infor.Header(),
			"status":      infor.Status(),
			"status code": infor.StatusCode(),
		}
		l.Debugf("response info: %s", strutil.TryGetJsonStr(m))
		f.headerPrinted = true
	}
	return f.DefaultResponseFilter.OnResponseChunk(ctx, infor, w, chunk)
}

func (f *LogHttp) OnResponseEOF(ctx context.Context, infor roundtrip.HttpInfor, w roundtrip.Writer, chunk []byte) error {
	if err := f.DefaultResponseFilter.OnResponseEOF(ctx, infor, w, chunk); err != nil {
		return err
	}
	if !strutil.Equal(os.Getenv("LOG_LEVEL"), "debug") {
		return nil
	}
	var l = ctx.Value(roundtrip.CtxKeyLogger{}).(logs.Logger)
	if httputil.HeaderContains(infor.Header(), httputil.ApplicationJson) || f.Len() <= 1024 {
		l.Debugf("received response content: %s", f.String())
	} else {
		var content = f.Buffer.String()
		if len(content) > 300 {
			content = content[:200] + "  ... and more ...  " + content[len(content)-100:]
		} else if len(content) > 200 {
			content = content[:200] + "  ... and more ...  "
		}
		l.Debugf("received response content length: %d, content excerpt: %s", f.Buffer.Len(), content)
	}
	return nil
}
