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

package http

import (
	"strings"

	nethttp "net/http"

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
