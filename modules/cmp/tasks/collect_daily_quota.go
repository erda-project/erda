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

package tasks

import (
	"context"

	"github.com/rancher/apiserver/pkg/types"

	"github.com/erda-project/erda-proto-go/cmp/dashboard/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/cmp/dbclient"
)

type DailyQuotaCollector struct {
	db  *dbclient.DBClient
	bdl bundle.Bundle
	cmp interface {
		ListSteveResource(ctx context.Context, req *apistructs.SteveRequest) ([]types.APIObject, error)
		GetNamespacesResources(ctx context.Context, nReq *pb.GetNamespacesResourcesRequest) (*pb.GetNamespacesResourcesResponse, error)
	}
}

func (d *DailyQuotaCollector) Task() error {
	var (
		projectsDaily []*apistructs.ProjectResourceDailyModel
		clusterDaily  []*apistructs.ClusterResourceDailyModel
	)
	_ = projectsDaily
	_ = clusterDaily
	// 1) 采集项目 quota

	// 2) 采集项目 request

	// 3) 采集 cluster quota

	// 4) 采集 cluster request

	return nil
}
