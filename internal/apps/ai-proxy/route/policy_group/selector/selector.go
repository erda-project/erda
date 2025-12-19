package selector

import (
	"strings"

	"github.com/erda-project/erda-proto-go/apps/aiproxy/policy_group/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/common_types"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/policy_group"
)

// MatchSelector filters instances by selector (AND over requirements).
func MatchSelector(instances []*policy_group.RoutingModelInstance, selector *pb.PolicySelector) []*policy_group.RoutingModelInstance {
	if selector == nil || len(selector.Requirements) == 0 {
		return instances
	}
	var matches []*policy_group.RoutingModelInstance
	for _, instance := range instances {
		if matchSelector(selector, instance.Labels) {
			matches = append(matches, instance)
		}
	}
	return matches
}

func matchSelector(sel *pb.PolicySelector, labels map[string]string) bool {
	if sel == nil || len(sel.Requirements) == 0 {
		return true
	}
	for _, r := range sel.Requirements {
		if !matchRequirement(r, labels) {
			return false
		}
	}
	return true
}

func matchRequirement(r *pb.PolicyRequirement, labels map[string]string) bool {
	if r == nil {
		return true
	}
	tp := r.Type
	switch tp {
	case common_types.PolicyBranchSelectorRequirementTypeLabel.String():
		return matchLabelRequirement(r.Label, labels)
	default:
		return false
	}
}

func matchLabelRequirement(lr *pb.LabelRequirement, labels map[string]string) bool {
	if lr == nil {
		return true
	}
	val, ok := "", false
	for k, v := range labels {
		if strings.EqualFold(k, lr.Key) {
			val, ok = v, true
			break
		}
	}
	switch lr.Operator {
	case common_types.PolicySelectorLabelOpIn.String():
		if !ok {
			return false
		}
		return strInSliceCaseInsensitive(val, lr.Values)
	case common_types.PolicySelectorLabelOpNotIn.String():
		if !ok {
			return true
		}
		return !strInSliceCaseInsensitive(val, lr.Values)
	case common_types.PolicySelectorLabelOpExists.String():
		return ok
	case common_types.PolicySelectorLabelOpDoesNotExist.String():
		return !ok
	default:
		return false
	}
}

func strInSliceCaseInsensitive(val string, list []string) bool {
	for _, item := range list {
		if strings.EqualFold(val, item) {
			return true
		}
	}
	return false
}

// BuildLabelSelectorForKVIn helper to build selector for a key=value.
func BuildLabelSelectorForKVIn(key, value string) *pb.PolicySelector {
	return &pb.PolicySelector{
		Requirements: []*pb.PolicyRequirement{
			{
				Type: common_types.PolicyBranchSelectorRequirementTypeLabel.String(),
				Label: &pb.LabelRequirement{
					Key:      key,
					Operator: common_types.PolicySelectorLabelOpIn.String(),
					Values:   []string{value},
				},
			},
		},
	}
}
