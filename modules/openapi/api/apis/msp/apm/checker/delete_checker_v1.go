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

package checker

import (
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/openapi/api/apis"
)

var DELETE_CHECKER_V1 = apis.ApiSpec{
	Path:        "/api/metrics/<id>",
	BackendPath: "/api/v1/msp/checkers/metrics/<id>",
	Host:        "msp.marathon.l4lb.thisdcos.directory:8080",
	Scheme:      "http",
	Method:      "DELETE",
	CheckLogin:  true,
	Doc:         "summary: 删除url监控指标",
	Audit:       auditOperateMicroserviceStatusPageMetric(apistructs.DeleteInitiativeMonitor),
}
