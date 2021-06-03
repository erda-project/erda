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

package bundle

import (
	"fmt"
	"strconv"
	"sync"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle/apierrors"
	"github.com/erda-project/erda/pkg/http/httpclient"
)

// gittar
func (b *Bundle) ListGittarFileTreeNodes(req apistructs.UnifiedFileTreeNodeListRequest, orgID uint64) (results []apistructs.UnifiedFileTreeNode, err error) {
	var request *httpclient.Request

	host, err := b.urls.GittarAdaptor()
	if err != nil {
		return nil, err
	}
	request = b.hc.Get(host).Path("/api/cicd-pipeline/filetree")

	var response apistructs.UnifiedFileTreeNodeListResponse
	resp, err := request.
		Param("scope", req.Scope).
		Param("scopeID", req.ScopeID).
		Param("pinode", req.Pinode).
		Header("Accept", "application/json").
		Header("Org-ID", strconv.Itoa(int(orgID))).
		Header("User-ID", req.UserID).
		Do().JSON(&response)
	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}
	if !resp.IsOK() || !response.Success {
		return nil, toAPIError(resp.StatusCode(), response.Error)
	}
	return response.Data, nil
}

func (b *Bundle) GetGittarFileTreeNode(req apistructs.UnifiedFileTreeNodeGetRequest, orgID uint64) (result *apistructs.UnifiedFileTreeNode, err error) {
	var request *httpclient.Request

	host, err := b.urls.GittarAdaptor()
	if err != nil {
		return nil, err
	}
	request = b.hc.Get(host).Path(fmt.Sprintf("/api/cicd-pipeline/filetree/%s", req.Inode))

	var response apistructs.UnifiedFileTreeNodeGetResponse
	resp, err := request.
		Header("Accept", "application/json").
		Header("Org-ID", strconv.Itoa(int(orgID))).
		Header("User-ID", req.UserID).
		Do().JSON(&response)
	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}
	if !resp.IsOK() || !response.Success {
		return nil, toAPIError(resp.StatusCode(), response.Error)
	}
	return response.Data, nil
}

func (b *Bundle) FuzzySearchGittarFileTreeNodes(query apistructs.UnifiedFileTreeNodeFuzzySearchRequest, orgID uint64) (result []apistructs.UnifiedFileTreeNode, err error) {

	if query.Scope != "" && query.Scope != apistructs.FileTreeScopeProjectApp {
		return nil, nil
	}

	var request *httpclient.Request
	var response apistructs.UnifiedFileTreeNodeFuzzySearchResponse

	req := query
	if req.Scope == "" {
		req.Scope = apistructs.FileTreeScopeProjectApp
	}

	host, err := b.urls.GittarAdaptor()
	if err != nil {
		return nil, err
	}
	request = b.hc.Get(host).Path("/api/cicd-pipeline/filetree/actions/fuzzy-search")

	resp, err := request.
		Param("scope", req.Scope).
		Param("scopeID", req.ScopeID).
		Param("fuzzy", req.Fuzzy).
		Header("Accept", "application/json").
		Header("Org-ID", strconv.Itoa(int(orgID))).
		Header("User-ID", req.UserID).
		Do().JSON(&response)
	if err != nil {
		return nil, err
	}
	if !resp.IsOK() || !response.Success {
		return nil, toAPIError(resp.StatusCode(), response.Error)
	}

	if response.Data == nil || len(response.Data) <= 0 {
		return
	}

	result = append(result, response.Data...)

	return result, nil
}

func (b *Bundle) CreateGittarFileTreeNodes(req apistructs.UnifiedFileTreeNodeCreateRequest, orgID uint64) (result *apistructs.UnifiedFileTreeNode, err error) {
	var request *httpclient.Request

	host, err := b.urls.GittarAdaptor()
	if err != nil {
		return nil, err
	}
	request = b.hc.Post(host).Path("/api/cicd-pipeline/filetree")

	var response apistructs.UnifiedFileTreeNodeCreateResponse
	resp, err := request.
		Header("Accept", "application/json").
		Header("Org-ID", strconv.Itoa(int(orgID))).
		Header("User-ID", req.UserID).
		JSONBody(req).
		Do().JSON(&response)
	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}
	if !resp.IsOK() || !response.Success {
		return nil, toAPIError(resp.StatusCode(), response.Error)
	}
	return response.Data, nil
}

