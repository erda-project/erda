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
	"net/http"
	"reflect"
	"testing"

	"bou.ke/monkey"
	"github.com/gorilla/schema"
	"github.com/stretchr/testify/assert"

	pipelinepb "github.com/erda-project/erda-proto-go/core/pipeline/pipeline/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/internal/apps/dop/services/permission"
	"github.com/erda-project/erda/pkg/mock"
)

func Test_shouldCheckPermission(t *testing.T) {
	type args struct {
		isInternalClient       bool
		isInternalActionClient bool
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "isInternalClient",
			args: args{
				isInternalClient:       true,
				isInternalActionClient: false,
			},
			want: false,
		},
		{
			name: "isInternalActionClient",
			args: args{
				isInternalClient:       false,
				isInternalActionClient: true,
			},
			want: true,
		},
		{
			name: "isInternalClient_and_isInternalActionClient",
			args: args{
				isInternalClient:       true,
				isInternalActionClient: true,
			},
			want: true,
		},
		{
			name: "otherClient",
			args: args{
				isInternalClient:       false,
				isInternalActionClient: false,
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := shouldCheckPermission(tt.args.isInternalClient, tt.args.isInternalActionClient); got != tt.want {
				t.Errorf("shouldCheckPermission() = %v, want %v", got, tt.want)
			}
		})
	}
}

// func TestUpdateCmsNsConfigsWhenUserNotExist(t *testing.T) {
// 	var bdl *bundle.Bundle
// 	monkey.PatchInstanceMethod(reflect.TypeOf(bdl), "GetMemberByUserAndScope",
// 		func(*bundle.Bundle, apistructs.ScopeType, string, uint64) ([]apistructs.Member, error) {
// 			return nil, nil
// 		})
// 	defer monkey.UnpatchAll()
// 	e := New()
// 	assert.Equal(t, "the member is not exist", e.UpdateCmsNsConfigs("1", 1).Error())
// }

func TestEndpoints_pipelineDetail(t *testing.T) {

}

func Test_getPipelineDetailAndCheckPermission(t *testing.T) {
	type args struct {
		req          apistructs.CICDPipelineDetailRequest
		identityInfo apistructs.IdentityInfo
		result       pipelinepb.PipelineDetailDTO
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "test",
			args: args{
				req: apistructs.CICDPipelineDetailRequest{
					PipelineID:               1,
					SimplePipelineBaseResult: false,
				},
				result: pipelinepb.PipelineDetailDTO{
					ID:            1,
					ApplicationID: 1,
					Branch:        "master",
				},
				identityInfo: apistructs.IdentityInfo{
					UserID: "1",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			var pipelineSvc = &mock.MockPipelineServiceServer{}
			patch := monkey.PatchInstanceMethod(reflect.TypeOf(pipelineSvc), "PipelineDetail", func(_ *mock.MockPipelineServiceServer, ctx context.Context, req *pipelinepb.PipelineDetailRequest) (*pipelinepb.PipelineDetailResponse, error) {
				assert.Equal(t, req.PipelineID, tt.args.req.PipelineID)
				assert.Equal(t, req.SimplePipelineBaseResult, tt.args.req.SimplePipelineBaseResult)

				return &pipelinepb.PipelineDetailResponse{Data: &tt.args.result}, nil
			})

			defer patch.Unpatch()

			var permissionChecker = &permission.Permission{}
			patch1 := monkey.PatchInstanceMethod(reflect.TypeOf(permissionChecker), "CheckRuntimeBranch", func(p *permission.Permission, identityInfo apistructs.IdentityInfo, appID uint64, branch string, action string) error {
				assert.Equal(t, tt.args.identityInfo, identityInfo)
				assert.Equal(t, tt.args.result.ApplicationID, appID)
				assert.Equal(t, tt.args.result.Branch, branch)
				return nil
			})
			defer patch1.Unpatch()

			got, err := getPipelineDetailAndCheckPermission(pipelineSvc, permissionChecker, tt.args.req, tt.args.identityInfo)
			if (err != nil) != tt.wantErr {
				t.Errorf("getPipelineDetailAndCheckPermission() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.NotNil(t, got)
			assert.Equal(t, got.ID, tt.args.result.ID)
		})
	}
}

func Test_pipelineList(t *testing.T) {
	bdl := bundle.New()
	pm1 := monkey.PatchInstanceMethod(reflect.TypeOf(bdl), "PageListPipeline", func(bdl *bundle.Bundle, req apistructs.PipelinePageListRequest) (*apistructs.PipelinePageListData, error) {
		return &apistructs.PipelinePageListData{
			Pipelines: []apistructs.PagePipeline{
				{
					ID: 1,
				},
			},
		}, nil
	})
	defer pm1.Unpatch()

	e := &Endpoints{
		bdl:                bdl,
		queryStringDecoder: schema.NewDecoder(),
	}
	r, err := http.NewRequest("GET", "/api/pipelines?statuses=Running&sources=dice", nil)
	assert.NoError(t, err)
	_, err = e.pipelineList(context.Background(), r, nil)
	assert.NoError(t, err)
}

func Test_makeProjectDefaultLevelCmsNs(t *testing.T) {
	projectNamespace1 := makeProjectDefaultLevelCmsNs(1)
	assert.Equal(t, "project-1-default", projectNamespace1[0])

	projectNamespace2 := makeProjectDefaultLevelCmsNs(2)
	assert.Equal(t, "project-2-default", projectNamespace2[0])
}

func Test_makeOrgDefaultLevelCmsNs(t *testing.T) {
	orgNamespace1 := makeOrgDefaultLevelCmsNs(1)
	assert.Equal(t, "org-1-default", orgNamespace1[0])

	orgNamespace2 := makeOrgDefaultLevelCmsNs(2)
	assert.Equal(t, "org-2-default", orgNamespace2[0])
}
