package handler_policy_group

import (
	"context"
	"testing"

	pb "github.com/erda-project/erda-proto-go/apps/aiproxy/policy_group/pb"
)

func TestListAllOfficialLabels(t *testing.T) {
	h := &Handler{}
	resp, err := h.ListAllOfficialLabels(context.Background(), &pb.ListOfficialPolicyGroupLabelsRequest{})
	if err != nil {
		t.Fatalf("ListAllOfficialLabels error: %v", err)
	}
	if len(resp.LabelKeys) == 0 {
		t.Fatal("expected non-empty label keys")
	}
	// ensure sorted
	for i := 1; i < len(resp.LabelKeys); i++ {
		if resp.LabelKeys[i] < resp.LabelKeys[i-1] {
			t.Fatalf("label keys not sorted: %v", resp.LabelKeys)
		}
	}
}
