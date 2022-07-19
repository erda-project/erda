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
	reflect "reflect"
	"testing"

	"bou.ke/monkey"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
)

func TestUser_GetOrgInfo(t *testing.T) {
	var bdl *bundle.Bundle
	monkey.PatchInstanceMethod(reflect.TypeOf(bdl), "GetOrg", func(b *bundle.Bundle, idOrName interface{}) (*apistructs.OrgDTO, error) {
		return &apistructs.OrgDTO{
			ID: 1,
		}, nil
	})

	monkey.PatchInstanceMethod(reflect.TypeOf(bdl), "GetDopOrgByDomain", func(b *bundle.Bundle, domain string, userID string) (*apistructs.OrgDTO, error) {
		return &apistructs.OrgDTO{
			ID: 1,
		}, nil
	})

	defer monkey.UnpatchAll()

	type args struct {
		orgHeader    string
		domainHeader string
		host         string
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		want1   uint64
		wantErr bool
	}{
		{
			args: args{
				orgHeader: "erda",
			},
			want:  false,
			want1: 1,
		},
		{
			args: args{
				orgHeader:    "-",
				domainHeader: "erda.cloud",
			},
			want:  false,
			want1: 1,
		},
		{
			args: args{
				orgHeader: "-",
				host:      "erda.cloud",
			},
			want:  false,
			want1: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &User{}
			got, got1, err := u.GetOrgInfo(tt.args.orgHeader, tt.args.domainHeader, tt.args.host)
			if (err != nil) != tt.wantErr {
				t.Errorf("User.GetOrgInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("User.GetOrgInfo() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("User.GetOrgInfo() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
