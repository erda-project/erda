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