func (b *Bundle) DeleteGittarFileTreeNodes(req apistructs.UnifiedFileTreeNodeDeleteRequest, orgID uint64) (result *apistructs.UnifiedFileTreeNode, err error) {
	var request *httpclient.Request

	host, err := b.urls.GittarAdaptor()
	if err != nil {
		return nil, err
	}
	request = b.hc.Delete(host).Path(fmt.Sprintf("/api/cicd-pipeline/filetree/%s", req.Inode))

	var response apistructs.UnifiedFileTreeNodeDeleteResponse
	resp, err := request.
		Header("Accept", "application/json").
		Header("Org-ID", strconv.Itoa(int(orgID))).
		Header("User-ID", req.UserID).
		Do().JSON(&response)
	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}
	if !resp.IsOK() || !response.Success {
		return nil, toAPIError(resp.StatusCode(), response.Error)
	}
	return response.Data, nil
}

func (b *Bundle) FindGittarFileTreeNodeAncestors(req apistructs.UnifiedFileTreeNodeFindAncestorsRequest, orgID uint64) (result []apistructs.UnifiedFileTreeNode, err error) {

	var request *httpclient.Request
	var response apistructs.UnifiedFileTreeNodeFindAncestorsResponse

	host, err := b.urls.GittarAdaptor()
	if err != nil {
		return nil, err
	}
	request = b.hc.Get(host).Path(fmt.Sprintf("/api/cicd-pipeline/filetree/%s/actions/find-ancestors", req.Inode))

	resp, err := request.
		Header("Accept", "application/json").
		Header("Org-ID", strconv.Itoa(int(orgID))).
		Header("User-ID", req.UserID).
		Do().JSON(&response)
	if err != nil {
		return nil, err
	}
	if !resp.IsOK() || !response.Success {
		return nil, toAPIError(resp.StatusCode(), response.Error)
	}

	if response.Data == nil || len(response.Data) <= 0 {
		return
	}

	result = append(result, response.Data...)

	return result, nil
}

// qa
func (b *Bundle) ListQaFileTreeNodes(req apistructs.UnifiedFileTreeNodeListRequest, orgID uint64) (results []apistructs.UnifiedFileTreeNode, err error) {
	var request *httpclient.Request

	host, err := b.urls.QA()
	if err != nil {
		return nil, err
	}
	request = b.hc.Get(host).Path("/api/autotests/filetree")

	var response apistructs.UnifiedFileTreeNodeListResponse
	resp, err := request.
		Param("scope", req.Scope).
		Param("scopeID", req.ScopeID).
		Param("pinode", req.Pinode).
		Header("Accept", "application/json").
		Header("Org-ID", strconv.Itoa(int(orgID))).
		Header("User-ID", req.UserID).
		Do().JSON(&response)
	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}
	if !resp.IsOK() || !response.Success {
		return nil, toAPIError(resp.StatusCode(), response.Error)
	}
	return response.Data, nil
}

func (b *Bundle) GetQaFileTreeNode(req apistructs.UnifiedFileTreeNodeGetRequest, orgID uint64) (result *apistructs.UnifiedFileTreeNode, err error) {
	var request *httpclient.Request

	host, err := b.urls.QA()
	if err != nil {
		return nil, err
	}
	request = b.hc.Get(host).Path(fmt.Sprintf("/api/autotests/filetree/%s", req.Inode))

	var response apistructs.UnifiedFileTreeNodeGetResponse
	resp, err := request.
		Header("Accept", "application/json").
		Header("Org-ID", strconv.Itoa(int(orgID))).
		Header("User-ID", req.UserID).
		Do().JSON(&response)
	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}
	if !resp.IsOK() || !response.Success {
		return nil, toAPIError(resp.StatusCode(), response.Error)
	}
	return response.Data, nil
}

