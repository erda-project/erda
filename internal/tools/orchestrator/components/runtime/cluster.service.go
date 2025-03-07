package runtime

import "context"
import clusterpb "github.com/erda-project/erda-proto-go/core/clustermanager/cluster/pb"

type ClusterService interface {
	GetCluster(context.Context, *clusterpb.GetClusterRequest) (*clusterpb.GetClusterResponse, error)
}
