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

package credential

import (
	"context"
	"fmt"
	http1 "net/http"
	"testing"
	"time"

	"bou.ke/monkey"
	"github.com/golang/mock/gomock"
	"google.golang.org/protobuf/types/known/timestamppb"

	grpc1 "github.com/erda-project/erda-infra/pkg/transport/grpc"
	"github.com/erda-project/erda-infra/pkg/transport/http"
	"github.com/erda-project/erda-infra/pkg/transport/http/encoding"
	tokenpb "github.com/erda-project/erda-proto-go/core/token/pb"
	"github.com/erda-project/erda-proto-go/msp/credential/pb"
	tenant "github.com/erda-project/erda-proto-go/msp/tenant/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/pkg/common/apis"
)

////go:generate mockgen -destination=./credential_register_test.go -package exporter github.com/erda-project/erda-infra/pkg/transport Register
////go:generate mockgen -destination=./credential_ak_test.go -package exporter github.com/erda-project/erda-proto-go/core/services/authentication/credentials/accesskey/pb AccessKeyServiceServer
////go:generate mockgen -destination=./credential_context_test.go -package exporter github.com/erda-project/erda-infra/base/servicehub Context
////go:generate mockgen -destination=./tenant_test.go -package exporter github.com/erda-project/erda-proto-go/msp/tenant/pb TenantServiceServer

func ownerContext() context.Context {
	return apis.WithUserIDContext(context.Background(), "user-1")
}

func patchProjectOwnerRole() {
	monkey.Patch((*bundle.Bundle).ScopeRoleAccess, func(_ *bundle.Bundle, _ string, _ *apistructs.ScopeRoleAccessRequest) (*apistructs.ScopeRole, error) {
		return &apistructs.ScopeRole{
			Access: true,
			Roles:  []string{bundle.RoleProjectOwner},
		}, nil
	})
}

func Test_accessKeyService_QueryAccessKeys(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	akService := NewMockTokenServiceServer(ctrl)
	tenantService := NewMockTenantServiceServer(ctrl)

	akService.EXPECT().CreateToken(gomock.Any(), gomock.Any()).AnyTimes().Return(&tokenpb.CreateTokenResponse{
		Data: &tokenpb.Token{},
	}, nil)
	tenantService.EXPECT().GetTenantProject(gomock.Any(), gomock.Any()).AnyTimes().Return(&tenant.GetTenantProjectResponse{
		Data: &tenant.TenantProjectData{
			Workspace: "PROD",
			ProjectId: "98",
		},
	}, nil)
	defer monkey.UnpatchAll()
	patchProjectOwnerRole()
	monkey.Patch((*bundle.Bundle).GetProject, func(bdl *bundle.Bundle, id uint64) (*apistructs.ProjectDTO, error) {
		return &apistructs.ProjectDTO{
			ID:                   89,
			Name:                 "ss",
			DisplayName:          "ss",
			DDHook:               "ss",
			OrgID:                21,
			Creator:              "ss",
			Logo:                 "ss",
			Desc:                 "ss",
			Owners:               nil,
			ActiveTime:           "ss",
			Joined:               false,
			CanUnblock:           nil,
			BlockStatus:          "ss",
			CanManage:            false,
			IsPublic:             false,
			Stats:                apistructs.ProjectStats{},
			ProjectResourceUsage: apistructs.ProjectResourceUsage{},
			ClusterConfig:        nil,
			ResourceConfig:       nil,
			RollbackConfig:       nil,
			CpuQuota:             0,
			MemQuota:             0,
			CreatedAt:            time.Time{},
			UpdatedAt:            time.Time{},
			Type:                 "",
		}, nil
	})

	pro := &provider{
		Cfg:                  &config{},
		Register:             NewMockRegister(ctrl),
		credentialKeyService: &accessKeyService{},
		TokenService:         akService,
		bdl:                  &bundle.Bundle{},
		audit:                nil,
		Tenant:               tenantService,
	}
	pro.credentialKeyService.p = pro
	_, err := pro.credentialKeyService.CreateAccessKey(ownerContext(), &pb.CreateAccessKeyRequest{
		SubjectType: pb.SubjectTypeEnum_SYSTEM,
		Subject:     "xddd",
		Description: "cdddd",
		ScopeId:     "tenant-1",
	})
	if err != nil {
		fmt.Println("should not err")
	}
}

