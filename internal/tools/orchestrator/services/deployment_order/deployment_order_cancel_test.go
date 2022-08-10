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

package deployment_order

import (
	"context"
	"reflect"
	"testing"

	"bou.ke/monkey"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	releasepb "github.com/erda-project/erda-proto-go/core/dicehub/release/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/internal/tools/orchestrator/dbclient"
)

func TestCancel(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	rss := NewMockReleaseServiceServer(ctrl)

	rss.EXPECT().GetRelease(gomock.Any(), gomock.Any()).AnyTimes().Return(&releasepb.ReleaseGetResponse{
		Data: &releasepb.ReleaseGetResponseData{
			ReleaseID: "202cb962ac59075b964b07152d234b70",
			Modes: map[string]*releasepb.ModeSummary{
				"default": {
					ApplicationReleaseList: []*releasepb.ReleaseSummaryArray{
						{
							List: []*releasepb.ApplicationReleaseSummary{
								{
									ReleaseID:       "202cb962ac59075b964b07152d234b70",
									ApplicationName: "app-demo",
									ApplicationID:   1,
								},
							},
						},
					},
				},
			},
		},
	}, nil)

	order := New(WithReleaseSvc(rss))

	defer monkey.UnpatchAll()
	monkey.PatchInstanceMethod(reflect.TypeOf(order.db), "GetDeploymentOrder", func(*dbclient.DBClient, string) (*dbclient.DeploymentOrder, error) {
		return &dbclient.DeploymentOrder{
			Modes: DefaultMode,
		}, nil
	})
	monkey.PatchInstanceMethod(reflect.TypeOf(order.db), "UpdateDeploymentOrder", func(*dbclient.DBClient, *dbclient.DeploymentOrder) error {
		return nil
	})
	monkey.PatchInstanceMethod(reflect.TypeOf(order.db), "GetRuntimeByDeployOrderId", func(*dbclient.DBClient, uint64, string) (*[]dbclient.Runtime, error) {
		return &[]dbclient.Runtime{}, nil
	})

	monkey.PatchInstanceMethod(reflect.TypeOf(order.bdl), "CheckPermission", func(*bundle.Bundle, *apistructs.PermissionCheckRequest) (*apistructs.PermissionCheckResponseData, error) {
		return &apistructs.PermissionCheckResponseData{
			Access: true,
		}, nil
	})

	_, err := order.Cancel(context.Background(), &apistructs.DeploymentOrderCancelRequest{
		DeploymentOrderId: "demo-order-id",
	})
	assert.NoError(t, err)
}
