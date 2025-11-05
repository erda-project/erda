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

package handler_model

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"

	clientmodelrelationpb "github.com/erda-project/erda-proto-go/apps/aiproxy/client_model_relation/pb"
	modelpb "github.com/erda-project/erda-proto-go/apps/aiproxy/model/pb"
	providerpb "github.com/erda-project/erda-proto-go/apps/aiproxy/service_provider/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/cache/cachetypes"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
)

type mockCacheManager struct {
	models    []*modelpb.Model
	relations []*clientmodelrelationpb.ClientModelRelation
	providers map[string]*providerpb.ServiceProvider
}

func (m *mockCacheManager) ListAll(ctx context.Context, itemType cachetypes.ItemType) (uint64, any, error) {
	switch itemType {
	case cachetypes.ItemTypeModel:
		return uint64(len(m.models)), m.models, nil
	case cachetypes.ItemTypeClientModelRelation:
		return uint64(len(m.relations)), m.relations, nil
	case cachetypes.ItemTypeProvider:
		providers := make([]*providerpb.ServiceProvider, 0, len(m.providers))
		for _, p := range m.providers {
			providers = append(providers, p)
		}
		return uint64(len(providers)), providers, nil
	default:
		return 0, nil, fmt.Errorf("unsupported item type: %v", itemType)
	}
}

func (m *mockCacheManager) GetByID(ctx context.Context, itemType cachetypes.ItemType, id string) (any, error) {
	switch itemType {
	case cachetypes.ItemTypeModel:
		for _, model := range m.models {
			if model.Id == id {
				return model, nil
			}
		}
	case cachetypes.ItemTypeProvider:
		if provider, ok := m.providers[id]; ok {
			return provider, nil
		}
	default:
		return nil, fmt.Errorf("unsupported item type: %v", itemType)
	}
	return nil, fmt.Errorf("item %s not found for type %v", id, itemType)
}

func (m *mockCacheManager) TriggerRefresh(ctx context.Context, itemTypes ...cachetypes.ItemType) {}

func TestModelHandler_pagingViaCache(t *testing.T) {
	handler := &ModelHandler{}
	now := time.Date(2024, 8, 1, 12, 0, 0, 0, time.UTC)

	modelAssigned := &modelpb.Model{
		Id:         "assigned-1",
		Name:       "SharedBeta",
		ClientId:   "",
		TemplateId: "tpl-2",
		Type:       modelpb.ModelType_image,
		ProviderId: "provider-2",
		Publisher:  "pub-2",
		IsEnabled:  boolPtr(false),
		UpdatedAt:  timestamppb.New(now.Add(-3 * time.Hour)),
	}
	modelOwnedOld := &modelpb.Model{
		Id:         "owned-1",
		Name:       "AlphaOne",
		ClientId:   "client-1",
		TemplateId: "tpl-1",
		Type:       modelpb.ModelType_text_generation,
		ProviderId: "provider-1",
		Publisher:  "pub-1",
		IsEnabled:  boolPtr(true),
		UpdatedAt:  timestamppb.New(now.Add(-2 * time.Hour)),
	}
	modelOwnedNew := &modelpb.Model{
		Id:         "owned-2",
		Name:       "GammaFocus",
		ClientId:   "client-1",
		TemplateId: "tpl-1",
		Type:       modelpb.ModelType_text_generation,
		ProviderId: "provider-1",
		Publisher:  "pub-1",
		IsEnabled:  boolPtr(true),
		UpdatedAt:  timestamppb.New(now.Add(-1 * time.Hour)),
	}

	manager := &mockCacheManager{
		models: []*modelpb.Model{
			modelAssigned,
			modelOwnedOld,
			modelOwnedNew,
		},
		relations: []*clientmodelrelationpb.ClientModelRelation{
			{ClientId: "client-1", ModelId: modelAssigned.Id},
		},
		providers: map[string]*providerpb.ServiceProvider{
			"provider-1": {Id: "provider-1"},
			"provider-2": {Id: "provider-2"},
		},
	}

	ctxWithManager := func() context.Context {
		ctx := ctxhelper.InitCtxMapIfNeed(context.Background())
		ctxhelper.PutCacheManager(ctx, manager)
		return ctx
	}

	t.Run("client only default ordering and pagination", func(t *testing.T) {
		ctx := ctxWithManager()
		req := &modelpb.ModelPagingRequest{
			ClientId:   "client-1",
			ClientOnly: true,
			PageNum:    1,
			PageSize:   1,
		}
		resp, err := handler.pagingViaCache(ctx, req)
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, int64(2), resp.Total)
		require.Len(t, resp.List, 1)
		assert.Equal(t, modelOwnedNew.Id, resp.List[0].Id)
	})

	t.Run("filters by multiple request fields", func(t *testing.T) {
		ctx := ctxWithManager()
		req := &modelpb.ModelPagingRequest{
			ClientId:   "client-1",
			Name:       "focus",
			NameFull:   "gammafocus",
			Ids:        []string{modelOwnedNew.Id},
			TemplateId: "tpl-1",
			Type:       modelpb.ModelType_text_generation,
			ProviderId: "provider-1",
			Publisher:  "pub-1",
			IsEnabled:  boolPtr(true),
			OrderBys:   []string{"name ASC"},
			PageNum:    1,
			PageSize:   5,
		}
		resp, err := handler.pagingViaCache(ctx, req)
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, int64(1), resp.Total)
		require.Len(t, resp.List, 1)
		assert.Equal(t, modelOwnedNew.Id, resp.List[0].Id)
	})
}

