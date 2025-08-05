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
)

func MustGetResponseChunkIndex(ctx context.Context) int {
	index, ok := GetResponseChunkIndex(ctx)
	if !ok {
		return -1
	}
	return index
}

func GetResponseChunkIndex(ctx context.Context) (int, bool) {
	value, ok := ctx.Value(CtxKeyMap{}).(*sync.Map).Load(vars.MapKeyResponseChunkIndex{})
	if !ok || value == nil {
		return 0, false
	}
	index, ok := value.(int)
	if !ok {
		return 0, false
	}
	return index, true
}

func PutResponseChunkIndex(ctx context.Context, index int) {
	m := ctx.Value(CtxKeyMap{}).(*sync.Map)
	m.Store(vars.MapKeyResponseChunkIndex{}, index)
}

func IncrementResponseChunkIndex(ctx context.Context) int {
	m := ctx.Value(CtxKeyMap{}).(*sync.Map)
	value, ok := m.Load(vars.MapKeyResponseChunkIndex{})
	if !ok || value == nil {
		m.Store(vars.MapKeyResponseChunkIndex{}, 1)
		return 1
	}
	index, ok := value.(int)
	if !ok {
		m.Store(vars.MapKeyResponseChunkIndex{}, 1)
		return 1
	}
	newIndex := index + 1
	m.Store(vars.MapKeyResponseChunkIndex{}, newIndex)
	return newIndex
}

func IsFirstResponseChunk(ctx context.Context) bool {
	index, ok := GetResponseChunkIndex(ctx)
	return !ok || index == 0
}
