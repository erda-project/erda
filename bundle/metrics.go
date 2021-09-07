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
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle/apierrors"
	"github.com/erda-project/erda/pkg/http/httputil"
)

// GetMetrics .
func (b *Bundle) GetMetrics(req apistructs.MetricsRequest) ([]apistructs.MetricsData, error) {
	host, err := b.urls.CMP()
	if err != nil {
		return nil, err
	}
	hc := b.hc

	var metricsResp apistructs.MetricsResponse
	resp, err := hc.Post(host).Path("/api/metrics").JSONBody(req).
		Header(httputil.InternalHeader, "bundle").
		Header("User-ID", req.UserID).
		Header("Org-ID", req.OrgID).
		Do().RAW()
	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}

	if err = json.Unmarshal(data, &metricsResp); err != nil {
		return nil, apierrors.ErrInvoke.InternalError(fmt.Errorf(string(data)))
	}

	return metricsResp.Data, nil
}
