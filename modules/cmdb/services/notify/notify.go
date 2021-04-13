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

package notify

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/i18n"
)

func (o *NotifyGroup) CreateNotify(locale *i18n.LocaleResource, createReq *apistructs.CreateNotifyRequest) (int64, error) {
	exist, err := o.db.CheckNotifyNameExist(createReq.ScopeType, createReq.ScopeID, createReq.Name, createReq.Label)
	if err != nil {
		return 0, err
	}
	if exist {
		return 0, errors.New(locale.Get("ErrCreateNotify.NameExist"))
	}
	if createReq.WithGroup {
		name := createReq.Name
		if len(createReq.GroupTargets) > 0 {
			name = o.getGroupTargetSummary(createReq.GroupTargets)
		}
		err := o.CheckNotifyGroupTarget(createReq.GroupTargets)
		if err != nil {
			return 0, err
		}
		notifyGroupID, err := o.db.CreateNotifyGroup(&apistructs.CreateNotifyGroupRequest{
			Name:        name,
			ScopeType:   createReq.ScopeType,
			ScopeID:     createReq.ScopeID,
			Targets:     createReq.GroupTargets,
			Creator:     createReq.Creator,
			OrgID:       createReq.OrgID,
			Label:       createReq.Label,
			ClusterName: createReq.ClusterName,
			AutoCreate:  true,
		})
		if err != nil {
			return 0, err
		}
		createReq.NotifyGroupID = notifyGroupID
		return o.db.CreateNotify(createReq)
	} else {
		return o.db.CreateNotify(createReq)
	}
}
func (o *NotifyGroup) CheckNotifyGroupTarget(targets []apistructs.NotifyTarget) error {
	for _, target := range targets {
		if len(target.Values) == 0 {
			return errors.New("target values is empty")
		}
	}
	return nil
}

func (o *NotifyGroup) CheckNotifyChannels(channelStr string) error {
	channels := strings.Split(channelStr, ",")
	for _, channel := range channels {
		if channel != "dingding" && channel != "sms" && channel != "email" && channel != "mbox" && channel != "webhook" {
			return errors.New("invalid channel: " + channel)
		}
	}
	return nil
}

func (o *NotifyGroup) GetNotify(notifyID int64, orgID int64) (*apistructs.NotifyDetail, error) {
	return o.db.GetNotifyDetail(notifyID, orgID)
}

func (o *NotifyGroup) UpdateNotify(req *apistructs.UpdateNotifyRequest) error {
	if req.WithGroup {
		req.GroupName = o.getGroupTargetSummary(req.GroupTargets)
	}
	return o.db.UpdateNotify(req)
}

func (o *NotifyGroup) UpdateNotifyEnable(id int64, enabled bool, orgID int64) error {
	return o.db.UpdateNotifyEnable(id, enabled, orgID)
}

func (o *NotifyGroup) QueryNotifies(request *apistructs.QueryNotifyRequest) (*apistructs.QueryNotifyData, error) {
	return o.db.QueryNotifies(request)
}

func (o *NotifyGroup) QueryNotifiesBySource(locale *i18n.LocaleResource, sourceType, sourceID, itemName string, orgId int64, clusterName string, label string) ([]*apistructs.NotifyDetail, error) {
	result, err := o.db.QueryNotifiesBySource(sourceType, sourceID, itemName, orgId, clusterName, label)
	if err != nil {
		return nil, err
	}
	for _, detail := range result {
		for _, item := range detail.NotifyItems {
			o.LocaleItem(locale, item)
		}
	}
	return result, nil
}

func (o *NotifyGroup) FuzzyQueryNotifiesBySource(req apistructs.FuzzyQueryNotifiesBySourceRequest) (*apistructs.QueryNotifyData, error) {
	result, total, err := o.db.FuzzyQueryNotifiesBySource(req)
	if err != nil {
		return nil, err
	}
	for _, detail := range result {
		for _, item := range detail.NotifyItems {
			o.LocaleItem(req.Locale, item)
		}
	}
	return &apistructs.QueryNotifyData{List: result, Total: total}, nil
}

func (o *NotifyGroup) DeleteNotify(notifyID int64, deleteGroup bool, orgID int64) error {
	return o.db.DeleteNotify(notifyID, deleteGroup, orgID)
}

func (o *NotifyGroup) getGroupTargetSummary(targets []apistructs.NotifyTarget) string {
	users := []string{}
	for _, target := range targets {
		if target.Type == apistructs.ExternalUserNotifyTarget {
			for _, valueStr := range target.Values {
				var userInfo apistructs.NotifyUser
				err := json.Unmarshal([]byte(valueStr.Receiver), &userInfo)
				if err == nil {
					if userInfo.Username != "" {
						users = append(users, userInfo.Username)
					} else if userInfo.Mobile != "" {
						users = append(users, userInfo.Mobile)
					} else if userInfo.Email != "" {
						users = append(users, userInfo.Email)
					}
				}
			}
		}
	}
	name := strings.Join(users, ",")
	if len(name) > 100 {
		name = name[0:100]
	}
	return name
}
