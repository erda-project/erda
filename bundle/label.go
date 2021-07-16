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

// GetLabel 通过id获取label
func (b *Bundle) GetLabel(id uint64) (*apistructs.ProjectLabel, error) {
	host, err := b.urls.CoreServices()
	if err != nil {
		return nil, err
	}
	hc := b.hc

	var labelResp apistructs.ProjectLabelGetByIDResponseData
	resp, err := hc.Get(host).Path(fmt.Sprintf("/api/labels/%d", id)).Header(httputil.InternalHeader, "bundle").
		Do().JSON(&labelResp)
	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}
	if !resp.IsOK() || !labelResp.Success {
		return nil, toAPIError(resp.StatusCode(), labelResp.Error)
	}

	return &labelResp.Data, nil
}

// ListLabelByNameAndProjectID list label by names and projectID
func (b *Bundle) ListLabelByNameAndProjectID(projectID uint64, names []string) ([]apistructs.ProjectLabel, error) {
	host, err := b.urls.CoreServices()
	if err != nil {
		return nil, err
	}
	hc := b.hc

	var rsp apistructs.ProjectLabelsResponse
	httpResp, err := hc.Get(host).Path(fmt.Sprintf("/api/labels/actions/list-by-projectID-and-names")).
		Header(httputil.InternalHeader, "bundle").
		Param("projectID", strconv.FormatUint(projectID, 10)).
		Params(map[string][]string{"name": names}).
		Do().JSON(&rsp)
	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}
	if !httpResp.IsOK() || !rsp.Success {
		return nil, toAPIError(httpResp.StatusCode(), rsp.Error)
	}

	return rsp.Data, nil
}

// ListLabel list label
func (b *Bundle) ListLabel(req apistructs.ProjectLabelListRequest) (*apistructs.ProjectLabelListResponseData, error) {
	host, err := b.urls.CoreServices()
	if err != nil {
		return nil, err
	}
	hc := b.hc

	var rsp apistructs.ProjectLabelListResponse
	httpResp, err := hc.Get(host).Path(fmt.Sprintf("/api/labels")).
		Header(httputil.InternalHeader, "bundle").
		Param("projectID", strconv.FormatUint(req.ProjectID, 10)).
		Param("type", string(req.Type)).
		Param("key", req.Key).
		Param("pageNo", strconv.FormatUint(req.PageNo, 10)).
		Param("pageSize", strconv.FormatUint(req.PageSize, 10)).
		Do().JSON(&rsp)
	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}
	if !httpResp.IsOK() || !rsp.Success {
		return nil, toAPIError(httpResp.StatusCode(), rsp.Error)
	}

	return rsp.Data, nil
}

// ListLabelByIDs list label by ids
func (b *Bundle) ListLabelByIDs(ids []uint64) ([]apistructs.ProjectLabel, error) {
	host, err := b.urls.CoreServices()
	if err != nil {
		return nil, err
	}
	hc := b.hc

	idStrList := make([]string, 0, len(ids))
	for _, v := range ids {
		idStrList = append(idStrList, strconv.FormatUint(v, 10))
	}
	var rsp apistructs.ProjectLabelsResponse
	httpResp, err := hc.Get(host).Path(fmt.Sprintf("/api/labels/actions/list-by-ids")).
		Header(httputil.InternalHeader, "bundle").
		Params(map[string][]string{"ids": idStrList}).
		Do().JSON(&rsp)
	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}
	if !httpResp.IsOK() || !rsp.Success {
		return nil, toAPIError(httpResp.StatusCode(), rsp.Error)
	}

	return rsp.Data, nil
}
