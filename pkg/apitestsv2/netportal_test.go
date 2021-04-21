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
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/customhttp"
)

func Test_handleCustomNetportalRequest(t *testing.T) {
	customhttp.SetInetAddr("netportal.default")
	req := apistructs.APIRequestInfo{
		URL:     "inet://staging.terminus.io?ssl=on/www.erda.cloud",
		Method:  http.MethodGet,
		Headers: http.Header{},
	}
	customReq, err := handleCustomNetportalRequest(&req)
	_ = customReq
	assert.NoError(t, err)
	assert.NotZero(t, len(req.Headers))
}
