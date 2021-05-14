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

package models

import (
	"errors"
	"time"

	"github.com/erda-project/erda/modules/gittar/pkg/gitmodule"
)

type Permission string

const (
	PermissionCreateBranch           Permission = "CREATE_BRANCH"
	PermissionDeleteBranch           Permission = "DELETE_BRANCH"
	PermissionCreateTAG              Permission = "CREATE_TAG"
	PermissionDeleteTAG              Permission = "DELETE_TAG"
	PermissionCloseMR                Permission = "CLOSE_MR"
	PermissionCreateMR               Permission = "CREATE_MR"
	PermissionMergeMR                Permission = "MERGE_MR"
	PermissionEditMR                 Permission = "EDIT_MR"
	PermissionArchive                Permission = "ARCHIVE"
	PermissionClone                  Permission = "CLONE"
	PermissionPush                   Permission = "PUSH"
	PermissionPushProtectBranch      Permission = "PUSH_PROTECT_BRANCH"
	PermissionPushProtectBranchForce Permission = "PUSH_PROTECT_BRANCH_FORCE"
	PermissionRepoLocked             Permission = "REPO_LOCKED"
)

var NO_PERMISSION_ERROR = errors.New("no permission")

var rolePermissions map[string][]Permission

func init() {
	rolePermissions = map[string][]Permission{}
	rolePermissions["Manager"] = []Permission{
		PermissionCreateBranch,
		PermissionDeleteBranch,
		PermissionCreateTAG,
		PermissionDeleteTAG,
		PermissionCloseMR,
		PermissionCreateMR,
		PermissionMergeMR,
		PermissionPush,
		PermissionPushProtectBranch,
		PermissionPushProtectBranchForce,
		PermissionArchive,
		PermissionClone,
	}
	rolePermissions["Developer"] = []Permission{
		PermissionCreateBranch,
		PermissionDeleteBranch,
		PermissionCreateTAG,
		PermissionDeleteTAG,
		PermissionCloseMR,
		PermissionCreateMR,
		PermissionPushProtectBranch,
		PermissionPush,
		PermissionArchive,
		PermissionClone,
	}
	rolePermissions["Tester"] = []Permission{
		PermissionPush,
		PermissionCreateBranch,
		PermissionArchive,
		PermissionClone,
	}

	//在app有权限的情况，只有界面只读接口权限
	rolePermissions["Guest"] = []Permission{}
}

// User struct for sender pusher and more
type User struct {
	Id       string `json:"id"`
	Name     string `json:"name"`
	NickName string `json:"nickname"`
	Email    string `json:"email"`
}

func NewInnerUser() *User {
	return &User{
		Id:       "0",
		Name:     "git_inner_user",
		NickName: "git_inner_user",
		Email:    "git_inner_user@gittar.com",
	}
}

func (user *User) ToGitSignature() *gitmodule.Signature {
	name := user.NickName
	if name == "" {
		name = user.Name
	}
	email := user.Email
	if email == "" {
		email = name + "@gittar.default"
	}
	return &gitmodule.Signature{
		Name:  name,
		Email: email,
		When:  time.Now(),
	}
}

func (user *User) IsInnerUser() bool {
	return user.Id == "0"
}
