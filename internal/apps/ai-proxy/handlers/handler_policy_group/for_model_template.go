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
	"fmt"
	"strings"

	"github.com/erda-project/erda-proto-go/apps/aiproxy/policy_group/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/cache/cachehelpers"
	"github.com/erda-project/erda/internal/apps/ai-proxy/cache/cachetypes"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/common_types"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
)

// UpsertForModelTemplate provides a simplified upsert for a model/template: instances + weights => policy group.
func (h *Handler) UpsertForModelTemplate(ctx context.Context, req *pb.PolicyGroupForModelTemplateUpsertRequest) (*pb.PolicyGroupForModelTemplateResponse, error) {
	// check client id
	if req.ClientId == "" {
		return nil, fmt.Errorf("missing client id")
	}
	// check template id
	_, err := cachehelpers.GetModelTemplate(ctx, req.TemplateId)
	if err != nil {
		return nil, fmt.Errorf("failed to get model template: %w", err)
	}
	// check instances
	instanceIDs := make([]string, 0, len(req.Instances))
	for _, instance := range req.Instances {
		instanceIDs = append(instanceIDs, instance.ModelInstanceId)
	}
	for _, instanceID := range instanceIDs {
		_, err := cachehelpers.GetOneClientModel(ctx, req.ClientId, instanceID, &cachehelpers.ClientModelConfig{OnlyEnabled: false})
		if err != nil {
			return nil, fmt.Errorf("failed to get model instance, id: %s, err: %w", instanceID, err)
		}
	}
	// convert instances to std branches
	branches := buildBranchesFromInstances(req.Instances)
	// convert sticky key
	if req.StickyHeaderKey != "" {
		req.StickyHeaderKey = common_types.StickyKeyPrefixFromReqHeader + strings.ToLower(req.StickyHeaderKey)
	}

	createReq := &pb.PolicyGroupCreateRequest{
		ClientId:  req.ClientId,
		Name:      req.TemplateId,
		Desc:      fmt.Sprintf("auto generated policy-group for model template: %s", req.TemplateId),
		Mode:      common_types.PolicyGroupModeWeighted.String(),
		StickyKey: req.StickyHeaderKey,
		Branches:  branches,
		Source:    common_types.PolicyGroupSourceForModelTemplate.String(),
	}

	// query exists
	existGroup, err := h.queryGroupByName(ctx, req.ClientId, req.TemplateId)
	if err != nil {
		return nil, err
	}
	// upsert
	var newGroup *pb.PolicyGroup
	if existGroup != nil {
		updated, err := h.DAO.PolicyGroupClient().Update(ctx, &pb.PolicyGroupUpdateRequest{
			Id:       existGroup.Id,
			Branches: createReq.Branches,
			ClientId: req.ClientId,
			Mode:     &createReq.Mode,   // overwrite mode
			Source:   &createReq.Source, // overwrite source
		})
		if err != nil {
			return nil, err
		}
		newGroup = updated
	} else { // do insert
		created, err := h.DAO.PolicyGroupClient().Create(ctx, createReq)
		if err != nil {
			return nil, err
		}
		newGroup = created
	}
	go ctxhelper.MustGetCacheManager(ctx).(cachetypes.Manager).TriggerRefresh(ctx, cachetypes.ItemTypePolicyGroup)
	return buildForModelTemplateResponse(newGroup), err
}

// GetForModelTemplate fetches a simplified policy group view by model template name.
func (h *Handler) GetForModelTemplate(ctx context.Context, req *pb.PolicyGroupForModelTemplateGetRequest) (*pb.PolicyGroupForModelTemplateResponse, error) {
	pg, err := h.queryGroupByName(ctx, req.ClientId, req.TemplateId)
	if err != nil {
		return nil, err
	}
	return buildForModelTemplateResponse(pg), nil
}

func (h *Handler) queryGroupByName(ctx context.Context, clientID string, name string) (*pb.PolicyGroup, error) {
	pagingResp, err := h.DAO.PolicyGroupClient().Paging(ctx, &pb.PolicyGroupPagingRequest{
		PageNum:  1,
		PageSize: 1,
		ClientId: clientID,
		NameFull: name,
	})
	if err != nil {
		return nil, err
	}
	if pagingResp.Total == 0 {
		return nil, fmt.Errorf("policy group not found: %s", name)
	}
	return pagingResp.List[0], nil
}

func buildBranchesFromInstances(instances []*pb.PolicyGroupInstanceWeight) []*pb.PolicyBranch {
	var branches []*pb.PolicyBranch
	for _, instance := range instances {
		selector := &pb.PolicySelector{
			Requirements: []*pb.PolicyRequirement{
				{
					Type: common_types.PolicyBranchSelectorRequirementTypeLabel.String(),
					Label: &pb.LabelRequirement{
						Key:      common_types.PolicyLabelKeyModelInstanceID,
						Operator: common_types.PolicySelectorLabelOpIn.String(),
						Values:   []string{instance.ModelInstanceId},
					},
				},
			},
		}
		branches = append(branches, &pb.PolicyBranch{
			Name:     instance.ModelInstanceId,
			Desc:     "database uuid of model instance",
			Selector: selector,
			Weight:   instance.Weight,
			Priority: 1,
			Strategy: common_types.PolicyGroupBranchStrategyRoundRobin.String(),
		})
	}
	return branches
}

func buildForModelTemplateResponse(group *pb.PolicyGroup) *pb.PolicyGroupForModelTemplateResponse {
	var instances []*pb.PolicyGroupForModelTemplateInstance
	for _, branch := range group.Branches {
		instances = append(instances, &pb.PolicyGroupForModelTemplateInstance{
			ModelInstanceId: branch.Name,
			Weight:          branch.Weight,
		})
	}
	return &pb.PolicyGroupForModelTemplateResponse{
		Id:              group.Id,
		CreatedAt:       group.CreatedAt,
		UpdatedAt:       group.UpdatedAt,
		TemplateId:      group.Name,
		PolicyGroupName: group.Name,
		Mode:            group.Mode,
		StickyHeaderKey: strings.TrimPrefix(group.StickyKey, common_types.StickyKeyPrefixFromReqHeader),
		Instances:       instances,
	}
}

func findInstanceIDFromSelector(sel *pb.PolicySelector) (string, bool) {
	for _, req := range sel.Requirements {
		if req.Type != common_types.PolicyBranchSelectorRequirementTypeLabel.String() || req.Label == nil {
			continue
		}
		if !strings.EqualFold(req.Label.Key, "model-instance-id") {
			continue
		}
		if req.Label.Operator == common_types.PolicySelectorLabelOpIn.String() && len(req.Label.Values) > 0 {
			return req.Label.Values[0], true
		}
	}
	return "", false
}
