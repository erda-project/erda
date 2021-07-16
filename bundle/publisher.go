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

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle/apierrors"
	"github.com/erda-project/erda/pkg/http/httputil"
)

// FetchPublisher 获取 publisher 详情
func (b *Bundle) FetchPublisher(publisherID uint64) (*apistructs.PublisherDTO, error) {
	host, err := b.urls.DOP()
	if err != nil {
		return nil, err
	}
	hc := b.hc
	var publisherResp apistructs.PublisherDetailResponse
	resp, err := hc.Get(host).Path(fmt.Sprintf("/api/publishers/%d", publisherID)).
		Header("Internal-Client", "bundle").
		Do().JSON(&publisherResp)
	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}
	if !resp.IsOK() || !publisherResp.Success {
		return nil, toAPIError(resp.StatusCode(), publisherResp.Error)
	}
	return &publisherResp.Data, nil
}

func (b *Bundle) GetUserRelationPublisher(userID string, orgID string) (*apistructs.PagingPublisherDTO, error) {
	host, err := b.urls.DOP()
	if err != nil {
		return nil, err
	}
	hc := b.hc
	var publisherResp apistructs.PublisherListResponse
	resp, err := hc.Get(host).Path(fmt.Sprintf("/api/publishers/actions/list-my-publishers")).
		Header("Internal-Client", "bundle").
		Header("USER-ID", userID).
		Header(httputil.OrgHeader, orgID).
		Do().JSON(&publisherResp)
	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}
	if !resp.IsOK() || !publisherResp.Success {
		return nil, toAPIError(resp.StatusCode(), publisherResp.Error)
	}
	return &publisherResp.Data, nil
}
