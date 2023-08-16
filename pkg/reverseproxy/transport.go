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
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/http/httpproxy"

	"github.com/erda-project/erda-infra/base/logs"
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
		t.Inner = BaseTransport
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
		t.Inner = BaseTransport
	}
	t.Logger.Sub(reflect.TypeOf(t).String()).
		Debug("generated cURL command:\n\t" + GenCurl(req))
	return t.Inner.RoundTrip(req)
}

// ProxyConfig 是正向代理配置, 即 transport 出口流量的代理配置
var ProxyConfig = &httpproxy.Config{
	HTTPProxy:  os.Getenv("FORWARD_HTTP_PROXY"),
	HTTPSProxy: os.Getenv("FORWARD_HTTPS_PROXY"),
	NoProxy:    os.Getenv("NO_PROXY"),
	CGI:        os.Getenv("REQUEST_METHOD") != "",
}

// BaseTransport 返回一个基础的 http.RoundTripper. 它检查 *http.Request 所请求的 host 是否在 FORWARD_PROXY_HOSTS 清单内,
// 如果在清单内, 则使用 ProxyConfig 的代理配置, 如不在清单内, 则使用默认的代理配置 http.ProxyFromEnvironment.
var BaseTransport http.RoundTripper = &http.Transport{
	Proxy: func(req *http.Request) (*url.URL, error) {
		hosts := strings.Split(os.Getenv("FORWARD_PROXY_HOSTS"), ",")
		for _, host := range hosts {
			if req.Host == host || req.URL.Host == host {
				return ProxyConfig.ProxyFunc()(req.URL)
			}
		}
		return http.ProxyFromEnvironment(req)
	},
	DialContext: (&net.Dialer{
		Timeout:   60 * time.Second,
		KeepAlive: 60 * time.Second,
	}).DialContext,
	Dial:                   nil,
	DialTLSContext:         nil,
	DialTLS:                nil,
	TLSClientConfig:        nil,
	TLSHandshakeTimeout:    10 * time.Second,
	DisableKeepAlives:      false,
	DisableCompression:     false,
	MaxIdleConns:           100,
	MaxIdleConnsPerHost:    0,
	MaxConnsPerHost:        0,
	IdleConnTimeout:        90 * time.Second,
	ResponseHeaderTimeout:  0,
	ExpectContinueTimeout:  1 * time.Second,
	TLSNextProto:           nil,
	ProxyConnectHeader:     nil,
	GetProxyConnectHeader:  nil,
	MaxResponseHeaderBytes: 0,
	WriteBufferSize:        0,
	ReadBufferSize:         0,
	ForceAttemptHTTP2:      false,
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
			curl += ` --data ` + strconv.Quote(buf.String())
			req.Body = io.NopCloser(buf)
		}
	}
	return curl
}