func Test_accessKeyService_DeleteAccessKey(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	tenantService := NewMockTenantServiceServer(ctrl)
	akService := NewMockTokenServiceServer(ctrl)
	monkey.Patch((*bundle.Bundle).GetProject, func(bdl *bundle.Bundle, id uint64) (*apistructs.ProjectDTO, error) {
		return &apistructs.ProjectDTO{
			ID:                   89,
			Name:                 "ss",
			DisplayName:          "ss",
			DDHook:               "ss",
			OrgID:                21,
			Creator:              "ss",
			Logo:                 "ss",
			Desc:                 "ss",
			Owners:               nil,
			ActiveTime:           "ss",
			Joined:               false,
			CanUnblock:           nil,
			BlockStatus:          "ss",
			CanManage:            false,
			IsPublic:             false,
			Stats:                apistructs.ProjectStats{},
			ProjectResourceUsage: apistructs.ProjectResourceUsage{},
			ClusterConfig:        nil,
			ResourceConfig:       nil,
			RollbackConfig:       nil,
			CpuQuota:             0,
			MemQuota:             0,
			CreatedAt:            time.Time{},
			UpdatedAt:            time.Time{},
			Type:                 "",
		}, nil
	})
	defer monkey.UnpatchAll()
	patchProjectOwnerRole()
	akService.EXPECT().GetToken(gomock.Any(), gomock.Any()).AnyTimes().Return(&tokenpb.GetTokenResponse{
		Data: &tokenpb.Token{
			Id:          "2",
			AccessKey:   "sfdfgfg",
			SecretKey:   "sdgds",
			Description: "ss",
			CreatedAt:   nil,
			Scope:       "ss",
			ScopeId:     "dfdfd",
			CreatorId:   "2",
		},
	}, nil)
	tenantService.EXPECT().GetTenantProject(gomock.Any(), gomock.Any()).AnyTimes().Return(&tenant.GetTenantProjectResponse{
		Data: &tenant.TenantProjectData{
			Workspace: "PROD",
			ProjectId: "98",
		},
	}, nil)
	akService.EXPECT().DeleteToken(gomock.Any(), gomock.Any()).AnyTimes().Return(nil, nil)
	pro := &provider{
		Cfg:                  &config{},
		Register:             NewMockRegister(ctrl),
		credentialKeyService: &accessKeyService{},
		TokenService:         akService,
		bdl:                  &bundle.Bundle{},
		audit:                nil,
		Tenant:               tenantService,
	}
	pro.credentialKeyService.p = pro
	_, err := pro.credentialKeyService.DeleteAccessKey(ownerContext(), &pb.DeleteAccessKeyRequest{
		Id: "13eef468-1d0b-42ce-aa7b-b499545bf92d",
	})
	if err != nil {
		fmt.Println("should not err")
	}
}

func Test_accessKeyService_QueryAccessKeys1(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	akService := NewMockTokenServiceServer(ctrl)
	tenantService := NewMockTenantServiceServer(ctrl)
	akService.EXPECT().QueryTokens(gomock.Any(), gomock.Any()).AnyTimes().Return(&tokenpb.QueryTokensResponse{
		Data: make([]*tokenpb.Token, 0),
	}, nil)
	tenantService.EXPECT().GetTenantProject(gomock.Any(), gomock.Any()).AnyTimes().Return(&tenant.GetTenantProjectResponse{
		Data: &tenant.TenantProjectData{ProjectId: "98"},
	}, nil)
	defer monkey.UnpatchAll()
	patchProjectOwnerRole()
	pro := &provider{
		Cfg:                  &config{},
		Register:             NewMockRegister(ctrl),
		credentialKeyService: &accessKeyService{},
		TokenService:         akService,
		bdl:                  &bundle.Bundle{},
		Tenant:               tenantService,
	}
	pro.credentialKeyService.p = pro
	_, err := pro.credentialKeyService.QueryAccessKeys(ownerContext(), &pb.QueryAccessKeysRequest{
		SubjectType: 3,
		Subject:     "22",
		PageSize:    1,
		PageNo:      3,
		ScopeId:     "tenant-1",
	})
	if err != nil {
		fmt.Println("should not err")
	}
}

func Test_accessKeyService_GetAccessKey(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	akService := NewMockTokenServiceServer(ctrl)
	tenantService := NewMockTenantServiceServer(ctrl)
	akService.EXPECT().GetToken(gomock.Any(), gomock.Any()).AnyTimes().Return(&tokenpb.GetTokenResponse{
		Data: &tokenpb.Token{},
	}, nil)
	tenantService.EXPECT().GetTenantProject(gomock.Any(), gomock.Any()).AnyTimes().Return(&tenant.GetTenantProjectResponse{
		Data: &tenant.TenantProjectData{ProjectId: "98"},
	}, nil)
	defer monkey.UnpatchAll()
	patchProjectOwnerRole()
	pro := &provider{
		Cfg:                  &config{},
		Register:             NewMockRegister(ctrl),
		credentialKeyService: &accessKeyService{},
		TokenService:         akService,
		bdl:                  &bundle.Bundle{},
		Tenant:               tenantService,
	}
	pro.credentialKeyService.p = pro
	_, err := pro.credentialKeyService.GetAccessKey(ownerContext(), &pb.GetAccessKeyRequest{
		Id: "13eef468-1d0b-42ce-aa7b-b499545bf92d",
	})
	if err != nil {
		fmt.Println("should not err")
	}
}

