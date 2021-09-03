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

package member

import (
	"bou.ke/monkey"
	"context"
	"fmt"
	"github.com/erda-project/erda-proto-go/msp/member/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/pkg/common/apis"
	"github.com/golang/mock/gomock"
	"reflect"
	"testing"
)

////go:generate mockgen -destination=./credential_register_test.go -package exporter github.com/erda-project/erda-infra/pkg/transport Register
func Test_memberService_ListMember(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	register := NewMockRegister(ctrl)
	pro := &provider{
		Cfg:           &config{},
		Register:      register,
		memberService: &memberService{},
		bdl:           &bundle.Bundle{},
	}
	var bdl *bundle.Bundle
	listMembersAndTotalData := &apistructs.MemberList{
		List: []apistructs.Member{
			{
				UserID: "sss",
				Email:  "ssdddd",
				Mobile: "www",
				Name:   "gg",
				Nick:   "33",
				Avatar: "s",
				Status: "f",
				Scope: apistructs.Scope{
					Type: "project",
					ID:   "dddd",
				},
				Roles:   nil,
				Labels:  nil,
				Removed: false,
				Deleted: false,
			},
		},
		Total: 1,
	}
	defer monkey.UnpatchAll()
	monkey.PatchInstanceMethod(reflect.TypeOf(bdl), "ListMembersAndTotal",
		func(_ *bundle.Bundle, _ apistructs.MemberListRequest) (*apistructs.MemberList, error) {
			return listMembersAndTotalData, nil
		})
	pro.memberService.p = pro
	_, err := pro.memberService.ListMember(context.Background(), &pb.ListMemberRequest{
		ScopeType: "project",
		ScopeId:   "15",
		PageNo:    1,
		PageSize:  1,
	})
	if err != nil {
		fmt.Println(err)
		fmt.Println("should not error")
	}
}

func Test_memberService_DeleteMember(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	register := NewMockRegister(ctrl)
	pro := &provider{
		Cfg:           &config{},
		Register:      register,
		memberService: &memberService{},
		bdl:           &bundle.Bundle{},
	}
	var bdl *bundle.Bundle
	defer monkey.UnpatchAll()
	monkey.PatchInstanceMethod(reflect.TypeOf(bdl), "DeleteMember",
		func(_ *bundle.Bundle, _ apistructs.MemberRemoveRequest) error {
			return nil
		})
	pro.memberService.p = pro
	_, err := pro.memberService.DeleteMember(context.Background(), &pb.DeleteMemberRequest{
		Scope: &pb.Scope{
			Id:   "12",
			Type: "project",
		},
		UserIds:        []string{"11"},
		UserID:         "ss",
		InternalClient: "ss",
	})
	if err != nil {
		fmt.Println(err)
		fmt.Println("should not err")
	}
}

func Test_memberService_ListMemberRoles(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	register := NewMockRegister(ctrl)
	pro := &provider{
		Cfg:           &config{},
		Register:      register,
		memberService: &memberService{},
		bdl:           &bundle.Bundle{},
	}
	var bdl *bundle.Bundle
	defer monkey.UnpatchAll()
	monkey.PatchInstanceMethod(reflect.TypeOf(bdl), "ListMemberRoles",
		func(_ *bundle.Bundle, _ apistructs.ListScopeManagersByScopeIDRequest, _ int64) (*apistructs.RoleList, error) {
			return &apistructs.RoleList{
				List: []apistructs.RoleInfo{
					{
						Role: "xxse",
						Name: "dfdf",
					},
				},
				Total: 1,
			}, nil
		})
	monkey.Patch(apis.GetOrgID, func(ctx context.Context) string {
		return "1"
	})
	pro.memberService.p = pro
	_, err := pro.memberService.ListMemberRoles(context.Background(), &pb.ListMemberRolesRequest{
		ScopeType: "project",
		ScopeID:   12,
	})
	if err != nil {
		fmt.Println(err)
		fmt.Println("should not err")
	}
}

func Test_memberService_CreateOrUpdateMember(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	register := NewMockRegister(ctrl)
	pro := &provider{
		Cfg:           &config{},
		Register:      register,
		memberService: &memberService{},
		bdl:           &bundle.Bundle{},
	}
	var bdl *bundle.Bundle
	defer monkey.UnpatchAll()
	monkey.Patch(apis.GetUserID, func(ctx context.Context) string {
		return "1"
	})
	monkey.PatchInstanceMethod(reflect.TypeOf(bdl), "AddMember",
		func(_ *bundle.Bundle, _ apistructs.MemberAddRequest, _ string) error {
			return nil
		})
	pro.memberService.p = pro
	_, err := pro.memberService.CreateOrUpdateMember(context.Background(), &pb.CreateOrUpdateMemberRequest{
		Scope: &pb.Scope{
			Id:   "sss",
			Type: "projuect",
		},
		TargetScopeType: "sss",
		TargetScopeIds:  []int64{12},
		Roles:           []string{"Dev"},
		UserIds:         []string{"12"},
		Options: &pb.MemberAddOptions{
			Rewrite: true,
		},
		Labels:     []string{"xxx"},
		VerifyCode: "sss",
	})
	if err != nil {
		fmt.Println("should not err")
	}
}

func Test_memberService_ScopeRoleAccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	register := NewMockRegister(ctrl)
	pro := &provider{
		Cfg:           &config{},
		Register:      register,
		memberService: &memberService{},
		bdl:           &bundle.Bundle{},
	}
	var bdl *bundle.Bundle
	defer monkey.UnpatchAll()
	monkey.Patch(apis.GetUserID, func(ctx context.Context) string {
		return "1"
	})
	monkey.PatchInstanceMethod(reflect.TypeOf(bdl), "ScopeRoleAccessList",
		func(_ *bundle.Bundle, _ string, _ *apistructs.ScopeRoleAccessRequest) (*apistructs.PermissionList, error) {
			return &apistructs.PermissionList{
				Access: false,
				Roles:  []string{"Dev"},
				PermissionList: []apistructs.ScopeResource{
					{
						Resource:     "xxx",
						Action:       "xxx",
						ResourceRole: "xxx",
					},
				},
				ResourceRoleList: []apistructs.ScopeResource{
					{
						Resource:     "sss",
						Action:       "sss",
						ResourceRole: "sss",
					},
				},
				Exist: false,
				ContactsWhenNoPermission: []string{
					"sss",
				},
			}, nil
		})
	pro.memberService.p = pro
	_, err := pro.memberService.ScopeRoleAccess(context.Background(), &pb.ScopeRoleAccessRequest{
		Scope: &pb.Scope{
			Id:   "sss",
			Type: "sss",
		},
	})
	if err != nil {
		fmt.Println("should not err")
	}
}
