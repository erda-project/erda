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

package apidocsvc

import (
	"encoding/json"
	"fmt"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/apim/bdl"
	"github.com/erda-project/erda/modules/apim/services/apierrors"
	"github.com/erda-project/erda/pkg/httpserver/errorresp"
	"github.com/erda-project/erda/pkg/swagger"
	"github.com/erda-project/erda/pkg/swagger/oas3"
	"github.com/erda-project/erda/pkg/swagger/oasconv"
)

const oas3Text = `{
  "openapi": "3.0.0",
  "info": {
    "title": "%s",
    "description": "# API 设计中心创建的 API 文档。\n\n请在『API 概况』中填写 API 文档的基本信息；在『API列表』新增接口描述；在『数据类型』中定义要引用的数据结构。\n",
    "version": "default"
  },
  "paths": {
    "/new-resource": {}
  },
  "tags": [
  	{
   		"name": "other"
  	}
 ]
}
`

func FetchAPIDocContent(orgID uint64, userID, inode string, specProtocol oasconv.Protocol) (*apistructs.FileTreeNodeRspData,
	*errorresp.APIError) {
	// 解析路径
	ft, err := bundle.NewGittarFileTree(inode)
	if err != nil {
		return nil, apierrors.GetNodeDetail.InternalError(err)
	}

	blob, err := bdl.Bdl.GetGittarBlobNodeInfo(ft.BlobPath(), strconv.FormatUint(orgID, 10))
	if err != nil {
		return nil, apierrors.GetNodeDetail.InternalError(err)
	}

	// gittar 仓库的文件是 .yaml, 根据参数要转换格式
	switch specProtocol {
	case oasconv.OAS3YAML:
	case oasconv.OAS3JSON:
		data, err := oasconv.YAMLToJSON([]byte(blob.Content))
		if err != nil {
			return nil, apierrors.GetNodeDetail.InternalError(errors.New("failed to convert doc to oas3-json"))
		}
		blob.Content = string(data)
		blob.Path = mustSuffix(blob.Path, ".json")
	case oasconv.OAS2YAML:
		v3, err := swagger.LoadFromData([]byte(blob.Content))
		if err == nil {
			return nil, apierrors.GetNodeDetail.InternalError(errors.New("failed to load oas3 from doc, content is invalid oas3 doc"))
		}
		v2, err := oasconv.OAS3ConvTo2(v3)
		if err != nil {
			return nil, apierrors.GetNodeDetail.InternalError(errors.Wrap(err, "failed to convert doc to oas2-yaml"))
		}
		data, err := json.Marshal(v2)
		if err != nil {
			return nil, apierrors.GetNodeDetail.InternalError(errors.Wrap(err, "failed to convert doc to oas2-yaml"))
		}
		data, err = oasconv.JSONToYAML(data)
		if err != nil {
			return nil, apierrors.GetNodeDetail.InternalError(errors.Wrap(err, "failed to convert doc to oas2-yaml"))
		}
		blob.Content = string(data)
		blob.Path = mustSuffix(blob.Path, ".yaml")
	case oasconv.OAS2JSON:
		v3, err := swagger.LoadFromData([]byte(blob.Content))
		if err == nil {
			return nil, apierrors.GetNodeDetail.InternalError(errors.Wrap(err, "failed to load oas3 from doc, content is invalid oas3 doc"))
		}
		v2, err := oasconv.OAS3ConvTo2(v3)
		if err != nil {
			return nil, apierrors.GetNodeDetail.InternalError(errors.Wrap(err, "failed to load oas3 from doc, content is invalid oas3 doc"))
		}
		data, err := json.Marshal(v2)
		if err != nil {
			return nil, apierrors.GetNodeDetail.InternalError(errors.New("failed to convert doc to oas2-json"))
		}
		blob.Content = string(data)
		blob.Path = mustSuffix(blob.Path, ".json")
	default:
		return nil, apierrors.GetNodeDetail.InvalidParameter(errors.Errorf("invalid specProtocol: %s", specProtocol))
	}

	rules, err := getRules(ft.ApplicationID())
	if err != nil {
		return nil, apierrors.GetNodeDetail.InvalidParameter(err)
	}

	meta := apistructs.APIDocMeta{
		Lock:     nil,
		Tree:     nil,
		Blob:     blob,
		ReadOnly: readOnly(orgID, userID, ft.ApplicationID(), ft.BranchName(), rules),
	}
	metaData, _ := json.Marshal(meta)

	var data = apistructs.FileTreeNodeRspData{
		Type:      "f",
		Inode:     inode,
		Pinode:    ft.Clone().DeletePathFromRepoRoot().Inode(),
		Scope:     "application",
		ScopeID:   ft.ApplicationID(),
		Name:      mustSuffix(path.Base(ft.PathFromRepoRoot()), ""),
		CreatorID: "",
		UpdaterID: "",
		Meta:      json.RawMessage(metaData),
	}

	return &data, nil
}

