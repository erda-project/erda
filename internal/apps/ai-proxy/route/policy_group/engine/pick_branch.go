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
	"sort"

	"github.com/erda-project/erda/internal/apps/ai-proxy/common/common_types"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/lb/algo"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/lb/state_store"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/policy_group"
	"github.com/erda-project/erda/pkg/strutil"
)

func (e *Engine) pickBranch(ctx context.Context, req policy_group.RouteRequest, candidates []BranchCandidate) (*BranchCandidate, error) {
	var picked *BranchCandidate
	var err error

	switch req.Group.Mode {
	case common_types.PolicyGroupModePriority.String():
		picked, err = pickBranchByPriority(ctx, e.store, req, candidates)
	case common_types.PolicyGroupModeWeighted.String():
		picked, err = pickBranchWeighted(ctx, e.store, req, candidates)
	default:
		err = fmt.Errorf("unsupported policy group mode: %q", req.Group.Mode)
	}
	return picked, err
}

func pickBranchByPriority(ctx context.Context, store state_store.LBStateStore, req policy_group.RouteRequest, candidates []BranchCandidate) (*BranchCandidate, error) {
	// only one branch
	if len(candidates) == 1 {
		return &candidates[0], nil
	}

	byPriority := make(map[uint64][]BranchCandidate)
	var priorities []uint64
	for _, c := range candidates {
		p := c.branch.Priority
		if p == 0 {
			p = defaultPriority
		}
		byPriority[p] = append(byPriority[p], c)
		priorities = append(priorities, p)
	}
	priorities = strutil.DedupUint64Slice(priorities)
	sort.Slice(priorities, func(i, j int) bool { return priorities[i] < priorities[j] })
	for _, p := range priorities {
		branches := byPriority[p]
		if len(branches) == 0 {
			continue
		}
		// use round-robin between branches with the same priority
		idx, err := algo.NextRoundRobinIndex(ctx, store, makeBranchCounterKey(req.ClientID, req.Group.Name, fmt.Sprintf("priority-%d", p)), len(branches))
		if err != nil {
			return nil, err
		}
		return &branches[idx], nil
	}
	return nil, wrapNoAvailableBranch(req.Group.Name)
}

func pickBranchWeighted(ctx context.Context, store state_store.LBStateStore, req policy_group.RouteRequest, candidates []BranchCandidate) (*BranchCandidate, error) {
	// only one branch
	if len(candidates) == 1 {
		return &candidates[0], nil
	}

	// Smooth-Weighted-RR for all weighted modes
	swrrKey := fmt.Sprintf("lb:swr/%s/%s", req.ClientID, req.Group.Name)
	items := make([]algo.WeightedItem, 0, len(candidates))
	for _, c := range candidates {
		w := int(c.branch.Weight)
		if w <= 0 {
			w = 1
		}
		items = append(items, algo.WeightedItem{ID: c.branch.Name, Weight: w})
	}
	swrr := getOrInitSWRR(swrrKey, items)
	item, ok := swrr.Next()
	if !ok {
		return nil, wrapNoAvailableBranch(req.Group.Name)
	}
	for i := range candidates {
		if candidates[i].branch.Name == item.ID {
			return &candidates[i], nil
		}
	}
	return nil, wrapNoAvailableBranch(req.Group.Name)
}

func getOrInitSWRR(key string, items []algo.WeightedItem) *algo.SmoothWeightedRR {
	swrrMu.Lock()
	defer swrrMu.Unlock()
	if rr, ok := swrrByID[key]; ok {
		if !sameWeightedItems(rr.Items(), items) {
			rr.UpdateItems(items)
		}
		return rr
	}
	rr := algo.NewSmoothWeightedRR(items)
	swrrByID[key] = rr
	return rr
}

func sameWeightedItems(a, b []algo.WeightedItem) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].ID != b[i].ID || a[i].Weight != b[i].Weight {
			return false
		}
	}
	return true
}