func Test_accessKeyService_DownloadAccessKeyFile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	akService := NewMockTokenServiceServer(ctrl)
	tenantService := NewMockTenantServiceServer(ctrl)
	akService.EXPECT().GetToken(gomock.Any(), gomock.Any()).AnyTimes().Return(&tokenpb.GetTokenResponse{
		Data: &tokenpb.Token{
			Id:          "ssss",
			AccessKey:   "dddd",
			SecretKey:   "dddd",
			Description: "aaa",
			CreatedAt:   &timestamppb.Timestamp{},
		},
	}, nil)
	tenantService.EXPECT().GetTenantProject(gomock.Any(), gomock.Any()).AnyTimes().Return(&tenant.GetTenantProjectResponse{
		Data: &tenant.TenantProjectData{ProjectId: "98"},
	}, nil)
	defer monkey.UnpatchAll()
	patchProjectOwnerRole()
	pro := &provider{
		Cfg:                  &config{},
		Register:             NewMockRegister(ctrl),
		credentialKeyService: &accessKeyService{},
		TokenService:         akService,
		bdl:                  &bundle.Bundle{},
		Tenant:               tenantService,
	}
	pro.credentialKeyService.p = pro
	_, err := pro.credentialKeyService.DownloadAccessKeyFile(ownerContext(), &pb.DownloadAccessKeyFileRequest{
		Id: "ssss",
	})
	if err != nil {
		fmt.Println("should not err")
	}
}

func Test_Init(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	akService := NewMockTokenServiceServer(ctrl)
	tenantService := NewMockTenantServiceServer(ctrl)
	pro := &provider{
		Cfg:                  &config{},
		Register:             NewMockRegister(ctrl),
		credentialKeyService: &accessKeyService{},
		TokenService:         akService,
		bdl:                  &bundle.Bundle{},
		audit:                nil,
		Tenant:               tenantService,
	}
	pro.credentialKeyService.p = pro
	defer monkey.UnpatchAll()
	patchProjectOwnerRole()
	monkey.Patch(encoding.EncodeResponse, func(w http1.ResponseWriter, r *http1.Request, out interface{}) error {
		return nil
	})
	monkey.Patch(pb.RegisterAccessKeyServiceHandler, func(r http.Router, srv pb.AccessKeyServiceHandler, opts ...http.HandleOption) {})
	monkey.Patch(pb.RegisterAccessKeyServiceServer, func(s grpc1.ServiceRegistrar, srv pb.AccessKeyServiceServer, opts ...grpc1.HandleOption) {})
	akService.EXPECT().GetToken(gomock.Any(), gomock.Any()).AnyTimes().Return(&tokenpb.GetTokenResponse{
		Data: &tokenpb.Token{
			Id:          "ssss",
			AccessKey:   "dddd",
			SecretKey:   "dddd",
			Description: "aaa",
			CreatedAt:   &timestamppb.Timestamp{},
		},
	}, nil)
	tenantService.EXPECT().GetTenantProject(gomock.Any(), gomock.Any()).AnyTimes().Return(&tenant.GetTenantProjectResponse{
		Data: &tenant.TenantProjectData{ProjectId: "98"},
	}, nil)
	pro.credentialKeyService.p = pro
	_, err := pro.credentialKeyService.DownloadAccessKeyFile(ownerContext(), &pb.DownloadAccessKeyFileRequest{
		Id: "ssss",
	})
	if err != nil {
		fmt.Println("should not err2")
	}
}

func Test_checkProjectOwnerOrLeadDeniesNonOwner(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	tenantService := NewMockTenantServiceServer(ctrl)
	tenantService.EXPECT().GetTenantProject(gomock.Any(), gomock.Any()).AnyTimes().Return(&tenant.GetTenantProjectResponse{
		Data: &tenant.TenantProjectData{ProjectId: "98"},
	}, nil)
	pro := &provider{
		Cfg:    &config{},
		bdl:    &bundle.Bundle{},
		Tenant: tenantService,
	}
	pro.credentialKeyService = &accessKeyService{p: pro}
	defer monkey.UnpatchAll()
	monkey.Patch((*bundle.Bundle).ScopeRoleAccess, func(_ *bundle.Bundle, _ string, _ *apistructs.ScopeRoleAccessRequest) (*apistructs.ScopeRole, error) {
		return &apistructs.ScopeRole{
			Access: true,
			Roles:  []string{"Dev"},
		}, nil
	})

	if _, err := pro.credentialKeyService.checkProjectOwnerOrLead(ownerContext(), "tenant-1"); err == nil {
		t.Errorf("checkProjectOwnerOrLead() should reject a project member without Owner/Lead role")
	}
}
