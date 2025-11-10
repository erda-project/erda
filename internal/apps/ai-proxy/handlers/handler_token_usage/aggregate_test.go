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

package handler_token_usage

import (
	"context"
	"fmt"
	"math"
	"testing"

	modelpb "github.com/erda-project/erda-proto-go/apps/aiproxy/model/pb"
	usagepb "github.com/erda-project/erda-proto-go/apps/aiproxy/usage/token/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/cache/cachetypes"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/metadata"
)

func TestAggregateTokenUsages_Basic(t *testing.T) {
	ctx := ctxhelper.InitCtxMapIfNeed(context.Background())

	models := map[string]*modelpb.Model{
		"m1": buildModelWithPricing(map[string]any{
			"prompt":     "0.001",
			"completion": "0.002",
			"unit":       "USD",
		}),
		"m2": buildModelWithPricing(map[string]any{
			"request": "0.5",
			"unit":    "USD",
		}),
	}

	usages := []*usagepb.TokenUsage{
		{Id: 1, ModelId: "m1", InputTokens: 10, OutputTokens: 5},
		{Id: 2, ModelId: "m2", InputTokens: 2, OutputTokens: 3},
	}

	handler := &TokenUsageHandler{Cache: &mockCache{models: models}}
	// inject cache manager into context to mimic normal initialized request context
	ctxhelper.PutCacheManager(ctx, handler.Cache)

	resp, err := handler.aggregateTokenUsages(ctx, usages, "")
	if err != nil {
		t.Fatalf("aggregateTokenUsages unexpected error: %v", err)
	}

	expectApprox(t, 0.52, resp.GetTotalCost(), "total cost")
	if got := resp.GetTotalInputTokens(); got != 12 {
		t.Fatalf("expected total input tokens 12, got %d", got)
	}
	if got := resp.GetTotalOutputTokens(); got != 8 {
		t.Fatalf("expected total output tokens 8, got %d", got)
	}
	if got := resp.GetTotalTokens(); got != 20 {
		t.Fatalf("expected total tokens 20, got %d", got)
	}
	if resp.GetCurrency() != "USD" {
		t.Fatalf("expected currency USD, got %q", resp.GetCurrency())
	}
	if len(resp.GetDetails()) != 2 {
		t.Fatalf("expected 2 details, got %d", len(resp.GetDetails()))
	}
	expectApprox(t, 0.02, resp.GetDetails()[0].GetCost(), "detail[0] cost")
	expectApprox(t, 0.5, resp.GetDetails()[1].GetCost(), "detail[1] cost")
	if resp.GetDetails()[0].GetTotalTokens() != 15 {
		t.Fatalf("expected detail[0] total tokens 15, got %d", resp.GetDetails()[0].GetTotalTokens())
	}
}

func TestAggregateTokenUsages_NoPricing(t *testing.T) {
	ctx := ctxhelper.InitCtxMapIfNeed(context.Background())

	models := map[string]*modelpb.Model{
		"m1": buildModelWithPricing(map[string]any{
			"prompt": "0.001",
			"unit":   "USD",
		}),
		"m2": {
			Id:       "m2",
			Metadata: (&metadata.Metadata{}).ToProtobuf(),
		},
	}

	usages := []*usagepb.TokenUsage{
		{Id: 1, ModelId: "m1", InputTokens: 10, OutputTokens: 0},
		{Id: 2, ModelId: "m2", InputTokens: 4, OutputTokens: 6},
	}

	handler := &TokenUsageHandler{Cache: &mockCache{models: models}}
	ctxhelper.PutCacheManager(ctx, handler.Cache)

	resp, err := handler.aggregateTokenUsages(ctx, usages, "")
	if err != nil {
		t.Fatalf("aggregateTokenUsages unexpected error: %v", err)
	}
	expectApprox(t, 0.01, resp.GetTotalCost(), "total cost with missing pricing")
	if detail := resp.GetDetails()[1]; detail.GetCost() != 0 {
		t.Fatalf("expected second detail cost 0, got %f", detail.GetCost())
	}
}

func TestAggregateTokenUsages_MixedCurrency(t *testing.T) {
	ctx := ctxhelper.InitCtxMapIfNeed(context.Background())

	models := map[string]*modelpb.Model{
		"m1": buildModelWithPricing(map[string]any{
			"prompt": "0.001",
			"unit":   "USD",
		}),
		"m2": buildModelWithPricing(map[string]any{
			"prompt": "0.002",
			"unit":   "CNY",
		}),
	}
	usages := []*usagepb.TokenUsage{
		{Id: 1, ModelId: "m1", InputTokens: 10},
		{Id: 2, ModelId: "m2", InputTokens: 5},
	}

	handler := &TokenUsageHandler{Cache: &mockCache{models: models}}
	ctxhelper.PutCacheManager(ctx, handler.Cache)

	_, err := handler.aggregateTokenUsages(ctx, usages, "")
	if err == nil {
		t.Fatalf("expected error for mixed currencies, got nil")
	}
}

func buildModelWithPricing(pricing map[string]any) *modelpb.Model {
	meta := metadata.Metadata{
		Public: map[string]any{
			"pricing": pricing,
		},
	}
	return &modelpb.Model{
		Metadata: meta.ToProtobuf(),
	}
}

func expectApprox(t *testing.T, expect, actual float64, label string) {
	t.Helper()
	if math.Abs(expect-actual) > 1e-9 {
		t.Fatalf("expected %s %.6f, got %.6f", label, expect, actual)
	}
}

type mockCache struct {
	models map[string]*modelpb.Model
}

func (m *mockCache) ListAll(ctx context.Context, itemType cachetypes.ItemType) (uint64, any, error) {
	return 0, nil, fmt.Errorf("not implemented")
}

func (m *mockCache) GetByID(ctx context.Context, itemType cachetypes.ItemType, id string) (any, error) {
	if itemType != cachetypes.ItemTypeModel {
		return nil, fmt.Errorf("unsupported item type: %v", itemType)
	}
	model, ok := m.models[id]
	if !ok {
		return nil, fmt.Errorf("model %s not found", id)
	}
	return model, nil
}

func (m *mockCache) TriggerRefresh(ctx context.Context, itemTypes ...cachetypes.ItemType) {
	// no-op for tests
}
