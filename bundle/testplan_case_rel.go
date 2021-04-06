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

func (b *Bundle) ListTestPlanCaseRel(testCaseIDs []uint64) ([]apistructs.TestPlanCaseRel, error) {
	host, err := b.urls.QA()
	if err != nil {
		return nil, err
	}
	hc := b.hc

	urlQueryStrings := make(map[string][]string)
	for _, tcID := range testCaseIDs {
		urlQueryStrings["id"] = append(urlQueryStrings["id"], fmt.Sprintf("%d", tcID))
	}

	var listResp apistructs.TestPlanCaseRelListResponse
	resp, err := hc.Get(host).Path("/api/testplans/testcase-relations/actions/internal-list").
		Header(httputil.InternalHeader, "bundle").
		Params(urlQueryStrings).
		Do().JSON(&listResp)
	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}
	if !resp.IsOK() || !listResp.Success {
		return nil, toAPIError(resp.StatusCode(), listResp.Error)
	}

	return listResp.Data, nil
}