func TestParseOrderBys(t *testing.T) {
	t.Run("defaults to updated_at when empty", func(t *testing.T) {
		orders := parseOrderBys(nil)
		require.Len(t, orders, 1)
		assert.Equal(t, "updated_at", orders[0].Field)
		assert.True(t, orders[0].Desc)
	})

	t.Run("parses field names and directions", func(t *testing.T) {
		orders := parseOrderBys([]string{"NAME desc", "   updated_at   asc  ", "", "publisher"})
		require.Len(t, orders, 3)
		assert.Equal(t, orderBy{Field: "name", Desc: true}, orders[0])
		assert.Equal(t, orderBy{Field: "updated_at", Desc: false}, orders[1])
		assert.Equal(t, orderBy{Field: "publisher", Desc: false}, orders[2])
	})
}

func TestSortModels(t *testing.T) {
	now := time.Date(2024, 8, 1, 12, 0, 0, 0, time.UTC)
	models := []*modelpb.Model{
		{Id: "3", Name: "beta", UpdatedAt: timestamppb.New(now.Add(-2 * time.Hour))},
		{Id: "1", Name: "Alpha", UpdatedAt: timestamppb.New(now)},
		{Id: "2", Name: "alpha", UpdatedAt: timestamppb.New(now.Add(-1 * time.Hour))},
	}

	sortModels(models, []string{"name ASC", "updated_at DESC"})

	got := []string{models[0].Id, models[1].Id, models[2].Id}
	assert.Equal(t, []string{"1", "2", "3"}, got)
}

func TestSortModelsDefaultOrder(t *testing.T) {
	now := time.Date(2024, 8, 1, 12, 0, 0, 0, time.UTC)
	models := []*modelpb.Model{
		{Id: "1", UpdatedAt: timestamppb.New(now.Add(-time.Hour))},
		{Id: "2", UpdatedAt: timestamppb.New(now)},
	}

	sortModels(models, nil)

	got := []string{models[0].Id, models[1].Id}
	assert.Equal(t, []string{"2", "1"}, got)
}

func TestGetUpdatedAt(t *testing.T) {
	assert.True(t, getUpdatedAt(&modelpb.Model{}).IsZero())

	ts := timestamppb.New(time.Unix(1719820800, 0))
	assert.Equal(t, ts.AsTime(), getUpdatedAt(&modelpb.Model{UpdatedAt: ts}))
}

func boolPtr(v bool) *bool {
	return &v
}
