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
	"github.com/erda-project/erda/pkg/httputil"
)

// Create or update component ingress
func (b *Bundle) CreateOrUpdateComponentIngress(req apistructs.ComponentIngressUpdateRequest) error {
	host, err := b.urls.Hepa()
	if err != nil {
		return err
	}
	var fetchResp apistructs.ComponentIngressUpdateResponse
	resp, err := b.hc.Put(host).Path("/api/gateway/component-ingress").Header(httputil.InternalHeader, "bundle").JSONBody(req).Do().JSON(&fetchResp)
	if err != nil {
		return apierrors.ErrInvoke.InternalError(err)
	}
	if !resp.IsOK() || !fetchResp.Success {
		return toAPIError(resp.StatusCode(), fetchResp.Error)
	}
	return nil
}
