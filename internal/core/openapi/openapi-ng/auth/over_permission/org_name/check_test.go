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

package org_name

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/erda-project/erda-proto-go/common/pb"
	"github.com/erda-project/erda/internal/core/openapi/openapi-ng/auth/over_permission/match"
)

type mockOpts struct {
	opts map[string]interface{}
}

func (m *mockOpts) Get(key string) interface{} {
	if val, ok := m.opts[key]; ok {
		return val
	}
	return nil
}

func (m *mockOpts) Set(key string, val interface{}) {
	if m.opts == nil {
		m.opts = make(map[string]interface{})
	}
	m.opts[key] = val
}

func TestWeightShouldByConfig(t *testing.T) {
	service := newOverPermissionOrgName(&provider{
		Cfg: &config{
			Weight: int64(1000),
		},
	})
	require.Equal(t, int64(1000), service.Weight())
}

func TestMatch(t *testing.T) {

	service := newOverPermissionOrgName(&provider{
		Cfg: &config{
			DefaultMatchOrg: []string{
				"query:scope", "query:scopeId",
			},
		},
	})
	tests := []struct {
		name   string
		opts   *pb.CheckOverPermission
		url    string
		isAim  bool
		result interface{}
	}{
		{
			name: "no_expr",
			url:  "localhost:9529/api/dashboard/blocks?scope=org&scopeId=erda-development222&pageSize=10&pageNo=1&createdAt=",
			opts: &pb.CheckOverPermission{
				OrgName: &pb.CheckOver{
					Enable: true,
				},
			},
			isAim:  true,
			result: map[string]interface{}{"scope": "org", "scopeId": "erda-development222"},
		},
		{
			name: "expr",
			url:  "localhost:9529/api/dashboard/blocks?ttt=123",
			opts: &pb.CheckOverPermission{
				OrgName: &pb.CheckOver{
					Enable: true,
					Expr:   []string{"query:ttt"},
				},
			},
			isAim:  true,
			result: map[string]interface{}{"ttt": "123"},
		},
		{
			name:   "no_aim_nil",
			opts:   nil,
			isAim:  false,
			result: nil,
		},
		{
			name: "no_aim_default_org",
			opts: &pb.CheckOverPermission{
				OrgName: &pb.CheckOver{},
			},
			isAim:  false,
			result: nil,
		},
		{
			name: "no_aim_default_no_enable",
			opts: &pb.CheckOverPermission{
				OrgName: &pb.CheckOver{
					Enable: false,
				},
			},
			isAim:  false,
			result: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			opts := mockOpts{}
			opts.Set(match.ProtoComponent, test.opts)
			request, _ := http.NewRequest("GET", test.url, nil)
			aim, got := service.Match(request, &opts)
			require.Equal(t, test.isAim, aim)
			require.Equal(t, test.result, got)
		})
	}
}
