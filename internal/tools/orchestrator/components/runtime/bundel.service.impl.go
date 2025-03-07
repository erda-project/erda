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

package runtime

import (
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/pkg/parser/diceyml"
)

type bundleServiceImpl struct {
	bdl *bundle.Bundle
}

func (b *bundleServiceImpl) GetProjectBranchRules(projectId uint64) ([]*apistructs.BranchRule, error) {
	return b.bdl.GetProjectBranchRules(projectId)
}

func (b *bundleServiceImpl) GetDiceYAML(releaseID string, workspace ...string) (*diceyml.DiceYaml, error) {
	return b.bdl.GetDiceYAML(releaseID, workspace...)
}

func (b *bundleServiceImpl) ListMembers(req apistructs.MemberListRequest) ([]apistructs.Member, error) {
	return b.bdl.ListMembers(req)
}

func (b *bundleServiceImpl) ListUsers(req apistructs.UserListRequest) (*apistructs.UserListResponseData, error) {
	return b.bdl.ListUsers(req)
}

func (b *bundleServiceImpl) CreateMboxNotify(templatename string, params map[string]string, locale string, orgid uint64, users []string) error {
	return b.bdl.CreateMboxNotify(templatename, params, locale, orgid, users)
}

func (b *bundleServiceImpl) CreateEmailNotify(templatename string, params map[string]string, locale string, orgid uint64, emailaddrs []string) error {
	return b.bdl.CreateEmailNotify(templatename, params, locale, orgid, emailaddrs)
}

func (b *bundleServiceImpl) GetAllValidBranchWorkspace(appId uint64, userID string) ([]apistructs.ValidBranch, error) {
	return b.bdl.GetAllValidBranchWorkspace(appId, userID)
}

func (b *bundleServiceImpl) GetAllProjects() ([]apistructs.ProjectDTO, error) {
	return b.bdl.GetAllProjects()
}

func (b *bundleServiceImpl) GetLog(orgName string, req apistructs.DashboardSpotLogRequest) (*apistructs.DashboardSpotLogData, error) {
	return b.bdl.GetLog(orgName, req)
}

func (b *bundleServiceImpl) GetGittarCommit(repo, ref, userID string) (*apistructs.Commit, error) {
	return b.bdl.GetGittarCommit(repo, ref, userID)
}

func (b *bundleServiceImpl) InvalidateOAuth2Token(req apistructs.OAuth2TokenInvalidateRequest) (*apistructs.OAuth2Token, error) {
	return b.bdl.InvalidateOAuth2Token(req)
}

func (b *bundleServiceImpl) GetProject(id uint64) (*apistructs.ProjectDTO, error) {
	return b.bdl.GetProject(id)
}

func (b *bundleServiceImpl) GetMyAppsByProject(userid string, orgid, projectID uint64, appName string) (*apistructs.ApplicationListResponseData, error) {
	return b.bdl.GetMyAppsByProject(userid, orgid, projectID, appName)
}

func (b *bundleServiceImpl) GetMyApps(userid string, orgid uint64) (*apistructs.ApplicationListResponseData, error) {
	return b.bdl.GetMyApps(userid, orgid)
}

func (b *bundleServiceImpl) GetApp(id uint64) (*apistructs.ApplicationDTO, error) {
	return b.bdl.GetApp(id)
}

func (b *bundleServiceImpl) InspectServiceGroupWithTimeout(namespace string, name string) (*apistructs.ServiceGroup, error) {
	return b.bdl.InspectServiceGroupWithTimeout(namespace, name)
}

func (b *bundleServiceImpl) GetCluster(name string) (*apistructs.ClusterInfo, error) {
	return b.bdl.GetCluster(name)
}

func (b *bundleServiceImpl) CheckPermission(req *apistructs.PermissionCheckRequest) (*apistructs.PermissionCheckResponseData, error) {
	return b.bdl.CheckPermission(req)
}

func NewBundleService() BundleService {
	return &bundleServiceImpl{
		bdl: bundle.New(
			bundle.WithErdaServer(),
			bundle.WithClusterManager(),
			bundle.WithScheduler(),
		),
	}
}
