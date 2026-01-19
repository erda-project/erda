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

package handler_cache

import (
	"context"
	"fmt"
	"strings"

	"github.com/erda-project/erda-proto-go/apps/aiproxy/cache/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/cache/cachetypes"
)

type CacheHandler struct {
	Cache cachetypes.Manager
}

func (h *CacheHandler) RefreshCache(ctx context.Context, req *pb.CacheRefreshRequest) (*pb.CacheRefreshResponse, error) {
	var typesToRefresh []cachetypes.ItemType
	if len(req.Types) == 0 {
		typesToRefresh = cachetypes.AllItemTypes()
	} else {
		var invalidTypeNames []string
		for _, typeName := range req.Types {
			if !cachetypes.IsValidItemType(typeName) {
				invalidTypeNames = append(invalidTypeNames, typeName)
				continue
			}
			typesToRefresh = append(typesToRefresh, cachetypes.ItemTypeStrToType[typeName])
		}
		if len(invalidTypeNames) > 0 {
			return nil, fmt.Errorf("found invalid cache types: %s", strings.Join(invalidTypeNames, ", "))
		}
	}

	// refresh specified types
	h.Cache.TriggerRefresh(ctx, typesToRefresh...)

	var refreshedTypeNames []string
	for _, t := range typesToRefresh {
		refreshedTypeNames = append(refreshedTypeNames, t.String())
	}

	return &pb.CacheRefreshResponse{
		Message:        "cache refresh triggered",
		RefreshedTypes: refreshedTypeNames,
	}, nil
}
