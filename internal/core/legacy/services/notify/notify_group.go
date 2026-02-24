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

package notify

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/jinzhu/gorm"

	userpb "github.com/erda-project/erda-proto-go/core/user/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/internal/core/legacy/dao"
	"github.com/erda-project/erda/internal/core/legacy/model"
	"github.com/erda-project/erda/pkg/common/apis"
	"github.com/erda-project/erda/pkg/discover"
	"github.com/erda-project/erda/pkg/i18n"
)

type NotifyGroup struct {
	db          *dao.DBClient
	bdl         *bundle.Bundle
	userService userpb.UserServiceServer
}

type Option func(*NotifyGroup)

func New(options ...Option) *NotifyGroup {
	o := &NotifyGroup{}
	for _, op := range options {
		op(o)
	}
	return o
}

// WithDBClient 配置 db client
func WithDBClient(db *dao.DBClient) Option {
	return func(o *NotifyGroup) {
		o.db = db
	}
}

func WithUserService(userService userpb.UserServiceServer) Option {
	return func(o *NotifyGroup) {
		o.userService = userService
	}
}

func (o *NotifyGroup) Create(locale *i18n.LocaleResource, createReq *apistructs.CreateNotifyGroupRequest) (int64, error) {
	exist, err := o.db.CheckNotifyGroupNameExist(createReq.ScopeType, createReq.ScopeID, createReq.Name)
	if err != nil {
		return 0, err
	}
	if exist {
		return 0, errors.New(locale.Get("ErrCreateNotifyGroup.NameExist"))
	}
	err = o.CheckNotifyGroupTarget(createReq.Targets)
	if err != nil {
		return 0, err
	}
	return o.db.CreateNotifyGroup(createReq)
}

func (o *NotifyGroup) Update(updateReq *apistructs.UpdateNotifyGroupRequest) error {
	err := o.CheckNotifyGroupTarget(updateReq.Targets)
	if err != nil {
		return err
	}
	return o.db.UpdateNotifyGroup(updateReq)
}

func (o *NotifyGroup) Get(id int64, orgID int64) (*apistructs.NotifyGroup, error) {
	return o.db.GetNotifyGroupByID(id, orgID)
}

func (o *NotifyGroup) BatchGet(ids []int64) ([]*apistructs.NotifyGroup, error) {
	return o.db.BatchGetNotifyGroup(ids)
}

func (o *NotifyGroup) Delete(id int64, orgID int64) error {
	notify, err := o.db.GetNotifyByGroupID(id)
	if err == gorm.ErrRecordNotFound {
		return o.db.DeleteNotifyGroup(id, orgID)
	}
	if err != nil {
		return err
	}
	return fmt.Errorf("存在关联通知无法删除 关联通知ID:%d 通知名:%s ", notify.ID, notify.Name)
}

func (o *NotifyGroup) Query(queryReq *apistructs.QueryNotifyGroupRequest, orgID int64) (*apistructs.QueryNotifyGroupData, error) {
	return o.db.QueryNotifyGroup(queryReq, orgID)
}

