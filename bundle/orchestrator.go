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

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle/apierrors"
	"github.com/erda-project/erda/pkg/http/httputil"
)

func (b *Bundle) GetRuntimes(name, applicationId, workspace, orgID, userID string) ([]apistructs.RuntimeSummaryDTO, error) {
	host, err := b.urls.Orchestrator()
	if err != nil {
		return nil, err
	}
	hc := b.hc

	var rsp struct {
		apistructs.Header
		Data []apistructs.RuntimeSummaryDTO
	}

	resp, err := hc.Get(host).
		Path(fmt.Sprintf("/api/runtimes?name=%s&applicationId=%s&workspace=%s", name, applicationId, workspace)).
		Header(httputil.OrgHeader, orgID).
		Header(httputil.UserHeader, userID).
		Do().JSON(&rsp)

	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}
	if !resp.IsOK() || !rsp.Success {
		return nil, toAPIError(resp.StatusCode(), rsp.Error)
	}
	if len(rsp.Data) == 0 {
		return nil, nil
	}
	return rsp.Data, nil
}

func (b *Bundle) CreateRuntime(req apistructs.RuntimeCreateRequest, orgID uint64, userID string) (*apistructs.DeploymentCreateResponseDTO, error) {
	host, err := b.urls.Orchestrator()
	if err != nil {
		return nil, err
	}
	hc := b.hc

	var rsp struct {
		apistructs.Header
		Data apistructs.DeploymentCreateResponseDTO
	}

	resp, err := hc.Post(host).Path(fmt.Sprintf("/api/runtimes")).
		Header(httputil.OrgHeader, strconv.FormatUint(orgID, 10)).
		Header(httputil.UserHeader, userID).
		JSONBody(req).Do().JSON(&rsp)

	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}
	if !resp.IsOK() || !rsp.Success {
		return nil, toAPIError(resp.StatusCode(), rsp.Error)
	}

	return &rsp.Data, nil
}
