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

	"github.com/erda-project/erda/internal/apps/ai-proxy/common/common_types"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/lb/algo"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/lb/state_store"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/policy_group"
)

func (e *Engine) pickInstance(ctx context.Context, req policy_group.RouteRequest, candidate *BranchCandidate) (*policy_group.RoutingModelInstance, error) {
	if len(candidate.instances) == 0 {
		return nil, ErrNoAvailableInstance
	}

	routingKey := req.GetRoutingKeyForBranch(candidate.branch.Name)
	switch candidate.branch.Strategy {
	case common_types.PolicyGroupBranchStrategyConsistentHash.String():
		if routingKey != "" {
			if inst := pickInstanceByConsistentHash(routingKey, candidate.instances); inst != nil {
				return inst, nil
			}
		}
		// fallback to rr
		fallthrough
	case common_types.PolicyGroupBranchStrategyRoundRobin.String():
		// Anti-jitter for sticky routing:
		// When the request carries a sticky value (routingKey != ""), multiple concurrent requests may arrive
		// before the branch/instance binding is established in the state store. In that "sticky miss" window,
		// pure RR/SWRR will jitter across instances.
		//
		// To reduce jitter while keeping RR semantics for non-sticky traffic, we deterministically select an
		// instance using a stable hash of the routingKey as the initial pick. Once the binding is written,
		// subsequent requests should hit the binding and bypass this path.
		if routingKey != "" {
			if inst := pickInstanceByConsistentHash(routingKey, candidate.instances); inst != nil {
				return inst, nil
			}
		}
		return pickInstanceByRoundRobin(ctx, e.store, req.ClientID, req.Group.Name, candidate.branch.Name, candidate.instances)
	default:
		return nil, fmt.Errorf("unsupported branch strategy: %q", candidate.branch.Strategy)
	}
}

func pickInstanceByConsistentHash(routingKey string, instances []*policy_group.RoutingModelInstance) *policy_group.RoutingModelInstance {
	if len(instances) == 0 {
		return nil
	}
	idx := algo.ConsistentHashIndex(routingKey, len(instances))
	if idx < 0 {
		return nil
	}
	return instances[idx]
}

func pickInstanceByRoundRobin(ctx context.Context, store state_store.LBStateStore, clientID, groupName, branchName string, instances []*policy_group.RoutingModelInstance) (*policy_group.RoutingModelInstance, error) {
	idx, err := algo.NextRoundRobinIndex(ctx, store, makeBranchCounterKey(clientID, groupName, branchName), len(instances))
	if err != nil {
		return nil, err
	}
	return instances[idx], nil
}
