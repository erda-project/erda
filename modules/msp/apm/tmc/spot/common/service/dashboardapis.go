// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package service

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	tmcconfig "github.com/erda-project/erda/modules/msp/apm/tmc/tmcconfig"
	"github.com/erda-project/erda/pkg/http/httpclient"
)

func ProxyBody(path string, request *http.Request, params map[string][]string, debug bool) (interface{}, error) {
	body, err := request.GetBody()
	if err != nil {
		return nil, err
	}
	arrBody, err := inputToByte(body)
	if err != nil {
		return nil, err
	}
	return Proxy(path, request, params, arrBody, debug)
}

func Proxy(path string, request *http.Request, params map[string][]string, body interface{}, debug bool) (interface{}, error) {
	if len(params) > 0 {
		for key, values := range request.URL.Query() {
			_, ok := params[key]
			if len(values) != 0 && !ok {
				params[key] = append(params[key], values...)
			}
		}
		return ProxyUrl(path, request.Method, params, &request.Header, body, debug)
	}
	return ProxyUrl(path, request.Method, request.URL.RawQuery, &request.Header, body, debug)
}

func ProxyUrl(path, method string, params interface{}, headers *http.Header, body interface{}, debug bool) (interface{}, error) {
	var urlValues url.Values
	if params != nil {
		switch v := params.(type) {
		case string:
			if v, err := url.ParseQuery(v); err != nil {
				return nil, err
			} else {
				urlValues = v
			}
		default:
			encodedParams := make(map[string][]string)
			switch query := params.(type) {
			case map[string][]string:
				for key, values := range query {
					if len(values) != 0 {
						for _, value := range values {
							if value != "" {
								encodeKey := encodeURIComponent(key)
								encodedParams[encodeKey] = append(encodedParams[encodeKey], encodeURIComponent(value))
							}
						}
					}
				}
			case map[string]interface{}:
				for key, v := range query {
					if v != nil {
						switch values := v.(type) {
						case []string:
							for _, value := range values {
								if value != "" {
									encodeKey := encodeURIComponent(key)
									encodedParams[encodeKey] = append(encodedParams[encodeKey], encodeURIComponent(value))
								}
							}
						case []interface{}:
							for _, value := range values {
								if value != nil {
									encodeKey := encodeURIComponent(key)
									encodedParams[encodeKey] = append(encodedParams[encodeKey], encodeURIComponent(value.(string)))
								}
							}
						default:
							encodeKey := encodeURIComponent(key)
							encodedParams[encodeKey] = append(encodedParams[encodeKey], encodeURIComponent(values.(string)))
						}
					}
				}
			}
			urlValues = encodedParams
		}
	}

	client := httpclient.New()
	var req *httpclient.Request
	switch method {
	case "GET":
		req = client.Get(tmcconfig.Conf.MonitorAddr)
	case "POST":
		req = client.Post(tmcconfig.Conf.MonitorAddr)
	}
	req = req.Path(path).
		Params(urlValues).
		Headers(*headers)

	if method == "POST" {
		req = req.RawBody(body.(io.Reader))
		req = req.Header("Content-Type", "application/json")
	}

	var buf bytes.Buffer
	_, err := req.Do().Body(&buf)
	if debug {
		println(req.GetUrl())
		println(buf.String())
	}
	if err != nil {
		return err, nil
	}
	return buf.String(), nil
}

func inputToByte(inStream io.ReadCloser) ([]byte, error) {
	return ioutil.ReadAll(inStream)
}

func encodeURIComponent(s string) string {
	if s == "" {
		return s
	}

	return strings.ReplaceAll(
		strings.ReplaceAll(
			strings.ReplaceAll(
				strings.ReplaceAll(
					strings.ReplaceAll(
						strings.ReplaceAll(string(s), "\\+", "%20"), "\\%21", "!"), "\\%27", "'"), "\\%28", "("), "\\%29", ")"), "\\%7E", "~")
}
