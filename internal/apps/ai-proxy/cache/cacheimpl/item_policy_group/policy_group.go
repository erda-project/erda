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

package item_policy_group

import (
	"context"
	"fmt"

	pb "github.com/erda-project/erda-proto-go/apps/aiproxy/policy_group/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/cache/cachetypes"
	"github.com/erda-project/erda/internal/apps/ai-proxy/cache/config"
	"github.com/erda-project/erda/internal/apps/ai-proxy/providers/dao"
)

// policyGroupCacheItem implements CacheItem for policy groups.
type policyGroupCacheItem struct {
	*cachetypes.BaseCacheItem
}

func NewPolicyGroupCacheItem(dao dao.DAO, cfg *config.Config) cachetypes.CacheItem {
	item := &policyGroupCacheItem{}
	item.BaseCacheItem = cachetypes.NewBaseCacheItem(cachetypes.ItemTypePolicyGroup, dao, cfg, item)
	return item
}

func (c *policyGroupCacheItem) QueryFromDB(ctx context.Context) (uint64, any, error) {
	resp, err := c.DBClient.PolicyGroupClient().Paging(ctx, &pb.PolicyGroupPagingRequest{
		PageNum:  1,
		PageSize: 99999,
	})
	if err != nil {
		return 0, nil, fmt.Errorf("failed to query policy groups: %w", err)
	}
	return uint64(resp.Total), resp.List, nil
}

func (c *policyGroupCacheItem) GetIDValue(item any) (string, error) {
	pg, ok := item.(*pb.PolicyGroup)
	if !ok {
		return "", fmt.Errorf("invalid policy group data type")
	}
	return pg.Id, nil
}
