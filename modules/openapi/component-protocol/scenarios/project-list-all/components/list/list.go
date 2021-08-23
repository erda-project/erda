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

package list

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/openapi/component-protocol/scenarios/project-list-all/i18n"
)

func GetOpsInfo(opsData interface{}) (*Meta, error) {
	if opsData == nil {
		err := fmt.Errorf("empty operation data")
		return nil, err
	}
	var op Operation
	cont, err := json.Marshal(opsData)
	if err != nil {
		logrus.Errorf("marshal inParams failed, content:%v, err:%v", opsData, err)
		return nil, err
	}
	err = json.Unmarshal(cont, &op)
	if err != nil {
		logrus.Errorf("unmarshal move out request failed, content:%v, err:%v", cont, err)
		return nil, err
	}
	meta := op.Meta
	return &meta, nil
}

func (i *ComponentList) RenItem(pro apistructs.ProjectDTO, orgDomain string) (ProItem, error) {
	i18nLocale := i.ctxBdl.Bdl.GetLocale(i.ctxBdl.Locale)
	activeTime, err := i.CountActiveTime(pro.ActiveTime)
	if err != nil {
		return ProItem{}, err
	}
	opClick := Operation{
		Key:    apistructs.ClickOperation.String(),
		Reload: false,
		Show:   false,
		Command: Command{
			Key:    "goto",
			Target: "project",
		},
	}
	//opManage := GenerateOperation{
	//	Key:    apistructs.ListProjectToManageOperationKey.String(),
	//	Reload: false,
	//	Text:   "管理",
	//	Show:   false,
	//	Command: Command{
	//		Key:    "goto",
	//		Target: "https://terminus-org.dev.terminus.io/orgCenter/projects",
	//	},
	//}
	//opExist := GenerateOperation{
	//	Key:     apistructs.ListProjectExistOperationKey.String(),
	//	Reload:  true,
	//	Text:    "退出",
	//	Confirm: "退出当前项目后，将不再有项目协作权限，如要再次加入需要项目管理员邀请，请确认是否退出？",
	//	Meta: Meta{
	//		ID: pro.ID,
	//	},
	//}
	//opApplyDeploy := GenerateOperation{
	//	Key:    apistructs.ApplyDeployProjectFilterOperation.String(),
	//	Reload: false,
	//	Text:   "申请部署",
	//	Meta: Meta{
	//		ProjectId:   pro.ID,
	//		ProjectName: pro.DisplayName,
	//	},
	//}

	item := ProItem{
		ID:          strconv.Itoa(int(pro.ID)),
		ProjectId:   pro.ID,
		Title:       pro.DisplayName,
		Description: pro.Desc,
		PrefixImg:   "frontImg_default_project_icon",
		ExtraInfos: []ExtraInfos{
			map[bool]ExtraInfos{
				true: {
					Icon: "unlock",
					Text: i18nLocale.Get(i18n.I18nPublicProject),
				},
				false: {
					Icon: "lock",
					Text: i18nLocale.Get(i18n.I18nPrivateProject),
				},
			}[pro.IsPublic],
			{
				Icon:    "application-one",
				Text:    strconv.Itoa(pro.Stats.CountApplications),
				Tooltip: i18nLocale.Get(i18n.I18nAppNumber),
			},
			{
				Icon:    "time",
				Text:    activeTime,
				Tooltip: pro.ActiveTime,
			},
		},
		Operations: map[string]interface{}{
			apistructs.ClickOperation.String(): opClick,
		},
	}
	if pro.Logo != "" {
		item.PrefixImg = pro.Logo
	}

	// joined
	if pro.Joined {
		item.ExtraInfos = append(item.ExtraInfos, ExtraInfos{
			Icon: "user",
			Text: i18nLocale.Get(i18n.I18nJoined),
		})
	}
	// 解封状态
	if pro.BlockStatus == "unblocking" {
		item.ExtraInfos = append(item.ExtraInfos, ExtraInfos{
			Icon:    "link-cloud-faild",
			Text:    i18nLocale.Get(i18n.I18nUnblocking),
			Type:    "warning",
			Tooltip: i18nLocale.Get(i18n.I18nUnblocking),
		})
	} else if pro.BlockStatus == "unblocked" {
		item.ExtraInfos = append(item.ExtraInfos, ExtraInfos{
			Icon: "link-cloud-sucess",
			Text: i18nLocale.Get(i18n.I18nUnblocked),
			Type: "success",
		})
	}
	//目前去掉关联按钮
	//// Owner Lead PM
	//if pro.CanManage {
	//	item.Operations[apistructs.ListProjectToManageOperationKey.String()] = opManage
	//}
	// only exist project in my project list
	//if pro.Joined {
	//	item.Operations[apistructs.ListProjectExistOperationKey.String()] = opExist
	//}
	// 根据解封状态判断
	//if pro.CanManage && (pro.BlockStatus == "unblocked" || pro.BlockStatus == "unblocking" || pro.BlockStatus == "blocked") {
	//	item.Operations[apistructs.ApplyDeployProjectFilterOperation.String()] = opApplyDeploy
	//}
	return item, nil
}

