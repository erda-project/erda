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

package item_clientmodelrelation

import (
	"context"
	"fmt"

	"github.com/erda-project/erda-proto-go/apps/aiproxy/client_model_relation/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/cache/cachetypes"
	"github.com/erda-project/erda/internal/apps/ai-proxy/cache/config"
	"github.com/erda-project/erda/internal/apps/ai-proxy/providers/dao"
)

// clientModelRelationCacheItem implements CacheItem for client-model-relation.
type clientModelRelationCacheItem struct {
	*cachetypes.BaseCacheItem
}

func NewClientModelRelationCacheItem(dao dao.DAO, config *config.Config) cachetypes.CacheItem {
	item := &clientModelRelationCacheItem{}
	item.BaseCacheItem = cachetypes.NewBaseCacheItem(cachetypes.ItemTypeClientModelRelation, dao, config, item)
	return item
}

func (c *clientModelRelationCacheItem) QueryFromDB(ctx context.Context) (uint64, any, error) {
	resp, err := c.DBClient.ClientModelRelationClient().Paging(ctx, &pb.PagingRequest{
		PageNum:  1,
		PageSize: 999,
	})
	if err != nil {
		return 0, nil, fmt.Errorf("failed to query client-model-relations: %w", err)
	}
	return uint64(resp.Total), resp.List, nil
}

func (c *clientModelRelationCacheItem) GetByIDFromData(ctx context.Context, data any, id string) (any, error) {
	relations, ok := data.([]*pb.ClientModelRelation)
	if !ok {
		return nil, fmt.Errorf("invalid client-relation-relation data type")
	}

	for _, relation := range relations {
		if relation.Id == id {
			return relation, nil
		}
	}

	return nil, fmt.Errorf("client-model-relation with ID %s not found", id)
}
