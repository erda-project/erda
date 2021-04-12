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

package testset

import (
	"net/http"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/openapi/api/apis"
)

var RECOVER_FROM_RECYCLE_BIN = apis.ApiSpec{
	Path:         "/api/testsets/<testSetID>/actions/recover-from-recycle-bin",
	BackendPath:  "/api/testsets/<testSetID>/actions/recover-from-recycle-bin",
	Host:         "qa.marathon.l4lb.thisdcos.directory:3033",
	Scheme:       "http",
	Method:       http.MethodPost,
	RequestType:  apistructs.TestSetRecoverFromRecycleBinRequest{},
	ResponseType: apistructs.TestSetRecoverFromRecycleBinResponse{},
	IsOpenAPI:    true,
	CheckLogin:   true,
	CheckToken:   true,
	Doc:          `summary: 从回收站恢复测试集(递归)`,
}
