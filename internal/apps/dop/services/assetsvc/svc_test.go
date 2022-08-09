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

package assetsvc

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	orgpb "github.com/erda-project/erda-proto-go/core/org/pb"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/internal/core/org"
	"github.com/erda-project/erda/internal/pkg/mock"
)

func TestNew(t *testing.T) {
	var (
		bdl = bundle.New()
	)
	New(
		WithBundle(bdl),
	)
}

type assetOrgMock struct {
	mock.OrgMock
}

func (m assetOrgMock) GetOrg(ctx context.Context, request *orgpb.GetOrgRequest) (*orgpb.GetOrgResponse, error) {
	if request.IdOrName == "1" {
		return nil, fmt.Errorf("error")
	}
	return &orgpb.GetOrgResponse{Data: &orgpb.Org{Name: "erda"}}, nil
}

func TestService_getOrg(t *testing.T) {
	type fields struct {
		org org.Interface
	}
	type args struct {
		ctx   context.Context
		orgID uint64
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
				org: assetOrgMock{},
			},
			args:    args{orgID: 1, ctx: context.Background()},
			want:    nil,
			wantErr: true,
		},
		{
			name: "test with no error",
			fields: fields{
				org: assetOrgMock{},
			},
			args:    args{orgID: 2, ctx: context.Background()},
			want:    &orgpb.Org{Name: "erda"},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			svc := &Service{
				org: tt.fields.org,
			}
			got, err := svc.getOrg(tt.args.ctx, tt.args.orgID)
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
