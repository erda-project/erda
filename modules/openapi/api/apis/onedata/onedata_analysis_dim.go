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

package onedata

import (
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/openapi/api/apis"
)

var ONEDATA_ANALYSIS_DIM = apis.ApiSpec{
	Path:         "/api/analysis/dim",
	BackendPath:  "/analysis/dim",
	Host:         "onedata-analysis.bigdata.svc.cluster.local:8080",
	Scheme:       "http",
	Method:       "GET",
	CheckLogin:   true,
	RequestType:  apistructs.OneDataAnalysisDimRequest{},
	ResponseType: apistructs.OneDataAnalysisDimRequest{},
	Doc:          `解析维度`,
}
