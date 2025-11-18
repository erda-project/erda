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

package body_util

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func SetBody(r *http.Request, newBody any) error {
	switch v := newBody.(type) {
	case []byte, string:
		b := v.([]byte)
		r.Body = io.NopCloser(bytes.NewBuffer(b))
		r.ContentLength = int64(len(b))
		r.Header.Set("Content-Length", fmt.Sprintf("%d", len(b)))
		return nil
	case io.Reader:
		r.Body = io.NopCloser(v)
		r.ContentLength = -1
		r.Header.Del("Content-Length")
		return nil
	case io.ReadCloser:
		r.Body = v
		r.ContentLength = -1
		r.Header.Del("Content-Length")
		return nil
	case proto.Message:
		// use protojson for pb message
		b, err := (protojson.MarshalOptions{
			UseProtoNames: false,
		}).Marshal(v)
		if err != nil {
			return fmt.Errorf("failed to marshal proto message: %w", err)
		}
		r.Body = io.NopCloser(bytes.NewBuffer(b))
		r.ContentLength = int64(len(b))
		r.Header.Set("Content-Length", fmt.Sprintf("%d", len(b)))
		r.Header.Set("Content-Type", "application/json")
		return nil
	default:
		b, err := json.Marshal(v)
		if err != nil {
			return fmt.Errorf("failed to marshal newBody: %w", err)
		}
		r.Body = io.NopCloser(bytes.NewBuffer(b))
		r.ContentLength = int64(len(b))
		r.Header.Set("Content-Length", fmt.Sprintf("%d", len(b)))
		return nil
	}
}
