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
	"context"
	"encoding/json"
	"net/http"
)

const (
	Continue Signal = iota
	Intercept
)

type Filter interface {
	RequestFilter
	ResponseFilter
}

type RequestFilter interface {
	OnHttpRequest(ctx context.Context, w http.ResponseWriter, r *http.Request) Signal
}

type ResponseFilter interface {
	OnHttpResponse(ctx context.Context, w http.ResponseWriter, r *http.Request) Signal
}

type DefaultFilter struct {
	Name   string          `json:"name" yaml:"name"`
	Config json.RawMessage `json:"config" yaml:"config"`
}

type Signal int

// RouteCtxKey 用以从 context.Context 中获取 route.Route 以获取 route.Route 的更多配置信息
type RouteCtxKey struct{}

type ProviderCtxKey struct{}
