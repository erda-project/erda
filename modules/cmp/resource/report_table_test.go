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

package resource_test

import (
	"context"
	"testing"

	"github.com/rancher/apiserver/pkg/types"

	"github.com/erda-project/erda-infra/providers/i18n"
	"github.com/erda-project/erda-proto-go/cmp/dashboard/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/cmp/resource"
)

type fakeCmp struct {
}

func (f fakeCmp) ListSteveResource(ctx context.Context, req *apistructs.SteveRequest) ([]types.APIObject, error) {
	return nil, nil
}

func (f fakeCmp) GetNamespacesResources(ctx context.Context, nReq *pb.GetNamespacesResourcesRequest) (*pb.GetNamespacesResourcesResponse, error) {
	return nil, nil
}

func TestNewReportTable(t *testing.T) {
	var bdl bundle.Bundle
	var cmp fakeCmp
	var trans i18n.Translator
	resource.NewReportTable(
		resource.ReportTableWithBundle(&bdl),
		resource.ReportTableWithCMP(cmp),
		resource.ReportTableWithTrans(trans),
	)
}

func TestAddResourceForEveryProject(t *testing.T) {
	var (
		namespaces = &apistructs.GetProjectsNamesapcesResponseData{
			Total: 0,
			List: []*apistructs.ProjectNamespaces{
				{
					ProjectID:          1,
					ProjectName:        "project-1",
					ProjectDisplayName: "project-1",
					ProjectDesc:        "",
					OwnerUserID:        1,
					OwnerUserName:      "user-1",
					OwnerUserNickname:  "user-",
					CPUQuota:           10,
					MemQuota:           10,
					Clusters: map[string][]string{
						"cluster-1": {"namespace-1", "namespace-2"},
					},
				},
			},
		}
		resources = &pb.GetNamespacesResourcesResponse{
			Total: 0,
			List: []*pb.ClusterResourceItem{
				{
					Success:     true,
					Err:         "",
					ClusterName: "cluster-1",
					List: []*pb.NamespaceResourceDetail{
						{
							Namespace:  "namespace-1",
							CpuRequest: 5,
							MemRequest: 5,
						},
					},
				},
			},
		}
	)
	data, _ := resource.AddResourceForEveryProject(namespaces, resources)
	t.Logf("data: %+v", data)
}
