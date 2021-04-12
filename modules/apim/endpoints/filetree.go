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

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/apim/services/apidocsvc"
	"github.com/erda-project/erda/modules/apim/services/apierrors"
	"github.com/erda-project/erda/modules/pkg/user"
	"github.com/erda-project/erda/pkg/httpserver"
)

func (e *Endpoints) CreateNode(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	identity, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.CreateNode.NotLogin().ToResp(), nil
	}
	orgID, err := user.GetOrgID(r)
	if err != nil {
		return apierrors.CreateNode.MissingParameter("Org-ID").ToResp(), nil
	}

	treeName, ok := vars["treeName"]
	if !ok || treeName != apidocsvc.TreeNameAPIDocs {
		return apierrors.CreateNode.NotFound().ToResp(), nil
	}

	var (
		req  apistructs.APIDocCreateNodeReq
		body apistructs.APIDocCreateUpdateNodeBody
	)
	if err = json.NewDecoder(r.Body).Decode(&body); err != nil {
		return apierrors.CreateNode.InvalidParameter(err).ToResp(), nil
	}

	req.Identity = &identity
	req.OrgID = orgID
	req.Body = &body
	req.URIParams = &apistructs.FileTreeDetailURI{TreeName: apidocsvc.TreeNameAPIDocs}

	data, err2 := e.fileTreeSvc.CreateNode(&req)
	if err2 != nil {
		return err2.ToResp(), nil
	}

	return httpserver.OkResp(data)
}

func (e *Endpoints) DeleteNode(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	identity, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.DeleteNode.NotLogin().ToResp(), nil
	}
	orgID, err := user.GetOrgID(r)
	if err != nil {
		return apierrors.DeleteNode.MissingParameter("Org-ID").ToResp(), nil
	}

	if treeName, ok := vars["treeName"]; !ok || treeName != apidocsvc.TreeNameAPIDocs {
		return apierrors.DeleteNode.ToResp(), nil
	}

	var (
		uriParams = apistructs.FileTreeDetailURI{
			TreeName: apidocsvc.TreeNameAPIDocs,
			Inode:    vars["inode"],
		}
		req = apistructs.APIDocDeleteNodeReq{
			OrgID:     orgID,
			Identity:  &identity,
			URIParams: &uriParams,
		}
	)

	if err2 := e.fileTreeSvc.DeleteNode(&req); err2 != nil {
		return err2.ToResp(), nil
	}

	return httpserver.OkResp(nil)
}

func (e *Endpoints) UpdateNode(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	identity, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.UpdateNode.NotLogin().ToResp(), nil
	}
	orgID, err := user.GetOrgID(r)
	if err != nil {
		return apierrors.UpdateNode.MissingParameter("Org-ID").ToResp(), nil
	}

	treeName, ok := vars["treeName"]
	if !ok || treeName != apidocsvc.TreeNameAPIDocs {
		return apierrors.UpdateNode.NotFound().ToResp(), nil
	}

	inode, ok := vars["inode"]
	if !ok || len(inode) == 0 {
		return apierrors.UpdateNode.NotFound().ToResp(), nil
	}

	var body apistructs.RenameAPIDocBody
	if err = json.NewDecoder(r.Body).Decode(&body); err != nil {
		return apierrors.UpdateNode.InvalidParameter(err).ToResp(), nil
	}

	var (
		uri = apistructs.FileTreeDetailURI{
			TreeName: treeName,
			Inode:    inode,
		}
		req = apistructs.APIDocUpdateNodeReq{
			OrgID:     orgID,
			Identity:  &identity,
			URIParams: &uri,
			Body:      &body,
		}
	)

	data, err2 := e.fileTreeSvc.UpdateNode(&req)
	if err2 != nil {
		return err2.ToResp(), nil
	}

	return httpserver.OkResp(data)
}

func (e *Endpoints) MvCpNode(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	identity, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.MoveNode.NotLogin().ToResp(), nil
	}
	orgID, err := user.GetOrgID(r)
	if err != nil {
		return apierrors.MoveNode.MissingParameter("Org-ID").ToResp(), nil
	}

	treeName, ok := vars["treeName"]
	if !ok || treeName != apidocsvc.TreeNameAPIDocs {
		return apierrors.MoveNode.NotFound().ToResp(), nil
	}

	var body apistructs.APIDocMvCpNodeReqBody
	if err = json.NewDecoder(r.Body).Decode(&body); err != nil {
		return apierrors.MoveNode.InvalidParameter(err).ToResp(), nil
	}

	inode, ok := vars["inode"]
	if !ok || len(inode) == 0 {
		return apierrors.MoveNode.NotFound().ToResp(), nil
	}

	action := vars["action"]

	var req = apistructs.APIDocMvCpNodeReq{
		OrgID:    orgID,
		Identity: &identity,
		URIParams: &apistructs.FileTreeActionURI{
			TreeName: treeName,
			Inode:    inode,
			Action:   action,
		},
		Body: &body,
	}

	if action == "move" {
		data, err2 := e.fileTreeSvc.MoveNode(&req)
		if err2 != nil {
			return err2.ToResp(), nil
		}
		return httpserver.OkResp(data)
	}

	if action == "copy" {
		data, err2 := e.fileTreeSvc.CopyNode(&req)
		if err2 != nil {
			return err2.ToResp(), nil
		}
		return httpserver.OkResp(data)
	}

	return apierrors.MoveNode.NotFound().ToResp(), nil
}

func (e *Endpoints) ListChildrenNodes(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	identity, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ListChildrenNodes.NotLogin().ToResp(), nil
	}
	orgID, err := user.GetOrgID(r)
	if err != nil {
		return apierrors.ListChildrenNodes.NotLogin().ToResp(), nil
	}

	var params apistructs.FileTreeQueryParameters
	if err = e.queryStringDecoder.Decode(&params, r.URL.Query()); err != nil {
		return apierrors.PagingAPIAssets.InvalidParameter(err).ToResp(), nil
	}

	var req = apistructs.APIDocListChildrenReq{
		OrgID:       orgID,
		Identity:    &identity,
		URIParams:   &apistructs.FileTreeDetailURI{TreeName: vars["treeName"]},
		QueryParams: &params,
	}

	data, err2 := e.fileTreeSvc.ListChildren(&req)
	if err2 != nil {
		return err2.ToResp(), nil
	}

	return httpserver.OkResp(data)
}

func (e *Endpoints) GetNodeDetail(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	identity, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.GetNodeDetail.NotLogin().ToResp(), nil
	}
	orgID, err := user.GetOrgID(r)
	if err != nil {
		return apierrors.GetNodeDetail.MissingParameter("Org-ID").ToResp(), nil
	}

	var req = apistructs.APIDocNodeDetailReq{
		OrgID:    orgID,
		Identity: &identity,
		URIParams: &apistructs.FileTreeDetailURI{
			TreeName: vars["treeName"],
			Inode:    vars["inode"],
		},
	}

	data, err2 := e.fileTreeSvc.GetNodeDetail(&req)
	if err2 != nil {
		return err2.ToResp(), nil
	}

	return httpserver.OkResp(data)
}
