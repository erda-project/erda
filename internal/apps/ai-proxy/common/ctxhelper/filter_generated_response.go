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

package ctxhelper

import (
	"context"
	"net/http"
	"sync"
)

type ctxKeyFilterGeneratedResponse struct{}

// PutRequestFilterGeneratedResponse stores filter-generated response in context
func PutRequestFilterGeneratedResponse(ctx context.Context, resp *http.Response) {
	m := ctx.Value(CtxKeyMap{}).(*sync.Map)
	m.Store(ctxKeyFilterGeneratedResponse{}, resp)
}

// GetRequestFilterGeneratedResponse retrieves filter-generated response from context
func GetRequestFilterGeneratedResponse(ctx context.Context) (*http.Response, bool) {
	value, ok := ctx.Value(CtxKeyMap{}).(*sync.Map).Load(ctxKeyFilterGeneratedResponse{})
	if !ok || value == nil {
		return nil, false
	}
	resp, ok := value.(*http.Response)
	if !ok {
		return nil, false
	}
	return resp, true
}
