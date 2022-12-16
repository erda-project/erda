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

package endpoints

import (
	"context"
	"reflect"
	"testing"

	"bou.ke/monkey"
	"golang.org/x/text/language"
	"golang.org/x/text/message"

	orgpb "github.com/erda-project/erda-proto-go/core/org/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/internal/pkg/mock"
)

type OrgMock struct {
	mock.OrgMock
}

func (m OrgMock) GetOrgClusterRelationsByOrg(ctx context.Context, request *orgpb.GetOrgClusterRelationsByOrgRequest) (*orgpb.GetOrgClusterRelationsByOrgResponse, error) {
	return &orgpb.GetOrgClusterRelationsByOrgResponse{
		Data: []*orgpb.OrgClusterRelation{
			{
				OrgID:       1,
				ClusterName: "cluster",
			},
			{
				OrgID:       2,
				ClusterName: "cluster-2",
			},
		},
	}, nil
}

func TestCloudResourcePermissionCheck(t *testing.T) {
	bdl := bundle.New()
	orgSvc := &OrgMock{}

	monkey.PatchInstanceMethod(reflect.TypeOf(bdl), "CheckPermission", func(_ *bundle.Bundle,
		req *apistructs.PermissionCheckRequest) (*apistructs.PermissionCheckResponseData, error) {
		access := true
		if req.UserID == "1001" && req.ScopeID == 2 {
			access = false
		}
		return &apistructs.PermissionCheckResponseData{
			Access: access,
		}, nil
	})

	defer monkey.UnpatchAll()

	ctx := context.WithValue(context.Background(), "i18nPrinter", message.NewPrinter(language.SimplifiedChinese))

	ep := Endpoints{
		org: orgSvc,
		bdl: bdl,
	}

	type args struct {
		clusterName string
		userId      string
		orgId       string
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "permission denied",
			args: args{
				clusterName: "cluster",
				userId:      "10001",
				orgId:       "1",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ep.CloudResourcePermissionCheck(ctx, tt.args.userId, tt.args.orgId,
				tt.args.clusterName, apistructs.GetAction); (err != nil) != tt.wantErr {
				t.Errorf("CheckPermission() error = %v, wantErr %v", err, tt.wantErr)
			} else if err != nil && tt.wantErr {
				t.Log(err)
			}
		})
	}
}
