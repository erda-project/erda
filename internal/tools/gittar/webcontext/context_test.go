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

package webcontext

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	orgpb "github.com/erda-project/erda-proto-go/core/org/pb"
	"github.com/erda-project/erda/internal/core/org"
	"github.com/erda-project/erda/internal/pkg/mock"
)

type orgMock struct {
	mock.OrgMock
}

func (m orgMock) GetOrg(ctx context.Context, request *orgpb.GetOrgRequest) (*orgpb.GetOrgResponse, error) {
	if request.IdOrName == "" {
		return nil, fmt.Errorf("error")
	}
	return &orgpb.GetOrgResponse{Data: &orgpb.Org{}}, nil
}

func (m orgMock) GetOrgByDomain(ctx context.Context, request *orgpb.GetOrgByDomainRequest) (*orgpb.GetOrgByDomainResponse, error) {
	if request.Domain == "" {
		return nil, fmt.Errorf("error")
	}
	return &orgpb.GetOrgByDomainResponse{Data: &orgpb.Org{}}, nil
}

func TestContext_GetOrg(t *testing.T) {
	type fields struct {
		orgClient org.ClientInterface
	}
	type args struct {
		orgID interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *orgpb.Org
		wantErr bool
	}{
		{
			name: "test with error1",
			fields: fields{
				orgClient: orgMock{},
			},
			args: args{
				orgID: nil,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "test with error2",
			fields: fields{
				orgClient: orgMock{},
			},
			args: args{
				orgID: "",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "test with no error",
			fields: fields{
				orgClient: orgMock{},
			},
			args: args{
				orgID: 1,
			},
			want:    &orgpb.Org{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Context{
				orgClient: tt.fields.orgClient,
			}
			got, err := c.GetOrg(tt.args.orgID)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetOrg() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetOrg() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestContext_GetOrgByDomain(t *testing.T) {
	type fields struct {
		orgClient org.ClientInterface
	}
	type args struct {
		domain string
		userID string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *orgpb.Org
		wantErr bool
	}{
		{
			name: "test with error",
			fields: fields{
				orgClient: orgMock{},
			},
			args: args{
				domain: "",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "test with no error",
			fields: fields{
				orgClient: orgMock{},
			},
			args: args{
				domain: "erda",
			},
			want:    &orgpb.Org{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Context{
				orgClient: tt.fields.orgClient,
			}
			got, err := c.GetOrgByDomain(tt.args.domain, tt.args.userID)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetOrgByDomain() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetOrgByDomain() got = %v, want %v", got, tt.want)
			}
		})
	}
}
