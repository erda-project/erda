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
	"github.com/erda-project/erda/pkg/http/httputil"
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
		Debug("generated cURL command:\n\t" + GenCurl(NewInfor(nil, req)))
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
	TLSHandshakeTimeout:   10 * time.Second,
	MaxIdleConns:          100,
	IdleConnTimeout:       90 * time.Second,
	ExpectContinueTimeout: 1 * time.Second,
	ForceAttemptHTTP2:     true,
}

func GenCurl(infor HttpInfor) string {
	var curl = fmt.Sprintf(`curl -v -N -X %s '%s://%s%s'`, infor.Method(), infor.URL().Scheme, infor.Host(), infor.URL().RequestURI())
	for k, vv := range infor.Header() {
		for _, v := range vv {
			if strings.EqualFold(k, httputil.HeaderKeyContentLength) {
				continue
			}
			curl += fmt.Sprintf(` -H '%s: %s'`, k, v)
		}
	}
	if bodyBuffer := infor.BodyBuffer(); bodyBuffer != nil {
		// handle multipart form format
		if strings.HasPrefix(infor.Header().Get(httputil.HeaderKeyContentType), httputil.ContentTypeMultiPartFormData) {
			return genCurlPartsForMultipartForm(curl, bodyBuffer)
		}
		// normal
		curl += ` --data ` + strconv.Quote(bodyBuffer.String())
	}
	return curl
}

func genCurlPartsForMultipartForm(curl string, bodyBuffer *bytes.Buffer) string {
	lines := strings.Split(bodyBuffer.String(), "\r\n")
	var fieldKey, value, fileName string
	for _, line := range lines {
		if strings.HasPrefix(line, "---") {
			if fieldKey == "" {
				continue
			}
			if value != "" {
				curl += fmt.Sprintf(` --form %s`, strconv.Quote(fieldKey+"="+value))
			}
			if fileName != "" {
				curl += fmt.Sprintf(` --form %s`, strconv.Quote(fieldKey+"=@"+fileName))
			}
			fieldKey = ""
			value = ""
			fileName = ""
			continue
		}
		if strings.HasPrefix(line, httputil.HeaderKeyContentDisposition+": form-data") {
			ss := strings.Split(line, ";")
			for _, s := range ss {
				s = strings.TrimSpace(s)
				if strings.HasPrefix(s, "name=") {
					fieldKey = strings.Trim(strings.TrimPrefix(s, "name="), `"`)
				}
				if strings.HasPrefix(s, "filename=") {
					fileName = strings.Trim(strings.TrimPrefix(s, "filename="), `"`)
				}
			}
		} else {
			if fileName == "" && line != "" {
				// not file field, treat coming lines as value. Or the coming lines are file content (data).
				if value != "" {
					value += "\n"
				}
				value += line
			}
		}
	}
	return curl
}
