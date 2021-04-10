package report

import (
	"github.com/erda-project/erda-infra/modcom/api"
	"github.com/erda-project/erda-infra/providers/httpserver"
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
