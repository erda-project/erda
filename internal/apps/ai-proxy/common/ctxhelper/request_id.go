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
	"sync"

	"github.com/erda-project/erda/internal/apps/ai-proxy/vars"
	"github.com/erda-project/erda/pkg/crypto/uuid"
)

func GetRequestID(ctx context.Context) (string, bool) {
	value, ok := ctx.Value(CtxKeyMap{}).(*sync.Map).Load(vars.MapKeyRequestID{})
	if !ok || value == nil {
		return "", false
	}
	requestID, ok := value.(string)
	if !ok {
		return "", false
	}
	return requestID, true
}

func MustGetRequestID(ctx context.Context) string {
	requestID, ok := GetRequestID(ctx)
	if !ok {
		panic("request ID not found in context")
	}
	return requestID
}

func PutRequestID(ctx context.Context, requestID string) {
	if requestID == "" {
		requestID = uuid.New()
	}
	m := ctx.Value(CtxKeyMap{}).(*sync.Map)
	m.Store(vars.MapKeyRequestID{}, requestID)
}

func GetGeneratedCallID(ctx context.Context) (string, bool) {
	value, ok := ctx.Value(CtxKeyMap{}).(*sync.Map).Load(vars.MapKeyGeneratedCallID{})
	if !ok || value == nil {
		return "", false
	}
	callID, ok := value.(string)
	if !ok {
		return "", false
	}
	return callID, true
}

func MustGetGeneratedCallID(ctx context.Context) string {
	callID, ok := GetGeneratedCallID(ctx)
	if !ok {
		panic("generated call ID not found in context")
	}
	return callID
}

func PutGeneratedCallID(ctx context.Context, callID string) {
	m := ctx.Value(CtxKeyMap{}).(*sync.Map)
	m.Store(vars.MapKeyGeneratedCallID{}, callID)
}
