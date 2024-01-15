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

package impl

import (
	"fmt"
	"reflect"
	"sort"
	"testing"

	"bou.ke/monkey"

	"github.com/erda-project/erda/internal/tools/orchestrator/hepa/common"
	"github.com/erda-project/erda/internal/tools/orchestrator/hepa/gateway/dto"
	"github.com/erda-project/erda/internal/tools/orchestrator/hepa/repository/orm"
	"github.com/erda-project/erda/internal/tools/orchestrator/hepa/repository/service"
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

func TestRuntimeData_checkValid(t *testing.T) {
	type fields struct {
		ReleaseId             string
		ServiceGroupNamespace string
		ServiceGroupName      string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{"case1", fields{ReleaseId: "qwer", ServiceGroupNamespace: "asdf", ServiceGroupName: "zxcv"}, false},
		{"case2", fields{ReleaseId: "", ServiceGroupNamespace: "asdf", ServiceGroupName: "zxcv"}, true},
		{"case3", fields{ReleaseId: "qwer", ServiceGroupNamespace: "", ServiceGroupName: "zxcv"}, true},
		{"case4", fields{ReleaseId: "qwer", ServiceGroupNamespace: "asdf", ServiceGroupName: ""}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := RuntimeData{
				ReleaseId:             tt.fields.ReleaseId,
				ServiceGroupNamespace: tt.fields.ServiceGroupNamespace,
				ServiceGroupName:      tt.fields.ServiceGroupName,
			}
			if err := data.checkValid(); (err != nil) != tt.wantErr {
				t.Errorf("checkValid() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGatewayDomainServiceImpl(t *testing.T) {
	impl := &GatewayDomainServiceImpl{
		runtimeDb:    &service.GatewayRuntimeServiceServiceImpl{},
		azDb:         &service.GatewayAzInfoServiceImpl{},
		kongDb:       &service.GatewayKongInfoServiceImpl{},
		packageAPIDB: &service.GatewayPackageApiServiceImpl{},
		domainDb:     &service.GatewayDomainServiceImpl{},
	}
	monkey.PatchInstanceMethod(reflect.TypeOf(impl.runtimeDb), "SelectByAny", func(*service.GatewayRuntimeServiceServiceImpl, *orm.GatewayRuntimeService) ([]orm.GatewayRuntimeService, error) {
		return []orm.GatewayRuntimeService{
			{
				ProjectId:    "8888",
				InnerAddress: "test.test",
				RuntimeId:    "22222",
				BaseRow: orm.BaseRow{
					Id: "22222",
				},
			},
		}, nil
	})

	monkey.PatchInstanceMethod(reflect.TypeOf(impl.azDb), "GetAzInfoByClusterName", func(svc *service.GatewayAzInfoServiceImpl, name string) (*orm.GatewayAzInfo, *service.ClusterInfoDto, error) {
		return &orm.GatewayAzInfo{
			Az:        "jicheng",
			OrgId:     "632",
			ProjectId: "8888",
		}, nil, nil
	})

	monkey.PatchInstanceMethod(reflect.TypeOf(impl.kongDb), "GetKongInfo", func(*service.GatewayKongInfoServiceImpl, *orm.GatewayKongInfo) (*orm.GatewayKongInfo, error) {
		return nil, nil
	})

	monkey.PatchInstanceMethod(reflect.TypeOf(impl.packageAPIDB), "SelectByOptions", func(svc *service.GatewayPackageApiServiceImpl, options []orm.SelectOption) ([]orm.GatewayPackageApi, error) {
		return []orm.GatewayPackageApi{
			{
				PackageId:        "1111",
				Method:           "GET",
				RedirectType:     "service",
				RedirectAddr:     "test.test",
				RuntimeServiceId: "22222",
			},
		}, nil
	})

	monkey.PatchInstanceMethod(reflect.TypeOf(*impl), "doesClusterSupportHttps", func(string) bool { return false })

	monkey.PatchInstanceMethod(reflect.TypeOf(impl.domainDb), "SelectByOptions", func(svc *service.GatewayDomainServiceImpl, options []orm.SelectOption) ([]orm.GatewayDomain, error) {
		return nil, nil
	})

	monkey.PatchInstanceMethod(reflect.TypeOf(impl.domainDb), "SelectByAny", func(*service.GatewayDomainServiceImpl, *orm.GatewayDomain) ([]orm.GatewayDomain, error) {
		return nil, nil
	})

	result, err := impl.GetRuntimeDomains("19986", 632)
	if err != nil {
		t.Error(err)
	}
	t.Log(result)
}
