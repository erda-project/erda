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
