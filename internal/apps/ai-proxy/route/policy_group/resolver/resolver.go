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

package resolver

import (
	"context"
	"fmt"

	"github.com/mohae/deepcopy"

	"github.com/erda-project/erda-proto-go/apps/aiproxy/policy_group/pb"
	policypb "github.com/erda-project/erda-proto-go/apps/aiproxy/policy_group/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/cache/cachehelpers"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/audit/audithelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/common_types"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/policy_group/selector"
)

// Resolver implements the multi-layer resolution:
// - user-defined: concrete policy group name, such as "cheapest"
// - runtime-interval
type Resolver struct{}

func NewResolver() *Resolver {
	return &Resolver{}
}

func (r *Resolver) Resolve(ctx context.Context, clientID, inputGroupName string) (group *policypb.PolicyGroup, err error) {
	if clientID == "" {
		return nil, fmt.Errorf("missing client id")
	}
	if inputGroupName == "" {
		return nil, fmt.Errorf("missing model")
	}

	// save the picked group into audit for debug
	defer func() {
		if group != nil {
			audithelper.Note(ctx, "policy-group.picked", group)
		}
	}()

	// user-defined
	group, err = r.resolveUserDefined(ctx, clientID, inputGroupName)
	if err != nil {
		return nil, err
	}
	if group != nil {
		return group, nil
	}

	// runtime-internal
	group, err = r.resolveRuntimeInternal(ctx, clientID, inputGroupName)
	if err != nil {
		return nil, err
	}
	if group != nil {
		return group, nil
	}

	return nil, fmt.Errorf("no policy group matched for model %q", inputGroupName)
}

func (r *Resolver) resolveUserDefined(ctx context.Context, clientID, inputGroupName string) (*policypb.PolicyGroup, error) {
	group, err := cachehelpers.TryGetPolicyGroupByName(ctx, clientID, inputGroupName)
	if err != nil {
		return nil, err
	}
	if group == nil {
		return nil, nil
	}
	return group, nil
}

func (r *Resolver) resolveRuntimeInternal(ctx context.Context, clientID, inputGroupName string) (*policypb.PolicyGroup, error) {
	runtimeInternalGroups, err := r.buildRuntimeInternalGroups(ctx, clientID)
	if err != nil {
		return nil, err
	}

	for _, group := range runtimeInternalGroups {
		if group.Name == inputGroupName {
			return group, nil
		}
	}

	return nil, nil
}

// buildRuntimeInternalGroups builds a map of acceptable names.
// availableName rules (priority high -> low):
//   - model-publisher/model-template-id
//   - model-template-id
//   - model-instance-id
//   - model-instance-name
//   - for-compatible
func (r *Resolver) buildRuntimeInternalGroups(ctx context.Context, clientID string) ([]*policypb.PolicyGroup, error) {
	allEnabledInstances, err := cachehelpers.ListAllClientModels(ctx, clientID, &cachehelpers.ClientModelConfig{OnlyEnabled: true})
	if err != nil {
		return nil, err
	}

	// prepare instance-instances for each type of group, in order
	var modelPublisherModelTemplateIDs []string
	var modelTemplateIDs []string
	var modelInstanceIDs []string
	var modelInstanceNames []string

	for _, instance := range allEnabledInstances {
		// model-publisher/instance-template-id
		modelPublisherModelTemplateIDs = append(modelPublisherModelTemplateIDs, instance.Publisher+"/"+instance.TemplateId)
		// model-template-id
		modelTemplateIDs = append(modelTemplateIDs, instance.TemplateId)
		// model-instance-id
		modelInstanceIDs = append(modelInstanceIDs, instance.Id)
		// model-instance-name, maybe the same under one client but in different providers
		modelInstanceNames = append(modelInstanceNames, instance.Name)
	}

	var allPolicyGroups []*policypb.PolicyGroup

	allPolicyGroups = append(allPolicyGroups, buildRuntimeInternalGroupForModelPublisherModelTemplateID(ctx, clientID, modelPublisherModelTemplateIDs)...)
	allPolicyGroups = append(allPolicyGroups, buildRuntimeInternalGroupForModelTemplateID(ctx, clientID, modelTemplateIDs)...)
	allPolicyGroups = append(allPolicyGroups, buildRuntimeInternalGroupForModelInstanceID(ctx, clientID, modelInstanceIDs)...)
	allPolicyGroups = append(allPolicyGroups, buildRuntimeInternalGroupForModelInstanceName(ctx, clientID, modelInstanceNames)...)
	allPolicyGroups = append(allPolicyGroups, buildRuntimeInternalGroupForCompatible(ctx, clientID, allEnabledInstances)...)

	return allPolicyGroups, nil
}

