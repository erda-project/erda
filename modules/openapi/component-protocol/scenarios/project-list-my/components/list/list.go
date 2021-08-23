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

package list

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
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

func RenItem(pro apistructs.ProjectDTO, orgDomain string) (ProItem, error) {
	activeTime, err := CountActiveTime(pro.ActiveTime)
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
	opExist := Operation{
		Key:     apistructs.ListProjectExistOperationKey.String(),
		Reload:  true,
		Text:    "退出",
		Confirm: "退出当前项目后，将不再有项目协作权限，如要再次加入需要项目管理员邀请，请确认是否退出？",
		Meta: Meta{
			ID: pro.ID,
		},
	}
	opApplyDeploy := Operation{
		Key:    apistructs.ApplyDeployProjectFilterOperation.String(),
		Reload: false,
		Text:   "申请部署",
		Meta: Meta{
			ProjectId:   pro.ID,
			ProjectName: pro.DisplayName,
		},
	}

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
					Text: "公开项目",
				},
				false: {
					Icon: "lock",
					Text: "私有项目",
				},
			}[pro.IsPublic],
			{
				Icon:    "application-one",
				Text:    strconv.Itoa(pro.Stats.CountApplications),
				Tooltip: "应用数",
			},
			{
				Icon:    "time",
				Text:    activeTime,
				Tooltip: pro.ActiveTime,
			},
		},
		Operations: map[string]interface{}{
			apistructs.ClickOperation.String():               opClick,
			apistructs.ListProjectExistOperationKey.String(): opExist,
		},
	}
	if pro.Logo != "" {
		item.PrefixImg = pro.Logo
	}
	// 解封状态
	if pro.BlockStatus == "unblocking" {
		item.ExtraInfos = append(item.ExtraInfos, ExtraInfos{
			Icon:    "link-cloud-faild",
			Text:    "解封处理中，请稍等",
			Type:    "warning",
			Tooltip: "解封处理中，请稍等",
		})
	} else if pro.BlockStatus == "unblocked" {
		item.ExtraInfos = append(item.ExtraInfos, ExtraInfos{
			Icon: "link-cloud-sucess",
			Text: "已解封",
			Type: "success",
		})
	}
	// 目前去掉管理按钮
	//// Owner Lead PM
	//if pro.CanManage {
	//	item.Operations[apistructs.ListProjectToManageOperationKey.String()] = opManage
	//}
	// 根据解封状态判断
	if pro.CanManage && (pro.BlockStatus == "unblocked" || pro.BlockStatus == "unblocking" || pro.BlockStatus == "blocked") {
		item.Operations[apistructs.ApplyDeployProjectFilterOperation.String()] = opApplyDeploy
	}
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
	}

	projectDTO, err := i.ctxBdl.Bdl.ListMyProject(i.ctxBdl.Identity.UserID, req)
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
			p, err := RenItem(v, org.Domain)
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

func CountActiveTime(ActiveTimeStr string) (string, error) {
	var subStr string
	nowTime := time.Now()
	activeTime, err := time.Parse("2006-01-02 15:04:05", ActiveTimeStr)
	if err != nil {
		return "", err
	}
	// 计算时间差值
	sub := nowTime.Sub(activeTime)
	sub = sub + 8*time.Hour
	if int64(sub.Hours()) >= 24*30*12 {
		subStr = strconv.FormatInt(int64(sub.Hours())/(24*30*12), 10) + " 年前"
	} else if int64(sub.Hours()) >= 24*30 {
		subStr = strconv.FormatInt(int64(sub.Hours())/(24*30), 10) + " 月前"
	} else if int64(sub.Hours()) >= 24 {
		subStr = strconv.FormatInt(int64(sub.Hours())/24, 10) + " 天前"
	} else if int64(sub.Hours()) > 0 {
		subStr = strconv.FormatInt(int64(sub.Hours()), 10) + " 小时前"
	} else if int64(sub.Minutes()) > 0 {
		subStr = strconv.FormatInt(int64(sub.Minutes()), 10) + " 分钟前"
	} else {
		subStr = "几秒前"
	}
	return subStr, nil
}

func (i *ComponentList) CheckVisible() (bool, error) {
	// 项目列表不为空
	if i.State.Total != 0 {
		return true, nil
	}
	// 没有加搜索条件
	if (i.State.Query != nil && i.State.Query["title"] == "") || i.State.Query == nil {
		return false, nil
	}
	// 如果加了搜索条件则需要去除搜索条件再判断一次
	orgID, err := strconv.ParseUint(i.ctxBdl.Identity.OrgID, 10, 64)
	if err != nil {
		return false, err
	}
	req := apistructs.ProjectListRequest{
		OrgID:    orgID,
		PageNo:   1,
		PageSize: 1,
		OrderBy:  "activeTime",
	}
	projectDTO, err := i.ctxBdl.Bdl.ListMyProject(i.ctxBdl.Identity.UserID, req)
	if err != nil {
		return false, err
	}
	if projectDTO == nil {
		return false, nil
	}
	if projectDTO.Total == 0 {
		return false, nil
	}

	return true, nil
}
