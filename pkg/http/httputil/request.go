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

package httputil

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"reflect"

	"github.com/mohae/deepcopy"
	"github.com/pkg/errors"
)

type RequestResponse interface {
	http.Request | ~*http.Request | http.Response | ~*http.Response
}

func DeepCopyRequestResponse[R RequestResponse](r R) (R, error) {
	var (
		result interface{}
		err    error
	)
	switch i := (interface{})(r).(type) {
	case http.Request:
		result, err = DeepCopyRequest(&i)
	case *http.Request:
		result, err = DeepCopyRequest(i)
	case http.Response:
		result, err = DeepCopyResponse(&i)
	case *http.Response:
		result, err = DeepCopyResponse(i)
	}
	return result.(R), err
}

func DeepCopyRequest(r *http.Request) (*http.Request, error) {
	clone := r.Clone(context.Background())
	if r.ContentLength == 0 || r.Body == nil {
		clone.Body = nil
		return clone, nil
	}
	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r.Body); err != nil {
		return nil, err
	}
	if err := r.Body.Close(); err != nil {
		return nil, err
	}
	r.Body = io.NopCloser(&buf)
	clone.Body = io.NopCloser(bytes.NewBuffer(buf.Bytes()))
	return clone, nil
}

func DeepCopyResponse(response *http.Response) (*http.Response, error) {
	request, err := DeepCopyRequest(response.Request)
	if err != nil {
		return nil, err
	}
	clone := deepcopy.Copy(response).(*http.Response)
	clone.Request = request
	if response.ContentLength == 0 || response.Body == nil {
		return clone, nil
	}
	var buf bytes.Buffer
	if _, err := buf.ReadFrom(response.Body); err != nil {
		return nil, err
	}
	if err := response.Body.Close(); err != nil {
		return nil, err
	}
	response.Body = io.NopCloser(&buf)
	clone.Body = io.NopCloser(bytes.NewBuffer(buf.Bytes()))
	return clone, err
}

func CopyHeader[R RequestResponse](r R) http.Header {
	var h = make(http.Header)
	field := reflect.ValueOf(r)
	if field.Kind() == reflect.Ptr {
		if field.IsNil() || field.IsZero() || !field.IsValid() {
			return h
		}
		field = field.Elem()
	}
	field = field.FieldByName("Header")
	if field.IsNil() || field.IsZero() || !field.IsValid() || field.Len() == 0 {
		return h
	}
	header := field.Interface().(http.Header)
	for k, vv := range header {
		for _, v := range vv {
			h.Add(k, v)
		}
	}
	return h
}

// NopCloseReadBody is equivalent to the following two functions
//
//	func NopCloseReadRequestBody(r *http.Request) ([]byte, error) {
//		if r == nil || r.Body == nil {
//			return nil, nil
//		}
//		var buf bytes.Buffer
//		if _, err := buf.ReadFrom(r.Body); err != nil {
//			return nil, err
//		}
//		if err := r.Body.Close(); err != nil {
//			return nil, err
//		}
//		r.Body = io.NopCloser(&buf)
//		return buf.Bytes(), nil
//	}
//
//	func NopCloseReadResponseBody(r *http.Response) ([]byte, error) {
//		if r == nil || r.Body == nil {
//			return nil, nil
//		}
//		var buf bytes.Buffer
//		if _, err := buf.ReadFrom(r.Body); err != nil {
//			return nil, err
//		}
//		if err := r.Body.Close(); err != nil {
//			return nil, err
//		}
//		r.Body = io.NopCloser(&buf)
//		return buf.Bytes(), nil
//	}
func NopCloseReadBody[R RequestResponse](r R) (*bytes.Buffer, error) {
	var buf bytes.Buffer
	if _, err := NopCloseReadBodyBuf(&buf, r); err != nil {
		return nil, err
	}
	return &buf, nil
}

func NopCloseReadBodyBuf[R RequestResponse](buf *bytes.Buffer, r R) (int64, error) {
	if buf == nil {
		return 0, errors.New("buf is nil")
	}
	value := reflect.ValueOf(r)
	if value.Kind() == reflect.Ptr {
		if value.IsNil() || value.IsZero() || !value.IsValid() {
			return 0, nil
		}
		value = value.Elem()
	}
	value = value.FieldByName("Body")
	if value.IsNil() || value.IsZero() || !value.IsValid() {
		return 0, nil
	}
	body := value.Interface().(io.ReadCloser)
	n, err := buf.ReadFrom(body)
	if err != nil {
		return 0, err
	}
	if err := body.Close(); err != nil {
		return 0, err
	}
	value.Set(reflect.ValueOf(io.NopCloser(bytes.NewBuffer(buf.Bytes()))))
	return n, nil
}
