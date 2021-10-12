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

package apis

import (
	"context"
	"fmt"
	"testing"

	"bou.ke/monkey"

	"github.com/erda-project/erda-proto-go/core/monitor/alert/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/pkg/common/apis"
)

func Test_alertService_TriggerConditions(t *testing.T) {
	defer monkey.UnpatchAll()
	monkey.Patch(apis.GetOrgID, func(ctx context.Context) string {
		return "1"
	})
	monkey.Patch(apis.GetUserID, func(ctx context.Context) string {
		return "2"
	})
	monkey.Patch((*bundle.Bundle).GetOrg, func(_ *bundle.Bundle, _ interface{}) (*apistructs.OrgDTO, error) {
		return &apistructs.OrgDTO{
			ID:   1,
			Name: "terminus",
		}, nil
	})
	monkey.Patch((*bundle.Bundle).GetOrgClusterRelationsByOrg, func(_ *bundle.Bundle, orgID uint64) ([]apistructs.OrgClusterRelationDTO, error) {
		return []apistructs.OrgClusterRelationDTO{
			{
				ClusterName: "terminus-dev",
			},
		}, nil
	})
	monkey.Patch((*bundle.Bundle).GroupHosts, func(_ *bundle.Bundle, req apistructs.ResourceRequest, orgName, userId string) ([]*apistructs.HostData, error) {
		return []*apistructs.HostData{
			{
				ClusterName: "terminus-dev",
				IP:          "125.12.34.3",
			},
		}, nil
	})
	monkey.Patch((*bundle.Bundle).GetAppsByProject, func(_ *bundle.Bundle, projectID, orgID uint64, userID string) (*apistructs.ApplicationListResponseData, error) {
		return &apistructs.ApplicationListResponseData{
			Total: 1,
			List: []apistructs.ApplicationDTO{
				{
					ID:   10,
					Name: "demo",
				},
			},
		}, nil
	})
	monkey.Patch((*bundle.Bundle).GetServices, func(_ *bundle.Bundle, _ string) ([]apistructs.ServiceDashboard, error) {
		return []apistructs.ServiceDashboard{
			{
				Id:   "5_develop_apm-demo-api",
				Name: "apm-demo-api",
			},
		}, nil
	})
	pro := &provider{
		bdl:          &bundle.Bundle{},
		alertService: &alertService{},
	}
	pro.alertService.p = pro
	_, err := pro.alertService.TriggerConditions(context.Background(), &pb.TriggerConditionsRequest{
		ScopeType: "org",
		ScopeId:   "",
		Tk:        "",
	})
	if err != nil {
		fmt.Println("should not err", err)
	}
	_, err = pro.alertService.TriggerConditions(context.Background(), &pb.TriggerConditionsRequest{
		ScopeType: "project",
		ScopeId:   "4",
		Tk:        "fc1f8c074e46a9df505a15c1a94d62cc",
	})
	if err != nil {
		fmt.Println("should not err", err)
	}
}
