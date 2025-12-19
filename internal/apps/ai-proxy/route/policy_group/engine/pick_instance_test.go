package engine

import (
	"context"
	"testing"

	pb "github.com/erda-project/erda-proto-go/apps/aiproxy/policy_group/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/common_types"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/lb/algo"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/lb/state_store"
	policygroup "github.com/erda-project/erda/internal/apps/ai-proxy/route/policy_group"
)

func TestPickInstance_ConsistentHash(t *testing.T) {
	store := state_store.NewMemoryStateStore()
	engine := NewEngine(store)

	group := &pb.PolicyGroup{
		Name:      "g1",
		StickyKey: "x-request-id",
	}
	req := policygroup.RouteRequest{
		ClientID: "c1",
		Group:    group,
		Meta:     policygroup.RequestMeta{Keys: map[string]string{"x-request-id": "sticky"}},
	}
	candidate := &BranchCandidate{
		branch: &pb.PolicyBranch{
			Name:     "b1",
			Strategy: common_types.PolicyGroupBranchStrategyConsistentHash.String(),
		},
		instances: []*policygroup.RoutingModelInstance{
			testInstance("i1", nil),
			testInstance("i2", nil),
			testInstance("i3", nil),
		},
	}

	got, err := engine.pickInstance(context.Background(), req, candidate)
	if err != nil {
		t.Fatalf("pickInstance failed: %v", err)
	}

	routingKey := req.GetRoutingKeyForBranch(candidate.branch.Name)
	wantIdx := algo.ConsistentHashIndex(routingKey, len(candidate.instances))
	if wantIdx < 0 {
		t.Fatalf("expected wantIdx >= 0, got %d", wantIdx)
	}
	wantID := candidate.instances[wantIdx].ModelWithProvider.Id
	if got.ModelWithProvider.Id != wantID {
		t.Fatalf("expected %s, got %s", wantID, got.ModelWithProvider.Id)
	}
}

func TestPickInstance_ConsistentHashFallbackToRoundRobin(t *testing.T) {
	store := state_store.NewMemoryStateStore()
	engine := NewEngine(store)

	req := policygroup.RouteRequest{
		ClientID: "c1",
		Group: &pb.PolicyGroup{
			Name:      "g2",
			StickyKey: "x-request-id",
		},
		Meta: policygroup.RequestMeta{Keys: map[string]string{}},
	}
	candidate := &BranchCandidate{
		branch: &pb.PolicyBranch{
			Name:     "b1",
			Strategy: common_types.PolicyGroupBranchStrategyConsistentHash.String(),
		},
		instances: []*policygroup.RoutingModelInstance{
			testInstance("i1", nil),
			testInstance("i2", nil),
		},
	}

	first, err := engine.pickInstance(context.Background(), req, candidate)
	if err != nil {
		t.Fatalf("pickInstance first failed: %v", err)
	}
	second, err := engine.pickInstance(context.Background(), req, candidate)
	if err != nil {
		t.Fatalf("pickInstance second failed: %v", err)
	}
	if first.ModelWithProvider.Id == second.ModelWithProvider.Id {
		t.Fatalf("expected RR to pick different instances, got %s then %s", first.ModelWithProvider.Id, second.ModelWithProvider.Id)
	}
}
