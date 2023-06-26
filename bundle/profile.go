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
	"fmt"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle/apierrors"
	"github.com/erda-project/erda/internal/tools/monitor/core/profile/query"
)

func (b *Bundle) ProfileRender(req *apistructs.ProfileRenderRequest) (*query.RenderResponse, error) {
	host, err := b.urls.Monitor()
	if err != nil {
		return nil, err
	}
	hc := b.hc

	res := query.ProfileRenderResponse{}
	resp, err := hc.Get(host).
		Path("/api/profile/render").
		Params(req.URLQueryString()).
		Header("Content-Type", "application/json").
		Do().JSON(&res)
	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}
	if !resp.IsOK() {
		return nil, apierrors.ErrInvoke.InternalError(
			fmt.Errorf("failed to render profile, error is %d", resp.StatusCode()))
	}
	return res.Data, nil
}