func buildRuntimeInternalGroupForModelPublisherModelTemplateID(ctx context.Context, clientID string, modelPublisherModelTemplateIDs []string) []*policypb.PolicyGroup {
	var groups []*policypb.PolicyGroup
	for _, modelPublisherModelTemplateID := range modelPublisherModelTemplateIDs {
		group := &policypb.PolicyGroup{
			ClientId:  clientID,
			Name:      modelPublisherModelTemplateID,
			Desc:      "runtime-internal auto route to target model-publisher/model-template-id",
			Mode:      common_types.PolicyGroupModeWeighted.String(),
			StickyKey: common_types.StickyKeyOfXRequestID,
			Source:    common_types.PolicyGroupSourceRuntimeInternal.String(),
			Branches: []*policypb.PolicyBranch{
				{
					Name:     "auto",
					Desc:     "auto route to target model-publisher/model-template-id",
					Weight:   1,
					Priority: 1,
					Strategy: common_types.PolicyGroupBranchStrategyRoundRobin.String(),
					Selector: selector.BuildLabelSelectorForKVIn(common_types.PolicyLabelKeyModelPublisherModelTemplateID, modelPublisherModelTemplateID),
				},
			},
		}
		groups = append(groups, group)
	}
	return groups
}

func buildRuntimeInternalGroupForModelTemplateID(ctx context.Context, clientID string, modelTemplateIDs []string) []*policypb.PolicyGroup {
	var groups []*policypb.PolicyGroup
	for _, modelTemplateID := range modelTemplateIDs {
		group := &policypb.PolicyGroup{
			ClientId:  clientID,
			Name:      modelTemplateID,
			Desc:      "runtime-internal auto route to target model-template-id",
			Mode:      common_types.PolicyGroupModeWeighted.String(),
			StickyKey: common_types.StickyKeyOfXRequestID,
			Source:    common_types.PolicyGroupSourceRuntimeInternal.String(),
			Branches: []*policypb.PolicyBranch{
				{
					Name:     "auto",
					Desc:     "auto route to target model-template-id",
					Weight:   1,
					Priority: 1,
					Strategy: common_types.PolicyGroupBranchStrategyRoundRobin.String(),
					Selector: selector.BuildLabelSelectorForKVIn(common_types.PolicyLabelKeyModelTemplateID, modelTemplateID),
				},
			},
		}
		groups = append(groups, group)
	}
	return groups
}

func buildRuntimeInternalGroupForModelInstanceID(ctx context.Context, clientID string, modelInstanceIDs []string) []*policypb.PolicyGroup {
	var groups []*policypb.PolicyGroup
	for _, modelInstanceID := range modelInstanceIDs {
		group := &policypb.PolicyGroup{
			ClientId:  clientID,
			Name:      modelInstanceID,
			Desc:      "runtime-internal auto route to target model-instance-id",
			Mode:      common_types.PolicyGroupModeWeighted.String(),
			StickyKey: common_types.StickyKeyOfXRequestID,
			Source:    common_types.PolicyGroupSourceRuntimeInternal.String(),
			Branches: []*policypb.PolicyBranch{
				{
					Name:     "auto",
					Weight:   1,
					Priority: 1,
					Strategy: common_types.PolicyGroupBranchStrategyRoundRobin.String(),
					Selector: selector.BuildLabelSelectorForKVIn(common_types.PolicyLabelKeyModelInstanceID, modelInstanceID),
				},
			},
		}
		// support format: [ID:xxx]
		group2 := deepcopy.Copy(group).(*policypb.PolicyGroup)
		group2.Name = fmt.Sprintf("[ID:%s]", modelInstanceID)
		groups = append(groups, group, group2)
	}
	return groups
}

func buildRuntimeInternalGroupForModelInstanceName(ctx context.Context, clientID string, modelInstanceNames []string) []*policypb.PolicyGroup {
	var groups []*policypb.PolicyGroup
	for _, modelInstanceName := range modelInstanceNames {
		group := &policypb.PolicyGroup{
			ClientId:  clientID,
			Name:      modelInstanceName,
			Desc:      "runtime-internal auto route to target model-instance-name",
			Mode:      common_types.PolicyGroupModeWeighted.String(),
			StickyKey: common_types.StickyKeyOfXRequestID,
			Source:    common_types.PolicyGroupSourceRuntimeInternal.String(),
			Branches: []*policypb.PolicyBranch{
				{
					Name:     "auto",
					Desc:     "auto route to target model-instance-name",
					Weight:   1,
					Priority: 1,
					Strategy: common_types.PolicyGroupBranchStrategyRoundRobin.String(),
					Selector: selector.BuildLabelSelectorForKVIn(common_types.PolicyLabelKeyModelInstanceName, modelInstanceName),
				},
			},
		}
		groups = append(groups, group)
	}
	return groups
}

