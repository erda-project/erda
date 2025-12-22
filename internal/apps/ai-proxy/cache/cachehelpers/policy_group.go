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

	"github.com/erda-project/erda-proto-go/apps/aiproxy/policy_group/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/cache/cachetypes"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
)

func GetPolicyGroupByName(ctx context.Context, clientID, name string) (*pb.PolicyGroup, error) {
	group, err := TryGetPolicyGroupByName(ctx, clientID, name)
	if err != nil {
		return nil, err
	}
	if group == nil {
		return nil, fmt.Errorf("policy group not found: %s", name)
	}
	return group, nil
}

// TryGetPolicyGroupByName returns the policy group with the given name under the client.
func TryGetPolicyGroupByName(ctx context.Context, clientID, name string) (*pb.PolicyGroup, error) {
	cache := ctxhelper.MustGetCacheManager(ctx).(cachetypes.Manager)
	_, allGroupsV, err := cache.ListAll(ctx, cachetypes.ItemTypePolicyGroup)
	if err != nil {
		return nil, err
	}
	for _, group := range allGroupsV.([]*pb.PolicyGroup) {
		if group.ClientId == clientID && group.Name == name {
			return group, nil
		}
	}
	return nil, nil
}
