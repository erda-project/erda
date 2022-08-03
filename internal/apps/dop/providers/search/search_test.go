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

package search

import (
	"context"
	"reflect"
	"testing"

	"bou.ke/monkey"
	"github.com/alecthomas/assert"

	"github.com/erda-project/erda-proto-go/common/pb"
	pb2 "github.com/erda-project/erda-proto-go/dop/search/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/internal/apps/dop/providers/search/handlers"
	"github.com/erda-project/erda/pkg/common/apis"
	"github.com/erda-project/erda/pkg/mock"
)

func TestSearch(t *testing.T) {
	bdl := bundle.New()
	pm1 := monkey.PatchInstanceMethod(reflect.TypeOf(bdl), "ListMyProject", func(_ *bundle.Bundle, userID string, req apistructs.ProjectListRequest) (*apistructs.PagingProjectDTO, error) {
		return &apistructs.PagingProjectDTO{
			Total: 2,
			List: []apistructs.ProjectDTO{
				{
					ID:   1,
					Name: "project1",
				},
				{
					ID:   2,
					Name: "project2",
				},
			},
		}, nil
	})
	defer pm1.Unpatch()

	pm2 := monkey.PatchInstanceMethod(reflect.TypeOf(bdl), "GetOrg", func(_ *bundle.Bundle, idOrName interface{}) (*apistructs.OrgDTO, error) {
		return &apistructs.OrgDTO{
			ID:   1,
			Name: "org1",
		}, nil
	})
	defer pm2.Unpatch()

	pm3 := monkey.Patch(apis.GetIdentityInfo, func(_ context.Context) *pb.IdentityInfo {
		return &pb.IdentityInfo{UserID: "1"}
	})
	defer pm3.Unpatch()

	pm4 := monkey.Patch(apis.GetIntOrgID, func(_ context.Context) (int64, error) {
		return 1, nil
	})
	defer pm4.Unpatch()

	pm5 := monkey.Patch(apis.GetOrgID, func(_ context.Context) string {
		return "1"
	})
	defer pm5.Unpatch()

	p := &provider{
		Query: &mock.MockIssueQuery{},
		bdl:   bdl,
		Cfg:   &config{},
	}
	service := &ServiceImpl{
		query: p.Query,
		bdl:   p.bdl,
	}
	res, err := service.Search(context.Background(), &pb2.SearchRequest{
		Query: "issue",
	})
	assert.NoError(t, err)
	assert.Equal(t, 1, len(res.Data))
	assert.Equal(t, handlers.SearchTypeIssue.String(), res.Data[0].Type)
	assert.Equal(t, 2, len(res.Data[0].Items))
}

func TestInit(t *testing.T) {
	p := &provider{}
	err := p.Init(nil)
	assert.NoError(t, err)
}

func Test_checkRequest(t *testing.T) {
	type args struct {
		req *pb2.SearchRequest
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "ok",
			args: args{
				req: &pb2.SearchRequest{
					Query: "query",
					IdentityInfo: &pb.IdentityInfo{
						UserID: "1",
						OrgID:  "1",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "missing query",
			args: args{
				req: &pb2.SearchRequest{
					Query: "",
				},
			},
			wantErr: true,
		},
		{
			name: "missing user id",
			args: args{
				req: &pb2.SearchRequest{
					Query:        "query",
					IdentityInfo: &pb.IdentityInfo{},
				},
			},
			wantErr: true,
		},
	}
	s := &ServiceImpl{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := s.checkRequest(tt.args.req); (got != nil) != tt.wantErr {
				t.Errorf("checkRequest() = %v, want %v", got, tt.wantErr)
			}
		})
	}
}
