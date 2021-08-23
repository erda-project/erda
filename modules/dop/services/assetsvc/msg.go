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

package assetsvc

import (
	"fmt"
	"path"
	"strconv"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/dop/bdl"
	"github.com/erda-project/erda/modules/dop/conf"
	"github.com/erda-project/erda/modules/dop/services/uc"
	"github.com/erda-project/erda/pkg/strutil"
)

const (
	contractProveTitle    = "API 调用申请处理结果"
	contractProveTemplate = `
管理员处理了您的 API 调用申请，详细信息如下：

- API 名称：{{assetName}}
- 客户端名：{{clientDisplayName}}({{clientName}})
- 客户端ID：{{clientID}}
- **处理结果: {{result}}** 

[点击此处快速跳转至申请列表查看详情]({{path}})
`

	requestAccessTitle    = "API 调用申请处理"
	requestAccessTemplate = ` 
用户发起对 API 调用申请，{{do}}。
- 申请事项: {{requestItem}}
- 申请用户：{{userName}}
- API 名称: {{assetName}} 

[点击此处快速跳转至访问管理页面处理]({{path}})
`

	versionDeprecatedTitle    = "【注意】您申请调用的 API 版本已被停用"
	versionDeprecatedTemplate = `
您好，您正在调用的 API 版本已被停用，
请及时申请最新版本调用，如有疑问可联系 API 管理员进行具体咨询，谢谢！

详细信息：
- API 名称：{{assetName}}
- API 版本：{{apiVersion}} {{major}}.{{minor}}.{{patch}}
- 客户端名：{{clientDisplayName}}({{clientName}})
- 客户端ID ：{{clientID}}
- 更新状态：已停用
`

	versionReAvailableTitle    = "您申请调用的 API 版本已恢复可用"
	versionReAvailableTemplate = `
您好，您正在调用的 API 版本已经恢复为可用，
如已经更换为其他版本，请忽略此信息，谢谢！

详细信息：
- API 名称：{{assetName}}
- API 版本：{{apiVersion}} {{major}}.{{minor}}.{{patch}}
- 客户端名：{{clientDisplayName}}({{clientName}})
- 客户端ID ：{{clientID}}
更新状态：已恢复可用
`
)

type ApprovalResult string
type RequestItem string

// approval results
func ApprovalResultSLAUpdated(name string) ApprovalResult {
	if name == "" {
		name = "空"
	}
	return ApprovalResult(fmt.Sprintf("SLA 更变为 %s", name))
}

const ManagerProvedContract ApprovalResult = "同意调用"

func ApprovalResultWhileDelete(status apistructs.ContractStatus) ApprovalResult {
	if status == apistructs.ContractApproved {
		return "撤销调用权限"
	}
	return "拒绝调用"
}

func ApprovalResultFromStatus(status apistructs.ContractStatus) ApprovalResult {
	switch status {
	case apistructs.ContractApproving:
		return "正在审批"
	case apistructs.ContractApproved:
		return "同意调用"
	case apistructs.ContractDisapproved:
		return "拒绝调用"
	case apistructs.ContractUnapproved:
		return "撤销调用权限"
	default:
		return ""
	}
}

func ManagerRewriteSLA(name string) ApprovalResult {
	return ApprovalResult(fmt.Sprintf("管理员更新了 SLA *%s*, 请及时查看", name))
}

const ManagerDeleteSLA ApprovalResult = "管理员撤销了你正在使用的 SLA, 请及时查看"

// RequestItemSLA
func RequestItemSLA(name string) RequestItem {
	return RequestItem(fmt.Sprintf("申请使用名称为 %s 的 SLA", name))
}

func RequestItemAPI(name, swaggerVersion string) RequestItem {
	return RequestItem(fmt.Sprintf("申请调用 API %s %s", name, swaggerVersion))
}