func (i *ComponentList) RenderList() error {
	orgID, err := strconv.ParseUint(i.ctxBdl.Identity.OrgID, 10, 64)
	if err != nil {
		return err
	}
	queryStr := ""
	_, ok := i.State.Query["title"]
	if ok {
		queryStr = i.State.Query["title"].(string)
	}

	req := apistructs.ProjectListRequest{
		OrgID:    orgID,
		Query:    queryStr,
		PageNo:   int(i.State.PageNo),
		PageSize: int(i.State.PageSize),
		OrderBy:  "activeTime",
		IsPublic: true,
	}

	projectDTO, err := i.ctxBdl.Bdl.ListPublicProject(i.ctxBdl.Identity.UserID, req)
	if err != nil {
		return err
	}

	org, err := i.ctxBdl.Bdl.GetOrg(orgID)
	if err != nil {
		return err
	}

	i.Data.List = make([]ProItem, 0)
	i.State.Total = 0
	if projectDTO != nil {
		for _, v := range projectDTO.List {
			p, err := i.RenItem(v, org.Domain)
			if err != nil {
				return err
			}
			i.Data.List = append(i.Data.List, p)
		}
		i.State.Total = uint64(projectDTO.Total)
	}
	return nil
}

func (i *ComponentList) RenderExist(ops interface{}) error {
	mate, err := GetOpsInfo(ops)
	if err != nil {
		return err
	}
	req := apistructs.MemberRemoveRequest{
		Scope: apistructs.Scope{
			Type: apistructs.ProjectScope,
			ID:   strconv.Itoa(int(mate.ID)),
		},
		UserIDs: []string{i.ctxBdl.Identity.UserID},
	}
	req.UserID = i.ctxBdl.Identity.UserID
	return i.ctxBdl.Bdl.DeleteMember(req)
}

func (i *ComponentList) RenderChangePageSize(ops interface{}) error {
	mate, err := GetOpsInfo(ops)
	if err != nil {
		return err
	}
	i.State.PageNo = 1
	i.State.PageSize = mate.PageSize.PageSize
	return nil
}

func (i *ComponentList) RenderChangePageNo(ops interface{}) error {
	mate, err := GetOpsInfo(ops)
	if err != nil {
		return err
	}
	i.State.PageNo = mate.PageNo.PageNo
	i.State.PageSize = mate.PageNo.PageSize
	return nil
}

func (i *ComponentList) CountActiveTime(ActiveTimeStr string) (string, error) {
	var subStr string
	i18nLocale := i.ctxBdl.Bdl.GetLocale(i.ctxBdl.Locale)
	nowTime := time.Now()
	activeTime, err := time.Parse("2006-01-02 15:04:05", ActiveTimeStr)
	if err != nil {
		return "", err
	}
	// 计算时间差值
	sub := nowTime.Sub(activeTime)
	sub = sub + 8*time.Hour
	if int64(sub.Hours()) >= 24*30*12 {
		subStr = strconv.FormatInt(int64(sub.Hours())/(24*30*12), 10) + " " + i18nLocale.Get(i18n.I18nYearAgo)
	} else if int64(sub.Hours()) >= 24*30 {
		subStr = strconv.FormatInt(int64(sub.Hours())/(24*30), 10) + " " + i18nLocale.Get(i18n.I18nMonthAgo)
	} else if int64(sub.Hours()) >= 24 {
		subStr = strconv.FormatInt(int64(sub.Hours())/24, 10) + " " + i18nLocale.Get(i18n.I18nDayAgo)
	} else if int64(sub.Hours()) > 0 {
		subStr = strconv.FormatInt(int64(sub.Hours()), 10) + " " + i18nLocale.Get(i18n.I18nHourAgo)
	} else if int64(sub.Minutes()) > 0 {
		subStr = strconv.FormatInt(int64(sub.Minutes()), 10) + " " + i18nLocale.Get(i18n.I18nMinuteAgo)
	} else {
		subStr = i18nLocale.Get(i18n.I18nSecondAgo)
	}
	return subStr, nil
}
