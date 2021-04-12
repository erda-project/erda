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

package bdl

import (
	"strconv"
	"strings"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/apim/dbclient"
)

var (
	OrgMRoles = []string{"manager"}             // 对企业有管理权限的角色列表
	ProMRoles = []string{"owner", "lead", "pm"} // 对项目有管理权限的角色列表
	AppMRoles = []string{"owner", "lead"}       // 对应用有管理权限的角色列表
)

type RolesSet struct {
	orgID                uint64
	userID               string
	orgs, projects, apps map[string][]string // key 为 scope id, value 为 role
}

func (p *RolesSet) OrgID() uint64 {
	return p.orgID
}

func (p *RolesSet) UserID() string {
	return p.userID
}

func (p *RolesSet) Orgs() map[string][]string {
	return p.orgs
}

func (p *RolesSet) Projects() map[string][]string {
	return p.projects
}

func (p *RolesSet) Apps() map[string][]string {
	return p.apps
}

// 传入角色, 返回 org id 列表 (由于返回结果不会太多, 所以都没有做去重)
func (p *RolesSet) RolesOrgs(roles ...string) []string {
	return p.sliceRoles(p.orgs, roles...)
}

// 传入角色, 返回 project id 列表 (由于返回结果不会太多, 所以都没有做去重)
func (p *RolesSet) RolesProjects(roles ...string) []string {
	return p.sliceRoles(p.projects, roles...)
}

// 传入角色, 返回 app id 列表 (由于返回结果不会太多, 所以都没有做去重)
func (p *RolesSet) RolesApps(roles ...string) []string {
	return p.sliceRoles(p.apps, roles...)
}

func (p *RolesSet) sliceRoles(scope map[string][]string, roles ...string) []string {
	var results []string
	if len(roles) == 0 || roles[0] == "*" {
		for id := range scope {
			results = append(results, id)
		}
		return results
	}
	for _, role := range roles {
		results = append(results, p.slice(role, scope)...)
	}
	return results
}

func (p *RolesSet) slice(role string, scope map[string][]string) []string {
	var results []string
	for id, roles := range scope {
		for _, v := range roles {
			if strings.EqualFold(v, role) {
				results = append(results, id)
				continue
			}
		}
	}
	return results
}

func FetchRolesSet(orgID uint64, userID string) *RolesSet {
	var permission = RolesSet{
		orgID:    orgID,
		userID:   userID,
		orgs:     make(map[string][]string),
		projects: make(map[string][]string),
		apps:     make(map[string][]string),
	}

	id := strconv.FormatUint(orgID, 10)
	if access, err := Bdl.ScopeRoleAccess(userID,
		&apistructs.ScopeRoleAccessRequest{Scope: apistructs.Scope{
			Type: apistructs.OrgScope,
			ID:   id,
		}}); err == nil && access.Access {
		permission.orgs[id] = access.Roles
	}

	var models []*apistructs.APIAssetsModel
	if err := dbclient.Sq().Where(map[string]interface{}{"org_id": orgID}).Find(&models).Error; err != nil {
		return &permission
	}

	for _, v := range models {
		if v.ProjectID != nil && *v.ProjectID != 0 {
			id := strconv.FormatUint(*v.ProjectID, 10)
			if _, ok := permission.projects[id]; !ok {
				if access, err := Bdl.ScopeRoleAccess(userID,
					&apistructs.ScopeRoleAccessRequest{Scope: apistructs.Scope{
						Type: apistructs.ProjectScope,
						ID:   id,
					}}); err == nil && access.Access {
					permission.projects[id] = access.Roles
				}
			}
		}
		if v.AppID != nil && *v.AppID != 0 {
			id := strconv.FormatUint(*v.AppID, 10)
			if _, ok := permission.apps[id]; !ok {
				if access, err := Bdl.ScopeRoleAccess(userID,
					&apistructs.ScopeRoleAccessRequest{Scope: apistructs.Scope{
						Type: apistructs.AppScope,
						ID:   id,
					}}); err == nil && access.Access {
					permission.apps[id] = access.Roles
				}
			}
		}
	}

	return &permission
}
