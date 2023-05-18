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

package reverseproxy

import (
	"bytes"
	"fmt"
	"github.com/erda-project/erda-infra/base/logs"
	"io"
	"net/http"
	"reflect"
	"time"
)

var (
	_ http.RoundTripper = (*DoNothingTransport)(nil)
)

type DoNothingTransport struct {
	Response *http.Response
}

func (d *DoNothingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if d.Response == nil {
		d.Response = &http.Response{
			Status:           "",
			StatusCode:       0,
			Proto:            "",
			ProtoMajor:       0,
			ProtoMinor:       0,
			Header:           make(http.Header),
			Body:             io.NopCloser(bytes.NewReader(nil)),
			ContentLength:    0,
			TransferEncoding: nil,
			Close:            false,
			Uncompressed:     false,
			Trailer:          nil,
			Request:          req,
			TLS:              nil,
		}
	}
	return d.Response, nil
}

type TimerTransport struct {
	Logger logs.Logger
	Inner  http.RoundTripper
}

func (t *TimerTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	start := time.Now()
	if t.Inner == nil {
		t.Inner = http.DefaultTransport
	}
	res, err := t.Inner.RoundTrip(req)
	t.Logger.Sub(reflect.TypeOf(t).String()).
		Debugf("RoundTrip costs: %s", time.Now().Sub(start).String())
	return res, err
}

type CurlPrinterTransport struct {
	Logger logs.Logger
	Inner  http.RoundTripper
}

func (t *CurlPrinterTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if t.Inner == nil {
		t.Inner = http.DefaultTransport
	}
	t.Logger.Sub(reflect.TypeOf(t).String()).
		Debug(GenCurl(req))
	return t.Inner.RoundTrip(req)
}

func GenCurl(req *http.Request) string {
	var curl = fmt.Sprintf(`curl -v -N -X %s '%s://%s%s'`, req.Method, req.URL.Scheme, req.Host, req.URL.RequestURI())
	for k, vv := range req.Header {
		for _, v := range vv {
			curl += fmt.Sprintf(` -H '%s: %s'`, k, v)
		}
	}
	if req.Body != nil {
		var buf = bytes.NewBuffer(nil)
		if _, err := buf.ReadFrom(req.Body); err == nil {
			_ = req.Body.Close()
			curl += ` --data '` + buf.String() + `'`
			req.Body = io.NopCloser(buf)
		}
	}
	return curl
}
