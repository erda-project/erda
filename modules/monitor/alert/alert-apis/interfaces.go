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

package apis

import (
	"github.com/erda-project/erda/apistructs"
	block "github.com/erda-project/erda/modules/monitor/dashboard/chart-block"
	"net/http"

	"github.com/erda-project/erda-infra/modcom/api"
	"github.com/erda-project/erda-infra/providers/i18n"
	"github.com/erda-project/erda/modules/monitor/alert/alert-apis/adapt"
)

type (
	MicroAlertAPI interface {
		GetMicroServiceFilterTags() map[string]bool

		// micro alert apis
		QueryAlertRule(r *http.Request, scope, scopeId string) (*adapt.AlertTypeRuleResp, error)
		QueryAlert(r *http.Request, scope, scopeId string, pageNum, pageSize uint64) ([]*adapt.Alert, error)
		GetAlert(lang i18n.LanguageCodes, id uint64) (*adapt.Alert, error)
		CountAlert(scope, scopeID string) (int, error)
		GetAlertDetail(r *http.Request, id uint64) (*adapt.Alert, error)
		CheckAlert(alert *adapt.Alert) interface{}
		CreateAlert(alert *adapt.Alert) (alertID uint64, err error)
		UpdateAlert(alertID uint64, alert *adapt.Alert) (err error)
		UpdateAlertEnable(id uint64, enable bool) (err error)
		DeleteAlert(id uint64) (err error)

		// micro custom alert apis
		CustomizeMetrics(lang i18n.LanguageCodes, scope, scopeID string, names []string) (*adapt.CustomizeMetrics, error)
		NotifyTargetsKeys(lang i18n.LanguageCodes, orgId string) []*adapt.DisplayKey
		CustomizeAlerts(lang i18n.LanguageCodes, scope, scopeID string, pageNo, pageSize int) ([]*adapt.CustomizeAlertOverview, int, error)
		CustomizeAlert(id uint64) (*adapt.CustomizeAlertDetail, error)
		CustomizeAlertDetail(id uint64) (*adapt.CustomizeAlertDetail, error)
		CheckCustomizeAlert(alert *adapt.CustomizeAlertDetail) error
		CreateCustomizeAlert(alertDetail *adapt.CustomizeAlertDetail) (alertID uint64, err error)
		UpdateCustomizeAlert(alertDetail *adapt.CustomizeAlertDetail) (err error)
		UpdateCustomizeAlertEnable(id uint64, enable bool) (err error)
		DeleteCustomizeAlert(id uint64) (err error)

		//micro custom alert records
		GetAlertRecordAttr(lang i18n.LanguageCodes, scope string) (*adapt.AlertRecordAttr, error)
		QueryAlertRecord(lang i18n.LanguageCodes, scope, scopeId string, alertGroup, alertState, alertType,
			handleState, handlerId []string, pageNo, pageSize int64) ([]*adapt.AlertRecord, error)
		CountAlertRecord(scope, scopeId string, alertGroups, alertStates, alertTypes, handleStates, handlerIDs []string) (int, error)
		GetAlertRecord(lang i18n.LanguageCodes, groupId string) (*adapt.AlertRecord, error)
		QueryAlertHistory(lang i18n.LanguageCodes, groupId string, start, end int64, limit uint) ([]*adapt.AlertHistory, error)
		CreateAlertRecordIssue(groupId string, issueCreate *apistructs.IssueCreateRequest) (uint64, error)
		UpdateAlertRecordIssue(groupId string, issueId uint64, request *apistructs.IssueUpdateRequest) error
		DashboardPreview(alert *adapt.CustomizeAlertDetail) (res *block.View, err error)
	}
)

func (p *provider) GetMicroServiceFilterTags() map[string]bool {
	return p.microServiceFilterTags
}

func (p *provider) QueryAlertRule(r *http.Request, scope, scopeId string) (*adapt.AlertTypeRuleResp, error) {
	return p.a.QueryAlertRule(api.Language(r), scope, scopeId)
}

func (p *provider) QueryAlert(r *http.Request, scope, scopeId string, pageNum, pageSize uint64) ([]*adapt.Alert, error) {
	return p.a.QueryAlert(api.Language(r), scope, scopeId, pageNum, pageSize)
}

func (p *provider) GetAlert(lang i18n.LanguageCodes, id uint64) (*adapt.Alert, error) {
	return p.a.GetAlert(lang, id)
}

func (p *provider) GetAlertDetail(r *http.Request, id uint64) (*adapt.Alert, error) {
	return p.a.GetAlertDetail(api.Language(r), id)
}

func (p *provider) CountAlert(scope, scopeID string) (int, error) {
	return p.a.CountAlert(scope, scopeID)
}

