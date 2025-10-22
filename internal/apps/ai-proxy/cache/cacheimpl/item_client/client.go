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

package item_client

import (
	"context"
	"fmt"

	"github.com/erda-project/erda-proto-go/apps/aiproxy/client/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/cache/cachetypes"
	"github.com/erda-project/erda/internal/apps/ai-proxy/cache/config"
	"github.com/erda-project/erda/internal/apps/ai-proxy/providers/dao"
)

// clientCacheItem implements CacheItem for clients
type clientCacheItem struct {
	*cachetypes.BaseCacheItem
}

func NewClientCacheItem(dao dao.DAO, config *config.Config) cachetypes.CacheItem {
	item := &clientCacheItem{}
	item.BaseCacheItem = cachetypes.NewBaseCacheItem(cachetypes.ItemTypeClient, dao, config, item)
	return item
}

func (c *clientCacheItem) QueryFromDB(ctx context.Context) (uint64, any, error) {
	resp, err := c.DBClient.ClientClient().Paging(ctx, &pb.ClientPagingRequest{
		PageNum:  1,
		PageSize: 999,
	})
	if err != nil {
		return 0, nil, fmt.Errorf("failed to query clients: %w", err)
	}
	return uint64(resp.Total), resp.List, nil
}

func (c *clientCacheItem) GetByIDFromData(ctx context.Context, data any, id string) (any, error) {
	clients, ok := data.([]*pb.Client)
	if !ok {
		return nil, fmt.Errorf("invalid client data type")
	}

	for _, client := range clients {
		if client.Id == id {
			return client, nil
		}
	}

	return nil, fmt.Errorf("client with ID %s not found", id)
}
