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

type ctxKeyMap struct{}

func InitCtxMapIfNeed(ctx context.Context) context.Context {
	if _, ok := ctx.Value(ctxKeyMap{}).(*sync.Map); ok {
		return ctx
	}
	return context.WithValue(ctx, ctxKeyMap{}, &sync.Map{})
}

// getFromMapKey generic function to get value from sync.Map in context
func getFromMapKey(ctx context.Context, key any) (any, bool) {
	m, ok := ctx.Value(ctxKeyMap{}).(*sync.Map)
	if !ok || m == nil {
		return nil, false
	}
	value, ok := m.Load(reflect.TypeOf(key))
	if !ok || value == nil {
		return nil, false
	}
	return value, true
}

func getFromMapKeyAs[T any](ctx context.Context, key any) (T, bool) {
	v, _ := getFromMapKey(ctx, key)
	x, ok := v.(T)
	return x, ok
}

// putToMapKey generic storage function to store value into sync.Map in context
func putToMapKey(ctx context.Context, key, value any) {
	m, ok := ctx.Value(ctxKeyMap{}).(*sync.Map)
	if !ok || m == nil {
		panic("sync.Map not found in context")
	}
	m.Store(reflect.TypeOf(key), value)
}
