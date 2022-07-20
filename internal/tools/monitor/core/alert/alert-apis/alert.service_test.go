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

package apis

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	orgpb "github.com/erda-project/erda-proto-go/core/org/pb"
	"github.com/erda-project/erda/internal/pkg/mock"
)

type orgMock struct {
	mock.OrgMock
}

func (m orgMock) GetOrg(ctx context.Context, request *orgpb.GetOrgRequest) (*orgpb.GetOrgResponse, error) {
	if request.IdOrName == "" {
		return nil, fmt.Errorf("the IdOrName is empty")
	}
	return &orgpb.GetOrgResponse{Data: &orgpb.Org{}}, nil
}

func Test_alertService_getOrg(t *testing.T) {
	type fields struct {
		p *provider
	}
	type args struct {
		orgIDOrName interface{}
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
				p: &provider{Org: orgMock{}},
			},
			args:    args{orgIDOrName: ""},
			want:    nil,
			wantErr: true,
		},
		{
			name: "test with error2",
			fields: fields{
				p: &provider{Org: orgMock{}},
			},
			args:    args{orgIDOrName: nil},
			want:    nil,
			wantErr: true,
		},
		{
			name: "test with no error",
			fields: fields{
				p: &provider{Org: orgMock{}},
			},
			args:    args{orgIDOrName: "1"},
			want:    &orgpb.Org{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &alertService{
				p: tt.fields.p,
			}
			got, err := m.GetOrg(tt.args.orgIDOrName)
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
