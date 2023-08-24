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

// 本文件是对 Infra 框架中注册 http 接口的补充, 实现了 encoder, 将 proto 格式接口的返回值编码为 http.ResponseWriter,
// 使得用户可以像使用 Infra http-server provider 一样将 proto 接口注册到 Mux 上来.
// InfraResponseEncoder 中可以添加 InfraEncodeMiddle, 以在编码前对 http.ResponseWriter 做一些额外的操作, 如 CORS 等.
// 而 Infra 默认的 Encoder 不支持这样的操作, 它会在获取到响应后立即编码 http.ResponseWriter, 由于已写入 response body,
// 导致在 http.Handler 上的 middles 已不能再操作 response header 了.
//
// This file complements the Infra framework for registering http interfaces by implementing encoder, which encodes the return value of a proto-formatted interface as http.ResponseWriter, // allowing users to register the proto interface with Mux as if it were an Infra http-server provider.
// ResponseWriter, allowing users to register the proto interface with Mux as if they were using the Infra http-server provider.
// An InfraEncodeMiddle can be added to the InfraResponseEncoder to perform additional operations on the http.ResponseWriter before encoding, such as CORS.
// This is not supported by Infra's default Encoder, which encodes the http.ResponseWriter as soon as it gets the response, since it's already written to the response body.
// ResponseWriter as soon as it gets the response, and since it's already written to the response body, // the middles on the http.Handler can no longer manipulate the response header.

package mux

import (
	"encoding/json"
	"fmt"
	"mime"
	"net/http"
	"net/url"
	"strings"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	"github.com/erda-project/erda-infra/pkg/transport"
	transhttp "github.com/erda-project/erda-infra/pkg/transport/http"
	"github.com/erda-project/erda-infra/pkg/urlenc"
)

var InfraCORS InfraEncodeMiddle = func(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
}

type InfraResponseEncoder struct {
	Middles []InfraEncodeMiddle
}

func (re *InfraResponseEncoder) Encoder() transhttp.EncodeResponseFunc {
	return func(w http.ResponseWriter, r *http.Request, out any) error {
		for _, m := range re.Middles {
			m(w)
		}

		if out == nil {
			return nil
		}
		accept := r.Header.Get("Accept")
		var acceptAny bool
		if len(accept) > 0 {
			// TODO select MediaType of max q
			for _, item := range strings.Split(accept, ",") {
				mtype, _, err := mime.ParseMediaType(item)
				if err != nil {
					return err
				}
				if mtype == "*/*" {
					acceptAny = true
					continue
				}
				ok, err := encodeResponse(mtype, w, out)
				if ok {
					if err != nil {
						return err
					}
					return nil
				}
			}
		} else {
			_, err := encodeResponse("application/json", w, out)
			return err
		}
		if acceptAny {
			contentType := r.Header.Get("Content-Type")
			if len(contentType) > 0 {
				mtype, _, err := mime.ParseMediaType(contentType)
				if err != nil {
					return err
				}
				ok, err := encodeResponse(mtype, w, out)
				if ok {
					if err != nil {
						return err
					}
					return nil
				}
			}
			_, err := encodeResponse("application/json", w, out)
			return err
		}
		return notSupportMediaTypeErr{text: fmt.Sprintf("not support media type: %s", accept)}
	}
}

func InfraEncoderOpt(middles ...any) transport.ServiceOption {
	var opts []transhttp.HandleOption
	var encoder = &InfraResponseEncoder{}
	for _, m := range middles {
		switch middle := m.(type) {
		case Middle:
			opts = append(opts, transhttp.WithHTTPInterceptor(middle.MiddleFunc()))
		case InfraEncodeMiddle:
			encoder.Middles = append(encoder.Middles, middle)
		}
	}
	opts = append(opts, transhttp.WithEncoder(encoder.Encoder()))
	return transport.WithHTTPOptions(opts...)
}

type InfraEncodeMiddle func(w http.ResponseWriter)

func encodeResponse(mtype string, w http.ResponseWriter, out any) (bool, error) {
	switch mtype {
	case "application/protobuf", "application/x-protobuf":
		if msg, ok := out.(proto.Message); ok {
			byts, err := proto.Marshal(msg)
			if err != nil {
				return false, err
			}
			w.Header().Set("Content-Type", "application/protobuf")
			_, err = w.Write(byts)
			return true, err
		}
	case "application/x-www-form-urlencoded", "multipart/form-data":
		if m, ok := out.(urlenc.URLValuesMarshaler); ok {
			vals := make(url.Values)
			w.Header().Set("Content-Type", "application/x-www-form-urlencoded")
			return true, m.MarshalURLValues("", vals)
		}
	default:
		if mtype == "application/json" || (strings.HasPrefix(mtype, "application/vnd.") && strings.HasSuffix(mtype, "+json")) {
			if msg, ok := out.(proto.Message); ok {
				byts, err := protojson.Marshal(msg)
				if err != nil {
					return false, err
				}
				w.Header().Set("Content-Type", "application/json")
				_, err = w.Write(byts)
				return true, err
			}
			w.Header().Set("Content-Type", "application/json")
			return true, json.NewEncoder(w).Encode(out)
		}
	}
	return false, nil
}

type notSupportMediaTypeErr struct {
	text string
}

func (e notSupportMediaTypeErr) HTTPStatus() int { return http.StatusNotAcceptable }

func (e notSupportMediaTypeErr) Error() string { return e.text }
