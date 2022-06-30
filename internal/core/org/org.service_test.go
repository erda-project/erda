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

package org

import (
	"context"
	"reflect"
	"testing"

	"bou.ke/monkey"

	"github.com/erda-project/erda-proto-go/core/org/pb"
	"github.com/erda-project/erda/internal/core/legacy/services/member"
	"github.com/erda-project/erda/pkg/common/apis"
)

func Test_provider_getOrgPermissions(t *testing.T) {
	type fields struct {
		member *member.Member
	}
	type args struct {
		ctx context.Context
		req *pb.ListOrgRequest
	}

	var me = &member.Member{}
	patch1 := monkey.PatchInstanceMethod(reflect.TypeOf(me), "IsAdmin", func(mem *member.Member, userID string) bool {
		return true
	})
	defer patch1.Unpatch()

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    bool
		want1   []int64
		wantErr bool
	}{
		{
			name: "test with getOrgPermissions",
			fields: fields{
				member: me,
			},
			args: args{
				ctx: apis.WithUserIDContext(context.Background(), "1"),
				req: &pb.ListOrgRequest{
					Q:        "",
					Key:      "",
					PageNo:   0,
					PageSize: 0,
					Org:      "",
				},
			},
			want:    true,
			want1:   nil,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &provider{
				member: tt.fields.member,
			}

			got, got1, err := p.getOrgPermissions(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("getOrgPermissions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("getOrgPermissions() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("getOrgPermissions() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