func (b *Bundle) FuzzySearchQaFileTreeNodes(query apistructs.UnifiedFileTreeNodeFuzzySearchRequest, orgID uint64) (result []apistructs.UnifiedFileTreeNode, err error) {
	var warn error
	var wait sync.WaitGroup

	for _, scope := range apistructs.AllScope {
		if scope == apistructs.FileTreeScopeProjectApp || (query.Scope != "" && scope != query.Scope) {
			continue
		}
		wait.Add(1)
		go func(scope string) {
			defer wait.Done()

			var request *httpclient.Request
			var response apistructs.UnifiedFileTreeNodeFuzzySearchResponse

			req := query
			req.Scope = scope

			host, err := b.urls.QA()
			if err != nil {
				warn = err
				return
			}
			request = b.hc.Get(host).Path("/api/autotests/filetree/actions/fuzzy-search")

			resp, err := request.
				Param("scope", req.Scope).
				Param("scopeID", req.ScopeID).
				Param("fuzzy", req.Fuzzy).
				Header("Accept", "application/json").
				Header("Org-ID", strconv.Itoa(int(orgID))).
				Header("User-ID", req.UserID).
				Do().JSON(&response)
			if err != nil {
				warn = err
				return
			}

			if !resp.IsOK() || !response.Success {
				warn = toAPIError(resp.StatusCode(), response.Error)
				return
			}

			if response.Data == nil || len(response.Data) <= 0 {
				return
			}

			result = append(result, response.Data...)
		}(scope)
	}

	wait.Wait()
	if warn != nil {
		return nil, warn
	}
	return result, nil
}

func (b *Bundle) CreateQaFileTreeNodes(req apistructs.UnifiedFileTreeNodeCreateRequest, orgID uint64) (result *apistructs.UnifiedFileTreeNode, err error) {
	var request *httpclient.Request

	host, err := b.urls.QA()
	if err != nil {
		return nil, err
	}
	request = b.hc.Post(host).Path("/api/autotests/filetree")

	var response apistructs.UnifiedFileTreeNodeCreateResponse
	resp, err := request.
		Header("Accept", "application/json").
		Header("Org-ID", strconv.Itoa(int(orgID))).
		Header("User-ID", req.UserID).
		JSONBody(req).
		Do().JSON(&response)
	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}
	if !resp.IsOK() || !response.Success {
		return nil, toAPIError(resp.StatusCode(), response.Error)
	}
	return response.Data, nil
}

func (b *Bundle) DeleteQaFileTreeNodes(req apistructs.UnifiedFileTreeNodeDeleteRequest, orgID uint64) (result *apistructs.UnifiedFileTreeNode, err error) {
	var request *httpclient.Request

	host, err := b.urls.QA()
	if err != nil {
		return nil, err
	}
	request = b.hc.Delete(host).Path(fmt.Sprintf("/api/autotests/filetree/%s", req.Inode))

	var response apistructs.UnifiedFileTreeNodeDeleteResponse
	resp, err := request.
		Header("Accept", "application/json").
		Header("Org-ID", strconv.Itoa(int(orgID))).
		Header("User-ID", req.UserID).
		Do().JSON(&response)
	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}
	if !resp.IsOK() || !response.Success {
		return nil, toAPIError(resp.StatusCode(), response.Error)
	}
	return response.Data, nil
}

func (b *Bundle) FindQaFileTreeNodeAncestors(req apistructs.UnifiedFileTreeNodeFindAncestorsRequest, orgID uint64) (result []apistructs.UnifiedFileTreeNode, err error) {

	var request *httpclient.Request
	var response apistructs.UnifiedFileTreeNodeFindAncestorsResponse

	host, err := b.urls.QA()
	if err != nil {
		return nil, err
	}
	request = b.hc.Get(host).Path(fmt.Sprintf("/api/autotests/filetree/%s/actions/find-ancestors", req.Inode))

	resp, err := request.
		Header("Accept", "application/json").
		Header("Org-ID", strconv.Itoa(int(orgID))).
		Header("User-ID", req.UserID).
		Do().JSON(&response)
	if err != nil {
		return nil, err
	}
	if !resp.IsOK() || !response.Success {
		return nil, toAPIError(resp.StatusCode(), response.Error)
	}

	if response.Data == nil || len(response.Data) <= 0 {
		return
	}

	result = append(result, response.Data...)

	return result, nil
}

// project

