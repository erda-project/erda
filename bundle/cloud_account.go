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
	"github.com/erda-project/erda/pkg/httputil"
)

func (b *Bundle) GetCloudAccount(accountID uint64) (*apistructs.CloudAccountAllInfo, error) {
	host, err := b.urls.CMDB()
	if err != nil {
		return nil, err
	}
	hc := b.hc

	var accountResp apistructs.CloudAccountGetResponse
	resp, err := hc.Get(host).Path(fmt.Sprintf("/api/cloud-accounts/%d", accountID)).Header(httputil.InternalHeader, "bundle").Do().JSON(&accountResp)
	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}
	if !resp.IsOK() || !accountResp.Success {
		return nil, toAPIError(resp.StatusCode(), accountResp.Error)
	}

	return &accountResp.Data, nil
}
