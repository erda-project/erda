package engine

import (
	modelpb "github.com/erda-project/erda-proto-go/apps/aiproxy/model/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/cache/cachehelpers"
	policygroup "github.com/erda-project/erda/internal/apps/ai-proxy/route/policy_group"
)

func testInstance(id string, labels map[string]string) *policygroup.RoutingModelInstance {
	return &policygroup.RoutingModelInstance{
		ModelWithProvider: &cachehelpers.ModelWithProvider{
			Model: &modelpb.Model{
				Id:   id,
				Name: id,
			},
		},
		Labels: labels,
	}
}
