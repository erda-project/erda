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

package transports

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
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/body_util"
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
	Inner http.RoundTripper
}

func (t *CurlPrinterTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if t.Inner == nil {
		t.Inner = BaseTransport
	}
	logger := ctxhelper.MustGetLogger(req.Context())
	logger.Sub(reflect.TypeOf(t).String()).
		Infof("generated cURL command:\n\t" + GenCurl(req))
	return t.Inner.RoundTrip(req)
}

// ProxyConfig is forward proxy configuration, i.e., proxy configuration for transport outbound traffic
var ProxyConfig = &httpproxy.Config{
	HTTPProxy:  os.Getenv("FORWARD_HTTP_PROXY"),
	HTTPSProxy: os.Getenv("FORWARD_HTTPS_PROXY"),
	NoProxy:    os.Getenv("NO_PROXY"),
	CGI:        os.Getenv("REQUEST_METHOD") != "",
}

// BaseTransport returns a basic http.RoundTripper. It checks whether the host requested by *http.Request is in the FORWARD_PROXY_HOSTS list,
// if it is in the list, it uses ProxyConfig's proxy configuration, if not in the list, it uses the default proxy configuration http.ProxyFromEnvironment.
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
	DisableCompression:    false,
}

func GenCurl(req *http.Request) string {
	var curl = fmt.Sprintf(`curl -v -N -X %s '%s://%s%s'`, req.Method, req.URL.Scheme, req.Host, req.URL.RequestURI())
	for k, vv := range req.Header {
		for _, v := range vv {
			if strings.EqualFold(k, httputil.HeaderKeyContentLength) {
				continue
			}
			curl += fmt.Sprintf(` -H '%s: %s'`, k, v)
		}
	}
	bodyCopy, err := body_util.SmartCloneBody(&req.Body, body_util.MaxSample)
	if err != nil {
		ctxhelper.MustGetLogger(req.Context()).Errorf("failed to clone request body, err: %v", err)
		return "no curl generated"
	}
	defer bodyCopy.Close()
	if bodyCopy.Size() > 0 {
		bodyBytes, err := io.ReadAll(bodyCopy)
		if err != nil {
			ctxhelper.MustGetLogger(req.Context()).Errorf("failed to read cloned request body, err: %v", err)
			return "no curl generated"
		}
		// handle multipart form format
		if strings.HasPrefix(req.Header.Get(httputil.HeaderKeyContentType), httputil.ContentTypeMultiPartFormData) {
			return genCurlPartsForMultipartForm(curl, bodyBytes)
		}
		// normal
		//curl += ` --data ` + strconv.Quote(string(bodyBytes))
		bodyStr := string(bodyBytes)
		bodyStr = strings.ReplaceAll(bodyStr, `'`, `'\''`)
		curl += " --data '" + bodyStr + "'"
	}
	return curl
}

func genCurlPartsForMultipartForm(curl string, bodyBytes []byte) string {
	lines := strings.Split(string(bodyBytes), "\r\n")
	var fieldKey, value, fileName string
	var processingField bool

	for _, line := range lines {
		// Detect boundary line (independent line starting with --)
		if strings.HasPrefix(line, "--") {
			// Process previous field
			if processingField && fieldKey != "" {
				if fileName != "" {
					curl += fmt.Sprintf(` -F %s=@%s`, fieldKey, fileName)
				} else if value != "" {
					curl += fmt.Sprintf(` -F %s=%s`, fieldKey, strconv.Quote(value))
				}
			}
			// Reset field state
			fieldKey = ""
			value = ""
			fileName = ""
			processingField = false
			continue
		}

		if strings.HasPrefix(line, "Content-Disposition: form-data") {
			ss := strings.Split(line, ";")
			for _, s := range ss {
				s = strings.TrimSpace(s)
				if strings.HasPrefix(s, "name=") {
					fieldKey = strings.Trim(strings.TrimPrefix(s, "name="), `"`)
					processingField = true
				}
				if strings.HasPrefix(s, "filename=") {
					fileName = strings.Trim(strings.TrimPrefix(s, "filename="), `"`)
				}
			}
		} else if processingField && !strings.HasPrefix(line, "Content-") && line != "" {
			// Process field value (non-file content)
			if fileName == "" {
				if value != "" {
					value += "\n"
				}
				value += line
			}
		}
	}

	// Process last field
	if processingField && fieldKey != "" {
		if fileName != "" {
			curl += fmt.Sprintf(` -F %s=@%s`, fieldKey, fileName)
		} else if value != "" {
			curl += fmt.Sprintf(` -F %s=%s`, fieldKey, strconv.Quote(value))
		}
	}

	return curl
}
