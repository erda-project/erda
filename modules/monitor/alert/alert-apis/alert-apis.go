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
	"github.com/erda-project/erda/modules/monitor/alert/alert-apis/adapt"
)

type listResult struct {
	List  interface{} `json:"list"`
	Total int         `json:"total"`
}

func (p *provider) queryAlertRule(r *http.Request, params struct {
	Scope   string `query:"scope" validate:"required"`
	ScopeID string `query:"scopeId" validate:"required"`
}) interface{} {
	data, err := p.a.QueryAlertRule(api.Language(r), params.Scope, params.ScopeID)
	if err != nil {
		return api.Errors.Internal(err)
	}
	return api.Success(data)
}

func (p *provider) queryAlert(r *http.Request, params struct {
	Scope    string `query:"scope" validate:"required"`
	ScopeID  string `query:"scopeId" validate:"required"`
	PageNo   int    `query:"pageNo" validate:"gte=1"`
	PageSize int    `query:"pageSize" validate:"gte=1,lte=100"`
}) interface{} {
	data, err := p.a.QueryAlert(api.Language(r), params.Scope, params.ScopeID, uint64(params.PageNo), uint64(params.PageSize))
	if err != nil {
		return api.Errors.Internal(err)
	}
	if data == nil {
		data = make([]*adapt.Alert, 0)
	}
	total, err := p.a.CountAlert(params.Scope, params.ScopeID)
	if err != nil {
		return api.Errors.Internal(err)
	}
	return api.Success(map[string]interface{}{
		"list":  data,
		"total": total,
	})
}

func (p *provider) getAlert(r *http.Request, params struct {
	ID int `param:"id" validate:"gte=1"`
}) interface{} {
	data, err := p.a.GetAlert(api.Language(r), uint64(params.ID))
	if err != nil {
		return api.Errors.Internal(err)
	}
	return api.Success(data)
}

func (p *provider) getAlertDetail(r *http.Request, params struct {
	ID int `param:"id" validate:"gte=1"`
}) interface{} {
	data, err := p.a.GetAlertDetail(api.Language(r), uint64(params.ID))
	if err != nil {
		return api.Errors.Internal(err)
	}
	return api.Success(data)
}

func (p *provider) createAlert(r *http.Request, alert adapt.Alert) interface{} {
	if err := p.checkAlert(&alert); err != nil {
		return err
	}
	orgID := alert.Attributes["dice_org_id"]
	org, err := p.bdl.GetOrg(orgID)
	if err != nil {
		return api.Errors.Internal(err)
	}
	alert.Attributes["org_name"] = org.Name
	id, err := p.a.CreateAlert(&alert)
	if err != nil {
		if adapt.IsInvalidParameterError(err) {
			return api.Errors.InvalidParameter(err)
		}
		if adapt.IsAlreadyExistsError(err) {
			return api.Errors.AlreadyExists(err)
		}
		return api.Errors.Internal(err)
	}
	return api.Success(id)
}

func (p *provider) checkAlert(alert *adapt.Alert) interface{} {
	if alert.Name == "" {
		return api.Errors.MissingParameter("alert name")
	}
	if alert.AlertScope == "" {
		return api.Errors.MissingParameter("alert scope")
	}
	if alert.AlertScopeID == "" {
		return api.Errors.MissingParameter("alert scopeId")
	}
	if len(alert.Rules) == 0 {
		return api.Errors.MissingParameter("alert rules")
	}
	if len(alert.Notifies) == 0 {
		return api.Errors.MissingParameter("alert notifies")
	}
	return nil
}

func (p *provider) updateAlert(r *http.Request, params struct {
	ID int `param:"id" validate:"required,gt=0"`
}, alert adapt.Alert) interface{} {
	if err := p.checkAlert(&alert); err != nil {
		return err
	}
	orgID := alert.Attributes["dice_org_id"]
	org, err := p.bdl.GetOrg(orgID)
	if err != nil {
		return api.Errors.Internal(err)
	}
	alert.Attributes["org_name"] = org.Name
	if err := p.a.UpdateAlert(uint64(params.ID), &alert); err != nil {
		if adapt.IsInvalidParameterError(err) {
			return api.Errors.InvalidParameter(err)
		}
		if adapt.IsAlreadyExistsError(err) {
			return api.Errors.AlreadyExists(err)
		}
		return api.Errors.Internal(err)
	}
	return api.Success(nil)
}

func (p *provider) updateAlertEnable(r *http.Request, params struct {
	ID     int  `param:"id" validate:"required,gt=0"`
	Enable bool `query:"enable"`
}) interface{} {
	err := p.a.UpdateAlertEnable(uint64(params.ID), params.Enable)
	if err != nil {
		return api.Errors.Internal(err)
	}
	return api.Success(nil)
}

func (p *provider) deleteAlert(r *http.Request, params struct {
	ID int `param:"id" validate:"required,gt=0"`
}) interface{} {
	data, _ := p.a.GetAlert(api.Language(r), uint64(params.ID))
	err := p.a.DeleteAlert(uint64(params.ID))
	if err != nil {
		return api.Errors.Internal(err)
	}
	if data != nil {
		return api.Success(map[string]interface{}{
			"name": data.Name,
		})
	}
	return api.Success(nil)
}
