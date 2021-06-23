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
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle/apierrors"
)

// CreateMBox 创建站内信记录
func (b *Bundle) CreateMBox(request *apistructs.CreateMBoxRequest) error {
	host, err := b.urls.CoreServices()
	if err != nil {
		return err
	}
	hc := b.hc

	var getResp apistructs.CreateMBoxResponse
	resp, err := hc.Post(host).Path("/api/mboxs").JSONBody(request).
		Do().JSON(&getResp)
	if err != nil {
		return apierrors.ErrInvoke.InternalError(err)
	}
	if !resp.IsOK() || !getResp.Success {
		return toAPIError(resp.StatusCode(), getResp.Error)
	}
	return nil
}