func (p *provider) CheckAlert(alert *adapt.Alert) interface{} {
	return p.checkAlert(alert)
}

func (p *provider) CreateAlert(alert *adapt.Alert) (alertID uint64, err error) {
	return p.a.CreateAlert(alert)
}

func (p *provider) UpdateAlert(alertID uint64, alert *adapt.Alert) (err error) {
	return p.a.UpdateAlert(alertID, alert)
}
func (p *provider) UpdateAlertEnable(id uint64, enable bool) (err error) {
	return p.a.UpdateAlertEnable(id, enable)
}

func (p *provider) DeleteAlert(id uint64) (err error) {
	return p.a.DeleteAlert(id)
}

func (p *provider) CustomizeMetrics(lang i18n.LanguageCodes, scope, scopeID string, names []string) (*adapt.CustomizeMetrics, error) {
	return p.a.CustomizeMetrics(lang, scope, scopeID, names)
}

func (p *provider) NotifyTargetsKeys(lang i18n.LanguageCodes, orgId string) []*adapt.DisplayKey {
	return p.a.NotifyTargetsKeys(lang, orgId)
}

func (p *provider) CustomizeAlerts(lang i18n.LanguageCodes, scope, scopeID string, pageNo, pageSize int) ([]*adapt.CustomizeAlertOverview, int, error) {
	return p.a.CustomizeAlerts(lang, scope, scopeID, pageNo, pageSize)
}

func (p *provider) CustomizeAlert(id uint64) (*adapt.CustomizeAlertDetail, error) {
	return p.a.CustomizeAlert(id)
}

func (p *provider) CustomizeAlertDetail(id uint64) (*adapt.CustomizeAlertDetail, error) {
	return p.a.CustomizeAlertDetail(id)
}

func (p *provider) CheckCustomizeAlert(alert *adapt.CustomizeAlertDetail) error {
	return p.checkCustomizeAlert(alert)
}

func (p *provider) CreateCustomizeAlert(alertDetail *adapt.CustomizeAlertDetail) (alertID uint64, err error) {
	return p.a.CreateCustomizeAlert(alertDetail)
}

func (p *provider) UpdateCustomizeAlert(alertDetail *adapt.CustomizeAlertDetail) (err error) {
	return p.a.UpdateCustomizeAlert(alertDetail)
}

func (p *provider) UpdateCustomizeAlertEnable(id uint64, enable bool) (err error) {
	return p.a.UpdateCustomizeAlertEnable(id, enable)
}

func (p *provider) DeleteCustomizeAlert(id uint64) (err error) {
	return p.a.DeleteCustomizeAlert(id)
}

func (p *provider) GetAlertRecordAttr(lang i18n.LanguageCodes, scope string) (*adapt.AlertRecordAttr, error) {
	return p.a.GetAlertRecordAttr(lang, scope)
}

func (p *provider) QueryAlertRecord(lang i18n.LanguageCodes, scope, scopeId string, alertGroup, alertState, alertType,
	handleState, handlerId []string, pageNo, pageSize int64) ([]*adapt.AlertRecord, error) {
	return p.a.QueryAlertRecord(lang, scope, scopeId, alertGroup, alertState, alertType, handleState,
		handlerId, uint(pageNo), uint(pageSize))
}

func (p *provider) CountAlertRecord(scope, scopeId string, alertGroups, alertStates, alertTypes, handleStates, handlerIDs []string) (int, error) {
	return p.a.CountAlertRecord(scope, scopeId, alertGroups, alertStates, alertTypes, handleStates, handlerIDs)
}

func (p *provider) GetAlertRecord(lang i18n.LanguageCodes, groupId string) (*adapt.AlertRecord, error) {
	return p.a.GetAlertRecord(lang, groupId)
}

func (p *provider) QueryAlertHistory(lang i18n.LanguageCodes, groupId string, start, end int64, limit uint) ([]*adapt.AlertHistory, error) {
	return p.a.QueryAlertHistory(lang, groupId, start, end, limit)
}

func (p *provider) CreateAlertRecordIssue(groupId string, issueCreate *apistructs.IssueCreateRequest) (uint64, error) {
	return p.a.CreateAlertIssue(groupId, *issueCreate)
}

func (p *provider) UpdateAlertRecordIssue(groupId string, issueId uint64, request *apistructs.IssueUpdateRequest) error {
	return p.a.UpdateAlertIssue(groupId, issueId, *request)
}

func (p *provider) DashboardPreview(alert *adapt.CustomizeAlertDetail) (res *block.View, err error) {
	return adapt.NewDashboard(p.a).GenerateDashboardPreView(alert)
}
