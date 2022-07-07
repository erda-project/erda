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
	"fmt"
	"reflect"
	"testing"

	orgpb "github.com/erda-project/erda-proto-go/core/org/pb"
	"github.com/erda-project/erda/internal/core/org"
	"github.com/erda-project/erda/internal/pkg/mock"
	"github.com/erda-project/erda/pkg/common/apis"
)

type orgMock struct {
	mock.OrgMock
}

func (m orgMock) ListOrg(ctx context.Context, request *orgpb.ListOrgRequest) (*orgpb.ListOrgResponse, error) {
	if apis.GetUserID(ctx) == "" {
		return nil, fmt.Errorf("error")
	}
	return &orgpb.ListOrgResponse{}, nil
}

func (m orgMock) ListPublicOrg(ctx context.Context, request *orgpb.ListOrgRequest) (*orgpb.ListOrgResponse, error) {
	if request.Org == "" {
		return nil, fmt.Errorf("error")
	}
	return &orgpb.ListOrgResponse{}, nil
}

func TestEndpoints_listOrg(t *testing.T) {
	type fields struct {
		orgClient org.ClientInterface
	}
	type args struct {
		ctx    context.Context
		userID string
		req    *orgpb.ListOrgRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *orgpb.ListOrgResponse
		wantErr bool
	}{
		{
			name:   "test with no userID",
			fields: fields{orgClient: orgMock{}},
			args: args{
				ctx:    context.Background(),
				userID: "",
				req:    nil,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name:   "test with userID",
			fields: fields{orgClient: orgMock{}},
			args: args{
				ctx:    context.Background(),
				userID: "1",
				req:    nil,
			},
			want:    &orgpb.ListOrgResponse{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &Endpoints{
				orgClient: tt.fields.orgClient,
			}
			got, err := e.listOrg(tt.args.ctx, tt.args.userID, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("listOrg() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("listOrg() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEndpoints_listPublicOrg(t *testing.T) {
	type fields struct {
		orgClient org.ClientInterface
	}
	type args struct {
		ctx    context.Context
		userID string
		req    *orgpb.ListOrgRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *orgpb.ListOrgResponse
		wantErr bool
	}{
		{
			name:   "test with err",
			fields: fields{orgClient: orgMock{}},
			args: args{
				ctx:    context.Background(),
				userID: "1",
				req:    &orgpb.ListOrgRequest{Org: ""},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name:   "test with no err",
			fields: fields{orgClient: orgMock{}},
			args: args{
				ctx:    context.Background(),
				userID: "1",
				req:    &orgpb.ListOrgRequest{Org: "erda"},
			},
			want:    &orgpb.ListOrgResponse{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &Endpoints{
				orgClient: tt.fields.orgClient,
			}
			got, err := e.listPublicOrg(tt.args.ctx, tt.args.userID, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("listPublicOrg() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("listPublicOrg() got = %v, want %v", got, tt.want)
			}
		})
	}
}
