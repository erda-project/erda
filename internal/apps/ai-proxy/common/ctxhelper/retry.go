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
	"reflect"
	"sync"
)

// ResetForRetry creates a fresh sync.Map context for a retry attempt.
//
// It carries over infrastructure-lifetime keys (DB, Logger, RequestID, etc.)
// from the old context, while clearing per-attempt keys (Model, AuditSink,
// Filter states).
func ResetForRetry(ctx context.Context) context.Context {
	oldMap, _ := ctx.Value(ctxKeyMap{}).(*sync.Map)
	if oldMap == nil {
		return ctx
	}

	newMap := &sync.Map{}
	newCtx := context.WithValue(ctx, ctxKeyMap{}, newMap)

	retryCarryOverKeys := []any{
		reflect.TypeOf(mapKeyDBClient{}),
		reflect.TypeOf(mapKeyLogger{}),
		reflect.TypeOf(mapKeyLoggerBase{}),
		reflect.TypeOf(mapKeyPathMatcher{}),
		reflect.TypeOf(mapKeyCacheManager{}),
		reflect.TypeOf(mapKeyAIProxyHandlers{}),
		reflect.TypeOf(mapKeyRequestID{}),
		reflect.TypeOf(mapKeyGeneratedCallID{}),
		reflect.TypeOf(mapKeyClient{}),
		reflect.TypeOf(mapKeyClientId{}),
		reflect.TypeOf(mapKeyClientToken{}),
		reflect.TypeOf(mapKeyIsAdmin{}),
		reflect.TypeOf(mapKeyAccessLang{}),
		// retry needs body copy and excluded models across attempts
		reflect.TypeOf(mapKeyReverseProxyRequestBodyBytes{}),
		reflect.TypeOf(mapKeyModelRetryExcludedModelIDs{}),
		reflect.TypeOf(mapKeyModelRetrySessionUnhealthyMarks{}),
		reflect.TypeOf(mapKeyModelRetryUnhealthyFallbackCount{}),
	}

	for _, key := range retryCarryOverKeys {
		if v, ok := oldMap.Load(key); ok {
			newMap.Store(key, v)
		}
	}
	return newCtx
}
