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

package common_types

import (
	"sort"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type PolicyGroupSource string

const (
	PolicyGroupSourceUserDefined      PolicyGroupSource = "user_defined"
	PolicyGroupSourceForModelTemplate PolicyGroupSource = "for_model_template"
	PolicyGroupSourceRuntimeInternal  PolicyGroupSource = "runtime_internal"
)

func (s PolicyGroupSource) String() string { return string(s) }

func (s PolicyGroupSource) IsValid() bool {
	return s == PolicyGroupSourceUserDefined || s == PolicyGroupSourceForModelTemplate || s == PolicyGroupSourceRuntimeInternal
}

type PolicyGroupMode string

const (
	PolicyGroupModeWeighted PolicyGroupMode = "weighted"
	PolicyGroupModePriority PolicyGroupMode = "priority"
)

func (m PolicyGroupMode) String() string { return string(m) }

func (m PolicyGroupMode) IsValid() bool {
	return m == PolicyGroupModeWeighted || m == PolicyGroupModePriority
}

type PolicyBranchStrategy string

const (
	PolicyGroupBranchStrategyRoundRobin     PolicyBranchStrategy = "round_robin"
	PolicyGroupBranchStrategyConsistentHash PolicyBranchStrategy = "consistent_hash"
)

func (s PolicyBranchStrategy) String() string { return string(s) }

func (s PolicyBranchStrategy) IsValid() bool {
	return s == PolicyGroupBranchStrategyRoundRobin || s == PolicyGroupBranchStrategyConsistentHash
}

type PolicySelectorRequirementType string

const (
	PolicyBranchSelectorRequirementTypeLabel PolicySelectorRequirementType = "label"
)

func (s PolicySelectorRequirementType) String() string { return string(s) }

func (s PolicySelectorRequirementType) IsValid() bool {
	return s == PolicyBranchSelectorRequirementTypeLabel
}

// PolicyLabelSelector refer to K8s label selector,
// see: https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/
type PolicyLabelSelector metav1.LabelSelectorOperator

const (
	PolicySelectorLabelOpIn           = PolicyLabelSelector(metav1.LabelSelectorOpIn)
	PolicySelectorLabelOpNotIn        = PolicyLabelSelector(metav1.LabelSelectorOpNotIn)
	PolicySelectorLabelOpExists       = PolicyLabelSelector(metav1.LabelSelectorOpExists)
	PolicySelectorLabelOpDoesNotExist = PolicyLabelSelector(metav1.LabelSelectorOpDoesNotExist)
)

func (s PolicyLabelSelector) String() string { return string(s) }

func (s PolicyLabelSelector) IsValid() bool {
	switch s {
	case PolicySelectorLabelOpIn, PolicySelectorLabelOpNotIn,
		PolicySelectorLabelOpExists, PolicySelectorLabelOpDoesNotExist:
		return true
	default:
		return false
	}
}

// Built-in policy group label keys (excluding custom/metadata-specific labels).
const (
	// model
	PolicyLabelKeyModelInstanceID               = "model-instance-id"
	PolicyLabelKeyModelInstanceName             = "model-instance-name"
	PolicyLabelKeyModelPublisher                = "model-publisher"
	PolicyLabelKeyModelTemplateID               = "model-template-id"
	PolicyLabelKeyTemplate                      = PolicyLabelKeyModelTemplateID
	PolicyLabelKeyModelIsEnabled                = "model-is-enabled" // true / false
	PolicyLabelKeyModelPublisherModelTemplateID = "model-publisher/model-template-id"

	// service-provider
	PolicyLabelKeyServiceProviderInstanceID   = "service-provider-instance-id"
	PolicyLabelKeyServiceProviderInstanceName = "service-provider-instance-name"
	PolicyLabelKeyServiceProviderType         = "service-provider-type"

	// location
	PolicyLabelKeyLocation = "location"
	PolicyLabelKeyRegion   = "region"
	PolicyLabelKeyCountry  = "country"
)

var OfficialPolicyGroupLabelKeys = []string{
	// Model
	PolicyLabelKeyModelInstanceID,
	PolicyLabelKeyModelInstanceName,
	PolicyLabelKeyModelPublisher,
	PolicyLabelKeyModelTemplateID,
	PolicyLabelKeyModelIsEnabled,
	PolicyLabelKeyModelPublisherModelTemplateID,

	// service-provider
	PolicyLabelKeyServiceProviderInstanceID,
	PolicyLabelKeyServiceProviderInstanceName,
	PolicyLabelKeyServiceProviderType,

	// location
	PolicyLabelKeyLocation,
	PolicyLabelKeyRegion,
}

// ListOfficialPolicyGroupLabelKeys returns a copy of built-in label keys.
func ListOfficialPolicyGroupLabelKeys() []string {
	out := make([]string, len(OfficialPolicyGroupLabelKeys))
	copy(out, OfficialPolicyGroupLabelKeys)
	sort.Strings(out)
	return out
}

type LabelPreviewSimpleOp string

const (
	LabelPreviewOpGroupBy LabelPreviewSimpleOp = "group-by"
	LabelPreviewOpFilter  LabelPreviewSimpleOp = "filter"
	LabelPreviewOpSplit   LabelPreviewSimpleOp = "split"
)

func (s LabelPreviewSimpleOp) String() string { return string(s) }

func (s LabelPreviewSimpleOp) IsValid() bool {
	return s == LabelPreviewOpGroupBy || s == LabelPreviewOpFilter || s == LabelPreviewOpSplit
}

const (
	StickyKeyPrefixFromReqHeader = "req.header."
	StickyKeyOfXRequestID        = StickyKeyPrefixFromReqHeader + "x-request-id"
)
