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

package userinfo

import (
	"net/http"
	"strings"
	"testing"

	"github.com/alecthomas/assert"
	gomock "github.com/golang/mock/gomock"

	apistructs "github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/http/httputil"
)

func Test_provider_userInfoRetriever(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	identitySvc := NewMockInterface(ctrl)
	identitySvc.EXPECT().GetUsers(gomock.Any(), gomock.Any()).AnyTimes().Return(map[string]apistructs.UserInfo{"1": {ID: "1"}}, nil)

	p := &provider{Identity: identitySvc}
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
	expected := []string{`"id":"1"`, `"id":"1"`}
	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := p.userInfoRetriever(tt.args.r, tt.args.data, tt.args.userIDs)
			assert.NotNil(t, body)
			assert.Equal(t, true, strings.Contains(string(body), expected[i]))
		})
	}
}
