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
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"sync"

	"github.com/erda-project/erda/pkg/http/httputil"
)

const (
	Continue Signal = iota
	Intercept
)

type Filter any

type RequestFilter interface {
	OnRequest(ctx context.Context, w http.ResponseWriter, infor HttpInfor) (signal Signal, err error)
}

type ResponseFilter interface {
	// OnResponseChunk 每被调用一次, 传入一个 response chunk, ResponseFilter 的实现者需要自行决定如何处理这些 chunk.
	// 对大多数情况来说, 实现者可以将这些 chunk 缓存到 filter 实例中, 待 response chunks 全部传完后整体处理.
	// 注意: w Writer 是将 response chunk 传入下一个 ResponseFilter 的句柄, 要将处理后的数据按需写入这个 Writer,
	// 不然后续的 ResponseFilter 都会丢失这部分数据.
	OnResponseChunk(ctx context.Context, infor HttpInfor, w Writer, chunk []byte) (signal Signal, err error)

	// OnResponseEOF 当 OnResponseEOF 被调用时, 表示这是最后一次传入 response chunk, OnResponseEOF 应当做一些收尾工作.
	// 比如 OnResponseChunk 截留了的数据, 可以在此时写入 w Writer.
	OnResponseEOF(ctx context.Context, infor HttpInfor, w Writer, chunk []byte) error
}
type HttpInfor interface {
	Method() string
	URL() *url.URL
	Status() string
	StatusCode() int
	Header() http.Header
	ContentLength() int64
	Host() string
	RemoteAddr() string
	// Body only for getting request body and only on request stage.
	Body() io.ReadCloser
	SetBody(body io.ReadCloser)
	// BodyBuffer only for getting request body and only on request stage.
	BodyBuffer() *bytes.Buffer
	Mutex() *sync.Mutex
}

func NewInfor[R httputil.RequestResponse](ctx context.Context, r R) HttpInfor {
	l, ok := ctx.Value(MutexCtxKey{}).(*sync.Mutex)
	if ok {
		return (&infor[R]{r: r}).WithMutex(l)
	}
	return (&infor[R]{r: r}).WithMutex(new(sync.Mutex))
}

type Writer interface {
	io.Writer
	io.ByteWriter
}

type infor[R httputil.RequestResponse] struct {
	r R
	l *sync.Mutex
}

func (r *infor[R]) Method() string {
	switch i := (interface{})(r.r).(type) {
	case http.Request:
		return i.Method
	case *http.Request:
		return i.Method
	case http.Response:
		if i.Request != nil {
			return i.Request.Method
		}
	case *http.Response:
		if i.Request != nil {
			return i.Request.Method
		}
	default:
		panic("not expected type")
	}
	return "UNKNOWN_METHOD"
}

func (r *infor[R]) RemoteAddr() string {
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
	default:
		panic("not expected type")
	}
	return ""
}

func (r *infor[R]) Host() string {
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
	default:
		panic("not expected type")
	}
	return ""
}

func (r *infor[R]) URL() *url.URL {
	switch i := (interface{})(r.r).(type) {
	case http.Request:
		return i.URL
	case *http.Request:
		return i.URL
	case http.Response:
		if i.Request != nil {
			return i.Request.URL
		}
	case *http.Response:
		if i.Request != nil {
			return i.Request.URL
		}
	default:
		panic("not expected type")
	}
	return nil
}

func (r *infor[R]) ContentLength() int64 {
	field := reflect.ValueOf(r.r)
	if field.Kind() == reflect.Ptr {
		field = field.Elem()
	}
	return field.FieldByName("ContentLength").Int()
}

func (r *infor[R]) Status() string {
	switch i := (interface{})(r.r).(type) {
	case http.Request, *http.Request:
	case http.Response:
		return i.Status
	case *http.Response:
		return i.Status
	default:
		panic("not expected type")
	}
	return ""
}

func (r *infor[R]) StatusCode() int {
	switch i := (interface{})(r.r).(type) {
	case http.Request, *http.Request:
	case http.Response:
		return i.StatusCode
	case *http.Response:
		return i.StatusCode
	default:
		panic("not expected type")
	}
	return 0
}

func (r *infor[R]) Header() http.Header {
	field := reflect.ValueOf(r.r)
	if field.Kind() == reflect.Ptr {
		field = field.Elem()
	}
	v := field.FieldByName("Header")
	if !v.IsValid() || v.IsNil() || v.IsZero() {
		return nil
	}
	return v.Interface().(http.Header)
}

// Body only for request body
func (r *infor[R]) Body() io.ReadCloser {
	switch i := (interface{})(r.r).(type) {
	case http.Request:
		return i.Body
	case *http.Request:
		return i.Body
	case http.Response, *http.Response:
		return nil
	default:
		panic("not expected type")
	}
}

// BodyBuffer only for request body
func (r *infor[R]) BodyBuffer() *bytes.Buffer {
	var request *http.Request
	switch i := (interface{})(r.r).(type) {
	case http.Request:
		request = &i
	case *http.Request:
		request = i
	case http.Response, *http.Response:
		return nil
	default:
		panic("not expected type")
	}

	if request.Body == nil {
		return nil
	}

	data, err := io.ReadAll(request.Body)
	if err != nil {
		return nil
	}
	_ = request.Body.Close()
	request.Body = io.NopCloser(bytes.NewReader(data))
	return bytes.NewBuffer(data)
}

func (r *infor[R]) SetBody(body io.ReadCloser) {
	field := reflect.ValueOf(r.r)
	if field.Kind() == reflect.Ptr {
		field = field.Elem()
	}
	v := field.FieldByName("Body")
	if v.IsValid() && !v.IsNil() && !v.IsZero() {
		i := v.Interface()
		_ = i.(io.Closer).Close()
	}
	v.Set(reflect.ValueOf(body))
}

func (r *infor[R]) Mutex() *sync.Mutex {
	return r.l
}

func (r *infor[R]) WithMutex(mutex *sync.Mutex) HttpInfor {
	r.l = mutex
	return r
}

type Signal int

func (s Signal) String() string {
	return strconv.FormatInt(int64(s), 10)
}

type FilterConfig struct {
	Name   string          `json:"name" yaml:"name"`
	Config json.RawMessage `json:"config" yaml:"config"`
}