func (b *Bundle) ListFileTreeNodes(req apistructs.UnifiedFileTreeNodeListRequest, orgID uint64) (results []apistructs.UnifiedFileTreeNode, err error) {
	switch req.Scope {
	case apistructs.FileTreeScopeProjectApp:
		return b.ListGittarFileTreeNodes(req, orgID)
	case apistructs.FileTreeScopeProject, apistructs.FileTreeScopeAutoTestPlan, apistructs.FileTreeScopeAutoTestConfigSheet, apistructs.FileTreeScopeAutoTest:
		return b.ListQaFileTreeNodes(req, orgID)
	}

	return nil, fmt.Errorf("not find this scope %s", req.Scope)
}

func (b *Bundle) GetFileTreeNode(req apistructs.UnifiedFileTreeNodeGetRequest, orgID uint64) (result *apistructs.UnifiedFileTreeNode, err error) {
	switch req.Scope {
	case apistructs.FileTreeScopeProjectApp:
		return b.GetGittarFileTreeNode(req, orgID)
	case apistructs.FileTreeScopeProject, apistructs.FileTreeScopeAutoTestPlan, apistructs.FileTreeScopeAutoTestConfigSheet, apistructs.FileTreeScopeAutoTest:
		return b.GetQaFileTreeNode(req, orgID)
	}

	return nil, fmt.Errorf("not find this scope %s", req.Scope)
}

func (b *Bundle) FuzzySearchFileTreeNodes(query apistructs.UnifiedFileTreeNodeFuzzySearchRequest, orgID uint64) (result []apistructs.UnifiedFileTreeNode, err error) {

	gittarList, err := b.FuzzySearchGittarFileTreeNodes(query, orgID)
	if err != nil {
		return nil, err
	}
	if gittarList != nil && len(gittarList) >= 0 {
		result = append(result, gittarList...)
	}

	qaList, err := b.FuzzySearchQaFileTreeNodes(query, orgID)
	if err != nil {
		return nil, err
	}

	if qaList != nil && len(qaList) >= 0 {
		result = append(result, qaList...)
	}

	return result, nil
}

func (b *Bundle) CreateFileTreeNodes(req apistructs.UnifiedFileTreeNodeCreateRequest, orgID uint64) (result *apistructs.UnifiedFileTreeNode, err error) {
	switch req.Scope {
	case apistructs.FileTreeScopeProjectApp:
		return b.CreateGittarFileTreeNodes(req, orgID)
	case apistructs.FileTreeScopeProject, apistructs.FileTreeScopeAutoTestPlan, apistructs.FileTreeScopeAutoTestConfigSheet, apistructs.FileTreeScopeAutoTest:
		return b.CreateQaFileTreeNodes(req, orgID)
	}

	return nil, fmt.Errorf("not find this scope %s", req.Scope)
}

func (b *Bundle) DeleteFileTreeNodes(req apistructs.UnifiedFileTreeNodeDeleteRequest, orgID uint64) (result *apistructs.UnifiedFileTreeNode, err error) {
	switch req.Scope {
	case apistructs.FileTreeScopeProjectApp:
		return b.DeleteGittarFileTreeNodes(req, orgID)
	case apistructs.FileTreeScopeProject, apistructs.FileTreeScopeAutoTestPlan, apistructs.FileTreeScopeAutoTestConfigSheet, apistructs.FileTreeScopeAutoTest:
		return b.DeleteQaFileTreeNodes(req, orgID)
	}

	return nil, fmt.Errorf("not find this scope %s", req.Scope)
}

func (b *Bundle) FindFileTreeNodeAncestors(req apistructs.UnifiedFileTreeNodeFindAncestorsRequest, orgID uint64) (result []apistructs.UnifiedFileTreeNode, err error) {
	switch req.Scope {
	case apistructs.FileTreeScopeProjectApp:
		return b.FindGittarFileTreeNodeAncestors(req, orgID)
	case apistructs.FileTreeScopeProject, apistructs.FileTreeScopeAutoTestPlan, apistructs.FileTreeScopeAutoTestConfigSheet, apistructs.FileTreeScopeAutoTest:
		return b.FindQaFileTreeNodeAncestors(req, orgID)
	}

	return nil, fmt.Errorf("not find this scope %s", req.Scope)
}
