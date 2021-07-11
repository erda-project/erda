// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package uc

import (
	"testing"
)

func TestLoginAuther_getAuthorizeURL(t *testing.T) {
	type args struct {
		callback string
		referer  string
	}
	tests := []struct {
		name string
		auth *LoginAuther
		args args
		want string
	}{
		{
			auth: &LoginAuther{
				clientID:    "test-client-id",
				ucPublicURL: "http://uc",
			},
			args: args{
				callback: "http://openapi",
				referer:  "http://erda.cloud",
			},
			want: "http://uc/oauth/authorize?client_id=test-client-id&redirect_uri=http%3A%2F%2Fopenapi%2Flogincb%3Freferer%3Dhttp%253A%252F%252Ferda.cloud&response_type=code&scope=public_profile",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.auth.getAuthorizeURL(tt.args.callback, tt.args.referer); got != tt.want {
				t.Errorf("LoginAuther.getAuthorizeURL() = %v, want %v", got, tt.want)
			}
		})
	}
}
