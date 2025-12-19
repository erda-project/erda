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
