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

package common

import (
	"context"
	"errors"
	"strings"

	"github.com/erda-project/erda-infra/pkg/transport/interceptor"
	jaegerpb "github.com/erda-project/erda-proto-go/oap/collector/receiver/jaeger/pb"
)

var (
	INVALID_MSP_ENV_ID        = errors.New("invalid msp.env.id field")
	INVALID_MSP_ACCESS_KEY    = errors.New("invalid msp.access.key field")
	INVALID_MSP_ACCESS_SECRET = errors.New("invalid msp.access.secret field")
)

func TagOverwrite(next interceptor.Handler) interceptor.Handler {
	return func(ctx context.Context, req interface{}) (interface{}, error) {
		if data, ok := req.(*jaegerpb.PostSpansRequest); ok {
			if data.Spans != nil {
				for _, span := range data.Spans {
					for k, v := range span.Attributes {
						if idx := strings.Index(k, "."); idx > -1 {
							span.Attributes[strings.Replace(k, ".", "_", -1)] = v
							delete(span.Attributes, k)
						}
					}
					delete(span.Attributes, TAG_MSP_AK_ID)
					delete(span.Attributes, TAG_MSP_AK_SECRET)
					if _, ok := span.Attributes[TAG_TERMINUS_KEY]; !ok {
						span.Attributes[TAG_TERMINUS_KEY] = span.Attributes[TAG_MSP_ENV_ID]
					}
				}
			}
		}
		return next(ctx, req)
	}
}

func Authentication(next interceptor.Handler) interceptor.Handler {
	return func(ctx context.Context, req interface{}) (interface{}, error) {
		if data, ok := req.(*jaegerpb.PostSpansRequest); ok {
			if data.Principal.Identity == "" {
				return nil, INVALID_MSP_ENV_ID
			}
			if data.Principal.AccessKey == "" {
				return nil, INVALID_MSP_ACCESS_KEY
			}
			if data.Principal.AccessSecret == "" {
				return nil, INVALID_MSP_ACCESS_SECRET
			}
		}
		return next(ctx, req)
	}
}
