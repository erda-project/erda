package endpoints

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/cmdb/services/apierrors"
	"github.com/erda-project/erda/modules/pkg/user"
	"github.com/erda-project/erda/pkg/httpserver"
	"github.com/erda-project/erda/pkg/httpserver/errorresp"
	"github.com/erda-project/erda/pkg/httputil"
)

func (e *Endpoints) ListFileTreeNodes(ctx context.Context, r *http.Request, vars map[string]string) (
	httpserver.Responser, error) {

	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrListFileTreeNodes.NotLogin().ToResp(), nil
	}

	var req apistructs.UnifiedFileTreeNodeListRequest
	if err := e.queryStringDecoder.Decode(&req, r.URL.Query()); err != nil {
		return apierrors.ErrListFileTreeNodes.InvalidParameter(err).ToResp(), nil
	}
	req.IdentityInfo = identityInfo

	// 获取企业id
	orgID, err := getOrgId(r)
	if err != nil {
		return apierrors.ErrListFileTreeNodes.MissingParameter("org id").ToResp(), nil
	}

	nodes, err := e.fileTree.ListFileTreeNodes(req, orgID)
	if err != nil {
		return errorresp.ErrResp(err)
	}

	return httpserver.OkResp(nodes)
}

func (e *Endpoints) GetFileTreeNode(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrGetFileTreeNode.NotLogin().ToResp(), nil
	}

	req := apistructs.UnifiedFileTreeNodeGetRequest{}
	if err := e.queryStringDecoder.Decode(&req, r.URL.Query()); err != nil {
		return apierrors.ErrListFileTreeNodes.InvalidParameter(err).ToResp(), nil
	}
	req.Inode = vars["inode"]
	req.IdentityInfo = identityInfo

	// 获取企业id
	orgID, err := getOrgId(r)
	if err != nil {
		return apierrors.ErrGetFileTreeNode.MissingParameter("org id").ToResp(), nil
	}

	// TODO: 鉴权

	node, err := e.fileTree.GetFileTreeNode(req, orgID)
	if err != nil {
		return errorresp.ErrResp(err)
	}

	return httpserver.OkResp(node)
}

func (e *Endpoints) FuzzySearchFileTreeNodes(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrFuzzySearchFileTreeNodes.NotLogin().ToResp(), nil
	}

	var req apistructs.UnifiedFileTreeNodeFuzzySearchRequest
	if err := e.queryStringDecoder.Decode(&req, r.URL.Query()); err != nil {
		return apierrors.ErrFuzzySearchFileTreeNodes.InvalidParameter(err).ToResp(), nil
	}
	req.IdentityInfo = identityInfo

	// 获取企业id
	orgID, err := getOrgId(r)
	if err != nil {
		return apierrors.ErrGetFileTreeNode.MissingParameter("org id").ToResp(), nil
	}

	// TODO: 鉴权

	nodes, err := e.fileTree.FuzzySearchFileTreeNodes(req, orgID)
	if err != nil {
		return errorresp.ErrResp(err)
	}

	return httpserver.OkResp(nodes)
}

func (e *Endpoints) CreateFileTreeNode(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrCreateFileTreeNode.NotLogin().ToResp(), nil
	}

	// 校验 body 合法性
	if r.ContentLength == 0 {
		return apierrors.ErrCreateFileTreeNode.InvalidParameter("missing request body").ToResp(), nil
	}
	var req apistructs.UnifiedFileTreeNodeCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return apierrors.ErrCreateFileTreeNode.InvalidParameter(err).ToResp(), nil
	}
	req.IdentityInfo = identityInfo

	// 获取企业id
	orgID, err := getOrgId(r)
	if err != nil {
		return apierrors.ErrCreateFileTreeNode.MissingParameter("org id").ToResp(), nil
	}

	// TODO: 鉴权

	unifiedNode, err := e.fileTree.CreateFileTreeNode(req, orgID)
	if err != nil {
		return errorresp.ErrResp(err)
	}

	return httpserver.OkResp(unifiedNode)
}

func (e *Endpoints) DeleteFileTreeNode(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrDeleteFileTreeNode.NotLogin().ToResp(), nil
	}

	req := apistructs.UnifiedFileTreeNodeDeleteRequest{
		Inode:        vars["inode"],
		IdentityInfo: identityInfo,
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return apierrors.ErrCreateFileTreeNode.InvalidParameter(err).ToResp(), nil
	}

	// 获取企业id
	orgID, err := getOrgId(r)
	if err != nil {
		return apierrors.ErrDeleteFileTreeNode.MissingParameter("org id").ToResp(), nil
	}

	// TODO: 鉴权

	unifiedNode, err := e.fileTree.DeleteFileTreeNode(req, orgID)
	if err != nil {
		return errorresp.ErrResp(err)
	}

	return httpserver.OkResp(unifiedNode)
}

func (e *Endpoints) FindFileTreeNodeAncestors(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrFindFileTreeNodeAncestors.NotLogin().ToResp(), nil
	}

	// 获取企业id
	orgID, err := getOrgId(r)
	if err != nil {
		return apierrors.ErrDeleteFileTreeNode.MissingParameter("org id").ToResp(), nil
	}

	// TODO: 鉴权

	req := apistructs.UnifiedFileTreeNodeFindAncestorsRequest{
		Inode:        vars["inode"],
		IdentityInfo: identityInfo,
	}

	if err := e.queryStringDecoder.Decode(&req, r.URL.Query()); err != nil {
		return apierrors.ErrListFileTreeNodes.InvalidParameter(err).ToResp(), nil
	}
	ancestors, err := e.fileTree.FindFileTreeNodeAncestors(req, orgID)
	if err != nil {
		return errorresp.ErrResp(err)
	}

	return httpserver.OkResp(ancestors)
}

func getOrgId(r *http.Request) (uint64, error) {
	orgIDStr := r.Header.Get(httputil.OrgHeader)
	if orgIDStr == "" {
		return 0, apierrors.ErrListFileTreeNodes.MissingParameter("org id")
	}
	orgID, err := strconv.ParseInt(orgIDStr, 10, 64)
	if err != nil {
		return 0, apierrors.ErrListFileTreeNodes.InvalidParameter(err)
	}
	return uint64(orgID), nil
}
