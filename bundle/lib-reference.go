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
	"github.com/erda-project/erda/pkg/http/httputil"
)

// PublisherItemRefered 根据发布内容 id 查看是否被库应用引用
func (b *Bundle) PublisherItemRefered(libID uint64) (uint64, error) {
	host, err := b.urls.DOP()
	if err != nil {
		return 0, err
	}
	hc := b.hc

	var libReferenceListResp apistructs.LibReferenceListResponse
	resp, err := hc.Get(host).Path("/api/lib-references").
		Param("libID", strconv.FormatUint(libID, 10)).
		Header("Accept", "application/json").
		Header(httputil.InternalHeader, "bundle").
		Do().JSON(&libReferenceListResp)
	if err != nil {
		return 0, apierrors.ErrInvoke.InternalError(err)
	}
	if !resp.IsOK() || !libReferenceListResp.Success {
		return 0, toAPIError(resp.StatusCode(), libReferenceListResp.Error)
	}

	return libReferenceListResp.Data.Total, nil
}
