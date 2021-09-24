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

	trace "github.com/erda-project/erda-proto-go/oap/trace/pb"
)

var (
	TAG_SERVICE_NAME        = "service_name"
	TAG_SERVICE_ID          = "service_id"
	TAG_SERVICE_INSTANCE_IP = "service_instance_ip"
	TAG_MSP_ENV_ID          = "msp_env_id"
	TAG_MSP_ENV_TOKEN       = "msp_env_token"
	TAG_TERMINUS_KEY        = "terminus_key"
	TAG_IP                  = "ip"
	TAG_HTTP_PATH           = "http_path"
	TAG_HTTP_URL            = "http_url"

	HEADER_MSP_ENV_ID    = "x-msp-env-id"
	HEADER_MSP_ENV_TOKEN = "x-msp-env-token"

	SCOPE_MSP_ENV = "msp_env"
)

type spanKey struct{}

func WithSpans(ctx context.Context, spans []*trace.Span) context.Context {
	return context.WithValue(ctx, spanKey{}, spans)
}

func Spans(ctx context.Context) ([]*trace.Span, bool) {
	val, ok := ctx.Value(spanKey{}).([]*trace.Span)
	if !ok {
		return nil, false
	}
	return val, true
}
