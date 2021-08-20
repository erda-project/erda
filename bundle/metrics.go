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
	"github.com/erda-project/erda-proto-go/core/monitor/metric/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle/apierrors"
	"github.com/erda-project/erda/pkg/http/httputil"
)

// GetMetrics .
func (b *Bundle) GetMetrics(req apistructs.MetricsRequest) (*pb.QueryWithInfluxFormatResponse, error) {
	host, err := b.urls.CMP()
	if err != nil {
		return nil, err
	}
	hc := b.hc

	var metricsResp apistructs.MetricsResponse
	resp, err := hc.Get(host).Path("/api/metrics").JSONBody(req).
		Header(httputil.InternalHeader, "bundle").
		Header("User-ID",req.UserID).
		Header("Org-ID",req.OrgID).
		Do().JSON(&metricsResp)
	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}
	if !resp.IsOK() || !metricsResp.Success {
		return nil, toAPIError(resp.StatusCode(), metricsResp.Error)
	}

	return metricsResp.Data, nil
}
