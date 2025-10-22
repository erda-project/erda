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

package cachehelpers

import (
	"context"
	"fmt"

	clientpb "github.com/erda-project/erda-proto-go/apps/aiproxy/client/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/cache/cachetypes"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
)

func GetClientByAK(ctx context.Context, ak string) (*clientpb.Client, error) {
	cache := ctxhelper.MustGetCacheManager(ctx).(cachetypes.Manager)
	_, clientsV, err := cache.ListAll(ctx, cachetypes.ItemTypeClient)
	if err != nil {
		return nil, err
	}
	clients := clientsV.([]*clientpb.Client)
	for _, client := range clients {
		if client.AccessKeyId == ak {
			return client, nil
		}
	}
	return nil, fmt.Errorf("client with access-key-id %s not found", ak)
}
