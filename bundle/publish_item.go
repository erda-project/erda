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
	"strconv"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle/apierrors"
)

func (b *Bundle) QueryPublishItems(req *apistructs.QueryPublishItemRequest) (*apistructs.QueryPublishItemData, error) {
	host, err := b.urls.DiceHub()
	if err != nil {
		return nil, err
	}
	hc := b.hc
	var getResp apistructs.QueryPublishItemResponse
	resp, err := hc.Get(host).Path("/api/publish-items").
		Param("pageSize", strconv.FormatInt(req.PageSize, 10)).
		Param("type", req.Type).
		Param("ids", req.Ids).
		Param("q", req.Q).
		Param("pageNo", strconv.FormatInt(req.PageNo, 10)).
		Param("publisherId", strconv.FormatInt(req.PublisherId, 10)).
		Param("name", req.Name).
		Header("Internal-Client", "bundle").
		Header("Org-ID", strconv.FormatInt(req.OrgID, 10)).
		Do().JSON(&getResp)
	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}
	if !resp.IsOK() || !getResp.Success {
		return nil, toAPIError(resp.StatusCode(), getResp.Error)
	}
	return &getResp.Data, nil
}
