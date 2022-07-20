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

package monitor

import (
	"context"
	"strconv"

	orgpb "github.com/erda-project/erda-proto-go/core/org/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/internal/core/openapi/legacy/api/apis"
	"github.com/erda-project/erda/internal/core/openapi/legacy/api/spec"
	"github.com/erda-project/erda/internal/core/org"
	comapis "github.com/erda-project/erda/pkg/common/apis"
	"github.com/erda-project/erda/pkg/discover"
)

var MONITOR_ORG_REPORT_TASK_CREATE = apis.ApiSpec{
	Path:        "/api/org/report/tasks",
	BackendPath: "/api/org/report/tasks",
	Host:        "monitor.marathon.l4lb.thisdcos.directory:7096",
	Scheme:      "http",
	Method:      "POST",
	CheckLogin:  true,
	CheckToken:  true,
	Doc:         "summary: 创建企业报表任务详情",
	Audit:       auditCreateOrgReportTask(apistructs.CreateOrgReportTasks),
}

func auditCreateOrgReportTask(tmp apistructs.TemplateName) func(ctx *spec.AuditContext) error {
	return func(ctx *spec.AuditContext) error {
		var requestBody struct {
			Name string `json:"name"`
		}
		if err := ctx.BindRequestData(&requestBody); err != nil {
			return err
		}

		orgResp, err := org.MustGetOrg().GetOrg(comapis.WithInternalClientContext(context.Background(), discover.SvcOpenapi), &orgpb.GetOrgRequest{
			IdOrName: strconv.FormatInt(ctx.OrgID, 10),
		})
		if err != nil {
			return err
		}
		org := orgResp.Data
		return ctx.CreateAudit(&apistructs.Audit{
			ScopeType:    apistructs.OrgScope,
			ScopeID:      uint64(ctx.OrgID),
			TemplateName: tmp,
			Context: map[string]interface{}{
				"reportName": requestBody.Name,
				"orgName":    org.Name,
			},
		})
	}
}