// buildRuntimeInternalGroupForCompatible
// - model-publisher/model-instance-name
// - service-provider-type/model-template-id
func buildRuntimeInternalGroupForCompatible(ctx context.Context, clientID string, allEnabledInstances []*cachehelpers.ModelWithProvider) []*policypb.PolicyGroup {
	var groups []*policypb.PolicyGroup
	for _, instance := range allEnabledInstances {
		baseGroup := &policypb.PolicyGroup{
			ClientId:  clientID,
			Desc:      "runtime-internal auto route for compatible",
			Mode:      common_types.PolicyGroupModeWeighted.String(),
			StickyKey: common_types.StickyKeyOfXRequestID,
			Source:    common_types.PolicyGroupSourceRuntimeInternal.String(),
		}
		// model-publisher/model-instance-name
		groupForModelPublisherModelInstanceName := deepcopy.Copy(baseGroup).(*policypb.PolicyGroup)
		groupForModelPublisherModelInstanceName.Name = instance.Publisher + "/" + instance.Name
		groupForModelPublisherModelInstanceName.Desc += ": model-publisher/model-instance-name"
		groupForModelPublisherModelInstanceName.Branches = []*policypb.PolicyBranch{
			{
				Name:     "auto",
				Desc:     "auto route to target model-publisher/model-instance-name",
				Weight:   1,
				Priority: 1,
				Strategy: common_types.PolicyGroupBranchStrategyRoundRobin.String(),
				Selector: &policypb.PolicySelector{
					Requirements: []*policypb.PolicyRequirement{
						{ // match model-publisher
							Type: common_types.PolicyBranchSelectorRequirementTypeLabel.String(),
							Label: &pb.LabelRequirement{
								Key:      common_types.PolicyLabelKeyModelPublisher,
								Operator: common_types.PolicySelectorLabelOpIn.String(),
								Values:   []string{instance.Publisher},
							},
						},
						{ // match model-instance-name
							Type: common_types.PolicyBranchSelectorRequirementTypeLabel.String(),
							Label: &pb.LabelRequirement{
								Key:      common_types.PolicyLabelKeyModelInstanceName,
								Operator: common_types.PolicySelectorLabelOpIn.String(),
								Values:   []string{instance.Name},
							},
						},
					},
				},
			},
		}
		groups = append(groups, groupForModelPublisherModelInstanceName)
		// service-provider-type/model-template-id
		groupForServiceProviderTypeModelTemplateID := deepcopy.Copy(baseGroup).(*policypb.PolicyGroup)
		groupForServiceProviderTypeModelTemplateID.Name = instance.Provider.Type + "/" + instance.TemplateId
		groupForServiceProviderTypeModelTemplateID.Desc += ": service-provider-type/model-template-id"
		groupForServiceProviderTypeModelTemplateID.Branches = []*policypb.PolicyBranch{
			{
				Name:     "auto",
				Desc:     "auto route to target model-instance-name",
				Weight:   1,
				Priority: 1,
				Strategy: common_types.PolicyGroupBranchStrategyRoundRobin.String(),
				Selector: &policypb.PolicySelector{
					Requirements: []*policypb.PolicyRequirement{
						{ // match model-publisher
							Type: common_types.PolicyBranchSelectorRequirementTypeLabel.String(),
							Label: &pb.LabelRequirement{
								Key:      common_types.PolicyLabelKeyServiceProviderType,
								Operator: common_types.PolicySelectorLabelOpIn.String(),
								Values:   []string{instance.Provider.Type},
							},
						},
						{ // match model-instance-name
							Type: common_types.PolicyBranchSelectorRequirementTypeLabel.String(),
							Label: &pb.LabelRequirement{
								Key:      common_types.PolicyLabelKeyModelTemplateID,
								Operator: common_types.PolicySelectorLabelOpIn.String(),
								Values:   []string{instance.TemplateId},
							},
						},
					},
				},
			},
		}
		groups = append(groups, groupForServiceProviderTypeModelTemplateID)
	}
	return groups
}
