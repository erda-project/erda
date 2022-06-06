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

package report

import (
	"github.com/erda-project/erda-infra/providers/httpserver"
	api "github.com/erda-project/erda/pkg/common/httpapi"
)

func (report *provider) initRoutes(routes httpserver.Router) error {
	routes.GET("/api/apm/report/settings", report.reportSettingsGetView, getReportSettingsPermission())
	routes.POST("/api/apm/report/settings", report.reportSettingsSaveView, saveReportSettingsPermission())
	return nil
}

func (report *provider) reportSettingsGetView(params reportSetting) interface{} {
	settings, err := report.GetReportSettings(params.ProjectId, params.Workspace)
	if err != nil {
		return api.Errors.Internal(err)
	}
	return api.Success(settings)
}

func (report *provider) reportSettingsSaveView(params reportSetting) interface{} {
	settings, err := report.SaveReportSettings(&params)
	if err != nil {
		return api.Errors.Internal(err)
	}
	return api.Success(settings)
}
