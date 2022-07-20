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

package auth

import (
	"context"
	"testing"

	orgpb "github.com/erda-project/erda-proto-go/core/org/pb"
	"github.com/erda-project/erda/internal/pkg/mock"
)

type orgMock struct {
	mock.OrgMock
}

func (m orgMock) GetOrg(ctx context.Context, request *orgpb.GetOrgRequest) (*orgpb.GetOrgResponse, error) {
	return &orgpb.GetOrgResponse{Data: &orgpb.Org{ID: 1}},nil
}

func TestUser_GetOrgInfo(t *testing.T) {
	type args struct {
		orgHeader    string
		domainHeader string
	}
	tests := []struct {
		name    string
		args    args
		want    uint64
		wantErr bool
	}{
		{
			args: args{
				orgHeader: "erda",
			},
			want: 1,
		},
		{
			args: args{
				orgHeader:    "-",
				domainHeader: "erda.cloud",
			},
			want: 0,
		},
		{
			args: args{
				orgHeader:    "",
				domainHeader: "erda.cloud",
			},
			want: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &User{
				org: orgMock{},
			}
			got, err := u.GetOrgInfo(tt.args.orgHeader, tt.args.domainHeader)
			if (err != nil) != tt.wantErr {
				t.Errorf("User.GetOrgInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("User.GetOrgInfo() got = %v, want %v", got, tt.want)
			}
		})
	}
}
