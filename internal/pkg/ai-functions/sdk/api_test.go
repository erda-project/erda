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

package sdk

import (
	"context"
	"reflect"
	"testing"

	"bou.ke/monkey"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
)

func TestGetUserInfo(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	userinfo := apistructs.UserInfo{
		ID:          "10000",
		Name:        "test",
		Nick:        "test",
		Avatar:      "",
		Phone:       "11111111111",
		Email:       "111222@333.com",
		Token:       "",
		LastLoginAt: "",
		PwdExpireAt: "",
		Source:      "",
	}
	tests := []struct {
		name    string
		args    args
		want    apistructs.UserInfo
		wantErr bool
	}{
		{
			name: "Test",
			args: args{
				ctx: context.WithValue(context.Background(), "user-id", "10000"),
			},
			want:    userinfo,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bdl := bundle.New(bundle.WithErdaServer())
			monkey.PatchInstanceMethod(reflect.TypeOf(bdl), "ListUsers", func(_ *bundle.Bundle,
				req apistructs.UserListRequest) (*apistructs.UserListResponseData, error) {
				return &apistructs.UserListResponseData{
					Users: []apistructs.UserInfo{userinfo},
				}, nil
			})

			got, err := GetUserInfo(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetUserInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetUserInfo() got = %v, want %v", got, tt.want)
			}
		})
	}
}
