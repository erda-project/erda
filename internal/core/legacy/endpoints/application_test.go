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
	"net/url"
	"reflect"
	"testing"

	"github.com/gorilla/schema"
	"github.com/stretchr/testify/assert"

	orgpb "github.com/erda-project/erda-proto-go/core/org/pb"
	"github.com/erda-project/erda/internal/core/org"
	"github.com/erda-project/erda/pkg/common/apis"
	"github.com/erda-project/erda/pkg/discover"
)

//func Test_transferAppsToApplicationDTOS(t *testing.T) {
//	var orgSvc = &org.Org{}
//	patch1 := monkey.PatchInstanceMethod(reflect.TypeOf(orgSvc), "ListOrgs", func(app *org.Org, orgIDs []int64, req *apistructs.OrgSearchRequest, all bool) (int, []model.Org, error) {
//		return 1, []model.Org{{BaseModel: model.BaseModel{ID: 1}}}, nil
//	})
//	defer patch1.Unpatch()
//
//	var pj = &project.Project{}
//	patch2 := monkey.PatchInstanceMethod(reflect.TypeOf(pj), "GetModelProjectsMap", func(project *project.Project, projectIDs []uint64) (map[int64]*model.Project, error) {
//		return map[int64]*model.Project{
//			1: {BaseModel: model.BaseModel{
//				ID: 1,
//			}},
//		}, nil
//	})
//	defer patch2.Unpatch()
//
//	var db = &dao.DBClient{}
//
//	ep := Endpoints{
//		org:     orgSvc,
//		project: pj,
//		db:      db,
//	}
//
//	apps := []model.Application{{BaseModel: model.BaseModel{ID: 1}, OrgID: 1, ProjectID: 1}}
//	_, err := ep.transferAppsToApplicationDTOS(true, apps, map[uint64]string{}, map[int64][]string{})
//	assert.NoError(t, err)
//}

func TestGetAppParams(t *testing.T) {
	// init Endpoints with queryStringDecoder
	queryStringDecoder := schema.NewDecoder()
	queryStringDecoder.IgnoreUnknownKeys(true)
	ep := &Endpoints{
		queryStringDecoder: queryStringDecoder,
	}

	req, err := http.NewRequest("GET", "https://baidu.com", nil)
	if err != nil {
		panic(err)
	}

	params := make(url.Values)
	params.Add("applicationID", "1")
	params.Add("applicationID", "2")
	req.URL.RawQuery = params.Encode()

	parsedReq, err := getListApplicationsParam(ep, req)
	assert.NoError(t, err)
	assert.Equal(t, parsedReq.ApplicationID, []uint64{1, 2})
}

func TestEndpoints_getOrg(t *testing.T) {
	type fields struct {
		org org.Interface
	}
	type args struct {
		ctx   context.Context
		orgID int64
	}
	mockOrg := MockOrg{}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *orgpb.Org
		wantErr bool
	}{
		{
			name: "test with userID",
			fields: fields{
				org: mockOrg,
			},
			args: args{
				ctx:   apis.WithUserIDContext(context.Background(), "1"),
				orgID: 1,
			},
			want:    &orgpb.Org{},
			wantErr: false,
		},
		{
			name: "test with internal-client",
			fields: fields{
				org: mockOrg,
			},
			args: args{
				ctx:   apis.WithInternalClientContext(context.Background(), discover.SvcErdaServer),
				orgID: 1,
			},
			want:    &orgpb.Org{},
			wantErr: false,
		},
		{
			name: "test with error",
			fields: fields{
				org: mockOrg,
			},
			args: args{
				ctx:   context.Background(),
				orgID: 1,
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := Endpoints{
				org: tt.fields.org,
			}
			got, err := e.getOrg(tt.args.ctx, tt.args.orgID)
			if (err != nil) != tt.wantErr {
				t.Errorf("getOrg() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getOrg() got = %v, want %v", got, tt.want)
			}
		})
	}
}
