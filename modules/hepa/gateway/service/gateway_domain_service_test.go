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
