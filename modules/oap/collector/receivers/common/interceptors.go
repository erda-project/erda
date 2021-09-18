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
	"github.com/erda-project/erda/modules/oap/collector/authentication"
	"github.com/erda-project/erda/pkg/common/apis"
)

var (
	INVALID_MSP_ENV_ID    = errors.New("invalid msp.env.id tag")
	INVALID_MSP_ENV_TOKEN = errors.New("invalid msp.env.token tag")
	AUTHENTICATION_FAILED = errors.New("authentication failed, please use the correct accessKey and accessKeySecret")
)

type Interceptors interface {
	Authentication(next interceptor.Handler) interceptor.Handler

	SpanTagOverwrite(next interceptor.Handler) interceptor.Handler
}

type interceptorImpl struct {
	validator authentication.Validator
}

func (i *interceptorImpl) SpanTagOverwrite(next interceptor.Handler) interceptor.Handler {
	return func(ctx context.Context, req interface{}) (interface{}, error) {
		spans, ok := Spans(ctx)
		if ok {
			for _, span := range spans {
				for k, v := range span.Attributes {
					if idx := strings.Index(k, "."); idx > -1 {
						span.Attributes[strings.Replace(k, ".", "_", -1)] = v
						delete(span.Attributes, k)
					}
				}
				delete(span.Attributes, TAG_MSP_ENV_TOKEN)
				if _, ok := span.Attributes[TAG_TERMINUS_KEY]; !ok {
					span.Attributes[TAG_TERMINUS_KEY] = span.Attributes[TAG_MSP_ENV_ID]
				}
			}
		}
		return next(ctx, req)
	}
}

func (i *interceptorImpl) Authentication(next interceptor.Handler) interceptor.Handler {
	return func(ctx context.Context, req interface{}) (interface{}, error) {
		envId := apis.GetHeader(ctx, HEADER_MSP_ENV_ID)
		token := apis.GetHeader(ctx, HEADER_MSP_ENV_TOKEN)

		if envId == "" {
			return nil, INVALID_MSP_ENV_ID
		}
		if token == "" {
			return nil, INVALID_MSP_ENV_TOKEN
		}

		if !i.validator.Validate(SCOPE_MSP_ENV, envId, token) {
			return nil, AUTHENTICATION_FAILED
		}

		return next(ctx, req)
	}
}
