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

// UpdatePipelineCron update pipeline cron
func (b *Bundle) UpdatePipelineCron(req apistructs.PipelineCronUpdateRequest) error {
	host, err := b.urls.Pipeline()
	if err != nil {
		return err
	}
	hc := b.hc

	var updateResp apistructs.PipelineCronUpdateResponse
	httpResp, err := hc.Put(host).Path(fmt.Sprintf("/api/pipeline-crons/%d", req.ID)).
		Header(httputil.InternalHeader, "bundle").
		JSONBody(&req).
		Do().JSON(&updateResp)
	if err != nil {
		return apierrors.ErrInvoke.InternalError(err)
	}
	if !httpResp.IsOK() || !updateResp.Success {
		return toAPIError(httpResp.StatusCode(), updateResp.Error)
	}
	return nil
}
