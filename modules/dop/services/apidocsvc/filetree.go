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

package apidocsvc

import (
	"encoding/json"
	"strconv"

	"github.com/pkg/errors"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/dop/services/apierrors"
	"github.com/erda-project/erda/pkg/http/httpserver/errorresp"
)

const (
	TreeNameAPIDocs = "api-docs"
	TreeNameSchemas = "schemas"
)

const (
	apiDocsPathFromRepoRoot    = ".dice/apidocs"
	migrationsPathFromRepoRoot = ".dice/migrations"
	suffixYml                  = ".yml"
	suffixYaml                 = ".yaml"
)

const (
	actionAdd    = "add"
	actionDelete = "delete"
)

func (svc *Service) CreateNode(req *apistructs.APIDocCreateNodeReq) (*apistructs.FileTreeNodeRspData, *errorresp.APIError) {
	if req.Body.Type != apistructs.NodeTypeFile {
		return nil, apierrors.CreateNode.InvalidParameter("节点类型错误, 只能创建文件类型(f)的节点")
	}

	var meta apistructs.CreateAPIDocMeta
	_ = json.Unmarshal(req.Body.Meta, &meta)

	// 主流程在 createDoc 中完成
	return svc.createDoc(req.OrgID, req.Identity.UserID, req.Body.Pinode, req.Body.Name, meta.Content)
}

func (svc *Service) DeleteNode(req *apistructs.APIDocDeleteNodeReq) *errorresp.APIError {
	switch req.URIParams.TreeName {
	case TreeNameAPIDocs:
		return svc.deleteAPIDoc(req.OrgID, req.Identity.UserID, req.URIParams.Inode)
	default:
		return apierrors.DeleteNode.NotFound()
	}
}

func (svc *Service) UpdateNode(req *apistructs.APIDocUpdateNodeReq) (*apistructs.FileTreeNodeRspData, *errorresp.APIError) {
	switch req.URIParams.TreeName {
	case TreeNameAPIDocs:
		return svc.renameAPIDoc(req.OrgID, req.Identity.UserID, req.URIParams.Inode, req.Body.Name)
	default:
		return nil, apierrors.UpdateNode.NotFound()
	}
}

func (svc *Service) MoveNode(req *apistructs.APIDocMvCpNodeReq) (*apistructs.FileTreeNodeRspData, *errorresp.APIError) {
	switch req.URIParams.TreeName {
	case TreeNameAPIDocs:
		return svc.moveAPIDco(req.OrgID, req.Identity.UserID, req.URIParams.Inode, req.Body.Pinode)
	default:
		return nil, apierrors.MoveNode.NotFound()
	}
}

func (svc *Service) CopyNode(req *apistructs.APIDocMvCpNodeReq) (*apistructs.FileTreeNodeRspData, *errorresp.APIError) {
	switch req.URIParams.TreeName {
	case TreeNameAPIDocs:
		return svc.copyAPIDoc(req.OrgID, req.Identity.UserID, req.URIParams.Inode, req.Body.Pinode)
	default:
		return nil, apierrors.CopyNode.NotFound()
	}
}

// 获取子节点列表
func (svc *Service) ListChildren(req *apistructs.APIDocListChildrenReq) ([]*apistructs.FileTreeNodeRspData, *errorresp.APIError) {
	// 对于 pinode == 0 的情况, 直接列出应用下的分支, 不用管要查找的是 API 文档还是 schema
	switch {
	case req.QueryParams.Pinode == "0" && req.QueryParams.Scope == "application":
		appID, err := strconv.ParseUint(req.QueryParams.ScopeID, 10, 64)
		if err != nil {
			return nil, apierrors.ListChildrenNodes.InvalidParameter("invalid appID")
		}
		return svc.listBranches(req.OrgID, appID, req.Identity.UserID)

	case req.QueryParams.Pinode == "0":
		return nil, apierrors.ListChildrenNodes.InvalidParameter(errors.Errorf("scope 错误, 本目录树仅支持应用层级, scope: %s", req.QueryParams.Scope))
	}

	// 对于 pinode != 0 的情况, 列出应用下的 "服务"
	// ps: 目前没有项目级的目录树, 所以不必讨论节点的层级, 直接查叶子节点即可; 以后有项目级目录树后, 需要讨论节点层级
	switch req.URIParams.TreeName {
	case TreeNameAPIDocs:
		return svc.listAPIDocs(req.OrgID, req.Identity.UserID, req.QueryParams.Pinode)

	case TreeNameSchemas:
		return svc.listSchemas(req.OrgID, req.Identity.UserID, req.QueryParams.Pinode)

	default:
		return nil, apierrors.ListChildrenNodes.NotFound()
	}
}

func (svc *Service) GetNodeDetail(req *apistructs.APIDocNodeDetailReq) (*apistructs.FileTreeNodeRspData, *errorresp.APIError) {
	switch req.URIParams.TreeName {
	case TreeNameAPIDocs:
		return svc.getAPIDocContent(req.OrgID, req.Identity.UserID, req.URIParams.Inode)

	case TreeNameSchemas:
		return svc.getSchemaContent(req.OrgID, req.URIParams.Inode)

	default:
		return nil, apierrors.GetNodeDetail.NotFound()
	}
}
