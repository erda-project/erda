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
	"net/url"
	"strings"

	"github.com/erda-project/erda-infra/pkg/transport"
	transhttp "github.com/erda-project/erda-infra/pkg/transport/http"
	"github.com/erda-project/erda-infra/pkg/transport/interceptor"
	"github.com/erda-project/erda/modules/oap/collector/authentication"
	"github.com/erda-project/erda/pkg/common/apis"
)

var (
	INVALID_MSP_ENV_ID    = errors.New("invalid erda.env.id tag")
	INVALID_MSP_ENV_TOKEN = errors.New("invalid erda.env.token tag")
	AUTHENTICATION_FAILED = errors.New("authentication failed, please use the correct accessKey and accessKeySecret")
)

type Interceptors interface {
	Authentication(next interceptor.Handler) interceptor.Handler

	SpanTagOverwrite(next interceptor.Handler) interceptor.Handler

	ExtractHttpHeaders(next interceptor.Handler) interceptor.Handler
}

type interceptorImpl struct {
	validator authentication.Validator
}

func (i *interceptorImpl) ExtractHttpHeaders(next interceptor.Handler) interceptor.Handler {
	return func(ctx context.Context, entity interface{}) (interface{}, error) {
		req := transhttp.ContextRequest(ctx)
		req.Header.Set("Accept", "application/json")

		ctxBaggage := transport.ContextHeader(ctx)

		if envId := req.Header.Get(HEADER_ERDA_ENV_ID); envId != "" {
			ctxBaggage.Set(HEADER_ERDA_ENV_ID, envId)
		}

		if token := req.Header.Get(HEADER_ERDA_ENV_TOKEN); token != "" {
			ctxBaggage.Set(HEADER_ERDA_ENV_TOKEN, token)
		}

		if orgName := req.Header.Get(HEADER_ERDA_ORG); orgName != "" {
			ctxBaggage.Set(HEADER_ERDA_ORG, orgName)
		}

		//if data, ok := entity.(*jaegerpb.PostSpansRequest); ok {
		//	ctx = common.WithSpans(ctx, data.Spans)
		//}
		return next(ctx, entity)
	}
}

func (i *interceptorImpl) SpanTagOverwrite(next interceptor.Handler) interceptor.Handler {
	return func(ctx context.Context, req interface{}) (interface{}, error) {
		envId := apis.GetHeader(ctx, HEADER_ERDA_ENV_ID)
		orgName := apis.GetHeader(ctx, HEADER_ERDA_ORG)
		if p, ok := req.(SpansProvider); ok && p.GetSpans() != nil {
			spans := p.GetSpans()
			for _, span := range spans {
				for k, v := range span.Attributes {
					key := k
					if idx := strings.Index(key, "."); idx > -1 {
						key = strings.Replace(k, ".", "_", -1)
						span.Attributes[key] = v
						delete(span.Attributes, k)
					}
					if idx := strings.Index(key, "erda_"); idx == 0 {
						span.Attributes[key[5:]] = v
						delete(span.Attributes, key)
					}
				}
				if _, ok := span.Attributes[TAG_ORG_NAME]; !ok {
					span.Attributes[TAG_ORG_NAME] = orgName
				}
				if _, ok := span.Attributes[TAG_ENV_ID]; !ok {
					span.Attributes[TAG_ENV_ID] = envId
				}
				if _, ok := span.Attributes[TAG_TERMINUS_KEY]; !ok {
					span.Attributes[TAG_TERMINUS_KEY] = span.Attributes[TAG_ENV_ID]
				}
				if _, ok := span.Attributes[TAG_SERVICE_ID]; !ok {
					span.Attributes[TAG_SERVICE_ID] = span.Attributes[TAG_SERVICE_NAME]
				}
				if _, ok := span.Attributes[TAG_IP]; ok {
					span.Attributes[TAG_SERVICE_INSTANCE_IP] = span.Attributes[TAG_IP]
					delete(span.Attributes, TAG_IP)
				}
				if _, ok := span.Attributes[TAG_HTTP_PATH]; !ok {
					if _, ok := span.Attributes[TAG_HTTP_URL]; ok {
						if u, err := url.Parse(span.Attributes[TAG_HTTP_URL]); err == nil {
							span.Attributes[TAG_HTTP_PATH] = u.Path
						}
					} else if _, ok := span.Attributes[TAG_HTTP_TARGET]; ok {
						if u, err := url.Parse(span.Attributes[TAG_HTTP_TARGET]); err == nil {
							span.Attributes[TAG_HTTP_PATH] = u.Path
						}
					}
				}
				if _, ok := span.Attributes[TAG_SERVICE_INSTANCE_ID]; !ok {
					if _, ok := span.Attributes[TAG_CLIENT_UUID]; ok {
						span.Attributes[TAG_SERVICE_INSTANCE_ID] = span.Attributes[TAG_CLIENT_UUID]
					}
				}

				delete(span.Attributes, TAG_ENV_TOKEN)
				delete(span.Attributes, TAG_ERDA_ENV_TOKEN)
			}
		}
		return next(ctx, req)
	}
}

func (i *interceptorImpl) Authentication(next interceptor.Handler) interceptor.Handler {
	return func(ctx context.Context, req interface{}) (interface{}, error) {
		envId := apis.GetHeader(ctx, HEADER_ERDA_ENV_ID)
		token := apis.GetHeader(ctx, HEADER_ERDA_ENV_TOKEN)

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
