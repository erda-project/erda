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
