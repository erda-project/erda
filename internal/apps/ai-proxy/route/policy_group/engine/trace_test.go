package engine

import (
	"testing"

	pb "github.com/erda-project/erda-proto-go/apps/aiproxy/policy_group/pb"
	policygroup "github.com/erda-project/erda/internal/apps/ai-proxy/route/policy_group"
)

func TestBuildRouteTrace_SetsFallbackFromSticky(t *testing.T) {
	req := policygroup.RouteRequest{
		ClientID: "c1",
		Group: &pb.PolicyGroup{
			Name:      "g1",
			StickyKey: "x-request-id",
		},
	}

	trace := buildRouteTrace(req, "v1", true, nil)
	if trace.Sticky == nil || trace.Sticky.Value != "v1" {
		t.Fatalf("expected sticky.value=v1, got %+v", trace.Sticky)
	}
	if trace.Sticky == nil || !trace.Sticky.FallbackFromSticky {
		t.Fatalf("expected fallbackFromSticky=true, got false")
	}
}
