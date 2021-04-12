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

package endpoints

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/gittar-adaptor/service/apierrors"
	"github.com/erda-project/erda/modules/pkg/user"
	"github.com/erda-project/erda/pkg/httpserver"
	"github.com/erda-project/erda/pkg/httpserver/errorresp"
	"github.com/erda-project/erda/pkg/httputil"
)

func (e *Endpoints) ListGittarFileTreeNodes(ctx context.Context, r *http.Request, vars map[string]string) (
	httpserver.Responser, error) {

	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrListGittarFileTreeNodes.NotLogin().ToResp(), nil
	}

	var req apistructs.UnifiedFileTreeNodeListRequest
	if err := e.queryStringDecoder.Decode(&req, r.URL.Query()); err != nil {
		return apierrors.ErrListGittarFileTreeNodes.InvalidParameter(err).ToResp(), nil
	}
	req.IdentityInfo = identityInfo

	// 获取企业id
	orgID, err := getOrgId(r)
	if err != nil {
		return apierrors.ErrListGittarFileTreeNodes.MissingParameter("org id").ToResp(), nil
	}

	nodes, err := e.fileTree.ListFileTreeNodes(req, orgID)
	if err != nil {
		return errorresp.ErrResp(err)
	}

	return httpserver.OkResp(nodes)
}

func (e *Endpoints) GetGittarFileTreeNode(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrGetGittarFileTreeNode.NotLogin().ToResp(), nil
	}

	req := apistructs.UnifiedFileTreeNodeGetRequest{
		Inode:        vars["inode"],
		IdentityInfo: identityInfo,
	}

	// 获取企业id
	orgID, err := getOrgId(r)
	if err != nil {
		return apierrors.ErrGetGittarFileTreeNode.MissingParameter("org id").ToResp(), nil
	}

	// TODO: 鉴权

	node, err := e.fileTree.GetFileTreeNode(req, orgID)
	if err != nil {
		return errorresp.ErrResp(err)
	}

	return httpserver.OkResp(node)
}

func (e *Endpoints) GetGittarFileByPipelineId(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	pipelineId := r.URL.Query().Get("pipelineId")
	pipelineIdInt, err := strconv.Atoi(pipelineId)
	if err != nil {
		return errorresp.ErrResp(err)
	}

	_, err = user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrFuzzySearchGittarFileTreeNodes.NotLogin().ToResp(), nil
	}

	// 获取企业id
	orgID, err := getOrgId(r)
	if err != nil {
		return apierrors.ErrGetGittarFileTreeNode.MissingParameter("org id").ToResp(), nil
	}
	node, err := e.fileTree.GetGittarFileByPipelineId(uint64(pipelineIdInt), orgID)
	if err != nil {
		return errorresp.ErrResp(err)
	}
	return httpserver.OkResp(node)
}

func (e *Endpoints) FuzzySearchGittarFileTreeNodes(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrFuzzySearchGittarFileTreeNodes.NotLogin().ToResp(), nil
	}

	var req apistructs.UnifiedFileTreeNodeFuzzySearchRequest
	if err := e.queryStringDecoder.Decode(&req, r.URL.Query()); err != nil {
		return apierrors.ErrFuzzySearchGittarFileTreeNodes.InvalidParameter(err).ToResp(), nil
	}
	req.IdentityInfo = identityInfo

	// 获取企业id
	orgID, err := getOrgId(r)
	if err != nil {
		return apierrors.ErrGetGittarFileTreeNode.MissingParameter("org id").ToResp(), nil
	}

	// TODO: 鉴权

	nodes, err := e.fileTree.FuzzySearchFileTreeNodes(req, orgID)
	if err != nil {
		return errorresp.ErrResp(err)
	}

	return httpserver.OkResp(nodes)
}

func getOrgId(r *http.Request) (uint64, error) {
	orgIDStr := r.Header.Get(httputil.OrgHeader)
	if orgIDStr == "" {
		return 0, apierrors.ErrListGittarFileTreeNodes.MissingParameter("org id")
	}
	orgID, err := strconv.ParseInt(orgIDStr, 10, 64)
	if err != nil {
		return 0, apierrors.ErrListGittarFileTreeNodes.InvalidParameter(err)
	}
	return uint64(orgID), nil
}

func (e *Endpoints) CreateGittarFileTreeNode(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrCreateGittarFileTreeNode.NotLogin().ToResp(), nil
	}

	// 校验 body 合法性
	if r.ContentLength == 0 {
		return apierrors.ErrCreateGittarFileTreeNode.InvalidParameter("missing request body").ToResp(), nil
	}
	var req apistructs.UnifiedFileTreeNodeCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return apierrors.ErrCreateGittarFileTreeNode.InvalidParameter(err).ToResp(), nil
	}
	req.IdentityInfo = identityInfo

	// 获取企业id
	orgID, err := getOrgId(r)
	if err != nil {
		return apierrors.ErrCreateGittarFileTreeNode.MissingParameter("org id").ToResp(), nil
	}

	// TODO: 鉴权

	unifiedNode, err := e.fileTree.CreateFileTreeNode(req, orgID)
	if err != nil {
		return errorresp.ErrResp(err)
	}

	return httpserver.OkResp(unifiedNode)
}

func (e *Endpoints) DeleteGittarFileTreeNode(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrDeleteGittarFileTreeNode.NotLogin().ToResp(), nil
	}

	req := apistructs.UnifiedFileTreeNodeDeleteRequest{
		Inode:        vars["inode"],
		IdentityInfo: identityInfo,
	}

	// 获取企业id
	orgID, err := getOrgId(r)
	if err != nil {
		return apierrors.ErrCreateGittarFileTreeNode.MissingParameter("org id").ToResp(), nil
	}

	// TODO: 鉴权

	unifiedNode, err := e.fileTree.DeleteFileTreeNode(req, orgID)
	if err != nil {
		return errorresp.ErrResp(err)
	}

	return httpserver.OkResp(unifiedNode)
}

func (e *Endpoints) FindGittarFileTreeNodeAncestors(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrFindGittarFileTreeNodeAncestors.NotLogin().ToResp(), nil
	}

	// TODO: 鉴权

	req := apistructs.UnifiedFileTreeNodeFindAncestorsRequest{
		Inode:        vars["inode"],
		IdentityInfo: identityInfo,
	}
	ancestors, err := e.fileTree.FindFileTreeNodeAncestors(req)
	if err != nil {
		return errorresp.ErrResp(err)
	}

	return httpserver.OkResp(ancestors)
}
