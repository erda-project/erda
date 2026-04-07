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

package handler_setting

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/erda-project/erda/internal/apps/ai-proxy/cache/cachetypes"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
)

func TestHandler_triggerSettingCacheRefresh_UsesInjectedCacheOnly(t *testing.T) {
	ctx := ctxhelper.InitCtxMapIfNeed(context.Background())
	ctxCache := &mockCacheManager{refreshCh: make(chan cachetypes.ItemType, 1)}
	ctxhelper.PutCacheManager(ctx, ctxCache)

	handler := &Handler{}
	handler.triggerSettingCacheRefresh(ctx)

	select {
	case itemType := <-ctxCache.refreshCh:
		t.Fatalf("unexpected context cache refresh for item type %s", itemType)
	case <-time.After(50 * time.Millisecond):
	}
}

func TestHandler_triggerSettingCacheRefresh_RefreshesInjectedCache(t *testing.T) {
	cache := &mockCacheManager{refreshCh: make(chan cachetypes.ItemType, 1)}
	handler := &Handler{Cache: cache}

	handler.triggerSettingCacheRefresh(context.Background())

	select {
	case itemType := <-cache.refreshCh:
		require.Equal(t, cachetypes.ItemTypeSetting, itemType)
	case <-time.After(time.Second):
		t.Fatal("expected injected cache refresh")
	}
}

type mockCacheManager struct {
	refreshCh chan cachetypes.ItemType
}

func (m *mockCacheManager) ListAll(ctx context.Context, itemType cachetypes.ItemType) (uint64, any, error) {
	return 0, nil, nil
}

func (m *mockCacheManager) GetByID(ctx context.Context, itemType cachetypes.ItemType, id string) (any, error) {
	return nil, nil
}

func (m *mockCacheManager) TriggerRefresh(ctx context.Context, itemTypes ...cachetypes.ItemType) {
	for _, itemType := range itemTypes {
		m.refreshCh <- itemType
	}
}
