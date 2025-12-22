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
	"testing"

	pb "github.com/erda-project/erda-proto-go/apps/aiproxy/policy_group/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/common_types"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/lb/state_store"
	policygroup "github.com/erda-project/erda/internal/apps/ai-proxy/route/policy_group"
)

func TestRoute_NilGroup(t *testing.T) {
	store := state_store.NewMemoryStateStore()
	engine := NewEngine(store)

	req := policygroup.RouteRequest{
		ClientID: "c1",
		Group:    nil,
	}

	_, _, err := engine.Route(context.Background(), req)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestRoute_NoAvailableBranch(t *testing.T) {
	store := state_store.NewMemoryStateStore()
	engine := NewEngine(store)

	group := &pb.PolicyGroup{
		Name: "g1",
		Mode: common_types.PolicyGroupModePriority.String(),
	}

	req := policygroup.RouteRequest{
		ClientID: "c1",
		Group:    group,
	}

	_, _, err := engine.Route(context.Background(), req)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, ErrNoAvailableBranch) {
		t.Fatalf("expected ErrNoAvailableBranch, got %v", err)
	}
}
