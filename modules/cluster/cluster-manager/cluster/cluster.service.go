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

package cluster

import (
	"context"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda-proto-go/core/clustermanager/cluster/pb"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/cluster/cluster-manager/cluster/db"
	"github.com/erda-project/erda/modules/cluster/cluster-manager/services/apierrors"
	"github.com/erda-project/erda/pkg/common/apis"
	"github.com/erda-project/erda/pkg/http/httputil"
	"github.com/erda-project/erda/pkg/strutil"
)

type ClusterService struct {
	db  *db.ClusterDB
	bdl *bundle.Bundle
}

func (c *ClusterService) ListCluster(ctx context.Context, req *pb.ListClusterRequest) (*pb.ListClusterResponse, error) {
	var (
		clusters []*pb.ClusterInfo
		err      error
	)
	if err = auth(ctx); err != nil {
		return nil, err
	}

	clusterType := req.ClusterType
	if clusterType != "" {
		clusters, err = c.ListClusterByType(clusterType)
	} else {
		clusters, err = c.List()
	}
	if err != nil {
		return nil, apierrors.ErrListCluster.InternalError(err)
	}

	if req.OrgID == 0 {
		return &pb.ListClusterResponse{Data: clusters}, nil
	}

	clusterRelation, err := c.bdl.GetOrgClusterRelationsByOrg(req.OrgID)
	if err != nil {
		return nil, apierrors.ErrListCluster.InternalError(err)
	}

	var clustersInOrg []*pb.ClusterInfo
	for _, relation := range clusterRelation {
		for _, cluster := range clusters {
			if uint64(cluster.Id) == relation.ClusterID {
				clustersInOrg = append(clustersInOrg, cluster)
				break
			}
		}
	}
	return &pb.ListClusterResponse{
		Data: clusters,
	}, nil
}

func (c *ClusterService) GetCluster(ctx context.Context, req *pb.GetClusterRequest) (*pb.GetClusterResponse, error) {
	if err := auth(ctx); err != nil {
		return nil, err
	}

	cluster, err := c.Get(req.IdOrName)
	if err != nil {
		if strutil.Contains(err.Error(), "not found") {
			return nil, apierrors.ErrGetCluster.NotFound()
		}
		return nil, apierrors.ErrGetCluster.InternalError(err)
	}

	return &pb.GetClusterResponse{Data: cluster}, nil
}

func (c *ClusterService) CreateCluster(ctx context.Context, req *pb.CreateClusterRequest) (*pb.CreateClusterResponse, error) {
	logrus.Infof("request body: %+v", req)
	if err := auth(ctx); err != nil {
		return nil, err
	}
	if req.UserID == "" {
		return nil, apierrors.ErrCreateCluster.MissingParameter("userID")
	}

	if err := c.CreateWithEvent(req); err != nil {
		return nil, apierrors.ErrCreateCluster.InternalError(err)
	}

	if err := c.bdl.CreateOrgClusterRelationsByOrg(req.Name, req.UserID, req.OrgID); err != nil {
		return nil, apierrors.ErrCreateCluster.InternalError(err)
	}
	return &pb.CreateClusterResponse{}, nil
}

func (c *ClusterService) UpdateCluster(ctx context.Context, req *pb.UpdateClusterRequest) (*pb.UpdateClusterResponse, error) {
	logrus.Infof("request body: %+v", req)
	if err := auth(ctx); err != nil {
		return nil, err
	}

	if err := c.UpdateWithEvent(req); err != nil {
		if strutil.Contains(err.Error(), "not found") {
			return nil, apierrors.ErrGetCluster.NotFound()
		}
		return nil, apierrors.ErrUpdateCluster.InvalidParameter(err)
	}

	return &pb.UpdateClusterResponse{}, nil
}

func (c *ClusterService) DeleteCluster(ctx context.Context, req *pb.DeleteClusterRequest) (*pb.DeleteClusterResponse, error) {
	if err := auth(ctx); err != nil {
		return nil, err
	}

	if err := c.DeleteWithEvent(req.ClusterName); err != nil {
		return nil, err
	}

	return &pb.DeleteClusterResponse{}, nil
}

func (c *ClusterService) PatchCluster(ctx context.Context, req *pb.PatchClusterRequest) (*pb.PatchClusterResponse, error) {
	logrus.Infof("request body: %+v", req)
	if err := auth(ctx); err != nil {
		return nil, err
	}

	if err := c.PatchWithEvent(req); err != nil {
		if strutil.Contains(err.Error(), "not found") {
			return nil, apierrors.ErrGetCluster.NotFound()
		}
		return nil, apierrors.ErrPatchCluster.InvalidParameter(err)
	}

	return &pb.PatchClusterResponse{}, nil
}

func auth(ctx context.Context) error {
	internalClient := apis.GetHeader(ctx, httputil.InternalHeader)
	if internalClient == "" {
		return apierrors.ErrPreCheckCluster.AccessDenied()
	}
	return nil
}
