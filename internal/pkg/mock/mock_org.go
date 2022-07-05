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

package mock

import (
	"context"

	orgpb "github.com/erda-project/erda-proto-go/core/org/pb"
)

type OrgMock struct {
}

func (m OrgMock) CreateOrg(ctx context.Context, request *orgpb.CreateOrgRequest) (*orgpb.CreateOrgResponse, error) {
	panic("implement me")
}

func (m OrgMock) UpdateOrg(ctx context.Context, request *orgpb.UpdateOrgRequest) (*orgpb.UpdateOrgResponse, error) {
	panic("implement me")
}

func (m OrgMock) GetOrg(ctx context.Context, request *orgpb.GetOrgRequest) (*orgpb.GetOrgResponse, error) {
	panic("implement me")
}

func (m OrgMock) DeleteOrg(ctx context.Context, request *orgpb.DeleteOrgRequest) (*orgpb.DeleteOrgResponse, error) {
	panic("implement me")
}

func (m OrgMock) ListOrg(ctx context.Context, request *orgpb.ListOrgRequest) (*orgpb.ListOrgResponse, error) {
	panic("implement me")
}

func (m OrgMock) ListPublicOrg(ctx context.Context, request *orgpb.ListOrgRequest) (*orgpb.ListOrgResponse, error) {
	panic("implement me")
}

func (m OrgMock) GetOrgByDomain(ctx context.Context, request *orgpb.GetOrgByDomainRequest) (*orgpb.GetOrgByDomainResponse, error) {
	panic("implement me")
}

func (m OrgMock) ChangeCurrentOrg(ctx context.Context, request *orgpb.ChangeCurrentOrgRequest) (*orgpb.ChangeCurrentOrgResponse, error) {
	panic("implement me")
}

func (m OrgMock) CreateOrgClusterRelation(ctx context.Context, request *orgpb.OrgClusterRelationCreateRequest) (*orgpb.OrgClusterRelationCreateResponse, error) {
	panic("implement me")
}

func (m OrgMock) ListOrgClusterRelation(ctx context.Context, request *orgpb.ListOrgClusterRelationRequest) (*orgpb.ListOrgClusterRelationResponse, error) {
	panic("implement me")
}

func (m OrgMock) SetReleaseCrossCluster(ctx context.Context, request *orgpb.SetReleaseCrossClusterRequest) (*orgpb.SetReleaseCrossClusterResponse, error) {
	panic("implement me")
}

func (m OrgMock) GenVerifyCode(ctx context.Context, request *orgpb.GenVerifyCodeRequest) (*orgpb.GenVerifyCodeResponse, error) {
	panic("implement me")
}

func (m OrgMock) SetNotifyConfig(ctx context.Context, request *orgpb.SetNotifyConfigRequest) (*orgpb.SetNotifyConfigResponse, error) {
	panic("implement me")
}

func (m OrgMock) GetNotifyConfig(ctx context.Context, request *orgpb.GetNotifyConfigRequest) (*orgpb.GetNotifyConfigResponse, error) {
	panic("implement me")
}

func (m OrgMock) GetOrgClusterRelationsByOrg(ctx context.Context, request *orgpb.GetOrgClusterRelationsByOrgRequest) (*orgpb.GetOrgClusterRelationsByOrgResponse, error) {
	panic("implement me")
}

func (m OrgMock) DereferenceCluster(ctx context.Context, request *orgpb.DereferenceClusterRequest) (*orgpb.DereferenceClusterResponse, error) {
	panic("implement me")
}
