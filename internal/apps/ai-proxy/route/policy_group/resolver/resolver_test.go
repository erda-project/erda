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

package resolver

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	clientmodelrelationpb "github.com/erda-project/erda-proto-go/apps/aiproxy/client_model_relation/pb"
	modelpb "github.com/erda-project/erda-proto-go/apps/aiproxy/model/pb"
	providerpb "github.com/erda-project/erda-proto-go/apps/aiproxy/service_provider/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/cache/cachetypes"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
)

// fakeCacheManager is a minimal cache manager for ListAllClientModels helper.
type fakeCacheManager struct {
	models    []*modelpb.Model
	providers map[string]*providerpb.ServiceProvider
}

func boolPtr(v bool) *bool {
	return &v
}

func (m *fakeCacheManager) ListAll(ctx context.Context, itemType cachetypes.ItemType) (uint64, any, error) {
	if itemType == cachetypes.ItemTypeModel {
		return uint64(len(m.models)), m.models, nil
	}
	if itemType == cachetypes.ItemTypeClientModelRelation {
		return 0, []*clientmodelrelationpb.ClientModelRelation{}, nil
	}
	return 0, nil, fmt.Errorf("unsupported list type %v", itemType)
}

func (m *fakeCacheManager) GetByID(ctx context.Context, itemType cachetypes.ItemType, id string) (any, error) {
	if itemType == cachetypes.ItemTypeModel {
		for _, mm := range m.models {
			if mm.Id == id {
				return mm, nil
			}
		}
		return nil, fmt.Errorf("model not found")
	}
	if itemType == cachetypes.ItemTypeProvider {
		if p, ok := m.providers[id]; ok {
			return p, nil
		}
		return nil, fmt.Errorf("provider not found")
	}
	return nil, fmt.Errorf("unsupported get type %v", itemType)
}

func (m *fakeCacheManager) TriggerRefresh(ctx context.Context, itemTypes ...cachetypes.ItemType) {}

// fakeModelDB implements only methods used in resolver.instance builtin path.
type fakeModelDB struct {
	models []*modelpb.Model
}

func (m *fakeModelDB) Get(ctx context.Context, req *modelpb.ModelGetRequest) (*modelpb.Model, error) {
	for _, mm := range m.models {
		if mm.Id == req.Id && (req.ClientId == "" || mm.ClientId == req.ClientId) {
			return mm, nil
		}
	}
	return nil, fmt.Errorf("not found")
}

func (m *fakeModelDB) Paging(ctx context.Context, req *modelpb.ModelPagingRequest) (*modelpb.ModelPagingResponse, error) {
	var list []*modelpb.Model
	for _, mm := range m.models {
		if req.ClientId != "" && !strings.EqualFold(mm.ClientId, req.ClientId) {
			continue
		}
		if req.NameFull != "" && !strings.EqualFold(req.NameFull, mm.Name) {
			continue
		}
		list = append(list, mm)
	}
	return &modelpb.ModelPagingResponse{List: list}, nil
}

// fakePolicyGroupDB unused but required by constructor.
type fakePolicyGroupDB struct{}

func TestResolverAvailableNamePriority(t *testing.T) {
	clientID := "c1"
	// prepare models and providers
	providerAzure := &providerpb.ServiceProvider{Id: "p-az", Type: "azure"}
	providerVolc := &providerpb.ServiceProvider{Id: "p-volc", Type: "volcengine"}
	modelAzure := &modelpb.Model{Id: "m-az", Name: "gpt-4o", TemplateId: "gpt-4o", Publisher: "openai", ClientId: clientID, ProviderId: providerAzure.Id, IsEnabled: boolPtr(true), UpdatedAt: timestamppb.New(time.Unix(10, 0))}
	// same name, later updated time
	modelVolc := &modelpb.Model{Id: "m-volc", Name: "gpt-4o", TemplateId: "gpt-4o", Publisher: "openai", ClientId: clientID, ProviderId: providerVolc.Id, IsEnabled: boolPtr(true), UpdatedAt: timestamppb.New(time.Unix(20, 0))}

	ctx := ctxhelper.InitCtxMapIfNeed(context.Background())
	ctxhelper.PutCacheManager(ctx, &fakeCacheManager{
		models: []*modelpb.Model{modelAzure, modelVolc},
		providers: map[string]*providerpb.ServiceProvider{
			providerAzure.Id: providerAzure,
			providerVolc.Id:  providerVolc,
		},
	})

	resolver := NewResolver()

	// 1) publisher/model should resolve
	if pg, err := resolver.resolveRuntimeInternal(ctx, clientID, "openai/gpt-4o"); err != nil || pg == nil || pg.Name != "openai/gpt-4o" {
		t.Fatalf("expected publisher/model hit, got %#v", pg)
	}

	// 2) model name alone should map to standard
	if pg, err := resolver.resolveRuntimeInternal(ctx, clientID, "gpt-4o"); err != nil || pg == nil || pg.Name != "gpt-4o" {
		t.Fatalf("expected model-template-id match gpt-4o, got %#v", pg)
	}

	// 3) provider.type/model should map
	if pg, err := resolver.resolveRuntimeInternal(ctx, clientID, "azure/gpt-4o"); err != nil || pg == nil || pg.Name != "azure/gpt-4o" {
		t.Fatalf("expected provider.type/model match azure/gpt-4o, got %#v", pg)
	}

	// 4) no cache entries -> short name map empty -> no match
	ctxEmpty := ctxhelper.InitCtxMapIfNeed(context.Background())
	ctxhelper.PutCacheManager(ctxEmpty, &fakeCacheManager{models: []*modelpb.Model{}, providers: map[string]*providerpb.ServiceProvider{}})
	resolverEmpty := NewResolver()
	if pg, err := resolverEmpty.resolveRuntimeInternal(ctxEmpty, clientID, "o1-preview"); err != nil || pg != nil {
		t.Fatalf("expected no match for o1-preview when cache empty, got %#v", pg)
	}
}
