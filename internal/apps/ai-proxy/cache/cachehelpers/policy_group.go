package cachehelpers

import (
	"context"
	"fmt"

	"github.com/erda-project/erda-proto-go/apps/aiproxy/policy_group/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/cache/cachetypes"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
)

func GetPolicyGroupByName(ctx context.Context, clientID, name string) (*pb.PolicyGroup, error) {
	group, err := TryGetPolicyGroupByName(ctx, clientID, name)
	if err != nil {
		return nil, err
	}
	if group == nil {
		return nil, fmt.Errorf("policy group not found: %s", name)
	}
	return group, nil
}

// TryGetPolicyGroupByName returns the policy group with the given name under the client.
func TryGetPolicyGroupByName(ctx context.Context, clientID, name string) (*pb.PolicyGroup, error) {
	cache := ctxhelper.MustGetCacheManager(ctx).(cachetypes.Manager)
	_, allGroupsV, err := cache.ListAll(ctx, cachetypes.ItemTypePolicyGroup)
	if err != nil {
		return nil, err
	}
	for _, group := range allGroupsV.([]*pb.PolicyGroup) {
		if group.ClientId == clientID && group.Name == name {
			return group, nil
		}
	}
	return nil, nil
}