// 异步发站内信和邮件, 通知 access 管理员
func (svc *Service) contractMsgToManager(orgID uint64, contractUserID string, asset *apistructs.APIAssetsModel, access *apistructs.APIAccessesModel, item RequestItem, autoed bool) {
	var username = contractUserID
	users, err := uc.GetUsers([]string{contractUserID, access.CreatorID})
	if err != nil {
		logrus.Errorf("failed to GetUsers, err: %v", err)
	}
	if user, ok := users[contractUserID]; ok {
		username = user.Nick
	}

	do := "请及时处理"
	if autoed {
		do = "已自动通过, 无须处理"
	}

	go func() {
		org, err := bdl.Bdl.GetOrg(orgID)
		if err != nil {
			logrus.Errorf("failed to GetOrg, err: %v", err)
			return
		}
		orgName := org.Name
		wildDomain := conf.WildDomain()
		host := fmt.Sprintf("%s-org.%s", orgName, wildDomain)
		pathName := accessDetailPath(access.ID)
		forwardURL := path.Join(host, pathName)
		params := map[string]string{
			"requestItem": string(item),
			"userName":    username,
			"assetName":   access.AssetName,
			"path":        "https://" + forwardURL,
			"do":          do,
		}
		if err := svc.EmailNotify(
			requestAccessTitle,
			requestAccessTemplate,
			params,
			"zh-CN",
			orgID,
			strutil.DedupSlice([]string{access.CreatorID, asset.CreatorID}),
		); err != nil {
			logrus.Errorf("failed to send email notify, err: %v", err)
		}
	}()
	go func() {
		params := map[string]string{
			"requestItem": string(item),
			"userName":    username,
			"assetName":   access.AssetName,
			"path":        accessDetailPath(access.ID),
			"do":          do,
		}
		if err := svc.MboxNotify(
			requestAccessTitle,
			requestAccessTemplate,
			params,
			"zh-CN",
			orgID,
			strutil.DedupSlice([]string{access.CreatorID, asset.CreatorID}),
		); err != nil {
			logrus.Errorf("failed to send mbox notify, err: %v", err)
		}
	}()
}

// 异步发送站内信和邮件, 通知调用申请人
func (svc *Service) contractMsgToUser(orgID uint64, contractUserID, assetName string, client *apistructs.ClientModel, result ApprovalResult) {

	go func() {
		org, err := bdl.Bdl.GetOrg(orgID)
		if err != nil {
			logrus.Errorf("failed to GetOrg, err: %v", err)
			return
		}
		orgName := org.Name
		wildDomain := conf.WildDomain()
		host := fmt.Sprintf("%s-org.%s", orgName, wildDomain)
		pathName := myClientsPath(client.ID)
		forwardURL := path.Join(host, pathName)
		if err := svc.EmailNotify(
			contractProveTitle,
			contractProveTemplate,
			map[string]string{
				"result":            string(result),
				"assetName":         assetName,
				"clientDisplayName": client.DisplayName,
				"clientName":        client.Name,
				"clientID":          client.ClientID,
				"path":              "https://" + forwardURL,
			},
			"zh-CN",
			orgID,
			[]string{contractUserID},
		); err != nil {
			logrus.Errorf("failed to send email notify, err: %v", err)
		}
	}()

	go func() {
		if err := svc.MboxNotify(
			contractProveTitle,
			contractProveTemplate,
			map[string]string{
				"result":            string(result),
				"assetName":         assetName,
				"clientDisplayName": client.DisplayName,
				"clientName":        client.Name,
				"clientID":          client.ClientID,
				"path":              myClientsPath(client.ID),
			},
			"zh-CN",
			orgID,
			[]string{contractUserID},
		); err != nil {
			logrus.Errorf("failed to send mbox notify, err: %v", err)
		}
	}()
}

func (svc *Service) updateVersionMsgToUser(orgID uint64, contractUserID, assetName string, version *apistructs.APIAssetVersionsModel, client *apistructs.ClientModel) {
	major := strconv.FormatUint(version.Major, 10)
	minor := strconv.FormatUint(version.Minor, 10)
	patch := strconv.FormatUint(version.Patch, 10)

	title := versionDeprecatedTitle
	template := versionDeprecatedTemplate
	if !version.Deprecated {
		title = versionReAvailableTitle
		template = versionReAvailableTemplate
	}

	go func() {
		if err := svc.EmailNotify(
			title,
			template,
			map[string]string{
				"assetName":         assetName,
				"apiVersion":        version.SwaggerVersion,
				"major":             major,
				"minor":             minor,
				"patch":             patch,
				"clientDisplayName": client.DisplayName,
				"clientName":        client.Name,
				"clientID":          client.ClientID,
			},
			"zh-CN",
			orgID,
			[]string{contractUserID},
		); err != nil {
			logrus.Errorf("failed to send email notify, err: %v", err)
		}
	}()

	go func() {
		if err := svc.MboxNotify(
			title,
			template,
			map[string]string{
				"assetName":         assetName,
				"apiVersion":        version.SwaggerVersion,
				"major":             major,
				"minor":             minor,
				"patch":             patch,
				"clientDisplayName": client.DisplayName,
				"clientName":        client.Name,
				"clientID":          client.ClientID,
			},
			"zh-CN",
			orgID,
			[]string{contractUserID},
		); err != nil {
			logrus.Errorf("failed to send email notify, err: %v", err)
		}
	}()
}
