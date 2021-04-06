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

func (b *Bundle) GetPipelineReportSet(pipelineID uint64, types []string) (*apistructs.PipelineReportSet, error) {
	host, err := b.urls.Pipeline()
	if err != nil {
		return nil, err
	}
	hc := b.hc

	var pipelineResp apistructs.PipelineReportSetGetResponse
	req := hc.Get(host).Path(fmt.Sprintf("/api/pipeline-reportsets/%d", pipelineID))
	if types != nil {
		for _, v := range types {
			req.Param("type", v)
		}
	}
	httpResp, err := req.Header(httputil.InternalHeader, "bundle").
		Do().JSON(&pipelineResp)

	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}

	if !httpResp.IsOK() || !pipelineResp.Success {
		return nil, toAPIError(httpResp.StatusCode(), pipelineResp.Error)
	}
	return pipelineResp.Data, nil
}
