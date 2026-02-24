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

//go:generate mockgen -destination=./mock_usersvc_test.go -package userinfo github.com/erda-project/erda-proto-go/core/user/pb UserServiceServer

package userinfo

import (
	"net/http"
	"strings"
	"testing"

	"github.com/alecthomas/assert"
	"github.com/golang/mock/gomock"

	commonpb "github.com/erda-project/erda-proto-go/common/pb"
	userpb "github.com/erda-project/erda-proto-go/core/user/pb"
	"github.com/erda-project/erda/pkg/http/httputil"
)

func Test_provider_userInfoRetriever(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	userSvc := NewMockUserServiceServer(ctrl)
	userSvc.EXPECT().FindUsers(gomock.Any(), gomock.Any()).AnyTimes().Return(&userpb.FindUsersResponse{Data: []*commonpb.UserInfo{{Id: "1"}}}, nil)

	p := &provider{UserSvc: userSvc}
	type args struct {
		r       *http.Request
		data    map[string]interface{}
		userIDs []string
		body    *[]byte
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "test",
			args: args{
				r: &http.Request{
					Header: http.Header{
						httputil.UserInfoDesensitizedHeader: []string{"true"},
					},
				},
				data:    map[string]interface{}{},
				userIDs: []string{"1"},
			},
		},
		{
			name: "test",
			args: args{
				r: &http.Request{
					Header: http.Header{
						httputil.UserInfoDesensitizedHeader: []string{"false"},
					},
				},
				data:    map[string]interface{}{},
				userIDs: []string{"1"},
			},
		},
	}
	expected := []string{`"id":""`, `"id":"1"`}
	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := p.userInfoRetriever(tt.args.r, tt.args.data, tt.args.userIDs)
			assert.NotNil(t, body)
			assert.Equal(t, true, strings.Contains(string(body), expected[i]))
		})
	}
}
