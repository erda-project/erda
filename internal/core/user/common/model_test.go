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

package common

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda-proto-go/core/user/pb"
	"github.com/erda-project/erda/apistructs"
)

func TestToPbUser(t *testing.T) {
	type args struct {
		user User
	}
	tests := []struct {
		name string
		args args
		want *pb.User
	}{
		{
			args: args{
				user: User{
					ID: "1",
				},
			},
			want: &pb.User{
				ID: "1",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ToPbUser(tt.args.user); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ToPbUser() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewUserInfoFromDTO(t *testing.T) {
	type args struct {
		dto *apistructs.UserInfoDto
	}
	tests := []struct {
		name string
		args args
		want *UserInfo
	}{
		{
			name: "nil",
			args: args{
				dto: nil,
			},
			want: nil,
		},
		{
			name: "normal",
			args: args{
				dto: &apistructs.UserInfoDto{
					AvatarURL: "imageURL",
					Email:     "a@b.com",
					UserID:    "1234",
					NickName:  "nickname",
					Phone:     "123456789",
					RealName:  "realname",
					Username:  "username",
				},
			},
			want: &UserInfo{
				ID:         USERID("1234"),
				Email:      "a@b.com",
				EmailExist: true,
				Phone:      "123456789",
				PhoneExist: true,
				AvatarUrl:  "imageURL",
				UserName:   "username",
				NickName:   "nickname",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, NewUserInfoFromDTO(tt.args.dto), "NewUserInfoFromDTO(%v)", tt.args.dto)
		})
	}
}
