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
	"io"
	"net/http"
	"os"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda/pkg/http/httputil"
	"github.com/erda-project/erda/pkg/reverseproxy"
	"github.com/erda-project/erda/pkg/strutil"
)

const (
	Name = "log-http"
)

var (
	_ reverseproxy.RequestFilter  = (*LogHttp)(nil)
	_ reverseproxy.ResponseFilter = (*LogHttp)(nil)
)

func init() {
	reverseproxy.RegisterFilterCreator(Name, New)
}

type LogHttp struct {
	*reverseproxy.DefaultResponseFilter

	headerPrinted bool
	lineCount     int
}

func New(_ json.RawMessage) (reverseproxy.Filter, error) {
	return &LogHttp{DefaultResponseFilter: reverseproxy.NewDefaultResponseFilter()}, nil
}

func (f *LogHttp) OnRequest(ctx context.Context, w http.ResponseWriter, infor reverseproxy.HttpInfor) (signal reverseproxy.Signal, err error) {
	if !strutil.Equal(os.Getenv("LOG_LEVEL"), "debug") {
		return reverseproxy.Continue, nil
	}
	var l = ctx.Value(reverseproxy.LoggerCtxKey{}).(logs.Logger).Sub("LogHttp").Sub("OnRequest")
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
		l.Debug("request body is nil")
		m["body"] = json.RawMessage("null")
	} else {
		l.Debugf("request body: %s", body.String())
		if httputil.HeaderContains(infor.Header(), httputil.ApplicationJson) {
			m["body"] = json.RawMessage(body.Bytes())
		} else {
			m["body"] = body.String()
		}
	}
	l.Debugf("request info: %s", strutil.TryGetJsonStr(m))
	return reverseproxy.Continue, nil
}

func (f *LogHttp) OnResponseChunk(ctx context.Context, infor reverseproxy.HttpInfor, w reverseproxy.Writer, chunk []byte) (signal reverseproxy.Signal, err error) {
	if !strutil.Equal(os.Getenv("LOG_LEVEL"), "debug") {
		return f.DefaultResponseFilter.OnResponseChunk(ctx, infor, w, chunk)
	}
	var l = ctx.Value(reverseproxy.LoggerCtxKey{}).(logs.Logger).Sub("LogHttp").Sub("OnResponseChunk")
	if !f.headerPrinted {
		var m = map[string]any{
			"headers":     infor.Header(),
			"status":      infor.Status(),
			"status code": infor.StatusCode(),
		}
		l.Debugf("response info: %s", strutil.TryGetJsonStr(m))
		f.headerPrinted = true
	}
	if _, err = w.Write(chunk); err != nil {
		return reverseproxy.Intercept, err
	}
	for _, c := range chunk {
		if err := f.WriteByte(c); err != nil {
			return reverseproxy.Intercept, err
		}
		if c == '\n' {
			data, err := io.ReadAll(f)
			if err != nil {
				return reverseproxy.Intercept, err
			}
			l.Debugf("response body line[%d]: %s", f.lineCount, string(data))
			f.lineCount++
		}
	}
	return reverseproxy.Continue, nil
}

func (f *LogHttp) OnResponseEOF(ctx context.Context, infor reverseproxy.HttpInfor, w reverseproxy.Writer, chunk []byte) error {
	if err := f.DefaultResponseFilter.OnResponseEOF(ctx, infor, w, chunk); err != nil {
		return err
	}
	if !strutil.Equal(os.Getenv("LOG_LEVEL"), "debug") {
		return nil
	}
	data, err := io.ReadAll(f)
	if err != nil {
		return err
	}
	var l = ctx.Value(reverseproxy.LoggerCtxKey{}).(logs.Logger).Sub("LogHttp").Sub("OnResponseEOF")
	l.Debugf("response body last lines: %s", string(data))
	return nil
}
