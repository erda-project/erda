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

package apitestsv2

import (
	"fmt"
	"net/http"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/customhttp"
)

func handleCustomNetportalRequest(apiReq *apistructs.APIRequestInfo) (*http.Request, error) {
	customReq, err := customhttp.NewRequest(apiReq.Method, apiReq.URL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to handle custom netportal request, err: %v", err)
	}
	for k, values := range customReq.Header {
		for _, v := range values {
			apiReq.Headers.Add(k, v)
		}
	}
	return customReq, nil
}
