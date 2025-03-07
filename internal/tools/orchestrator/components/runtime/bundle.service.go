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
	"github.com/erda-project/erda/pkg/parser/diceyml"
)

type BundleService interface {
	CheckPermission(req *apistructs.PermissionCheckRequest) (*apistructs.PermissionCheckResponseData, error)
	GetCluster(name string) (*apistructs.ClusterInfo, error)
	InspectServiceGroupWithTimeout(namespace string, name string) (*apistructs.ServiceGroup, error)
	GetApp(id uint64) (*apistructs.ApplicationDTO, error)
	GetProject(id uint64) (*apistructs.ProjectDTO, error)
	GetMyAppsByProject(userid string, orgid, projectID uint64, appName string) (*apistructs.ApplicationListResponseData, error)
	GetMyApps(userid string, orgid uint64) (*apistructs.ApplicationListResponseData, error)
	GetProjectBranchRules(projectId uint64) ([]*apistructs.BranchRule, error)
	GetDiceYAML(releaseID string, workspace ...string) (*diceyml.DiceYaml, error)
	ListMembers(req apistructs.MemberListRequest) ([]apistructs.Member, error)
	ListUsers(req apistructs.UserListRequest) (*apistructs.UserListResponseData, error)
	CreateMboxNotify(templatename string, params map[string]string, locale string, orgid uint64, users []string) error
	CreateEmailNotify(templatename string, params map[string]string, locale string, orgid uint64, emailaddrs []string) error
	GetAllValidBranchWorkspace(appId uint64, userID string) ([]apistructs.ValidBranch, error)
	GetAllProjects() ([]apistructs.ProjectDTO, error)
	GetLog(orgName string, req apistructs.DashboardSpotLogRequest) (*apistructs.DashboardSpotLogData, error)
	GetGittarCommit(repo, ref, userID string) (*apistructs.Commit, error)
	InvalidateOAuth2Token(req apistructs.OAuth2TokenInvalidateRequest) (*apistructs.OAuth2Token, error)
}
