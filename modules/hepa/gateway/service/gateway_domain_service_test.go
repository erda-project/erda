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

package service

import (
	"fmt"
	"reflect"
	"sort"
	"testing"

	"github.com/erda-project/erda/modules/hepa/common"
	"github.com/erda-project/erda/modules/hepa/gateway/dto"
	"github.com/erda-project/erda/modules/hepa/repository/orm"
)

func Test_sortPackage(t *testing.T) {
	list := []dto.PackageInfoDto{
		{
			PackageDto: dto.PackageDto{
				Scene: "openapi",
			},
			CreateAt: "1",
		},
		{
			PackageDto: dto.PackageDto{
				Scene: "unity",
			},
			CreateAt: "2",
		},
	}
	sortList := dto.SortBySceneList(list)
	sort.Sort(sortList)
	fmt.Printf("sort list:%+v\n", common.NewPages(sortList, 2))
}

func Test_diffDomains(t *testing.T) {
	type args struct {
		reqDomains   []dto.EndpointDomainDto
		existDomains []orm.GatewayDomain
	}
	tests := []struct {
		name     string
		args     args
		wantAdds []dto.EndpointDomainDto
		wantDels []orm.GatewayDomain
	}{
		{
			name: "case1",
			args: args{
				reqDomains: []dto.EndpointDomainDto{
					{Domain: "1.com"},
					{Domain: "2.com"},
				},
				existDomains: []orm.GatewayDomain{
					{Domain: "1.com"},
					{Domain: "3.com"},
				},
			},
			wantAdds: []dto.EndpointDomainDto{
				{Domain: "2.com"},
			},
			wantDels: []orm.GatewayDomain{{Domain: "3.com"}},
		},
		{
			name: "case2",
			args: args{
				reqDomains: []dto.EndpointDomainDto{
					{Domain: "1.com"},
					{Domain: "2.com"},
				},
				existDomains: []orm.GatewayDomain{},
			},
			wantAdds: []dto.EndpointDomainDto{
				{Domain: "1.com"},
				{Domain: "2.com"},
			},
			wantDels: []orm.GatewayDomain{},
		},
		{
			name: "case3",
			args: args{
				reqDomains: []dto.EndpointDomainDto{},
				existDomains: []orm.GatewayDomain{
					{Domain: "1.com"},
					{Domain: "3.com"},
				},
			},
			wantAdds: nil,
			wantDels: []orm.GatewayDomain{
				{Domain: "1.com"},
				{Domain: "3.com"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotAdds, gotDels, _ := diffDomains(tt.args.reqDomains, tt.args.existDomains)
			if !reflect.DeepEqual(gotAdds, tt.wantAdds) {
				t.Errorf("diffDomains() gotAdds = %v, want %v", gotAdds, tt.wantAdds)
			}
			if !reflect.DeepEqual(gotDels, tt.wantDels) {
				t.Errorf("diffDomains() gotDels = %v, want %v", gotDels, tt.wantDels)
			}
		})
	}
}