func (o *NotifyGroup) GetDetail(id int64, orgID int64) (*apistructs.NotifyGroupDetail, error) {
	result := &apistructs.NotifyGroupDetail{}
	group, err := o.db.GetNotifyGroupByID(id, orgID)
	if err != nil {
		return nil, err
	}
	for _, target := range group.Targets {
		var err error
		switch target.Type {
		case apistructs.UserNotifyTarget:
			var userIDs []string
			for _, value := range target.Values {
				userIDs = append(userIDs, value.Receiver)
			}
			notifyUsers, err := o.getNotifyUsersByIDs(userIDs)
			if err != nil {
				return nil, err
			}
			result.Users = append(result.Users, notifyUsers...)
		case apistructs.ExternalUserNotifyTarget:
			// 外部用户
			for _, valueStr := range target.Values {
				var user apistructs.NotifyUser
				err := json.Unmarshal([]byte(valueStr.Receiver), &user)
				if err == nil {
					result.Users = append(result.Users, user)
				}
			}
		case apistructs.DingdingNotifyTarget:
			for _, dingding := range target.Values {
				result.DingdingList = append(result.DingdingList, dingding)
			}
		case apistructs.DingdingWorkNoticeNotifyTarget:
			for _, dingding := range target.Values {
				result.DingdingWorkNoticeList = append(result.DingdingWorkNoticeList, dingding)
			}
		case apistructs.WebhookNotifyTarget:
			for _, webhookUrl := range target.Values {
				result.WebHookList = append(result.WebHookList, webhookUrl.Receiver)
			}
		case apistructs.RoleNotifyTarget:
			var roles []string
			for _, r := range target.Values {
				roles = append(roles, r.Receiver)
			}
			var members []model.Member
			if group.ScopeType == apistructs.MSPScope {
				label := make(map[string]string)
				err = json.Unmarshal([]byte(group.Label), &label)
				if err != nil {
					return nil, err
				}
				scopeID, err := strconv.ParseInt(label[apistructs.MSPMemberScopeId], 10, 32)
				if err != nil {
					return nil, err
				}
				_, members, err = o.db.GetMembersByParam(&apistructs.MemberListRequest{
					ScopeType: apistructs.ScopeType(label[apistructs.MSPMemberScope]),
					ScopeID:   scopeID,
					Roles:     roles,
					Q:         "",
					PageNo:    1,
					PageSize:  1000,
				})
			} else {
				scopeID, err := strconv.ParseInt(group.ScopeID, 10, 32)
				if err != nil {
					return nil, err
				}
				_, members, err = o.db.GetMembersByParam(&apistructs.MemberListRequest{
					ScopeType: apistructs.ScopeType(group.ScopeType),
					ScopeID:   scopeID,
					Roles:     roles,
					Q:         "",
					PageNo:    1,
					PageSize:  1000,
				})
			}
			if err != nil {
				return nil, err
			}
			var userIDs []string
			for _, member := range members {
				userIDs = append(userIDs, member.UserID)
			}
			notifyUsers, err := o.getNotifyUsersByIDs(userIDs)
			if err != nil {
				return nil, err
			}
			result.Users = append(result.Users, notifyUsers...)
		}
		if err != nil {
			return nil, err
		}
	}

	result.ID = group.ID
	result.Name = group.Name
	result.ScopeType = group.ScopeType
	result.ScopeID = group.ScopeID
	result.Targets = group.Targets
	result.DingdingList = uniqueTargetList(result.DingdingList)
	result.DingdingWorkNoticeList = uniqueTargetList(result.DingdingWorkNoticeList)
	result.Label = group.Label
	return result, nil
}

func uniqueTargetList(list []apistructs.Target) []apistructs.Target {
	var result []apistructs.Target
	keyMap := map[string]apistructs.Target{}
	for _, v := range list {
		keyMap[v.Receiver] = v
	}
	for _, target := range keyMap {
		result = append(result, target)
	}
	return result
}

func (o *NotifyGroup) getNotifyUsersByIDs(userIds []string) ([]apistructs.NotifyUser, error) {
	findUsersResp, err := o.userService.FindUsers(
		apis.WithInternalClientContext(context.Background(), discover.SvcCoreServices),
		&userpb.FindUsersRequest{IDs: userIds},
	)
	if err != nil {
		return nil, err
	}
	var notifyUsers []apistructs.NotifyUser
	for _, user := range findUsersResp.Data {
		user := apistructs.NotifyUser{
			ID:       user.Id,
			Type:     "inner",
			Email:    user.Email,
			Mobile:   user.Phone,
			Username: user.Name,
		}
		notifyUsers = append(notifyUsers, user)
	}
	return notifyUsers, nil
}
