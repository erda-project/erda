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
	"sort"
	"strings"

	"github.com/erda-project/erda-proto-go/apps/aiproxy/policy_group/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/common_types"
	policygroup "github.com/erda-project/erda/internal/apps/ai-proxy/route/policy_group"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/policy_group/selector"
)

// PreviewLabelGroups groups/filter/splits current client model instances by label and returns selectors.
func (h *Handler) PreviewLabelGroups(ctx context.Context, req *pb.PreviewLabelGroupsRequest) (*pb.PreviewLabelGroupsResponse, error) {
	if req.ClientId == "" {
		return nil, fmt.Errorf("missing client id")
	}

	// check label value with the simple operator
	if req.SimpleOperator == common_types.LabelPreviewOpFilter.String() || req.SimpleOperator == common_types.LabelPreviewOpSplit.String() {
		if req.LabelValue == "" {
			return nil, fmt.Errorf("missing label value for simple operator: %s", req.SimpleOperator)
		}
	}

	routingInstances, err := policygroup.BuildRoutingInstancesForClient(ctx, req.ClientId)
	if err != nil {
		return nil, err
	}

	groups, err := previewLabelGroups(routingInstances, req.BaseSelector, req.LabelKey, req.SimpleOperator, req.LabelValue)
	if err != nil {
		return nil, err
	}

	return &pb.PreviewLabelGroupsResponse{Groups: groups}, nil
}

func previewLabelGroups(instances []*policygroup.RoutingModelInstance, baseSelector *pb.PolicySelector, labelKey, simpleOperator, labelValue string) ([]*pb.PreviewLabelGroup, error) {
	filtered := instances
	if baseSelector != nil {
		filtered = selector.MatchSelector(instances, baseSelector)
	}
	return previewBySimpleOp(filtered, labelKey, simpleOperator, labelValue)
}

func previewBySimpleOp(instances []*policygroup.RoutingModelInstance, labelKey, simpleOperator, labelValue string) ([]*pb.PreviewLabelGroup, error) {
	switch common_types.LabelPreviewSimpleOp(simpleOperator) {
	case common_types.LabelPreviewOpGroupBy:
		return previewGroupBy(labelKey, instances), nil
	case common_types.LabelPreviewOpFilter:
		return previewFilter(labelKey, labelValue, instances), nil
	case common_types.LabelPreviewOpSplit:
		return previewSplit(labelKey, labelValue, instances), nil
	default:
		return nil, fmt.Errorf("unsupported simple operator: %s", simpleOperator)
	}
}

func previewGroupBy(labelKey string, instances []*policygroup.RoutingModelInstance) []*pb.PreviewLabelGroup {
	type bucket struct {
		value     string
		instances []*policygroup.RoutingModelInstance
	}
	grouped := make(map[string]*bucket)
	var keys []string
	for _, instance := range instances {
		labelValue, _ := findLabelValue(instance.Labels, labelKey)
		mapKey := labelValue
		if _, exists := grouped[mapKey]; !exists {
			grouped[mapKey] = &bucket{value: mapKey}
			keys = append(keys, mapKey)
		}
		grouped[mapKey].instances = append(grouped[mapKey].instances, instance)
	}
	sort.Strings(keys)

	var resp []*pb.PreviewLabelGroup
	for _, key := range keys {
		b := grouped[key]
		displayName := fmt.Sprintf("%s = %s", labelKey, b.value)
		labelSelector := buildLabelSelector(labelKey, common_types.PolicySelectorLabelOpIn.String(), b.value)
		if key == "" {
			displayName = fmt.Sprintf("%s = <empty>", labelKey)
			labelSelector = buildLabelSelector(labelKey, common_types.PolicySelectorLabelOpDoesNotExist.String(), "")
		}
		resp = append(resp, buildPreviewGroup(displayName, labelSelector, b.instances))
	}
	return resp
}

func previewFilter(labelKey, labelValue string, instances []*policygroup.RoutingModelInstance) []*pb.PreviewLabelGroup {
	var matched []*policygroup.RoutingModelInstance
	for _, instance := range instances {
		if v, ok := findLabelValue(instance.Labels, labelKey); ok && v == labelValue {
			matched = append(matched, instance)
		}
	}
	if len(matched) == 0 {
		return nil
	}
	displayName := fmt.Sprintf("%s = %s", labelKey, labelValue)
	labelSelector := buildLabelSelector(labelKey, common_types.PolicySelectorLabelOpIn.String(), labelValue)
	return []*pb.PreviewLabelGroup{buildPreviewGroup(displayName, labelSelector, matched)}
}

func previewSplit(labelKey, labelValue string, instances []*policygroup.RoutingModelInstance) []*pb.PreviewLabelGroup {
	labelValue = strings.TrimSpace(labelValue)
	var hit, miss []*policygroup.RoutingModelInstance
	for _, instance := range instances {
		if v, ok := findLabelValue(instance.Labels, labelKey); ok && v == labelValue {
			hit = append(hit, instance)
		} else {
			miss = append(miss, instance)
		}
	}
	displayHit := fmt.Sprintf("%s = %s", labelKey, labelValue)
	displayMiss := fmt.Sprintf("%s != %s", labelKey, labelValue)
	selectorHit := buildLabelSelector(labelKey, common_types.PolicySelectorLabelOpIn.String(), labelValue)
	selectorMiss := buildLabelSelector(labelKey, common_types.PolicySelectorLabelOpNotIn.String(), labelValue)
	return []*pb.PreviewLabelGroup{
		buildPreviewGroup(displayHit, selectorHit, hit),
		buildPreviewGroup(displayMiss, selectorMiss, miss),
	}
}

func buildPreviewGroup(display string, selector *pb.PolicySelector, instances []*policygroup.RoutingModelInstance) *pb.PreviewLabelGroup {
	var instanceIDs []string
	var sampleInstanceID string
	for i, instance := range instances {
		if i == 0 {
			sampleInstanceID = instance.ModelWithProvider.Id
		}
		instanceIDs = append(instanceIDs, instance.ModelWithProvider.Id)
	}
	group := &pb.PreviewLabelGroup{
		DisplayName:      display,
		InstanceCount:    int64(len(instances)),
		InstanceIds:      instanceIDs,
		SampleInstanceId: sampleInstanceID,
		Selector:         selector,
	}
	return group
}

func buildLabelSelector(key string, operator string, value string) *pb.PolicySelector {
	req := &pb.PolicyRequirement{
		Type: common_types.PolicyBranchSelectorRequirementTypeLabel.String(),
		Label: &pb.LabelRequirement{
			Key:      key,
			Operator: operator,
		},
	}
	if value != "" && (operator == common_types.PolicySelectorLabelOpIn.String() || operator == common_types.PolicySelectorLabelOpNotIn.String()) {
		req.Label.Values = []string{value}
	}
	return &pb.PolicySelector{Requirements: []*pb.PolicyRequirement{req}}
}

func findLabelValue(labels map[string]string, key string) (string, bool) {
	for k, v := range labels {
		if strings.EqualFold(k, key) {
			return v, true
		}
	}
	return "", false
}
