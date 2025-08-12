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
	"testing"
)

func TestInitCtxMapIfNeed(t *testing.T) {
	// Test with empty context
	ctx := context.Background()
	ctx = InitCtxMapIfNeed(ctx)

	// Verify sync.Map is created
	if m, ok := ctx.Value(ctxKeyMap{}).(*sync.Map); !ok || m == nil {
		t.Fatal("InitCtxMapIfNeed should create a sync.Map")
	}

	// Test with context that already has sync.Map
	ctx2 := InitCtxMapIfNeed(ctx)

	// Should return the same context (no new context created)
	if ctx != ctx2 {
		t.Error("InitCtxMapIfNeed should reuse existing context with sync.Map")
	}

	// Verify the sync.Map is still there and usable
	if m, ok := ctx2.Value(ctxKeyMap{}).(*sync.Map); !ok || m == nil {
		t.Fatal("sync.Map should still exist after second call")
	}
}

func TestPutAndGetFromMapKey(t *testing.T) {
	ctx := context.Background()
	ctx = InitCtxMapIfNeed(ctx)

	// Test string value
	key1 := struct{ string }{}
	value1 := "test-value"
	putToMapKey(ctx, key1, value1)

	retrieved, ok := getFromMapKeyAs[string](ctx, key1)
	if !ok {
		t.Fatal("Failed to retrieve stored string value")
	}
	if retrieved != value1 {
		t.Errorf("Expected %q, got %q", value1, retrieved)
	}

	// Test int value
	key2 := struct{ int }{}
	value2 := 42
	putToMapKey(ctx, key2, value2)

	retrieved2, ok := getFromMapKeyAs[int](ctx, key2)
	if !ok {
		t.Fatal("Failed to retrieve stored int value")
	}
	if retrieved2 != value2 {
		t.Errorf("Expected %d, got %d", value2, retrieved2)
	}
}

func TestPutToMapKeyPanic(t *testing.T) {
	// Test that putToMapKey panics when no sync.Map in context
	ctx := context.Background()

	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic when sync.Map not found in context")
		}
	}()

	key := struct{ string }{}
	putToMapKey(ctx, key, "value")
}
