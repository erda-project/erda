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

package runtime

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/metadata"

	"github.com/erda-project/erda-proto-go/orchestrator/runtime/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/orchestrator/components/runtime/mock"
	"github.com/erda-project/erda/modules/orchestrator/dbclient"
	"github.com/erda-project/erda/pkg/database/dbengine"
	"github.com/erda-project/erda/pkg/parser/diceyml"
)

func TestService_GetRuntime(t *testing.T) {
	assert := require.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	bdlSvc := mock.NewMockBundleService(ctrl)
	dbSvc := mock.NewMockDBService(ctrl)
	svc := NewRuntimeService(WithBundleService(bdlSvc), WithDBService(dbSvc))

	md := metadata.New(map[string]string{
		"user-id": "2",
		"org-id":  "4",
	})
	resp, err := svc.GetRuntime(context.Background(), &pb.GetRuntimeRequest{})
	// not login
	assert.NotNil(err)
	ctx := metadata.NewIncomingContext(context.Background(), md)
	// invalid param
	resp, err = svc.GetRuntime(ctx, &pb.GetRuntimeRequest{})
	assert.NotNil(err)

	runtime := &dbclient.Runtime{BaseModel: dbengine.BaseModel{
		ID: 1,
	},
		ClusterName:   "foo",
		ApplicationID: 1,
		ScheduleName: dbclient.ScheduleName{
			Namespace: "ns",
			Name:      "n",
		},
	}

	dbSvc.
		EXPECT().
		GetRuntimeAllowNil(gomock.Eq(uint64(1))).
		Return(runtime, nil).MinTimes(1)

	bdlSvc.
		EXPECT().
		CheckPermission(gomock.Any()).
		Return(&apistructs.PermissionCheckResponseData{Access: true}, nil).MinTimes(1)

	dbSvc.
		EXPECT().
		FindLastDeployment(gomock.Eq(uint64(1))).
		Return(&dbclient.Deployment{
			RuntimeId: uint64(10),
			Status:    apistructs.DeploymentStatusDeploying,
			Dice: `{
"services": {
  "sa": {
    "image": "latest",
    "expose": [1],
    "ports": [{"port": 1, "expose": true}]
  }
}
}`,
		}, nil).MinTimes(1)

	dbSvc.
		EXPECT().
		FindDomainsByRuntimeId(gomock.Eq(uint64(1))).
		Return([]dbclient.RuntimeDomain{
			{
				Domain: "foo.com",
			},
		}, nil).MinTimes(1)

	bdlSvc.
		EXPECT().
		GetCluster(gomock.Eq("foo")).
		Return(&apistructs.ClusterInfo{
			Name: "foo",
		}, nil).MinTimes(1)

	bdlSvc.
		EXPECT().
		GetApp(gomock.Eq(uint64(1))).
		Return(&apistructs.ApplicationDTO{
			Name: "foo",
		}, nil).MinTimes(1)

	bdlSvc.
		EXPECT().
		InspectServiceGroupWithTimeout(gomock.Eq("ns"), gomock.Eq("n")).
		Return(&apistructs.ServiceGroup{
			StatusDesc: apistructs.StatusDesc{
				Status: apistructs.StatusFailed,
			},
			Dice: apistructs.Dice{
				Services: []apistructs.Service{
					{
						Name: "sa",
						Ports: []diceyml.ServicePort{
							{
								Port:   1,
								Expose: true,
							},
						},
						Vip: "vip",
					},
				},
			},
		}, nil).
		MinTimes(1)

	resp, err = svc.GetRuntime(ctx, &pb.GetRuntimeRequest{
		NameOrID:  "1",
		AppID:     "2",
		Workspace: "dev",
	})

	assert.Nil(err)
	assert.NotNil(resp)

	dbSvc.
		EXPECT().
		FindRuntime(gomock.Any()).
		Return(runtime, nil).MinTimes(1)

	resp, err = svc.GetRuntime(ctx, &pb.GetRuntimeRequest{
		NameOrID:  "name",
		AppID:     "2",
		Workspace: "dev",
	})
	assert.Nil(err)
	assert.NotNil(resp)
}
