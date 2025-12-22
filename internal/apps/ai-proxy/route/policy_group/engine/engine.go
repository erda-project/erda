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
	"errors"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/erda-project/erda-proto-go/apps/aiproxy/policy_group/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/lb/algo"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/lb/state_store"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/policy_group"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/policy_group/selector"
)

const (
	defaultStickyTTL = time.Minute * 10
	defaultPriority  = uint64(100)
)

var (
	ErrNoAvailableBranch   = errors.New("no available branch")
	ErrNoAvailableInstance = errors.New("no available instance")
)

type BranchCandidate struct {
	branch    *pb.PolicyBranch
	instances []*policy_group.RoutingModelInstance
}

func wrapNoAvailableBranch(groupName string) error {
	if groupName == "" {
		return ErrNoAvailableBranch
	}
	return fmt.Errorf("policy group %q: %w", groupName, ErrNoAvailableBranch)
}

type HealthFilter func(instances []*policy_group.RoutingModelInstance) []*policy_group.RoutingModelInstance

type Engine struct {
	store        state_store.LBStateStore
	stickyTTL    time.Duration
	healthFilter HealthFilter
}

var (
	swrrMu   sync.Mutex
	swrrByID map[string]*algo.SmoothWeightedRR
)

func init() {
	swrrByID = make(map[string]*algo.SmoothWeightedRR)
}

type Option func(*Engine)

func WithStickyTTL(ttl time.Duration) Option {
	return func(e *Engine) {
		if ttl > 0 {
			e.stickyTTL = ttl
		}
	}
}

func WithHealthFilter(filter HealthFilter) Option {
	return func(e *Engine) {
		if filter != nil {
			e.healthFilter = filter
		}
	}
}

func NewEngine(store state_store.LBStateStore, opts ...Option) *Engine {
	if store == nil {
		panic("nil state store")
	}
	e := &Engine{
		store:     store,
		stickyTTL: defaultStickyTTL,
		healthFilter: func(instances []*policy_group.RoutingModelInstance) []*policy_group.RoutingModelInstance {
			return instances
		},
	}
	for _, opt := range opts {
		opt(e)
	}
	return e
}

func (e *Engine) Route(ctx context.Context, req policy_group.RouteRequest) (*policy_group.RoutingModelInstance, *policy_group.RouteTrace, error) {
	if req.Group == nil {
		return nil, nil, errors.New("nil policy group")
	}

	// ensure deterministic ordering of instances across requests
	sort.SliceStable(req.Instances, func(i, j int) bool {
		return req.Instances[i].ModelWithProvider.Id < req.Instances[j].ModelWithProvider.Id
	})

	availableBranches := e.buildBranchCandidates(req.Group, req.Instances)
	if len(availableBranches) == 0 {
		return nil, nil, wrapNoAvailableBranch(req.Group.Name)
	}

	stickyValue := req.GetStickyValue()

	var fallbackFromSticky bool
	if stickyValue != "" {
		br, inst, ok, err := e.directPickBySticky(ctx, req, availableBranches, stickyValue)
		if err != nil {
			return nil, nil, err
		}
		if ok {
			return inst, buildRouteTrace(req, stickyValue, false, br), nil
		}
		// want sticky but failed, fallback to normal pick
		fallbackFromSticky = true
	}

	branchCandidate, err := e.pickBranch(ctx, req, availableBranches)
	if err != nil {
		return nil, nil, err
	}
	if branchCandidate == nil {
		return nil, nil, wrapNoAvailableBranch(req.Group.Name)
	}
	instance, err := e.pickInstance(ctx, req, branchCandidate)
	if err != nil {
		return nil, nil, err
	}

	// save sticky info into state-store
	if stickyValue != "" && instance != nil {
		_ = e.store.SetBinding(ctx, makeGroupBindingKey(req.ClientID, req.Group.Name), stickyValue, encodeBinding(branchCandidate.branch.Name, instance.ModelWithProvider.Id), e.stickyTTL)
	}

	return instance, buildRouteTrace(req, stickyValue, fallbackFromSticky, branchCandidate), nil
}

func (e *Engine) buildBranchCandidates(group *pb.PolicyGroup, instances []*policy_group.RoutingModelInstance) []BranchCandidate {
	var ret []BranchCandidate
	for i := range group.Branches {
		br := group.Branches[i]
		matched := selector.MatchSelector(instances, br.Selector)
		matched = e.healthFilter(matched)
		if len(matched) == 0 {
			continue
		}
		ret = append(ret, BranchCandidate{
			branch:    br,
			instances: matched,
		})
	}
	return ret
}
