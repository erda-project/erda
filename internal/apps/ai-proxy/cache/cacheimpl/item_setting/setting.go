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

package item_setting

import (
	"context"
	"fmt"

	settingpb "github.com/erda-project/erda-proto-go/apps/aiproxy/setting/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/cache/cachetypes"
	"github.com/erda-project/erda/internal/apps/ai-proxy/cache/config"
	"github.com/erda-project/erda/internal/apps/ai-proxy/providers/dao"
)

type settingCacheItem struct {
	*cachetypes.BaseCacheItem
}

func NewSettingCacheItem(dao dao.DAO, config *config.Config) cachetypes.CacheItem {
	item := &settingCacheItem{}
	item.BaseCacheItem = cachetypes.NewBaseCacheItem(cachetypes.ItemTypeSetting, dao, config, item)
	return item
}

func (c *settingCacheItem) QueryFromDB(ctx context.Context) (uint64, any, error) {
	list, err := c.DBClient.SettingClient().ListAll(ctx)
	if err != nil {
		return 0, nil, fmt.Errorf("failed to query settings: %w", err)
	}
	pbList := make([]*settingpb.Setting, 0, len(list))
	for _, item := range list {
		pbList = append(pbList, item.ToProtobuf())
	}
	return uint64(len(pbList)), pbList, nil
}

func (c *settingCacheItem) GetIDValue(item any) (string, error) {
	setting, ok := item.(*settingpb.Setting)
	if !ok {
		return "", fmt.Errorf("invalid setting data type")
	}
	return setting.Id, nil
}
