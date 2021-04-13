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
	"strconv"
	"strings"
	"time"

	"github.com/erda-project/erda-infra/providers/httpserver"
	apm "github.com/erda-project/erda/modules/monitor/apm/common"
	"github.com/erda-project/erda/modules/monitor/common/permission"
)

func getReportSettingsPermission() httpserver.Interceptor {
	return permission.Intercepter(
		permission.ScopeProject, permission.ProjectIdFromParams(),
		apm.Report, permission.ActionGet,
	)
}

func saveReportSettingsPermission() httpserver.Interceptor {
	return permission.Intercepter(
		permission.ScopeProject, permission.ProjectIdFromParams(),
		apm.Report, permission.ActionCreate,
	)
}

type reportSetting struct {
	Id                 string    `json:"id"`
	ProjectId          string    `json:"project_id" query:"project_id" validate:"required"`
	ProjectName        string    `json:"project_name"`
	Workspace          string    `json:"workspace" query:"workspace" validate:"required"`
	Created            time.Time `json:"created"`
	WeeklyReportConfig string    `json:"weekly_report_config"`
	WeeklyReportEnable bool      `json:"weekly_report_enable"`
	DailyReportConfig  string    `json:"daily_report_config"`
	DailyReportEnable  bool      `json:"daily_report_enable"`
}

func (report *provider) GetReportSettings(projectId string, workspace string) (*reportSetting, error) {
	upperWorkspace := strings.ToUpper(workspace)
	var reportSetting reportSetting
	err := report.db.Select("*").
		Table("sp_report_settings").
		Where("project_id = ?", projectId).
		Where("workspace = ?", upperWorkspace).
		Find(&reportSetting).
		Error
	return &reportSetting, err
}

func (report *provider) SaveReportSettings(params *reportSetting) (*reportSetting, error) {
	settings, err := report.GetReportSettings(params.ProjectId, params.Workspace)
	if settings != nil && settings.Id == "" {
		err = nil
		id, err := strconv.Atoi(params.ProjectId)
		if err != nil {
			return nil, err
		}
		project, err := report.bundle.GetProject(uint64(id))
		if err != nil {
			return nil, err
		}
		if params.WeeklyReportConfig == "" {
			params.WeeklyReportConfig = ""
			params.WeeklyReportEnable = false
		}
		if params.DailyReportConfig == "" {
			params.DailyReportConfig = ""
			params.DailyReportEnable = false
		}
		params.ProjectName = project.Name
		params.Created = time.Now()
		err = report.db.Table("sp_report_settings").Create(params).Error
		if err != nil {
			return nil, err
		}
	} else if settings != nil {
		if params.WeeklyReportConfig == "" {
			params.WeeklyReportConfig = settings.WeeklyReportConfig
			params.WeeklyReportEnable = settings.WeeklyReportEnable
		}
		if params.DailyReportConfig == "" {
			params.DailyReportConfig = settings.DailyReportConfig
			params.DailyReportEnable = settings.DailyReportEnable
		}
		err = report.db.Table("sp_report_settings").
			Where("project_id = ?", params.ProjectId).
			Where("workspace = ?", params.Workspace).
			Update("weekly_report_enable", params.WeeklyReportEnable).
			Update("weekly_report_config", params.WeeklyReportConfig).
			Update("daily_report_enable", params.DailyReportEnable).
			Update("daily_report_config", params.DailyReportConfig).
			Error
		if err != nil {
			return nil, err
		}
	}
	return params, err
}
