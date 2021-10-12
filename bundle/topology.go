// Copyright (c) 2021 Terminus, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package bundle

import (
	"strconv"
	"time"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle/apierrors"
)

func (b *Bundle) GetServices(terminusKey string) ([]apistructs.ServiceDashboard, error) {
	host, err := b.urls.Monitor()
	if err != nil {
		return nil, err
	}
	hc := b.hc

	var serviceResp apistructs.ServiceDashboardResponse
	h, _ := time.ParseDuration("-15m")
	start := time.Now().Add(h).UnixNano() / 1e6
	resp, err := hc.Get(host).Path("/api/apm/topology/services").
		Param("terminusKey", terminusKey).
		Param("start", strconv.Itoa(int(start))).
		Param("end", strconv.Itoa(int(time.Now().UnixNano()/1e6))).
		Do().JSON(&serviceResp)
	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}
	if !resp.IsOK() || !serviceResp.Success {
		return nil, toAPIError(resp.StatusCode(), serviceResp.Error)
	}
	return serviceResp.Data.Data, nil
}