// 提交 API 文档内容到 gittar
// 如果文档内容为空, 则填充默认内容;
// 前端提交的文档格式为 json, 转换为 yaml 后再提交.
func CommitAPIDocContent(orgID uint64, userID, repo, commitMessage, serviceName, content, branch string) error {
	// 如果 content 为空, 给一个默认的 content
	if content == "" {
		content = defaultOAS3Content(serviceName)
	}

	// 前端传来的 content 是 json, 转换为 yaml
	data, err := oasconv.JSONToYAML([]byte(content))
	if err != nil {
		return errors.Wrap(err, "failed to JSONToYAML")
	}

	// 校验文档合法性 仅作反序列化校验
	if _, err := oas3.LoadFromData(data); err != nil {
		return err
	}

	content = string(data)

	// 统一更改文件名和路径后缀
	serviceName = mustSuffix(serviceName, suffixYaml)
	filenameFromRepoRoot := filepath.Join(apiDocsPathFromRepoRoot, serviceName)

	// 在 gittar 仓库对应分支下创建文档: .dice/apidocs/{docName}
	var commit = apistructs.GittarCreateCommitRequest{
		Message: commitMessage,
		Actions: []apistructs.EditActionItem{{
			Action:   actionAdd,
			Content:  content,
			Path:     filenameFromRepoRoot,
			PathType: "blob",
		}},
		Branch: branch,
	}

	resp, err := bdl.Bdl.CreateGittarCommitV3(orgID, userID, repo, &commit)
	if err != nil {
		return errors.Wrap(err, "failed to CreateGittarCommitV3")
	}
	if !resp.Success {
		return errors.Errorf("not success to CreateGittarCommitV3, err: %+v", resp.Error)
	}
	return nil
}

func mustSuffix(filename, suffix string) string {
	return strings.TrimSuffix(filename, filepath.Ext(filename)) + suffix
}

func matchSuffix(filename string, suffix ...string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	for _, s := range suffix {
		if s == ext {
			return true
		}
	}
	return false
}

func defaultOAS3Content(title string) string {
	return fmt.Sprintf(oas3Text, title)
}

func readOnly(orgID uint64, userID, appID, branchName string, rules []*apistructs.BranchRule) bool {
	// 没有分支规则 ==> 非只读
	if len(rules) == 0 {
		logrus.Infoln("len(rules) == 0")
		return false
	}

	// 当前用户有应用管理权限 ==> 非只读
	rolesSet := bdl.FetchRolesSet(orgID, userID)
	apps := rolesSet.RolesApps(bdl.AppMRoles...)
	for _, id := range apps {
		if id == appID {
			return false
		}
	}

	for _, rule := range rules {
		if rule.IsProtect {
			branches := strings.Split(rule.Rule, ",")
			for _, branchPattern := range branches {
				if match, _ := filepath.Match(branchPattern, branchName); match {
					return true
				}
			}
		}
	}

	return false
}

func getRules(applicationID string) ([]*apistructs.BranchRule, error) {
	// 查询文档是否只读
	appID, err := strconv.ParseUint(applicationID, 10, 64)
	if err != nil {
		return nil, errors.New("invalid inode: can not parse app id to uint")
	}
	rules, err := bdl.Bdl.GetAppBranchRules(appID)
	if err != nil {
		return nil, err
	}
	return rules, nil
}
