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

package item_clienttoken

import (
	"context"
	"fmt"

	"github.com/erda-project/erda-proto-go/apps/aiproxy/client_token/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/cache/cachetypes"
	"github.com/erda-project/erda/internal/apps/ai-proxy/cache/config"
	"github.com/erda-project/erda/internal/apps/ai-proxy/providers/dao"
)

// clientTokenCacheItem implements CacheItem for clientTokens
type clientTokenCacheItem struct {
	*cachetypes.BaseCacheItem
}

func NewClientTokenCacheItem(dao dao.DAO, config *config.Config) cachetypes.CacheItem {
	item := &clientTokenCacheItem{}
	item.BaseCacheItem = cachetypes.NewBaseCacheItem(cachetypes.ItemTypeClientToken, dao, config, item)
	return item
}

func (c *clientTokenCacheItem) QueryFromDB(ctx context.Context) (uint64, any, error) {
	resp, err := c.DBClient.ClientTokenClient().Paging(ctx, &pb.ClientTokenPagingRequest{
		PageNum:  1,
		PageSize: 999,
	})
	if err != nil {
		return 0, nil, fmt.Errorf("failed to query client-tokens: %w", err)
	}
	return uint64(resp.Total), resp.List, nil
}

func (c *clientTokenCacheItem) GetIDValue(item any) (string, error) {
	clientToken, ok := item.(*pb.ClientToken)
	if !ok {
		return "", fmt.Errorf("invalid clientToken data type")
	}
	return clientToken.Id, nil
}
