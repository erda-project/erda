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

package engine

import (
	"context"
	"fmt"
	"strings"

	"github.com/erda-project/erda/internal/apps/ai-proxy/route/lb/state_store"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/policy_group"
)

func (e *Engine) directPickBySticky(ctx context.Context, req policy_group.RouteRequest, candidates []BranchCandidate, stickyValue string) (*BranchCandidate, *policy_group.RoutingModelInstance, bool, error) {
	key := makeGroupBindingKey(req.ClientID, req.Group.Name)
	val, ok, err := e.store.GetBinding(ctx, key, stickyValue)
	if err != nil || !ok {
		return nil, nil, false, err
	}
	boundBranchName, boundInstanceID, ok := decodeBindingValue(val)
	if !ok {
		return nil, nil, false, fmt.Errorf("failed to decode binding value: %s", val)
	}
	for i := range candidates {
		if candidates[i].branch.Name != boundBranchName {
			continue
		}
		found := findInstanceByID(candidates[i].instances, boundInstanceID)
		if found == nil {
			continue
		}
		return &candidates[i], found, true, nil
	}
	return nil, nil, false, nil
}

func makeGroupBindingKey(clientID, groupName string) state_store.BindingKey {
	return fmt.Sprintf("client:%s:group:%s", clientID, groupName)
}

func makeBranchCounterKey(clientID, groupName, branchName string) state_store.CounterKey {
	return fmt.Sprintf("client:%s:group:%s:branch:%s", clientID, groupName, branchName)
}

func encodeBinding(branchName, instanceID string) string {
	return branchName + "|" + instanceID
}

func decodeBindingValue(v string) (string, string, bool) {
	idx := strings.IndexByte(v, '|')
	if idx <= 0 || idx == len(v)-1 {
		return "", "", false
	}
	return v[:idx], v[idx+1:], true
}

func findInstanceByID(instances []*policy_group.RoutingModelInstance, inputInstanceID string) *policy_group.RoutingModelInstance {
	for _, instance := range instances {
		if instance.ModelWithProvider.Id == inputInstanceID {
			return instance
		}
	}
	return nil
}
