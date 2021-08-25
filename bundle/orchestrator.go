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
