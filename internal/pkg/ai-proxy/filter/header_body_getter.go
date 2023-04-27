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

package filter

import (
	"bytes"
	"context"
	"net/http"
	"net/url"
	"reflect"
	"sync"

	"github.com/erda-project/erda/pkg/http/httputil"
)

type HttpInfor interface {
	URL() *url.URL
	Status() string
	StatusCode() int
	Header() http.Header
	Body() (*bytes.Buffer, error)
	ContentLength() int64
	Host() string
	RemoteAddr() string
}

func NewInfor[R httputil.RequestResponse](ctx context.Context, r R) (HttpInfor, error) {
	i, err := httputil.DeepCopyRequestResponse(r)
	if err != nil {
		return nil, err
	}
	return headerBodyGetter[R]{r: i, l: ctx.Value(MutexCtxKey{}).(*sync.Mutex)}, nil // panic early
}

type headerBodyGetter[R httputil.RequestResponse] struct {
	r R
	l *sync.Mutex
}

func (r headerBodyGetter[R]) RemoteAddr() string {
	switch i := (interface{})(r.r).(type) {
	case http.Request:
		return i.RemoteAddr
	case *http.Request:
		return i.RemoteAddr
	case http.Response:
		if i.Request != nil {
			return i.Request.RemoteAddr
		}
	case *http.Response:
		if i.Request != nil {
			return i.Request.RemoteAddr
		}
	}
	return ""
}

func (r headerBodyGetter[R]) Host() string {
	switch i := (interface{})(r.r).(type) {
	case http.Request:
		return i.Host
	case *http.Request:
		return i.Host
	case http.Response:
		if i.Request != nil {
			return i.Request.Host
		}
	case *http.Response:
		if i.Request != nil {
			return i.Request.Host
		}
	}
	return ""
}

func (r headerBodyGetter[R]) URL() *url.URL {
	switch i := (interface{})(r.r).(type) {
	case http.Request:
		return i.Clone(context.Background()).URL
	case *http.Request:
		return i.Clone(context.Background()).URL
	case http.Response:
		if i.Request != nil {
			return i.Request.Clone(context.Background()).URL
		}
	case *http.Response:
		if i.Request != nil {
			return i.Request.Clone(context.Background()).URL
		}
	}
	return &url.URL{}
}

func (r headerBodyGetter[R]) ContentLength() int64 {
	field := reflect.ValueOf(r.r)
	if field.Kind() == reflect.Ptr {
		field = field.Elem()
	}
	return field.FieldByName("ContentLength").Int()
}

func (r headerBodyGetter[R]) Status() string {
	field := reflect.ValueOf(r.r)
	if field.Kind() == reflect.Ptr {
		field = field.Elem()
	}
	return field.FieldByName("Status").String()
}

func (r headerBodyGetter[R]) StatusCode() int {
	field := reflect.ValueOf(r.r)
	if field.Kind() == reflect.Ptr {
		field = field.Elem()
	}
	return int(field.FieldByName("StatusCode").Int())
}

func (r headerBodyGetter[R]) Header() http.Header {
	r.l.Lock()
	header := httputil.CopyHeader(r.r)
	r.l.Unlock()
	return header
}

func (r headerBodyGetter[R]) Body() (*bytes.Buffer, error) {
	r.l.Lock()
	buf, err := httputil.NopCloseReadBody(r.r)
	r.l.Unlock()
	return buf, err
}
