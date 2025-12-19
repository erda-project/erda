package handler_policy_group

import (
	"context"

	pb "github.com/erda-project/erda-proto-go/apps/aiproxy/policy_group/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/common_types"
)

// ListAllOfficialLabels returns built-in (non-custom) label keys available for policy groups.
func (h *Handler) ListAllOfficialLabels(ctx context.Context, _ *pb.ListOfficialPolicyGroupLabelsRequest) (*pb.ListOfficialPolicyGroupLabelsResponse, error) {
	keys := common_types.ListOfficialPolicyGroupLabelKeys()
	return &pb.ListOfficialPolicyGroupLabelsResponse{LabelKeys: keys}, nil
}
