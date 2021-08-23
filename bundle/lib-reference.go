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
