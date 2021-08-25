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

package http

import (
	nethttp "net/http"
	"strings"

	"github.com/pkg/errors"
)

var (
	ErrAcceptHeader            = errors.New("Accept header error; should look like `application/vnd.dice+json; version=1.0`")
	ErrAcceptHeaderNotExist    = errors.New("Accept header not exist")
	ErrRequestIDHeaderNotExist = errors.New("Request-ID header not exist")
	ErrContentTypeHeader       = errors.New("Content-Type header error; it should be `application/json`")
)

func ValidateRequest(req *nethttp.Request) (*nethttp.Request, error) {
	if err := acceptHeader(req.Header); err != nil {
		return nil, err
	}
	if err := requestIDHeader(req.Header); err != nil {
		return nil, err
	}
	if err := contentTypeHeader(req.Header); err != nil {
		return nil, err
	}
	return req, nil
}

func ValidateResponse(res *nethttp.Response) (*nethttp.Response, error) {
	return res, nil
}

// all dice API should return JSON result, so
// `accept` should look like `Accept: application/vnd.dice+json; version=1.0`
func acceptHeader(headers nethttp.Header) error {
	accept := headers.Get("Accept")
	if accept == "" {
		return ErrAcceptHeaderNotExist
	}
	if !strings.Contains(accept, "version") {
		return ErrAcceptHeader
	}
	if !strings.Contains(accept, "application/vnd.dice+json") {
		return ErrAcceptHeader
	}
	return nil
}

// all http request should carry a request-id
func requestIDHeader(headers nethttp.Header) error {
	requestID := headers.Get("Request-ID")
	if requestID == "" {
		return ErrRequestIDHeaderNotExist
	}
	return nil
}

// if Content-Type Header exist in request, then it should be json type (maybe other type),
// error if it is form type
func contentTypeHeader(headers nethttp.Header) error {
	contentType := headers.Get("Content-Type")
	if contentType == "" {
		return nil
	}
	if strings.Contains(contentType, "application/x-www-form-urlencoded") {
		return ErrContentTypeHeader
	}
	if strings.Contains(contentType, "multipart/form-data") {
		return ErrContentTypeHeader
	}
	return nil
}
