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

package endpoints

import (
	"context"
	"fmt"

	"github.com/erda-project/erda-proto-go/core/org/pb"
	userpb "github.com/erda-project/erda-proto-go/core/user/pb"
	"github.com/erda-project/erda/internal/core/legacy/services/member"
	"github.com/erda-project/erda/internal/core/legacy/services/permission"
	"github.com/erda-project/erda/pkg/common/apis"
)

type MockOrg struct{}

func (m MockOrg) CreateOrg(ctx context.Context, request *pb.CreateOrgRequest) (*pb.CreateOrgResponse, error) {
	panic("implement me")
}

func (m MockOrg) UpdateOrg(ctx context.Context, request *pb.UpdateOrgRequest) (*pb.UpdateOrgResponse, error) {
	panic("implement me")
}

func (m MockOrg) GetOrg(ctx context.Context, request *pb.GetOrgRequest) (*pb.GetOrgResponse, error) {
	info := apis.GetIdentityInfo(ctx)
	if info == nil {
		return nil, fmt.Errorf("authentication failed")
	}
	return &pb.GetOrgResponse{Data: &pb.Org{}}, nil
}

func (m MockOrg) DeleteOrg(ctx context.Context, request *pb.DeleteOrgRequest) (*pb.DeleteOrgResponse, error) {
	panic("implement me")
}

func (m MockOrg) ListOrg(ctx context.Context, request *pb.ListOrgRequest) (*pb.ListOrgResponse, error) {
	panic("implement me")
}

func (m MockOrg) ListPublicOrg(ctx context.Context, request *pb.ListOrgRequest) (*pb.ListOrgResponse, error) {
	panic("implement me")
}

func (m MockOrg) GetOrgByDomain(ctx context.Context, request *pb.GetOrgByDomainRequest) (*pb.GetOrgByDomainResponse, error) {
	panic("implement me")
}

func (m MockOrg) ChangeCurrentOrg(ctx context.Context, request *pb.ChangeCurrentOrgRequest) (*pb.ChangeCurrentOrgResponse, error) {
	panic("implement me")
}

func (m MockOrg) CreateOrgClusterRelation(ctx context.Context, request *pb.OrgClusterRelationCreateRequest) (*pb.OrgClusterRelationCreateResponse, error) {
	panic("implement me")
}

func (m MockOrg) ListOrgClusterRelation(ctx context.Context, request *pb.ListOrgClusterRelationRequest) (*pb.ListOrgClusterRelationResponse, error) {
	panic("implement me")
}

func (m MockOrg) SetReleaseCrossCluster(ctx context.Context, request *pb.SetReleaseCrossClusterRequest) (*pb.SetReleaseCrossClusterResponse, error) {
	panic("implement me")
}

func (m MockOrg) GenVerifyCode(ctx context.Context, request *pb.GenVerifyCodeRequest) (*pb.GenVerifyCodeResponse, error) {
	panic("implement me")
}

func (m MockOrg) SetNotifyConfig(ctx context.Context, request *pb.SetNotifyConfigRequest) (*pb.SetNotifyConfigResponse, error) {
	panic("implement me")
}

func (m MockOrg) GetNotifyConfig(ctx context.Context, request *pb.GetNotifyConfigRequest) (*pb.GetNotifyConfigResponse, error) {
	panic("implement me")
}

func (m MockOrg) GetOrgClusterRelationsByOrg(ctx context.Context, request *pb.GetOrgClusterRelationsByOrgRequest) (*pb.GetOrgClusterRelationsByOrgResponse, error) {
	panic("implement me")
}

func (m MockOrg) DereferenceCluster(ctx context.Context, request *pb.DereferenceClusterRequest) (*pb.DereferenceClusterResponse, error) {
	panic("implement me")
}

func (m MockOrg) WithUc(uc userpb.UserServiceServer) {
	panic("implement me")
}

func (m MockOrg) WithMember(member *member.Member) {
	panic("implement me")
}

func (m MockOrg) WithPermission(permission *permission.Permission) {
	panic("implement me")
}

func (m MockOrg) ListOrgs(ctx context.Context, orgIDs []int64, req *pb.ListOrgRequest, all bool) (int, []*pb.Org, error) {
	panic("implement me")
}
